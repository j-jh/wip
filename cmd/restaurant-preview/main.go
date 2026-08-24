// Command restaurant-preview walks search → menu → add → preview with prompts.
//
// Stops after quote — no payment / submit.
//
// Run: go run ./cmd/restaurant-preview
// Starts with dd-cli login, then prompts through search → preview (no charge).
// Search often needs lat/lng if no default address resolves.
package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"wip/internal/ddcli"
)

func main() {
	doorDashClient := &ddcli.CLIClient{}
	inputReader := bufio.NewReader(os.Stdin)

	if err := runRestaurantPreview(doorDashClient, inputReader); err != nil {
		log.Fatal(err)
	}
}

// runRestaurantPreview is the interactive no-charge path:
// login → search → pick store → menu → pick item → add → preview.
func runRestaurantPreview(doorDashClient *ddcli.CLIClient, inputReader *bufio.Reader) error {
	ctx := context.Background()
	userGoal := "order from a restaurant and see a quote before paying"

	fmt.Println("=== restaurant → preview (no checkout) ===")
	fmt.Println("You will choose a store and item; we stop after the quote.")
	fmt.Println()

	// --- Step 0: ensure signed in (browser flow if needed) ---
	fmt.Println("=== login ===")
	fmt.Println("Running dd-cli login (opens browser if you are not already signed in)…")
	if err := doorDashClient.Login(ctx); err != nil {
		return err
	}
	fmt.Println("Login step finished.")
	fmt.Println()

	// --- Step 1: search ---
	query := promptLine(inputReader, "Search query (e.g. thai)", "thai")
	limitText := promptLine(inputReader, "Max results", "5")
	limit, err := strconv.Atoi(limitText)
	if err != nil || limit <= 0 {
		limit = 5
	}

	searchOptions := ddcli.SearchRestaurantsOptions{
		Query: query,
		Limit: limit,
	}
	// Without coords, CLI may return success with an empty store list (needs_address).
	coordsText := promptLine(inputReader, "lat,lng (optional, e.g. 37.76,-122.48)", "")
	if latitude, longitude, ok := parseLatLng(coordsText); ok {
		searchOptions.Latitude = &latitude
		searchOptions.Longitude = &longitude
	}

	searchResult, err := doorDashClient.SearchRestaurants(
		ctx,
		ddcli.BuildIntent("find restaurants for an interactive preview demo", userGoal),
		searchOptions,
	)
	if err != nil {
		return err
	}
	if len(searchResult.Stores) == 0 {
		return fmt.Errorf("no restaurants returned — try again with lat,lng near your address")
	}

	fmt.Println()
	fmt.Println("=== stores ===")
	for index, store := range searchResult.Stores {
		fmt.Printf("  %d) %s  (%s)\n", index+1, store.Name, store.StoreID)
	}
	storeIndex, err := promptChoice(inputReader, "Pick store number", len(searchResult.Stores))
	if err != nil {
		return err
	}
	selectedStore := searchResult.Stores[storeIndex]

	// --- Step 2: menu ---
	fmt.Println()
	fmt.Println("Loading menu for", selectedStore.Name, "…")
	menuResult, err := doorDashClient.GetMenu(
		ctx,
		ddcli.BuildIntent("load menu so the user can pick one item", userGoal),
		selectedStore.StoreID,
	)
	if err != nil {
		return err
	}
	if len(menuResult.Items) == 0 {
		return fmt.Errorf("menu returned no items")
	}

	fmt.Println()
	fmt.Println("=== menu ===")
	fmt.Println("menu_id", menuResult.MenuID)
	for index, item := range menuResult.Items {
		line := fmt.Sprintf("  %d) %s  (%s)", index+1, item.Name, item.ItemID)
		if item.Price != 0 {
			line += fmt.Sprintf("  $%.2f", item.Price)
		}
		fmt.Println(line)
	}
	fmt.Println("Tip: prefer a simple item (few/no modifiers) for this demo.")
	itemIndex, err := promptChoice(inputReader, "Pick item number", len(menuResult.Items))
	if err != nil {
		return err
	}
	selectedItem := menuResult.Items[itemIndex]

	quantityText := promptLine(inputReader, "Quantity", "1")
	quantity, err := strconv.Atoi(quantityText)
	if err != nil || quantity <= 0 {
		quantity = 1
	}
	fulfillment := promptLine(inputReader, "Fulfillment (delivery|pickup, empty=CLI default)", "pickup")

	// --- Step 3: add to cart (ListOpenCarts + reuse is inside AddCartItems) ---
	fmt.Println()
	fmt.Println("Adding to cart…")
	addItemsResult, err := doorDashClient.AddCartItems(
		ctx,
		ddcli.BuildIntent("add the chosen menu item for preview", userGoal),
		ddcli.AddCartItemsOptions{
			StoreID:     selectedStore.StoreID,
			MenuID:      menuResult.MenuID,
			Fulfillment: fulfillment,
			Items: []ddcli.CartAddItem{
				{
					ItemID:   selectedItem.ItemID,
					ItemName: selectedItem.Name,
					Quantity: quantity,
				},
			},
		},
	)
	if err != nil {
		return fmt.Errorf("add item failed (item may need modifiers via restaurant-item-details): %w", err)
	}
	fmt.Println("cart_uuid", addItemsResult.CartUUID)
	if addItemsResult.Message != "" {
		fmt.Println(addItemsResult.Message)
	}

	// --- Step 4: preview (stop — no charge) ---
	fmt.Println()
	fmt.Println("Loading quote…")
	previewResult, err := doorDashClient.PreviewOrder(
		ctx,
		ddcli.BuildIntent("preview pricing without placing the order", userGoal),
		addItemsResult.CartUUID,
	)
	if err != nil {
		return err
	}

	printOrderPreview(previewResult)
	fmt.Println()
	fmt.Println("Stopped before checkout. To abandon the cart:")
	fmt.Printf("  go run ./cmd/delete-cart -cart-uuid %s\n", addItemsResult.CartUUID)
	return nil
}

