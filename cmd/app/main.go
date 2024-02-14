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
)

func main() {
	var wait time.Duration
	cfg := savannah.LoadConfig()
	fmt.Println(cfg)
	server := savannah.NewServer(*cfg)
	srve := http.Server{
		Addr:         fmt.Sprintf("0.0.0.0:%s", cfg.Port),
		Handler:      server.Router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}
	fmt.Println("serving on port:", cfg.Port)
	fmt.Println(cfg)
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
