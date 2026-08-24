// Command delete-cart abandons an open cart (dd-cli cart delete).
//
// Run:
//
//	go run ./cmd/delete-cart -cart-uuid <uuid>
//
// Needs dd-cli login. After this, treat the uuid as invalid.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"

	"wip/internal/ddcli"
)

func main() {
	cartUUID := flag.String("cart-uuid", "", "open cart uuid to abandon (required)")
	flag.Parse()

	if *cartUUID == "" {
		log.Fatal("missing -cart-uuid")
	}

	deleteCartResult, err := (&ddcli.CLIClient{}).DeleteCart(
		context.Background(),
		ddcli.BuildIntent("abandon an open cart after a manual test", "clear this cart"),
		*cartUUID,
	)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("=== cart delete (abandoned) ===")
	fmt.Println("cart_uuid", deleteCartResult.CartUUID)
	if deleteCartResult.Message != "" {
		fmt.Println(deleteCartResult.Message)
	}
}
