package ddcli

import (
	"context"
	"fmt"
)

// MenuItem is one sellable row from a restaurant menu.
//
// Fields:
//   - ItemID (string) — id for cart add-items / restaurant-item-details
//   - Name (string) — display name
//   - Description (string) — short description when present
//   - Price (float64) — unit price when present (CLI may omit)
type MenuItem struct {
	ItemID      string  `json:"item_id"`
	Name        string  `json:"name"`
	Description string  `json:"description,omitempty"`
	Price       float64 `json:"price,omitempty"`
}

// GetMenuResult is the structuredContent payload for menu.
//
// Fields:
//   - Success (bool) — CLI reported success
//   - Message (string) — optional status / error text from CLI
//   - MenuID (string) — menu id required by cart add-items
//   - Items ([]MenuItem) — menu rows
type GetMenuResult struct {
	Success bool       `json:"success"`
	Message string     `json:"message,omitempty"`
	MenuID  string     `json:"menu_id"`
	Items   []MenuItem `json:"items"`
}

// GetMenu loads the restaurant menu for a store.
//
// Params:
//   - ctx (context.Context) — cancel / timeout for the CLI call
//   - intentText (string) — required DoorDash intent blob; use BuildIntent() to build it
//   - storeID (string) — from SearchRestaurants (or order history)
//
// Returns:
//   - *GetMenuResult — menu_id + items on success
//   - error — missing args, CLI failure, bad JSON, or success:false from CLI
//
// Notes: runs `dd-cli --json-output menu --store-id …`. Needs prior `dd-cli login`.
// Pass MenuID and Items[].ItemID into AddCartItems.
func (client *CLIClient) GetMenu(ctx context.Context, intentText string, storeID string) (*GetMenuResult, error) {
	if intentText == "" {
		return nil, fmt.Errorf("ddcli: intent is required")
	}
	if storeID == "" {
		return nil, fmt.Errorf("ddcli: storeID is required")
	}

	cliStdout, err := client.RunCLICommand(ctx, "menu", "--store-id", storeID, "--intent", intentText)
	if err != nil {
		return nil, err
	}

	var menuResult GetMenuResult
	if err := decodeStructuredContent(cliStdout, &menuResult); err != nil {
		return nil, err
	}

	if !menuResult.Success {
		failureMessage := menuResult.Message
		if failureMessage == "" {
			failureMessage = "menu failed"
		}
		return &menuResult, fmt.Errorf("ddcli: %s", failureMessage)
	}

	return &menuResult, nil
}
