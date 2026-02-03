package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

// End point for health check
func healthCheck(w http.ResponseWriter, r *http.Request) {
	log.Println("✅ HealthCheck Endpoint hit!")

	// Log basic request info
	log.Println("Method:", r.Method)
	log.Println("RemoteAddr:", r.RemoteAddr)

	// IMPORTANT: Pub/Sub expects 2xx
	w.WriteHeader(http.StatusOK)
	_, err := fmt.Fprintln(w, "ok")
	if err != nil {
		return
	}
}

func main() {
	port := "8080"

	app, err := NewApp(
		os.Getenv("GMAIL_CLIENT_ID"),
		os.Getenv("GMAIL_CLIENT_SECRET"),
		os.Getenv("GMAIL_REFRESH_TOKEN"),
	)

	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/health", healthCheck)
	http.HandleFunc("/handlePubSub", app.handlePubSubMessage)
	http.HandleFunc("/telegramCallback", app.handleTelegramCallback)

	log.Println("🚀 Listening on port", port)

	log.Fatal(http.ListenAndServe(":"+port, nil))
}

//
