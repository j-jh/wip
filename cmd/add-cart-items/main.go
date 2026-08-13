// Command add-cart-items runs dd-cli cart add-items for one simple item.
//
// Run:
//
//	go run ./cmd/add-cart-items \
//	  -store-id 123 -menu-id 456 -item-id 789 -item-name "Pad Thai" -quantity 1
//
// Optional: -cart-uuid (must still be open), -fulfillment delivery|pickup (new carts only).
// Needs dd-cli login. If an open cart already exists at the store, add reuses that session
// (even when -cart-uuid is omitted or points at a deleted cart).
package main

import (
	"context"
	"flag"
	"fmt"
	"log"

	"wip/internal/ddcli"
)

func main() {
	storeID := flag.String("store-id", "", "store id (required)")
	menuID := flag.String("menu-id", "", "menu id from get-menu (required)")
	itemID := flag.String("item-id", "", "menu item id (required)")
	itemName := flag.String("item-name", "", "item display name for CLI JSON (required)")
	quantity := flag.Int("quantity", 1, "how many to add")
	cartUUID := flag.String("cart-uuid", "", "optional existing cart uuid")
	fulfillment := flag.String("fulfillment", "", "optional delivery|pickup when creating a cart")
	flag.Parse()

	if *storeID == "" || *menuID == "" || *itemID == "" || *itemName == "" {
		log.Fatal("need -store-id -menu-id -item-id -item-name")
	}

	addItemsResult, err := (&ddcli.CLIClient{}).AddCartItems(
		context.Background(),
		ddcli.BuildIntent("add one menu item for a manual preview test", "add this to my cart"),
		ddcli.AddCartItemsOptions{
			StoreID:     *storeID,
			MenuID:      *menuID,
			CartUUID:    *cartUUID,
			Fulfillment: *fulfillment,
			Items: []ddcli.CartAddItem{
				{
					ItemID:   *itemID,
					ItemName: *itemName,
					Quantity: *quantity,
				},
			},
		},
	)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("=== cart add-items ===")
	fmt.Println("cart_uuid", addItemsResult.CartUUID)
	if addItemsResult.Message != "" {
		fmt.Println(addItemsResult.Message)
	}
	for _, itemError := range addItemsResult.ItemErrors {
		fmt.Println("item_error:", itemError.ErrorMessage)
	}
}
