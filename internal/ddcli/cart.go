package ddcli

import (
	"context"
	"fmt"
)

// OpenCartItem is one line item on an open cart from cart list.
//
// Fields:
//   - LineID (string) — cart line id
//   - ItemID (string) — menu/item id
//   - Name (string) — display name
//   - Quantity (int) — count in cart
//   - Price (float64) — unit/line price in USD as returned by CLI
type OpenCartItem struct {
	LineID   string  `json:"id"`
	ItemID   string  `json:"item_id"`
	Name     string  `json:"name"`
	Quantity int     `json:"quantity"`
	Price    float64 `json:"price"`
}

// OpenCart is one active (unsubmitted) cart from cart list.
//
// Fields:
//   - CartUUID (string) — id for preview / submit / cart show
//   - StoreID (string) — store id
//   - StoreName (string) — human-readable store name
//   - Items ([]OpenCartItem) — line items
//   - ItemsCount (int) — line count
//   - CreatedAtMs (int64) — epoch ms; may be zero
//   - UpdatedAtMs (int64) — epoch ms; may be zero
type OpenCart struct {
	CartUUID    string         `json:"cart_uuid"`
	StoreID     string         `json:"store_id"`
	StoreName   string         `json:"store_name"`
	Items       []OpenCartItem `json:"items"`
	ItemsCount  int            `json:"items_count"`
	CreatedAtMs int64          `json:"created_at"`
	UpdatedAtMs int64          `json:"updated_at"`
}

// ListOpenCartsResult is the structuredContent payload for cart list.
//
// Fields:
//   - Success (bool) — CLI reported success
//   - Message (string) — optional status / error text from CLI
//   - Carts ([]OpenCart) — open carts (possibly filtered by store)
type ListOpenCartsResult struct {
	Success bool       `json:"success"`
	Message string     `json:"message,omitempty"`
	Carts   []OpenCart `json:"carts"`
}

// ListOpenCarts lists active (open / unsubmitted) carts.
//
// Params:
//   - ctx (context.Context) — cancel / timeout for the CLI call
//   - intentText (string) — required DoorDash intent blob; use BuildIntent() to build it
//   - storeID (string) — optional store filter; empty omits --store-id
//
// Returns:
//   - *ListOpenCartsResult — open carts plus success flag
//   - error — missing intent, CLI failure, bad JSON, or success:false from CLI
//
// Notes: runs `dd-cli --json-output cart list …`. Needs prior `dd-cli login`.
// One open cart per store — use before order reorder to detect collisions.
func (client *CLIClient) ListOpenCarts(ctx context.Context, intentText string, storeID string) (*ListOpenCartsResult, error) {
	if intentText == "" {
		return nil, fmt.Errorf("ddcli: intent is required")
	}

	cliArgs := []string{"cart", "list", "--intent", intentText}
	// Only add optional flags when the caller provided a value (zero/empty = CLI default).
	if storeID != "" {
		cliArgs = append(cliArgs, "--store-id", storeID)
	}

	cliStdout, err := client.RunCLICommand(ctx, cliArgs...)
	if err != nil {
		return nil, err
	}

	var openCartsResult ListOpenCartsResult
	if err := decodeStructuredContent(cliStdout, &openCartsResult); err != nil {
		return nil, err
	}

	if !openCartsResult.Success {
		failureMessage := openCartsResult.Message
		if failureMessage == "" {
			failureMessage = "cart list failed"
		}
		return &openCartsResult, fmt.Errorf("ddcli: %s", failureMessage)
	}

	return &openCartsResult, nil
}
