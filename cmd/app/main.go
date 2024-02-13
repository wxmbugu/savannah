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

	"github.com/gorilla/mux"
)

func main() {
	mux := mux.NewRouter()
	var wait time.Duration
	server := savannah.NewServer()
	srve := http.Server{
		Addr:         server.ServerAddress,
		Handler:      mux,
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
