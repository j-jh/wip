// Command reorder-preview runs the phase-1 reorder flow through quote preview.
//
// Flow: order history → cart collision check → reorder (or reuse cart) → preview.
// Stops before payment / submit — no charge.
//
// Run: go run ./cmd/reorder-preview
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
	runReorderFlow(doorDashClient)
}

// runReorderFlow walks phase-1 reorder through preview and stops (no payment / submit).
//
// If an open cart already exists at the past order's store, it reuses that cart for
// preview instead of calling reorder again (CLI allows only one open cart per store).
func runReorderFlow(doorDashClient *ddcli.CLIClient) {
	ctx := context.Background()
	userGoal := "reorder something I've ordered before"

	// Step 1: load recent history and pick the first reorderable order.
	orderHistoryResult, err := doorDashClient.ListOrderHistory(
		ctx,
		ddcli.BuildIntent("find a recent reorderable order", userGoal),
		ddcli.ListOrderHistoryOptions{
			MaxOrders:    20,
			LookbackDays: 365,
		},
	)
	if err != nil {
		log.Fatal(err)
	}

	selectedOrder := firstReorderableOrder(orderHistoryResult.Orders)
	if selectedOrder == nil {
		log.Fatal("no reorderable orders in history window")
	}

	fmt.Println("=== reorder flow: selected order ===")
	fmt.Println(selectedOrder.OrderDate, selectedOrder.StoreName, selectedOrder.OrderUUID)

	// Step 2: DoorDash allows only one open cart per store — check before reorder.
	openCartsResult, err := doorDashClient.ListOpenCarts(
		ctx,
		ddcli.BuildIntent("check for an existing open cart at the store before reorder", userGoal),
		selectedOrder.StoreID,
	)
	if err != nil {
		log.Fatal(err)
	}

	// Step 3: reuse an open cart, or create one via reorder.
	cartUUID, err := cartUUIDForReorder(ctx, doorDashClient, userGoal, selectedOrder, openCartsResult.Carts)
	if err != nil {
		log.Fatal(err)
	}

	// Step 4: quote only — do not submit or charge.
	previewResult, err := doorDashClient.PreviewOrder(
		ctx,
		ddcli.BuildIntent("preview pricing for the reordered cart without placing the order", userGoal),
		cartUUID,
	)
	if err != nil {
		log.Fatal(err)
	}

	printOrderPreview(previewResult)
	printPossibleItemDrops(selectedOrder, previewResult)
}

// firstReorderableOrder returns a pointer into orders, or nil if none qualify.
func firstReorderableOrder(orders []ddcli.PastOrder) *ddcli.PastOrder {
	for i := range orders {
		// Index into the slice so the pointer stays valid after the loop.
		pastOrder := &orders[i]
		if pastOrder.IsReorderable {
			return pastOrder
		}
	}
	return nil
}

// cartUUIDForReorder reuses an open cart at the store, or calls reorder to create one.
func cartUUIDForReorder(
	ctx context.Context,
	doorDashClient *ddcli.CLIClient,
	userGoal string,
	selectedOrder *ddcli.PastOrder,
	openCarts []ddcli.OpenCart,
) (string, error) {
	if len(openCarts) > 0 {
		existingCart := openCarts[0]
		fmt.Println("=== open cart collision — reusing existing cart (skip reorder) ===")
		fmt.Println(existingCart.CartUUID, existingCart.StoreName, "items=", existingCart.ItemsCount)
		return existingCart.CartUUID, nil
	}

	reorderResult, err := doorDashClient.ReorderPastOrder(
		ctx,
		ddcli.BuildIntent("create a cart from a past order for preview", userGoal),
		selectedOrder.OrderUUID,
	)
	if err != nil {
		return "", err
	}

	fmt.Println("=== reordered into new cart ===")
	fmt.Println(reorderResult.CartUUID)
	if reorderResult.Message != "" {
		fmt.Println(reorderResult.Message)
	}
	return reorderResult.CartUUID, nil
}

func printOrderPreview(previewResult *ddcli.PreviewOrderResult) {
	fmt.Println("=== order preview (stop — no submit) ===")
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

// printPossibleItemDrops warns when preview has fewer of an item than the past order.
// Reorder can omit unavailable items without failing the CLI call.
func printPossibleItemDrops(selectedOrder *ddcli.PastOrder, previewResult *ddcli.PreviewOrderResult) {
	previewQuantityByName := map[string]int{}
	for _, quoteItem := range previewResult.Quote.Items {
		previewQuantityByName[quoteItem.Name] += quoteItem.Quantity
	}

	for _, historyItem := range selectedOrder.Items {
		if previewQuantityByName[historyItem.Name] < historyItem.Quantity {
			fmt.Printf(
				"possible drop/missing: %s (history qty %d)\n",
				historyItem.Name,
				historyItem.Quantity,
			)
		}
	}
}
