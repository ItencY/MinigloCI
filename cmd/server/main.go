// Package main is the entry point for the MinigloCI application.
package main

import (
	"log"
	"net/http"
	"time"
)

func main() {
	const port = "8080"
	mux := http.NewServeMux()
	mux.HandleFunc("POST /run", handlerRun)
	server := &http.Server{
		Handler:           mux,
		Addr:              ":" + port,
		ReadHeaderTimeout: 3 * time.Second,
	}
	log.Println("The server is running on port :8080.")
	log.Fatal(server.ListenAndServe())
}
