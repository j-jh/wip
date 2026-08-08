package main

import (
	"context"
	"fmt"
	"log"

	"wip/internal/ddcli"
)

func main() {
	doorDashClient := &ddcli.CLIClient{}
	runListDeliveryAddresses(doorDashClient)
}

// runListDeliveryAddresses prints saved delivery addresses via ListDeliveryAddresses.
func runListDeliveryAddresses(doorDashClient *ddcli.CLIClient) {
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
