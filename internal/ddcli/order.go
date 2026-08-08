package ddcli

import (
	"context"
	"fmt"
	"strconv"
)

// ListOrderHistoryOptions holds optional filters for order history.
//
// Fields:
//   - MaxOrders (int) — max orders to return (CLI 1–100). Zero omits the flag (CLI default 50).
//   - LookbackDays (int) — how many days back to search (CLI 0–365). Zero omits the flag (CLI default 90).
type ListOrderHistoryOptions struct {
	MaxOrders    int
	LookbackDays int
}

// PastOrderItem is one line item on a past order.
//
// Fields:
//   - ItemID (string) — menu/item id
//   - Name (string) — display name
//   - Quantity (int) — count ordered
type PastOrderItem struct {
	ItemID   string `json:"item_id"`
	Name     string `json:"name"`
	Quantity int    `json:"quantity"`
}

// PastOrder is one entry from order history.
//
// Fields:
//   - OrderUUID (string) — id for receipt / reorder / status
//   - StoreID (string) — store id for downstream store commands
//   - StoreName (string) — human-readable store name
//   - StoreImageURL (string) — best-effort image path/url
//   - IsReorderable (bool) — false means skip reorder for this order
//   - OrderDate (string) — timestamp from CLI (ISO-8601 string)
//   - BusinessVerticalID (int) — merchant vertical id
//   - BusinessID (int) — business id
//   - FulfillmentType (string) — e.g. FULFILLMENT_TYPE_PICKUP
//   - OrderTarget (string) — e.g. ORDER_TARGET_RESTAURANT
//   - Items ([]PastOrderItem) — line items on the order
type PastOrder struct {
	OrderUUID          string          `json:"order_uuid"`
	StoreID            string          `json:"store_id"`
	StoreName          string          `json:"store_name"`
	StoreImageURL      string          `json:"store_image_url"`
	IsReorderable      bool            `json:"is_reorderable"`
	OrderDate          string          `json:"order_date"`
	BusinessVerticalID int             `json:"business_vertical_id"`
	BusinessID         int             `json:"business_id"`
	FulfillmentType    string          `json:"fulfillment_type"`
	OrderTarget        string          `json:"order_target"`
	Items              []PastOrderItem `json:"items"`
}

// ListOrderHistoryResult is the structuredContent payload for order history.
//
// Fields:
//   - Success (bool) — CLI reported success
//   - Message (string) — optional status / error text from CLI
//   - PageFull (bool) — true if more results may exist; widen max/days
//   - Orders ([]PastOrder) — past orders in the window
type ListOrderHistoryResult struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	PageFull bool       `json:"page_full"`
	Orders  []PastOrder `json:"orders"`
}

// ListOrderHistory lists recent orders for the signed-in account.
//
// Params:
//   - ctx (context.Context) — cancel / timeout for the CLI call
//   - intentText (string) — required DoorDash intent blob; use BuildIntent() to build it
//   - options (ListOrderHistoryOptions) — optional max/days; zero fields omit those flags
//
// Returns:
//   - *ListOrderHistoryResult — orders plus success / page_full flags
//   - error — missing intent, CLI failure, bad JSON, or success:false from CLI
//
// Notes: runs `dd-cli --json-output order history …`. Needs prior `dd-cli login`.
// CLI defaults when flags omitted: --max 50, --days 90.
func (client *CLIClient) ListOrderHistory(ctx context.Context, intentText string, options ListOrderHistoryOptions) (*ListOrderHistoryResult, error) {
	if intentText == "" {
		return nil, fmt.Errorf("ddcli: intent is required")
	}

	cliArgs := []string{"order", "history", "--intent", intentText}
	if options.MaxOrders != 0 {
		cliArgs = append(cliArgs, "--max", strconv.Itoa(options.MaxOrders))
	}
	if options.LookbackDays != 0 {
		cliArgs = append(cliArgs, "--days", strconv.Itoa(options.LookbackDays))
	}

	cliStdout, err := client.RunCLICommand(ctx, cliArgs...)
	if err != nil {
		return nil, err
	}

	var orderHistoryResult ListOrderHistoryResult
	if err := decodeStructuredContent(cliStdout, &orderHistoryResult); err != nil {
		return nil, err
	}
	if !orderHistoryResult.Success {
		failureMessage := orderHistoryResult.Message
		if failureMessage == "" {
			failureMessage = "order history failed"
		}
		return &orderHistoryResult, fmt.Errorf("ddcli: %s", failureMessage)
	}
	return &orderHistoryResult, nil
}
