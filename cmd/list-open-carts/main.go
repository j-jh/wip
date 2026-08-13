// Command list-open-carts runs dd-cli cart list.
//
// Run:
//
//	go run ./cmd/list-open-carts
//	go run ./cmd/list-open-carts -store-id 123456
//
// Needs dd-cli login. Use before add-cart-items (one open cart per store).
package main

import (
	"context"
	"flag"
	"fmt"
	"log"

	"wip/internal/ddcli"
)

func main() {
	storeID := flag.String("store-id", "", "optional store filter; empty lists all open carts")
	flag.Parse()

	openCartsResult, err := (&ddcli.CLIClient{}).ListOpenCarts(
		context.Background(),
		ddcli.BuildIntent("check open carts before adding items", "what carts do I have open?"),
		*storeID,
	)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("=== open carts ===")
	if len(openCartsResult.Carts) == 0 {
		fmt.Println("(none)")
		return
	}
	for _, cart := range openCartsResult.Carts {
		fmt.Println(cart.CartUUID, cart.StoreID, cart.StoreName, "items=", cart.ItemsCount)
	}
}
