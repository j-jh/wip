package ddcli

import (
	"context"
	"encoding/json"
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

// CartAddItem is one entry in cart add-items --items-json.
//
// Fields:
//   - ItemID (string) — menu item id (from GetMenu)
//   - ItemName (string) — display name (required by CLI JSON schema)
//   - Quantity (int) — how many to add (APPENDs / sums with existing same modifiers)
type CartAddItem struct {
	ItemID   string `json:"item_id"`
	ItemName string `json:"item_name"`
	Quantity int    `json:"quantity"`
}

// AddCartItemsOptions holds required and optional flags for cart add-items.
//
// Fields:
//   - StoreID (string) — required store id
//   - MenuID (string) — required menu id (from GetMenu)
//   - Items ([]CartAddItem) — required; at least one item
//   - CartUUID (string) — optional existing cart; empty creates/appends per CLI rules
//   - Fulfillment (string) — optional "delivery" or "pickup"; only applies when creating a cart
type AddCartItemsOptions struct {
	StoreID     string
	MenuID      string
	Items       []CartAddItem
	CartUUID    string
	Fulfillment string
}

// CartAddItemError is one per-item failure from cart add-items.
//
// Fields:
//   - ErrorMessage (string) — why this item failed
type CartAddItemError struct {
	ErrorMessage string `json:"error_message"`
}

// AddCartItemsResult is the structuredContent payload for cart add-items.
//
// Fields:
//   - Success (bool) — CLI reported success
//   - Message (string) — optional status / error text
//   - CartUUID (string) — cart to use for preview / further add-items
//   - ItemErrors ([]CartAddItemError) — partial failures when present
type AddCartItemsResult struct {
	Success    bool               `json:"success"`
	Message    string             `json:"message,omitempty"`
	CartUUID   string             `json:"cart_uuid"`
	ItemErrors []CartAddItemError `json:"item_errors,omitempty"`
}

// AddCartItems adds items to a store cart (creates one if needed).
//
// Params:
//   - ctx (context.Context) — cancel / timeout for the CLI call
//   - intentText (string) — required DoorDash intent blob; use BuildIntent() to build it
//   - options (AddCartItemsOptions) — store, menu, items; cart uuid / fulfillment optional
//
// Returns:
//   - *AddCartItemsResult — cart_uuid on success
//   - error — missing args, CLI failure, bad JSON, or success:false from CLI
//
// Notes: runs `dd-cli --json-output cart add-items …`. Mutation.
// Pre-flight with ListOpenCarts(storeID): one open cart per store.
func (client *CLIClient) AddCartItems(ctx context.Context, intentText string, options AddCartItemsOptions) (*AddCartItemsResult, error) {
	if intentText == "" {
		return nil, fmt.Errorf("ddcli: intent is required")
	}
	if options.StoreID == "" {
		return nil, fmt.Errorf("ddcli: storeID is required")
	}
	if options.MenuID == "" {
		return nil, fmt.Errorf("ddcli: menuID is required")
	}
	if len(options.Items) == 0 {
		return nil, fmt.Errorf("ddcli: items are required")
	}

	itemsJSON, err := json.Marshal(options.Items)
	if err != nil {
		return nil, fmt.Errorf("ddcli: encode items-json: %w", err)
	}

	cliArgs := []string{
		"cart", "add-items",
		"--store-id", options.StoreID,
		"--menu-id", options.MenuID,
		"--items-json", string(itemsJSON),
		"--intent", intentText,
	}
	if options.CartUUID != "" {
		cliArgs = append(cliArgs, "--cart-uuid", options.CartUUID)
	}
	if options.Fulfillment != "" {
		cliArgs = append(cliArgs, "--fulfillment", options.Fulfillment)
	}

	cliStdout, err := client.RunCLICommand(ctx, cliArgs...)
	if err != nil {
		return nil, err
	}

	var addItemsResult AddCartItemsResult
	if err := decodeStructuredContent(cliStdout, &addItemsResult); err != nil {
		return nil, err
	}

	if !addItemsResult.Success {
		failureMessage := addItemsResult.Message
		if failureMessage == "" {
			failureMessage = "cart add-items failed"
		}
		return &addItemsResult, fmt.Errorf("ddcli: %s", failureMessage)
	}

	return &addItemsResult, nil
}
