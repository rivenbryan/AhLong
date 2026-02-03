package main

import (
	"regexp"
	"strconv"
	"strings"
)

type Transaction struct {
	Amount, Recipient string
}
type Expenses struct {
	category, amount, recipient string
}

func extractTransactionDetails(htmlBody string) Transaction {
	amountRe := regexp.MustCompile(`Amount:</td>\s*<td[^>]*>(.*?)</td>`)
	recipientRe := regexp.MustCompile(`To:</td>\s*<td[^>]*>(.*?)</td>`)

	var amount string
	var recipient string

	if match := amountRe.FindStringSubmatch(htmlBody); len(match) > 1 {
		amount = match[1]
	}

	if match := recipientRe.FindStringSubmatch(htmlBody); len(match) > 1 {
		recipient = match[1]
	}

	return Transaction{
		amount, recipient,
	}
}

func extractTelegramResponse(response TelegramResponse) Expenses {
	parts := strings.Split(response.CallbackQuery.Data, "|")

	category := parts[0]
	amount := parts[1]
	recipient := parts[2]

	return Expenses{category, amount, recipient}
}

func processAmount(amount string) (float64, error) {
	amountStr := strings.TrimPrefix(amount, "SGD")
	amt, err := strconv.ParseFloat(amountStr, 64)
	return amt, err
}
