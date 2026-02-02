package main

import "regexp"

func extractTransactionDetails(htmlBody string) (amount string, recipient string) {
	amountRe := regexp.MustCompile(`Amount:</td>\s*<td>(.*?)</td>`)
	recipientRe := regexp.MustCompile(`To:</td>\s*<td>(.*?)</td>`)

	if match := amountRe.FindStringSubmatch(htmlBody); len(match) > 1 {
		amount = match[1]
	}

	if match := recipientRe.FindStringSubmatch(htmlBody); len(match) > 1 {
		recipient = match[1]
	}

	return amount, recipient
}
