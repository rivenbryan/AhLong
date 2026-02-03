package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

type TelegramResponse struct {
	CallbackQuery struct {
		ID      string `json:"id"`
		Data    string `json:"data"`
		Message struct {
			MessageID int `json:"message_id"`
			Chat      struct {
				ID int64 `json:"id"`
			} `json:"chat"`
		} `json:"message"`
	} `json:"callback_query"`
}

func (app *App) sendTransactionPrompts(transactionBuf []Transaction) {
	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", botToken)

	for _, txn := range transactionBuf {
		body, err := createTelegramPayload(txn)
		if err != nil {
			log.Println(err)
			continue
		}

		resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
		if err != nil {
			log.Println(err)
			continue
		}
		defer resp.Body.Close()

		log.Printf("Telegram response status: %s", resp.Status)
	}
}

func (app *App) handleTelegramCallback(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var response TelegramResponse
	if err := json.Unmarshal(body, &response); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	log.Printf("%+v", response)

	expenses := extractTelegramResponse(response)
	err = updateNotionDatabase(expenses) // Send to Notion API
	if err != nil {
		log.Println(err)
	}

}

func createTelegramPayload(txn Transaction) ([]byte, error) {
	chatID := os.Getenv("TELEGRAM_CHAT_ID")
	payload := map[string]interface{}{
		"chat_id": chatID,
		"text":    fmt.Sprintf("💸 PayLah/PayNow Transaction Detected\n\nTo: %s\nAmount: %s", txn.Recipient, txn.Amount),
		"reply_markup": map[string]interface{}{
			"inline_keyboard": [][]map[string]string{
				{
					{"text": "🍔 Food", "callback_data": fmt.Sprintf("food|%s|%s", txn.Amount, txn.Recipient)},
					{"text": "🧑 Personal", "callback_data": fmt.Sprintf("personal|%s|%s", txn.Amount, txn.Recipient)},
					{"text": "🚗 Transportation", "callback_data": fmt.Sprintf("transportation|%s|%s", txn.Amount, txn.Recipient)},
				},
			},
		},
	}

	return json.Marshal(payload)

}
