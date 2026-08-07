package ddcli

import (
	"context"
	"encoding/json"
	"fmt"
)

// Address is one saved delivery address from address list.
//
// Fields:
//   - AddressID (string) — id for address set / matching
//   - PrintableAddress (string) — street + city + state
//   - Label (*string) — Home/Work/etc; nil when unset
//   - IsDefault (bool) — best-effort default flag (may be false for all)
//   - Lat, Lng (float64) — coordinates for location-aware commands
type Address struct {
	AddressID        string  `json:"address_id"`
	PrintableAddress string  `json:"printable_address"`
	Label            *string `json:"label"`
	IsDefault        bool    `json:"is_default"`
	Lat              float64 `json:"lat"`
	Lng              float64 `json:"lng"`
}

// AddressListResult is the JSON envelope for address list.
//
// Fields:
//   - Success (bool) — CLI reported success
//   - Message (string) — optional status / error text from CLI
//   - Addresses ([]Address) — saved delivery addresses
type AddressListResult struct {
	Success   bool      `json:"success"`
	Message   string    `json:"message,omitempty"`
	Addresses []Address `json:"addresses"`
}

// AddressList lists the signed-in account’s saved delivery addresses.
//
// Params:
//   - ctx (context.Context) — cancel / timeout for the CLI call
//   - intent (string) — required DoorDash intent blob; use Intent() to build it
//
// Returns:
//   - *AddressListResult — addresses plus success flag
//   - error — missing intent, CLI failure, bad JSON, or success:false from CLI
//
// Notes: runs `dd-cli --json-output address list --intent …`. Needs prior `dd-cli login`.
func (c *Client) AddressList(ctx context.Context, intent string) (*AddressListResult, error) {
	if intent == "" {
		return nil, fmt.Errorf("ddcli: intent is required")
	}

	raw, err := c.Run(ctx, "address", "list", "--intent", intent)
	if err != nil {
		return nil, err
	}

	var res AddressListResult
	if err := json.Unmarshal(raw, &res); err != nil {
		return nil, fmt.Errorf("ddcli: decode address list: %w", err)
	}
	if !res.Success {
		msg := res.Message
		if msg == "" {
			msg = "address list failed"
		}
		return &res, fmt.Errorf("ddcli: %s", msg)
	}
	return &res, nil
}
