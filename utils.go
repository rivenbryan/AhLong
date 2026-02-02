package main

import "regexp"

type Transaction struct {
	Amount, Recipient string
}

func extractTransactionDetails(htmlBody string) Transaction {
	amountRe := regexp.MustCompile(`Amount:</td>\s*<td>(.*?)</td>`)
	recipientRe := regexp.MustCompile(`To:</td>\s*<td>(.*?)</td>`)

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
