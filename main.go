package main

import (
	"fmt"
	"log"
	"net/http"
)

func webhookHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("✅ Webhook hit!")

	// Log basic request info
	log.Println("Method:", r.Method)
	log.Println("RemoteAddr:", r.RemoteAddr)

	// IMPORTANT: Pub/Sub expects 2xx
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "ok")
}

func main() {
	http.HandleFunc("/webhook", webhookHandler)

	port := "8080"
	log.Println("🚀 Listening on port", port)

	log.Fatal(http.ListenAndServe(":"+port, nil))
}
