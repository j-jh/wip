package ddcli

import (
	"context"
	"fmt"
	"strconv"
)

// FindNearbyStoresOptions holds optional filters for find-nearby-stores.
//
// Fields:
//   - Vertical (string) — grocery|alcohol|convenience|pets|retail|nv. Empty omits (CLI default grocery).
//   - MaxStores (int) — max stores to return. Zero omits the flag (CLI default 10).
//   - Latitude / Longitude (*float64) — override search location. Nil omits (CLI uses default delivery address).
type FindNearbyStoresOptions struct {
	Vertical   string
	MaxStores  int
	Latitude   *float64
	Longitude  *float64
}

// NearbyStore is one non-restaurant store from find-nearby-stores.
//
// Fields:
//   - StoreID (string) — id for downstream store / cart commands
//   - Name (string) — store display name
//   - MenuID (string) — menu id when present (often empty for retail/grocery)
//   - ImageURL (string) — best-effort image url
//   - BusinessID (int) — business id
//   - BusinessVerticalID (int) — numeric merchant-type id
//   - DistanceMeters (float64) — distance in meters (divide by 1609 for miles)
//   - DeliveryTime (string) — best-effort ETA string
//   - AvailabilityStatus (string) — e.g. available, opening_soon
//   - PrintableAddress (*string) — address when present; nil when unset
type NearbyStore struct {
	StoreID            string  `json:"store_id"`
	Name               string  `json:"name"`
	MenuID             string  `json:"menu_id"`
	ImageURL           string  `json:"image_url"`
	BusinessID         int     `json:"business_id"`
	BusinessVerticalID int     `json:"business_vertical_id"`
	DistanceMeters     float64 `json:"distance_meters"`
	DeliveryTime       string  `json:"delivery_time"`
	AvailabilityStatus string  `json:"availability_status"`
	PrintableAddress   *string `json:"printable_address"`
}

// FindNearbyStoresResult is the structuredContent payload for find-nearby-stores.
//
// Fields:
//   - Success (bool) — CLI reported success
//   - Message (string) — optional status / error text from CLI
//   - Stores ([]NearbyStore) — nearby non-restaurant stores
type FindNearbyStoresResult struct {
	Success bool          `json:"success"`
	Message string        `json:"message,omitempty"`
	Stores  []NearbyStore `json:"stores"`
}

// FindNearbyStores lists nearby non-restaurant stores within the CLI’s fixed radius.
//
// Params:
//   - ctx (context.Context) — cancel / timeout for the CLI call
//   - intentText (string) — required DoorDash intent blob; use BuildIntent() to build it
//   - options (FindNearbyStoresOptions) — optional vertical/max/lat/lng; empty/nil omit those flags
//
// Returns:
//   - *FindNearbyStoresResult — stores plus success flag
//   - error — missing intent, CLI failure, bad JSON, or success:false from CLI
//
// Notes: runs `dd-cli --json-output find-nearby-stores …`. Needs prior `dd-cli login`.
// For restaurants use search (not yet wrapped). Default vertical when omitted: grocery.
func (client *CLIClient) FindNearbyStores(ctx context.Context, intentText string, options FindNearbyStoresOptions) (*FindNearbyStoresResult, error) {
	if intentText == "" {
		return nil, fmt.Errorf("ddcli: intent is required")
	}

	cliArgs := []string{"find-nearby-stores", "--intent", intentText}
	if options.Vertical != "" {
		cliArgs = append(cliArgs, "--vertical", options.Vertical)
	}
	if options.MaxStores != 0 {
		cliArgs = append(cliArgs, "--max", strconv.Itoa(options.MaxStores))
	}
	if options.Latitude != nil && options.Longitude != nil {
		cliArgs = append(cliArgs,
			"--lat", strconv.FormatFloat(*options.Latitude, 'f', -1, 64),
			"--lng", strconv.FormatFloat(*options.Longitude, 'f', -1, 64),
		)
	}

	cliStdout, err := client.RunCLICommand(ctx, cliArgs...)
	if err != nil {
		return nil, err
	}

	var nearbyStoresResult FindNearbyStoresResult
	if err := decodeStructuredContent(cliStdout, &nearbyStoresResult); err != nil {
		return nil, err
	}
	if !nearbyStoresResult.Success {
		failureMessage := nearbyStoresResult.Message
		if failureMessage == "" {
			failureMessage = "find nearby stores failed"
		}
		return &nearbyStoresResult, fmt.Errorf("ddcli: %s", failureMessage)
	}
	return &nearbyStoresResult, nil
}
