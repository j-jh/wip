package ddcli

import (
	"context"
	"fmt"
	"strconv"
)

// SearchRestaurantsOptions holds filters for restaurant search.
//
// Fields:
//   - Query (string) — required search text (e.g. "thai near me")
//   - Limit (int) — max restaurants. Zero omits the flag (CLI default 5).
//   - Latitude / Longitude (*float64) — search location. Nil omits
//     (CLI uses DD_LAT/DD_LNG, then a Cupertino default — prefer explicit coords).
type SearchRestaurantsOptions struct {
	Query     string
	Limit     int
	Latitude  *float64
	Longitude *float64
}

// RestaurantStore is one restaurant from search.
//
// Fields:
//   - StoreID (string) — id for menu / cart add-items
//   - Name (string) — display name
//   - Description (string) — short blurb when present
type RestaurantStore struct {
	StoreID     string `json:"store_id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// SearchRestaurantsResult is the structuredContent payload for search.
//
// Fields:
//   - Success (bool) — CLI reported success
//   - Message (string) — optional status / error text from CLI
//   - Stores ([]RestaurantStore) — matching restaurants
type SearchRestaurantsResult struct {
	Success bool              `json:"success"`
	Message string            `json:"message,omitempty"`
	Stores  []RestaurantStore `json:"stores"`
}

// SearchRestaurants finds nearby restaurants by query text.
//
// Params:
//   - ctx (context.Context) — cancel / timeout for the CLI call
//   - intentText (string) — required DoorDash intent blob; use BuildIntent() to build it
//   - options (SearchRestaurantsOptions) — query required; limit/lat/lng optional
//
// Returns:
//   - *SearchRestaurantsResult — stores plus success flag
//   - error — missing args, CLI failure, bad JSON, or success:false from CLI
//
// Notes: runs `dd-cli --json-output search …`. Needs prior `dd-cli login`.
// Restaurant-focused — grocery/retail queries often return no stores (use FindNearbyStores).
func (client *CLIClient) SearchRestaurants(ctx context.Context, intentText string, options SearchRestaurantsOptions) (*SearchRestaurantsResult, error) {
	if intentText == "" {
		return nil, fmt.Errorf("ddcli: intent is required")
	}
	if options.Query == "" {
		return nil, fmt.Errorf("ddcli: query is required")
	}

	cliArgs := []string{"search", "--query", options.Query, "--intent", intentText}
	if options.Limit != 0 {
		cliArgs = append(cliArgs, "--limit", strconv.Itoa(options.Limit))
	}
	if options.Latitude != nil && options.Longitude != nil {
		latitudeText := strconv.FormatFloat(*options.Latitude, 'f', -1, 64)
		longitudeText := strconv.FormatFloat(*options.Longitude, 'f', -1, 64)
		cliArgs = append(cliArgs, "--lat", latitudeText, "--lng", longitudeText)
	}

	cliStdout, err := client.RunCLICommand(ctx, cliArgs...)
	if err != nil {
		return nil, err
	}

	var searchResult SearchRestaurantsResult
	if err := decodeStructuredContent(cliStdout, &searchResult); err != nil {
		return nil, err
	}

	if !searchResult.Success {
		failureMessage := searchResult.Message
		if failureMessage == "" {
			failureMessage = "search failed"
		}
		return &searchResult, fmt.Errorf("ddcli: %s", failureMessage)
	}

	return &searchResult, nil
}