func printOrderPreview(previewResult *ddcli.PreviewOrderResult) {
	fmt.Println("=== order preview (no submit) ===")
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

// promptLine prints a label and reads one line; empty input keeps defaultValue.
func promptLine(inputReader *bufio.Reader, label string, defaultValue string) string {
	if defaultValue != "" {
		fmt.Printf("%s [%s]: ", label, defaultValue)
	} else {
		fmt.Printf("%s: ", label)
	}
	line, err := inputReader.ReadString('\n')
	if err != nil {
		return defaultValue
	}
	line = strings.TrimSpace(line)
	if line == "" {
		return defaultValue
	}
	return line
}

// promptChoice asks for a 1-based index into a list of listLength items.
func promptChoice(inputReader *bufio.Reader, label string, listLength int) (int, error) {
	for {
		rawChoice := promptLine(inputReader, fmt.Sprintf("%s (1-%d)", label, listLength), "1")
		choiceNumber, err := strconv.Atoi(rawChoice)
		if err != nil || choiceNumber < 1 || choiceNumber > listLength {
			fmt.Println("Enter a number in range.")
			continue
		}
		return choiceNumber - 1, nil
	}
}

// parseLatLng accepts "lat,lng" (comma or space). ok is false when input is empty/invalid.
func parseLatLng(coordsText string) (latitude float64, longitude float64, ok bool) {
	coordsText = strings.TrimSpace(coordsText)
	if coordsText == "" {
		return 0, 0, false
	}
	coordsText = strings.ReplaceAll(coordsText, " ", "")
	parts := strings.Split(coordsText, ",")
	if len(parts) != 2 {
		return 0, 0, false
	}
	latitude, latErr := strconv.ParseFloat(parts[0], 64)
	longitude, lngErr := strconv.ParseFloat(parts[1], 64)
	if latErr != nil || lngErr != nil {
		return 0, 0, false
	}
	return latitude, longitude, true
}
