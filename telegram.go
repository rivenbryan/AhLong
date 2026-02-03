package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

const deleteDelay = 10 * time.Second

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
	if err = json.Unmarshal(body, &response); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	expenses := extractTelegramResponse(response)

	if err = updateNotionDatabase(expenses); err != nil {
		log.Println(err)
		return
	}

	msg := fmt.Sprintf("✅ Logged: %s\nRecipient: %s\nAmount: %s", expenses.category, expenses.recipient, expenses.amount)
	messageID, err := sendTelegramMessage(msg)
	if err != nil {
		log.Println(err)
		return
	}

	// The first message: We trigger from our Go Server
	if err = deleteTelegramMessage(response.CallbackQuery.Message.MessageID); err != nil {
		log.Println(err)
		return
	}

	// The second message: Successful message that has to be wrapped into go-routine (we don't want to block and wait for req to finish)
	go func(messageID int) {
		time.Sleep(deleteDelay)
		if err = deleteTelegramMessage(messageID); err != nil {
			log.Println(err)
		}
	}(*messageID)
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

func sendTelegramMessage(msg string) (*int, error) {
	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	chatID := os.Getenv("TELEGRAM_CHAT_ID")
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", botToken)

	payload := map[string]string{
		"chat_id": chatID,
		"text":    msg,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Result struct {
			MessageID int `json:"message_id"`
		} `json:"result"`
	}

	if err = json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return &result.Result.MessageID, nil

}

func deleteTelegramMessage(messageID int) error {
	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	chatID := os.Getenv("TELEGRAM_CHAT_ID")
	url := fmt.Sprintf("https://api.telegram.org/bot%s/deleteMessage", botToken)

	payload := map[string]interface{}{
		"chat_id":    chatID,
		"message_id": messageID,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	resp.Body.Close()

	return nil
}
