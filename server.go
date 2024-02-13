package savannah

import (
	"bytes"
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/gorilla/mux"
	"golang.org/x/oauth2"
)

type Server struct {
	Services      Service
	Router        *mux.Router
	oathConfig    *oauth2.Config
	provider      *oidc.Provider
	ctx           context.Context
	ServerAddress string
}

var (
	clientID      = "387480342800-7o0qjebp71dkqakq4uju6kj885b9t194.apps.googleusercontent.com"
	clientSecret  = "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"
	dbURL         = "postgresql://postgres:secret@localhost:5432/savannah?sslmode=disable"
	accProvider   = "https://accounts.google.com"
	redirectURL   = "http://127.0.0.1:9000/auth/google/callback"
	serverAddress = "localhost:9000"
)

func randString(nByte int) (string, error) {
	b := make([]byte, nByte)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func setCallbackCookie(w http.ResponseWriter, r *http.Request, name, value string) {
	c := &http.Cookie{
		Name:     name,
		Value:    value,
		MaxAge:   int(time.Hour.Seconds()),
		Secure:   r.TLS != nil,
		HttpOnly: true,
	}
	http.SetCookie(w, c)
}
func SetupDb(conn string) *sql.DB {
	db, err := sql.Open("postgres", conn)
	if err != nil {
		log.Fatal(err)
	}
	db.Ping()
	db.SetMaxOpenConns(35)
	db.SetMaxIdleConns(35)
	db.SetConnMaxLifetime(time.Hour)
	return db
}

func NewServer() *Server {

	ctx := context.Background()
	mux := mux.NewRouter()

	conn := SetupDb(dbURL)
	provider, err := oidc.NewProvider(ctx, accProvider)
	if err != nil {
		log.Fatal(err)
	}
	config := oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Endpoint:     provider.Endpoint(),
		RedirectURL:  redirectURL,
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
	}

	services := NewService(conn)
	server := Server{
		Router:        mux,
		Services:      services,
		oathConfig:    &config,
		provider:      provider,
		ctx:           ctx,
		ServerAddress: serverAddress,
	}

	server.Routes()
	return &server
}

func (server *Server) Routes() {
	http.Handle("/", server.Router)
	server.Router.Use(corsmiddleware)
	server.Router.Use(jsonmiddleware)
	server.Router.HandleFunc("/", server.setCallbackCookie).Methods("GET", "OPTIONS")
	server.Router.HandleFunc("/auth/google/callback", server.googleCallback).Methods("GET", "OPTIONS")
	authroutes := server.Router.PathPrefix("/v1").Subrouter()
	authroutes.Use(server.authmiddleware)
	authroutes.HandleFunc("/customers", server.createCustomer).Methods("POST", "OPTIONS")
	authroutes.HandleFunc("/customers/{id}", server.getCustomer).Methods("GET", "OPTIONS")
	authroutes.HandleFunc("/orders", server.createOrder).Methods("POST", "OPTIONS")
	authroutes.HandleFunc("/orders/{id}", server.getOrder).Methods("GET", "OPTIONS")
	authroutes.HandleFunc("/items/{id}", server.getItem).Methods("GET", "OPTIONS")
}

