// Command get-menu runs dd-cli menu and prints menu_id plus item ids/names.
//
// Run:
//
//	go run ./cmd/get-menu -store-id 123456
//
// Needs dd-cli login. Copy a store-id from search-restaurants.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"

	"wip/internal/ddcli"
)

func main() {
	storeID := flag.String("store-id", "", "restaurant store id from search (required)")
	flag.Parse()

	if *storeID == "" {
		log.Fatal("missing -store-id")
	}

	menuResult, err := (&ddcli.CLIClient{}).GetMenu(
		context.Background(),
		ddcli.BuildIntent("load a restaurant menu for a manual cart test", "show me the menu"),
		*storeID,
	)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("=== menu ===")
	fmt.Println("menu_id", menuResult.MenuID)
	for _, item := range menuResult.Items {
		fmt.Printf("%s  %s", item.ItemID, item.Name)
		if item.Price != 0 {
			fmt.Printf("  $%.2f", item.Price)
		}
		fmt.Println()
	}
}
