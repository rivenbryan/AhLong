package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

var categoryMap = map[string]string{
	"food":           "183de7c8610e8095ac4fd48ce0005e65",
	"personal":       "183de7c8610e80ec9cfcffd8d6ccd192",
	"transportation": "183de7c8610e80648072d0a9c6c54e57",
}

func updateNotionDatabase(expenses Expenses) error {
	notionKey := os.Getenv("NOTION_API_KEY")

	body, err := createNotionPayload(expenses)

	if err != nil {
		return err
	}

	req, err := createNotionRequest(notionKey, body)

	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("notion API error: %s — %s", resp.Status, string(respBody))
	}

	return nil
}

func createNotionPayload(expenses Expenses) ([]byte, error) {

	categoryID := categoryMap[expenses.category]
	amount, err := processAmount(expenses.amount)

	if err != nil {
		return nil, err
	}

	payload := map[string]interface{}{
		"parent": map[string]string{
			"database_id": "184de7c8610e80218412ee866c33d11d",
		},
		"properties": map[string]interface{}{
			"Name": map[string]interface{}{
				"title": []map[string]interface{}{
					{"text": map[string]string{"content": expenses.recipient}},
				},
			},
			"Amount": map[string]interface{}{
				"number": amount,
			},
			"Date": map[string]interface{}{
				"date": map[string]string{"start": time.Now().Format("2006-01-02")},
			},
			"Category": map[string]interface{}{
				"relation": []map[string]string{
					{"id": categoryID},
				},
			},
		},
	}
	return json.Marshal(payload)
}

func createNotionRequest(notionKey string, body []byte) (req *http.Request, err error) {
	req, err = http.NewRequest("POST", "https://api.notion.com/v1/pages", bytes.NewBuffer(body))

	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+notionKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Notion-Version", "2022-06-28")

	return req, nil
}
