// Command order-status checks whether a submitted order cleared processing.
//
// Run:
//
//	go run ./cmd/order-status -order-uuid ORDER_UUID
//
// Optional: -poll repeats while status is pending (bounded).
//
// Needs dd-cli login. Report the order as placed only when status is successful.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"wip/internal/ddcli"
)

func main() {
	orderUUID := flag.String("order-uuid", "", "order uuid from submit-order (required)")
	poll := flag.Bool("poll", false, "re-check while status is pending (max 6 tries, 5s apart)")
	flag.Parse()

	if *orderUUID == "" {
		log.Fatal("missing -order-uuid")
	}

	doorDashClient := &ddcli.CLIClient{}
	ctx := context.Background()
	intentText := ddcli.BuildIntent(
		"check whether a submitted order cleared processing",
		"did my order go through?",
	)

	maxAttempts := 1
	if *poll {
		maxAttempts = 6
	}

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		statusResult, err := doorDashClient.GetOrderStatus(ctx, intentText, *orderUUID)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println("=== order status ===")
		fmt.Println("order_uuid", statusResult.OrderUUID)
		if statusResult.OrderUUID == "" {
			fmt.Println("order_uuid", *orderUUID)
		}
		fmt.Println("status", statusResult.Status)
		if statusResult.ErrorMessage != "" {
			fmt.Println("error_message", statusResult.ErrorMessage)
		}
		if statusResult.Message != "" {
			fmt.Println("message", statusResult.Message)
		}

		switch statusResult.Status {
		case "pending":
			if attempt == maxAttempts {
				fmt.Println("still pending — re-run with -poll or check again later")
				return
			}
			fmt.Println("pending — waiting 5s…")
			time.Sleep(5 * time.Second)
		case "successful":
			fmt.Println("order created successfully")
			return
		case "action_required":
			fmt.Println("finish verification in the DoorDash app or website")
			return
		case "failed", "not_found":
			fmt.Println("order did not go through (terminal)")
			return
		default:
			fmt.Println("unrecognized status — treat as not confirmed")
			return
		}
	}
}
