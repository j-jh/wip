package ddcli

import (
	"context"
	"fmt"
)

// DeliveryAddress is one saved delivery address from address list.
//
// Fields:
//   - AddressID (string) — id for address set / matching
//   - PrintableAddress (string) — street + city + state
//   - Label (*string) — Home/Work/etc; nil when unset
//   - IsDefault (bool) — best-effort default flag (may be false for all)
//   - Latitude, Longitude (float64) — coordinates for location-aware commands
type DeliveryAddress struct {
	AddressID        string  `json:"address_id"`
	PrintableAddress string  `json:"printable_address"`
	Label            *string `json:"label"`
	IsDefault        bool    `json:"is_default"`
	Latitude         float64 `json:"lat"`
	Longitude        float64 `json:"lng"`
}

// ListDeliveryAddressesResult is the structuredContent payload for address list.
//
// Fields:
//   - Success (bool) — CLI reported success
//   - Message (string) — optional status / error text from CLI
//   - Addresses ([]DeliveryAddress) — saved delivery addresses
type ListDeliveryAddressesResult struct {
	Success   bool              `json:"success"`
	Message   string            `json:"message,omitempty"`
	Addresses []DeliveryAddress `json:"addresses"`
}

// ListDeliveryAddresses lists the signed-in account’s saved delivery addresses.
//
// Params:
//   - ctx (context.Context) — cancel / timeout for the CLI call
//   - intentText (string) — required DoorDash intent blob; use BuildIntent() to build it
//
// Returns:
//   - *ListDeliveryAddressesResult — addresses plus success flag
//   - error — missing intent, CLI failure, bad JSON, or success:false from CLI
//
// Notes: runs `dd-cli --json-output address list --intent …`. Needs prior `dd-cli login`.
func (client *CLIClient) ListDeliveryAddresses(ctx context.Context, intentText string) (*ListDeliveryAddressesResult, error) {
	if intentText == "" {
		return nil, fmt.Errorf("ddcli: intent is required")
	}

	cliStdout, err := client.RunCLICommand(ctx, "address", "list", "--intent", intentText)
	if err != nil {
		return nil, err
	}

	var addressListResult ListDeliveryAddressesResult
	if err := decodeStructuredContent(cliStdout, &addressListResult); err != nil {
		return nil, err
	}

	// CLI can exit 0 but still report failure inside the JSON payload.
	if !addressListResult.Success {
		failureMessage := addressListResult.Message
		if failureMessage == "" {
			failureMessage = "address list failed"
		}
		return &addressListResult, fmt.Errorf("ddcli: %s", failureMessage)
	}

	return &addressListResult, nil
}
