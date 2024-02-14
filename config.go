package savannah

import "os"

type Config struct {
	ClientID      string
	ClientSecret  string
	DBURL         string
	AccProvider   string
	RedirectURL   string
	ServerAddress string
	AtalkingAPI   string
	AUsername     string
}

func LoadConfig() *Config {
	return &Config{
		ClientID:      os.Getenv("CLIENTID"),
		ClientSecret:  os.Getenv("CLIENTSECRET"),
		DBURL:         os.Getenv("DBURL"),
		AccProvider:   os.Getenv("ACCPROVIDER"),
		RedirectURL:   os.Getenv("REDIRECTURL"),
		ServerAddress: os.Getenv("SERVERADDRESS"),
		AtalkingAPI:   os.Getenv("ATALKINGAPI"),
		AUsername:     os.Getenv("AUSERNAME"),
	}
}
