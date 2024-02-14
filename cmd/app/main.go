package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"savannah"
	"time"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	var wait time.Duration
	cfg := savannah.LoadConfig()
	server := savannah.NewServer(*cfg)
	srve := http.Server{
		Addr:         server.ServerAddress,
		Handler:      server.Router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}
	fmt.Println(srve.Addr)
	go func() {
		if err := srve.ListenAndServe(); err != nil {
			log.Println(err)
		}
	}()
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
	ctx, cancel := context.WithTimeout(context.Background(), wait)
	defer cancel()
	srve.Shutdown(ctx)
	log.Println("shutting down")
	os.Exit(0)

}
