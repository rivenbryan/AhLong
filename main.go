package main

import (
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"google.golang.org/api/gmail/v1"
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

// Endpoint to handlePubSubMessage from Google API
func (app *App) handlePubSubMessage(w http.ResponseWriter, r *http.Request) {
	log.Println("Handle Endpoint hit!")

	_, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Fetch the Gmail History List and send the starting history ID
	historyBuf := app.fetchGmailHistory(app.CurrentHistoryId)
	// Process the Gmail History
	transactionBuf := app.processGmailHistories(historyBuf)
	// Sends Transaction Slice to Telegram Bot
	log.Println(transactionBuf)

}

func (app *App) fetchGmailHistory(historyId uint64) []*gmail.History {
	history, err := app.GmailService.Users.History.List("me").StartHistoryId(historyId).HistoryTypes("messageAdded").LabelId("Label_1547646376766633227").Do()
	if err != nil {
		log.Println(err)
	}
	return history.History

}

func (app *App) processGmailHistories(historySlice []*gmail.History) []Transaction {
	var transactionBuf []Transaction
	// Loop through each History
	for _, history := range historySlice {
		for _, messageAdded := range history.MessagesAdded {
			messageID := messageAdded.Message.Id
			msg, err := app.GmailService.Users.Messages.Get("me", messageID).Do()
			if err != nil {
				log.Println(err)
			}
			// An email can have multiple parts (plain-text, HTML)
			for _, part := range msg.Payload.Parts {
				if part.MimeType == "text/html" { // We want the HTML
					data, err := base64.URLEncoding.DecodeString(part.Body.Data)
					if err != nil {
						log.Println(err)
						continue
					}
					transaction := extractTransactionDetails(string(data))
					transactionBuf = append(transactionBuf, transaction)
					log.Printf("Amount: %s, To: %s", transaction.Amount, transaction.Recipient)
				}
			}

		}
		app.CurrentHistoryId = history.Id
	}
	return transactionBuf
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
