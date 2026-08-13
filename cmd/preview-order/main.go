// Command preview-order runs dd-cli order preview for a cart (no charge).
//
// Run:
//
//	go run ./cmd/preview-order -cart-uuid <uuid>
//
// Needs dd-cli login. Copy cart-uuid from add-cart-items (or list-open-carts).
package main

import (
	"context"
	"flag"
	"fmt"
	"log"

	"wip/internal/ddcli"
)

func main() {
	cartUUID := flag.String("cart-uuid", "", "cart uuid to quote (required)")
	flag.Parse()

	if *cartUUID == "" {
		log.Fatal("missing -cart-uuid")
	}

	previewResult, err := (&ddcli.CLIClient{}).PreviewOrder(
		context.Background(),
		ddcli.BuildIntent("preview pricing for a manually built cart", "how much is this cart?"),
		*cartUUID,
	)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("=== order preview (no submit) ===")
	fmt.Println("cart", previewResult.CartUUID)
	fmt.Println("fulfillment", previewResult.Quote.FulfillmentType)
	fmt.Println("address", previewResult.Quote.DeliveryAddress)
	fmt.Println("dashpass", previewResult.Quote.IsDashPassApplied)
	fmt.Println("total_before_tip", previewResult.Quote.NetTotalBeforeTip.DisplayString)
	for _, quoteItem := range previewResult.Quote.Items {
		fmt.Printf("  %dx %s %s\n", quoteItem.Quantity, quoteItem.Name, quoteItem.UnitPriceDisplay)
	}
	for _, lineItem := range previewResult.Quote.LineItems {
		fmt.Printf("  %s %s\n", lineItem.Label, lineItem.FinalMoney.DisplayString)
	}
}
