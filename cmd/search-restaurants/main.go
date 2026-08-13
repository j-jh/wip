// Command search-restaurants runs dd-cli search and prints store ids/names.
//
// Run:
//
//	go run ./cmd/search-restaurants -query "thai" -limit 5
//
// Optional: -lat / -lng (both required together). Needs dd-cli login.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"

	"wip/internal/ddcli"
)

func main() {
	query := flag.String("query", "", "search text (required), e.g. thai")
	limit := flag.Int("limit", 5, "max restaurants (CLI default is 5)")
	lat := flag.Float64("lat", 0, "optional latitude (pass with -lng)")
	lng := flag.Float64("lng", 0, "optional longitude (pass with -lat)")
	flag.Parse()

	if *query == "" {
		log.Fatal("missing -query")
	}

	options := ddcli.SearchRestaurantsOptions{
		Query: *query,
		Limit: *limit,
	}
	if *lat != 0 || *lng != 0 {
		if *lat == 0 || *lng == 0 {
			log.Fatal("pass both -lat and -lng together")
		}
		latitude := *lat
		longitude := *lng
		options.Latitude = &latitude
		options.Longitude = &longitude
	}

	searchResult, err := (&ddcli.CLIClient{}).SearchRestaurants(
		context.Background(),
		ddcli.BuildIntent("find restaurants for a manual menu/cart test", *query),
		options,
	)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("=== search restaurants ===")
	for _, store := range searchResult.Stores {
		fmt.Println(store.StoreID, store.Name)
	}
}
