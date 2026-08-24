// Command list-payment-methods runs dd-cli payment-method list.
//
// Run:
//
//	go run ./cmd/list-payment-methods
//
// Needs dd-cli login. Cards only — wallets are not listed. Empty cards[] does not
// mean no payment methods on file.
package main

import (
	"context"
	"fmt"
	"log"

	"wip/internal/ddcli"
)

func main() {
	paymentMethodsResult, err := (&ddcli.CLIClient{}).ListPaymentMethods(
		context.Background(),
		ddcli.BuildIntent("list saved payment cards before checkout", "what cards are on my account?"),
	)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("=== payment cards ===")
	if len(paymentMethodsResult.Cards) == 0 {
		fmt.Println("(no cards returned — wallets may still be on file)")
		return
	}

	defaultID := paymentMethodsResult.DefaultPaymentMethodID
	for _, card := range paymentMethodsResult.Cards {
		marker := ""
		if card.PaymentMethodID != "" && card.PaymentMethodID == defaultID {
			marker = " (default)"
		}
		fmt.Printf("%s %s •••• %s exp %s/%s%s\n",
			card.PaymentMethodID,
			card.Brand,
			card.Last4,
			card.ExpMonth,
			card.ExpYear,
			marker,
		)
	}
}
