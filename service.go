package savannah

import (
	"database/sql"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type Service struct {
	service database
	// small api to handle sending of sms to users
	sms *ATalkingService
}

// africas talking service
type ATalkingService struct {
	Username string
	APIKey   string
	env      string
}

func NewATalkingService(username, apiKey string) *ATalkingService {
	var env string
	if username == "sandbox" {
		env = "sandbox"
	}
	return &ATalkingService{username, apiKey, env}
}

func GetAPIHost(env string) string {
	if env == "sandbox" {
		return "https://api.sandbox.africastalking.com"
	} else {
		return "https://api.africastalking.com"
	}

}

func GetSmsURL(env string) string {
	return GetAPIHost(env) + "/version1/messaging"
}
func (service ATalkingService) Send(to, message string) error {
	values := url.Values{}
	values.Set("username", service.Username)
	values.Set("to", to)
	values.Set("message", message)

	smsURL := GetSmsURL(service.env)
	headers := make(map[string]string)
	headers["Content-Type"] = "application/x-www-form-urlencoded"

	_, err := service.newPostRequest(smsURL, values, headers)
	if err != nil {
		return err
	}
	return nil
}

func (service ATalkingService) newPostRequest(url string, values url.Values, headers map[string]string) (*http.Response, error) {
	reader := strings.NewReader(values.Encode())

	req, err := http.NewRequest(http.MethodPost, url, reader)
	if err != nil {
		return nil, err
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}
	req.Header.Set("Content-Length", strconv.Itoa(reader.Len()))
	req.Header.Set("apikey", service.APIKey)
	req.Header.Set("Accept", "application/json")
	client := &http.Client{Timeout: 10 * time.Second}
	return client.Do(req)
}

func NewService(conn *sql.DB, username, apikey string) Service {
	db := Newdb(conn)
	asms := NewATalkingService(username, apikey)
	return Service{
		service: db,
		sms:     asms,
	}
}
