// Command list-addresses prints saved DoorDash delivery addresses.
//
// Run: go run ./cmd/list-addresses
// Needs: dd-cli on PATH (or DD_CLI_BIN) and a prior `dd-cli login`.
package main

import (
	"context"
	"fmt"
	"log"

	"wip/internal/ddcli"
)

func main() {
	doorDashClient := &ddcli.CLIClient{}

	// context.Background() is a non-cancelable context — fine for a short demo script.
	addressListResult, err := doorDashClient.ListDeliveryAddresses(
		context.Background(),
		ddcli.BuildIntent(
			"list saved delivery addresses",
			"where do I have saved?",
		),
	)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("=== delivery addresses ===")
	for _, deliveryAddress := range addressListResult.Addresses {
		fmt.Println(deliveryAddress.AddressID, deliveryAddress.PrintableAddress)
	}
}