func (server *Server) setCallbackCookie(w http.ResponseWriter, r *http.Request) {
	state, err := randString(16)
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}
	nonce, err := randString(16)
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}
	setCallbackCookie(w, r, "state", state)
	setCallbackCookie(w, r, "nonce", nonce)

	http.Redirect(w, r, server.oathConfig.AuthCodeURL(state, oidc.Nonce(nonce)), http.StatusFound)

}
func (server *Server) googleCallback(w http.ResponseWriter, r *http.Request) {
	state, err := r.Cookie("state")
	if err != nil {
		http.Error(w, "state not found", http.StatusBadRequest)
		return
	}
	if r.URL.Query().Get("state") != state.Value {
		http.Error(w, "state did not match", http.StatusBadRequest)
		return
	}

	oauth2Token, err := server.oathConfig.Exchange(server.ctx, r.URL.Query().Get("code"))
	if err != nil {
		http.Error(w, "Failed to exchange token: "+err.Error(), http.StatusInternalServerError)
		return
	}
	userInfo, err := server.provider.UserInfo(server.ctx, oauth2.StaticTokenSource(oauth2Token))
	if err != nil {
		http.Error(w, "Failed to get userinfo: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	user, err := server.Services.service.FindUserbyEmail(userInfo.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			user, err = server.Services.service.CreateUser(User{Code: userInfo.Subject, Email: userInfo.Email})
			if err != nil {
				serializeResponse(w, http.StatusInternalServerError, Errorjson{"error": err.Error()})
				return
			}
		} else {
			serializeResponse(w, http.StatusInternalServerError, Errorjson{"error": err.Error()})
			return
		}
	}
	_, err = server.provider.Verifier(&oidc.Config{ClientID: clientID}).Verify(r.Context(), oauth2Token.Extra("id_token").(string))
	if err != nil {
		serializeResponse(w, http.StatusUnauthorized, Errorjson{"error": err.Error()})
		return
	}
	response := struct {
		Access_Token string `json:"access_token"`
		User         *User  `json:"user"`
	}{oauth2Token.Extra("id_token").(string), user}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (server *Server) createCustomer(w http.ResponseWriter, r *http.Request) {
	var customer User
	err := json.NewDecoder(r.Body).Decode(&customer)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	createdCustomer, err := server.Services.service.CreateUser(customer)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdCustomer)
}

func (server *Server) getCustomer(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	idStr := params["id"]
	v := w.Header().Get("Authorization")
	fmt.Println("token", v)

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	customer, err := server.Services.service.FindUser(id)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Customer not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(customer)
}

func (server *Server) createOrder(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value("claims").(*Claims)
	if !ok {
		http.Error(w, "Claims not found in context", http.StatusInternalServerError)
		return
	}
	var order Orders
	err := json.NewDecoder(r.Body).Decode(&order)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	user, err := server.Services.service.FindUserbyEmail(claims.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			serializeResponse(w, http.StatusNotFound, Errorjson{"error": "No such user"})
			return
		} else {
			serializeResponse(w, http.StatusInternalServerError, Errorjson{"error": err.Error()})
			return
		}
	}
	order.Time = time.Now()
	order.UserId = user.ID
	createdOrder, err := server.Services.service.CreateOrders(order)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdOrder)
}

func (server *Server) getOrder(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	idStr := params["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}
	order, err := server.Services.service.FindOrders(id)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Order not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(order)
}

func (server *Server) getItem(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	idStr := params["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}
	item, err := server.Services.service.FindItem(id)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Order not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(item)
}
func corsmiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS,PUT")
		w.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization")
		w.Header().Set("Access-Control-Expose-Headers", "Authorization")
		w.Header().Set("Content-Type", "application/json")
		if r.Method == "OPTIONS" {
			return
		}

		next.ServeHTTP(w, r)
	})
}

const (
	authHeaderKey = "authorization"
	token         = "token"
	claims        = "claims"
)

func serializeResponse(w http.ResponseWriter, statuscode int, data interface{}) {
	w.WriteHeader(statuscode)
	reqBodyBytes := new(bytes.Buffer)
	json.NewEncoder(reqBodyBytes).Encode(data)
	w.Write(reqBodyBytes.Bytes())
}

type Errorjson map[string]string
type Claims struct {
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
}

func (server *Server) authmiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqtoken := r.Header.Get(authHeaderKey)
		tokenvalue := strings.Split(reqtoken, "Bearer ")
		if len(tokenvalue) == 0 {
			serializeResponse(w, http.StatusUnauthorized, Errorjson{"error": "authorization header not provided"})
			return
		}
		if len(tokenvalue) < 2 {
			serializeResponse(w, http.StatusUnauthorized, Errorjson{"error": "invalid authorization header format"})
			return
		}
		reqtoken = tokenvalue[1]
		verifier := server.provider.Verifier(&oidc.Config{ClientID: clientID})
		idToken, err := verifier.Verify(r.Context(), reqtoken)
		if err != nil {
			serializeResponse(w, http.StatusUnauthorized, Errorjson{"error": err.Error()})
			return
		}
		var claims Claims
		err = idToken.Claims(&claims)
		if err != nil {
			fmt.Println(err)
		}
		ctx := context.WithValue(r.Context(), claims, &claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func jsonmiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}
