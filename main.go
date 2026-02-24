package main

import (
	"log"
	"net/http"
	"time"
)

func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "healthy"}`))
}

func main() {
		mux := http.NewServeMux()
		mux.HandleFunc("/health", healthCheckHandler)

		server := &http.Server{
				Addr:         ":6969",
				Handler:      mux,
				ReadTimeout:  5 * time.Second,
				WriteTimeout: 10 * time.Second,
				IdleTimeout:  120 * time.Second,
		}

		log.Fatal(server.ListenAndServe())
}
