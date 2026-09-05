// Command submit-order places a cart with an explicit human gate (charges money).
//
// DESTRUCTIVE — charges the default payment method on file.
//
// Run (after preview + payment list):
//
//	go run ./cmd/submit-order \
//	  -cart-uuid CART_UUID \
//	  -tip-cents 0 \
//	  -confirm-submit
//
// Without -confirm-submit the command prints a quote + card summary and exits.
// Tip is in CENTS (500 = $5.00). Use 0 for pickup / explicit no tip.
//
// Needs dd-cli login. After success, poll with cmd/order-status.
package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"wip/internal/ddcli"
)

func main() {
	cartUUID := flag.String("cart-uuid", "", "cart uuid to submit (required)")
	tipCents := flag.Int("tip-cents", -1, "Dasher tip in CENTS (required; use 0 for pickup / no tip)")
	fulfillment := flag.String("fulfillment", "", "optional delivery|pickup; empty keeps cart mode")
	confirmSubmit := flag.Bool("confirm-submit", false, "required: acknowledge this will charge money")
	skipTypedYes := flag.Bool("yes", false, "skip typing yes after the summary (still needs -confirm-submit)")
	flag.Parse()

	if *cartUUID == "" {
		log.Fatal("missing -cart-uuid")
	}
	if *tipCents < 0 {
		log.Fatal("missing -tip-cents (cents, not dollars; use 0 for pickup / no tip)")
	}

	doorDashClient := &ddcli.CLIClient{}
	ctx := context.Background()
	userGoal := "place a human-confirmed order after preview"

	// Always show quote + payment before any charge path.
	previewResult, err := doorDashClient.PreviewOrder(
		ctx,
		ddcli.BuildIntent("preview cart before gated submit", userGoal),
		*cartUUID,
	)
	if err != nil {
		log.Fatal(err)
	}
	printPreview(previewResult)

	paymentLabel := paymentConfirmationLabel(ctx, doorDashClient, userGoal)
	fmt.Println()
	fmt.Println("=== payment ===")
	fmt.Println(paymentLabel)
	fmt.Printf("tip_cents %d ($%.2f)\n", *tipCents, float64(*tipCents)/100)

	if !*confirmSubmit {
		fmt.Println()
		fmt.Println("Stopped — no charge. Re-run with -confirm-submit after you approve.")
		fmt.Println("Example:")
		fmt.Printf("  go run ./cmd/submit-order -cart-uuid %s -tip-cents %d -confirm-submit\n",
			*cartUUID, *tipCents)
		return
	}

	if !*skipTypedYes {
		fmt.Println()
		fmt.Print(`Type "yes" to place the order and charge (anything else cancels): `)
		line, readErr := bufio.NewReader(os.Stdin).ReadString('\n')
		if readErr != nil || strings.TrimSpace(line) != "yes" {
			fmt.Println("Cancelled — no charge.")
			return
		}
	}

	fmt.Println()
	fmt.Println("Submitting…")
	submitResult, err := doorDashClient.SubmitOrder(
		ctx,
		ddcli.BuildIntent("submit cart after explicit human confirmation", userGoal),
		ddcli.SubmitOrderOptions{
			CartUUID:    *cartUUID,
			TipCents:    *tipCents,
			Fulfillment: *fulfillment,
		},
	)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("=== order submit (accepted into processing) ===")
	fmt.Println("order_uuid", submitResult.OrderUUID)
	if submitResult.Message != "" {
		fmt.Println(submitResult.Message)
	}
	fmt.Println()
	fmt.Println("Poll until not pending:")
	fmt.Printf("  go run ./cmd/order-status -order-uuid %s\n", submitResult.OrderUUID)
}

func printPreview(previewResult *ddcli.PreviewOrderResult) {
	fmt.Println("=== order preview ===")
	fmt.Println("cart", previewResult.CartUUID)
	fmt.Println("fulfillment", previewResult.Quote.FulfillmentType)
	fmt.Println("address", previewResult.Quote.DeliveryAddress)
	fmt.Println("total_before_tip", previewResult.Quote.NetTotalBeforeTip.DisplayString)
	for _, quoteItem := range previewResult.Quote.Items {
		fmt.Printf("  %dx %s %s\n", quoteItem.Quantity, quoteItem.Name, quoteItem.UnitPriceDisplay)
	}
	for _, lineItem := range previewResult.Quote.LineItems {
		fmt.Printf("  %s %s\n", lineItem.Label, lineItem.FinalMoney.DisplayString)
	}
}

// paymentConfirmationLabel names the default card when possible; otherwise wallet-aware text.
func paymentConfirmationLabel(ctx context.Context, doorDashClient *ddcli.CLIClient, userGoal string) string {
	paymentMethodsResult, err := doorDashClient.ListPaymentMethods(
		ctx,
		ddcli.BuildIntent("name the default card before submit", userGoal),
	)
	if err != nil {
		return "default on file (card or wallet) — payment-method list failed; confirm in app if unsure"
	}

	defaultID := paymentMethodsResult.DefaultPaymentMethodID
	for _, card := range paymentMethodsResult.Cards {
		if card.PaymentMethodID != "" && card.PaymentMethodID == defaultID {
			return fmt.Sprintf("default card: %s ending %s", card.Brand, card.Last4)
		}
	}
	if len(paymentMethodsResult.Cards) == 1 {
		card := paymentMethodsResult.Cards[0]
		return fmt.Sprintf("card on file: %s ending %s (no default_payment_method_id)", card.Brand, card.Last4)
	}
	if len(paymentMethodsResult.Cards) == 0 {
		return "DoorDash will charge whatever default is on file (card or wallet) — cards[] was empty"
	}
	return "DoorDash will charge the account default (could not match default_payment_method_id to cards[])"
}
