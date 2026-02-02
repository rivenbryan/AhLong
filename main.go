package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"google.golang.org/api/gmail/v1"
	"io"
	"log"
	"net/http"
	"os"
)

type PubSubMessage struct {
	Message struct {
		Data string `json:"data"`
		ID   string `json:"messageId"`
	} `json:"message"`
	Subscription string `json:"subscription"`
}

type GmailWatchNotification struct {
	EmailAddress string
	HistoryId    uint64
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	log.Println("✅ HealthCheck Endpoint hit!")

	// Log basic request info
	log.Println("Method:", r.Method)
	log.Println("RemoteAddr:", r.RemoteAddr)

	// IMPORTANT: Pub/Sub expects 2xx
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "ok")
}

func (app *App) handlePubSubMessage(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Parse the content and return gmailWatchNotification
	gmailWatchNotification, err := parsePubSubMessage(body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.Println(gmailWatchNotification)
	// Fetch the Gmail History List and send the starting history ID
	historySlice := app.fetchGmailHistory(6268636)
	fmt.Println(historySlice)

}

func parsePubSubMessage(body []byte) (*GmailWatchNotification, error) {
	var msg PubSubMessage
	// Turn the body into a type of PubSubMessage
	if err := json.Unmarshal(body, &msg); err != nil {
		return nil, err
	}
	// Decode the base64 encoding to a human-readable string
	data, err := base64.StdEncoding.DecodeString(msg.Message.Data)
	if err != nil {
		return nil, err
	}

	var notification GmailWatchNotification
	if err := json.Unmarshal(data, &notification); err != nil {
		return nil, err
	}
	return &notification, err
}

func (app *App) fetchGmailHistory(historyId uint64) []*gmail.History {
	history, err := app.GmailService.Users.History.List("me").StartHistoryId(historyId).HistoryTypes("messageAdded").Do()
	if err != nil {
		log.Println(err)
	}
	return history.History

}

func main() {
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
	port := "8080"
	log.Println("🚀 Listening on port", port)

	log.Fatal(http.ListenAndServe(":"+port, nil))
}

//
