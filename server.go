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
	"github.com/go-playground/validator"
	"github.com/gorilla/mux"
	"golang.org/x/oauth2"
)

type Server struct {
	Services      Service
	Router        *mux.Router
	oathConfig    *oauth2.Config
	provider      *oidc.Provider
	ctx           context.Context
	Cfg           *Config
	ServerAddress string
	validator     *validator.Validate
}

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

func NewServer(cfg Config) *Server {

	ctx := context.Background()
	mux := mux.NewRouter()

	conn := SetupDb(cfg.DBURL)
	provider, err := oidc.NewProvider(ctx, cfg.AccProvider)
	if err != nil {
		log.Fatal(err)
	}
	config := oauth2.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		Endpoint:     provider.Endpoint(),
		RedirectURL:  cfg.RedirectURL,
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
	}
	validate := validator.New()
	services := NewService(conn, cfg.AUsername, cfg.AtalkingAPI)
	server := Server{
		Router:        mux,
		Services:      services,
		oathConfig:    &config,
		provider:      provider,
		ctx:           ctx,
		ServerAddress: cfg.ServerAddress,
		Cfg:           &cfg,
		validator:     validate,
	}

	server.Routes()
	return &server
}

func (server *Server) Routes() {
	http.Handle("/", server.Router)
	server.Router.Use(corsmiddleware)
	server.Router.Use(jsonmiddleware)
	server.Router.HandleFunc("/login", server.setCallbackCookie).Methods("GET", "OPTIONS")
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
		serializeResponse(w, http.StatusInternalServerError, Errorjson{"error": err.Error()})
		return
	}
	nonce, err := randString(16)
	if err != nil {
		serializeResponse(w, http.StatusInternalServerError, Errorjson{"error": err.Error()})
		return
	}
	setCallbackCookie(w, r, "state", state)
	setCallbackCookie(w, r, "nonce", nonce)

	http.Redirect(w, r, server.oathConfig.AuthCodeURL(state, oidc.Nonce(nonce)), http.StatusFound)

}
func (server *Server) googleCallback(w http.ResponseWriter, r *http.Request) {
	state, err := r.Cookie("state")
	if err != nil {
		serializeResponse(w, http.StatusBadRequest, Errorjson{"error": "state not found"})
		return
	}
	if r.URL.Query().Get("state") != state.Value {
		serializeResponse(w, http.StatusBadRequest, Errorjson{"error": "state did not match"})
		return
	}

	oauth2Token, err := server.oathConfig.Exchange(server.ctx, r.URL.Query().Get("code"))
	if err != nil {
		serializeResponse(w, http.StatusInternalServerError, Errorjson{"error": err.Error()})
		return
	}
	userInfo, err := server.provider.UserInfo(server.ctx, oauth2.StaticTokenSource(oauth2Token))
	if err != nil {
		serializeResponse(w, http.StatusInternalServerError, Errorjson{"error": err.Error()})
		return
	}
	if err != nil {
		serializeResponse(w, http.StatusInternalServerError, Errorjson{"error": err.Error()})
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
	_, err = server.provider.Verifier(&oidc.Config{ClientID: server.Cfg.ClientID}).Verify(r.Context(), oauth2Token.Extra("id_token").(string))
	if err != nil {
		serializeResponse(w, http.StatusUnauthorized, Errorjson{"error": err.Error()})
		return
	}
	response := struct {
		Access_Token string `json:"access_token"`
		User         *User  `json:"user"`
	}{oauth2Token.Extra("id_token").(string), user}
	serializeResponse(w, http.StatusOK, response)
}

func (server *Server) createCustomer(w http.ResponseWriter, r *http.Request) {
	var customer User
	err := json.NewDecoder(r.Body).Decode(&customer)
	if err != nil {
		serializeResponse(w, http.StatusBadRequest, Errorjson{"error": err.Error()})
		return
	}
	if err := server.validator.Struct(customer); err != nil {
		serializeResponse(w, http.StatusBadRequest, Errorjson{"error": err.Error()})
		return
	}
	createdCustomer, err := server.Services.service.CreateUser(customer)
	if err != nil {
		serializeResponse(w, http.StatusInternalServerError, Errorjson{"error": err.Error()})
		return
	}
	serializeResponse(w, http.StatusCreated, createdCustomer)
}

func (server *Server) getCustomer(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	idStr := params["id"]

	id, err := strconv.Atoi(idStr)
	if err != nil {
		serializeResponse(w, http.StatusBadRequest, Errorjson{"error": "Invalid ID"})
		return
	}

	customer, err := server.Services.service.FindUser(id)
	if err != nil {
		if err == sql.ErrNoRows {
			serializeResponse(w, http.StatusNotFound, Errorjson{"error": "Customer not found"})
			return
		}
		serializeResponse(w, http.StatusInternalServerError, Errorjson{"error": err.Error()})
		return
	}
	serializeResponse(w, http.StatusOK, customer)
}

func (server *Server) createOrder(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value("claims").(*Claims)
	if !ok {
		serializeResponse(w, http.StatusInternalServerError, Errorjson{"error": "Claims not found in context"})
		return
	}
	var order Orders
	err := json.NewDecoder(r.Body).Decode(&order)
	if err != nil {
		serializeResponse(w, http.StatusBadRequest, Errorjson{"error": err.Error()})
		return
	}

	if err := server.validator.Struct(order); err != nil {
		serializeResponse(w, http.StatusBadRequest, Errorjson{"error": err.Error()})
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
		serializeResponse(w, http.StatusInternalServerError, Errorjson{"error": err.Error()})
		return
	}
	item, err := server.Services.service.FindItem(order.ItemID)
	if err != nil {
		serializeResponse(w, http.StatusInternalServerError, Errorjson{"error": err.Error()})
		return
	}

	orderConfirmationMessage := fmt.Sprintf("Thank you for your order! You've successfully created an order for %s. You ordered %d %s(s), totaling an amount of $%.2f. We appreciate your business!", item.Name, order.Qty, item.Name, item.Price*float32(order.Qty))
	err = server.Services.sms.Send(order.Contact, orderConfirmationMessage)
	if err != nil {
		serializeResponse(w, http.StatusInternalServerError, Errorjson{"error": err.Error()})
		return
	}
	serializeResponse(w, http.StatusCreated, createdOrder)
}

func (server *Server) getOrder(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	idStr := params["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		serializeResponse(w, http.StatusBadRequest, Errorjson{"error": "Invalid ID"})
		return
	}
	order, err := server.Services.service.FindOrders(id)
	if err != nil {
		if err == sql.ErrNoRows {
			serializeResponse(w, http.StatusNotFound, "Order not found")
			return
		}
		serializeResponse(w, http.StatusInternalServerError, Errorjson{"error": err.Error()})
		return
	}
	serializeResponse(w, http.StatusOK, order)
}

func (server *Server) getItem(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	idStr := params["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		serializeResponse(w, http.StatusBadRequest, Errorjson{"error": "Invalid ID"})
		return
	}
	item, err := server.Services.service.FindItem(id)
	if err != nil {
		if err == sql.ErrNoRows {
			serializeResponse(w, http.StatusNotFound, "Item not found")
			return
		}
		serializeResponse(w, http.StatusInternalServerError, Errorjson{"error": err.Error()})
		return
	}
	serializeResponse(w, http.StatusOK, item)
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
	claimsKey     = "claims"
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
		verifier := server.provider.Verifier(&oidc.Config{ClientID: server.Cfg.ClientID})
		idToken, err := verifier.Verify(r.Context(), reqtoken)
		if err != nil {
			serializeResponse(w, http.StatusUnauthorized, Errorjson{"error": err.Error()})
			return
		}
		var claims Claims
		err = idToken.Claims(&claims)
		if err != nil {
			serializeResponse(w, http.StatusInternalServerError, Errorjson{"error": err.Error()})
		}
		ctx := context.WithValue(r.Context(), claimsKey, &claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func jsonmiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}
