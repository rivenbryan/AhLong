package main

import (
	"encoding/base64"
	"io"
	"log"
	"net/http"

	"google.golang.org/api/gmail/v1"
)

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
	app.sendTransactionPrompts(transactionBuf)

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
