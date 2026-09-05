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
	Success  bool        `json:"success"`
	Message  string      `json:"message,omitempty"`
	PageFull bool        `json:"page_full"`
	Orders   []PastOrder `json:"orders"`
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

// ReorderPastOrderResult is the structuredContent payload for order reorder.
//
// Fields:
//   - Success (bool) — CLI reported success
//   - CartUUID (string) — new cart created from the past order
//   - FailReason (string) — why reorder failed when success is false
//   - Message (string) — optional status text from CLI
type ReorderPastOrderResult struct {
	Success    bool   `json:"success"`
	CartUUID   string `json:"cart_uuid"`
	FailReason string `json:"fail_reason"`
	Message    string `json:"message,omitempty"`
}

// ReorderPastOrder creates a new cart from a past order.
//
// Params:
//   - ctx (context.Context) — cancel / timeout for the CLI call
//   - intentText (string) — required DoorDash intent blob; use BuildIntent() to build it
//   - orderUUID (string) — past order id from ListOrderHistory
//
// Returns:
//   - *ReorderPastOrderResult — new cart_uuid on success
//   - error — missing args, CLI failure, bad JSON, or success:false from CLI
//
// Notes: runs `dd-cli --json-output order reorder …`. Mutation — creates a cart.
// Pre-flight with ListOpenCarts(storeID): one open cart per store. Items may drop silently.
func (client *CLIClient) ReorderPastOrder(ctx context.Context, intentText string, orderUUID string) (*ReorderPastOrderResult, error) {
	if intentText == "" {
		return nil, fmt.Errorf("ddcli: intent is required")
	}
	if orderUUID == "" {
		return nil, fmt.Errorf("ddcli: orderUUID is required")
	}

	cliStdout, err := client.RunCLICommand(
		ctx,
		"order", "reorder",
		"--order-uuid", orderUUID,
		"--intent", intentText,
	)
	if err != nil {
		return nil, err
	}

	var reorderResult ReorderPastOrderResult
	if err := decodeStructuredContent(cliStdout, &reorderResult); err != nil {
		return nil, err
	}

	if !reorderResult.Success {
		failureMessage := reorderResult.Message
		if failureMessage == "" {
			failureMessage = reorderResult.FailReason
		}
		if failureMessage == "" {
			failureMessage = "order reorder failed"
		}
		return &reorderResult, fmt.Errorf("ddcli: %s", failureMessage)
	}

	return &reorderResult, nil
}

// MoneyAmount is a displayable money value from order preview.
//
// Fields:
//   - UnitAmount (int) — minor units (e.g. cents); unsigned in some fields — do not invent totals from these alone
//   - Currency (string) — e.g. USD
//   - DisplayString (string) — formatted amount for humans
type MoneyAmount struct {
	UnitAmount    int    `json:"unit_amount"`
	Currency      string `json:"currency"`
	DisplayString string `json:"display_string"`
}

// PreviewLineItem is one quote breakdown row (subtotal, fees, tax, …).
//
// Fields:
//   - Label (string) — row label
//   - FinalMoney (MoneyAmount) — amount shown for the row
type PreviewLineItem struct {
	Label      string      `json:"label"`
	FinalMoney MoneyAmount `json:"final_money"`
}

// PreviewQuoteItem is one cart item as seen on the preview quote.
//
// Fields:
//   - Name (string) — item display name
//   - Quantity (int) — count
//   - UnitPriceDisplay (string) — formatted unit price when present
type PreviewQuoteItem struct {
	Name             string
	Quantity         int
	UnitPriceDisplay string
}

// PreviewOrderQuote is the subset of quote fields we surface for phase 1.
//
// Fields:
//   - NetTotalBeforeTip (MoneyAmount) — authoritative total before tip
//   - IsDashPassApplied (bool) — whether DashPass applied
//   - FulfillmentType (string) — DELIVERY or PICKUP
//   - DeliveryAddress (string) — printable delivery address when present
//   - LineItems ([]PreviewLineItem) — fee/tax/subtotal breakdown
//   - Items ([]PreviewQuoteItem) — items that landed in the cart quote
type PreviewOrderQuote struct {
	NetTotalBeforeTip MoneyAmount
	IsDashPassApplied bool
	FulfillmentType   string
	DeliveryAddress   string
	LineItems         []PreviewLineItem
	Items             []PreviewQuoteItem
}

// PreviewOrderResult is the decoded order preview for callers.
//
// Fields:
//   - Success (bool) — CLI reported success
//   - CartUUID (string) — cart that was previewed
//   - Message (string) — optional status / error text
//   - Quote (PreviewOrderQuote) — simplified quote summary
type PreviewOrderResult struct {
	Success  bool
	CartUUID string
	Message  string
	Quote    PreviewOrderQuote
}

// The types below mirror nested CLI JSON for order preview.
// Callers never see them — PreviewOrder copies into the flatter types above.

type previewOrderRawJSON struct {
	Success  bool              `json:"success"`
	CartUUID string            `json:"cart_uuid"`
	Message  string            `json:"message"`
	Quote    *previewQuoteJSON `json:"quote"`
}

type previewQuoteJSON struct {
	LineItems         []PreviewLineItem           `json:"line_items"`
	IsDashPassApplied bool                        `json:"is_dashpass_applied"`
	NetTotalBeforeTip MoneyAmount                 `json:"net_total_before_tip"`
	DeliveryAddress   *previewDeliveryAddressJSON `json:"delivery_address"`
	StoreOrderCart    *previewStoreOrderCartJSON  `json:"store_order_cart"`
}

type previewDeliveryAddressJSON struct {
	PrintableAddress string `json:"printable_address"`
}

type previewStoreOrderCartJSON struct {
	FulfillmentType string                  `json:"fulfillment_type"`
	Orders          []previewStoreOrderJSON `json:"orders"`
}

type previewStoreOrderJSON struct {
	OrderItems []previewOrderItemJSON `json:"order_items"`
}

type previewOrderItemJSON struct {
	Quantity                int                     `json:"quantity"`
	UnitPriceMonetaryFields *MoneyAmount            `json:"unit_price_monetary_fields"`
	Item                    *previewCatalogItemJSON `json:"item"`
}

type previewCatalogItemJSON struct {
	Name string `json:"name"`
}

// PreviewOrder loads pricing for a cart without charging.
//
// Params:
//   - ctx (context.Context) — cancel / timeout for the CLI call
//   - intentText (string) — required DoorDash intent blob; use BuildIntent() to build it
//   - cartUUID (string) — cart from reorder or cart add-items
//
// Returns:
//   - *PreviewOrderResult — simplified quote (total, lines, items, fulfillment)
//   - error — missing args, CLI failure, bad JSON, or success:false from CLI
//
// Notes: runs `dd-cli --json-output order preview --cart-uuid …` with no fulfillment
// mutation flags. Does not submit or touch payment methods.
func (client *CLIClient) PreviewOrder(ctx context.Context, intentText string, cartUUID string) (*PreviewOrderResult, error) {
	if intentText == "" {
		return nil, fmt.Errorf("ddcli: intent is required")
	}
	if cartUUID == "" {
		return nil, fmt.Errorf("ddcli: cartUUID is required")
	}

	cliStdout, err := client.RunCLICommand(
		ctx,
		"order", "preview",
		"--cart-uuid", cartUUID,
		"--intent", intentText,
	)
	if err != nil {
		return nil, err
	}

	var rawPreview previewOrderRawJSON
	if err := decodeStructuredContent(cliStdout, &rawPreview); err != nil {
		return nil, err
	}

	if !rawPreview.Success {
		failureMessage := rawPreview.Message
		if failureMessage == "" {
			failureMessage = "order preview failed"
		}
		return &PreviewOrderResult{
			Success:  false,
			CartUUID: rawPreview.CartUUID,
			Message:  rawPreview.Message,
		}, fmt.Errorf("ddcli: %s", failureMessage)
	}

	previewResult := &PreviewOrderResult{
		Success:  true,
		CartUUID: rawPreview.CartUUID,
		Message:  rawPreview.Message,
	}
	if rawPreview.Quote != nil {
		previewResult.Quote = simplifyPreviewQuote(rawPreview.Quote)
	}
	return previewResult, nil
}

// simplifyPreviewQuote copies nested CLI quote JSON into the flatter PreviewOrderQuote.
func simplifyPreviewQuote(rawQuote *previewQuoteJSON) PreviewOrderQuote {
	quote := PreviewOrderQuote{
		LineItems:         rawQuote.LineItems,
		IsDashPassApplied: rawQuote.IsDashPassApplied,
		NetTotalBeforeTip: rawQuote.NetTotalBeforeTip,
	}

	if rawQuote.DeliveryAddress != nil {
		quote.DeliveryAddress = rawQuote.DeliveryAddress.PrintableAddress
	}

	if rawQuote.StoreOrderCart == nil {
		return quote
	}

	quote.FulfillmentType = rawQuote.StoreOrderCart.FulfillmentType
	for _, storeOrder := range rawQuote.StoreOrderCart.Orders {
		for _, orderItem := range storeOrder.OrderItems {
			quoteItem := PreviewQuoteItem{Quantity: orderItem.Quantity}
			if orderItem.Item != nil {
				quoteItem.Name = orderItem.Item.Name
			}
			if orderItem.UnitPriceMonetaryFields != nil {
				quoteItem.UnitPriceDisplay = orderItem.UnitPriceMonetaryFields.DisplayString
			}
			quote.Items = append(quote.Items, quoteItem)
		}
	}
	return quote
}

// SubmitOrderOptions holds inputs for placing an order (charges default payment).
//
// Fields:
//   - CartUUID (string) — required; cart from add-items / reorder
//   - TipCents (int) — Dasher tip in CENTS (not dollars). 0 = no tip (required for pickup)
//   - Fulfillment (string) — optional "delivery" or "pickup"; empty keeps cart mode
//   - ScheduledTime (string) — optional ISO-8601 UTC; empty = ASAP
//   - Priority (bool) — express delivery upgrade; delivery only
//   - NoApplyCredits (bool) — opt out of DoorDash credits (must match preview)
type SubmitOrderOptions struct {
	CartUUID       string
	TipCents       int
	Fulfillment    string
	ScheduledTime  string
	Priority       bool
	NoApplyCredits bool
}

// SubmitOrderResult is the structuredContent payload for order submit.
//
// Fields:
//   - Success (bool) — CLI reported success (accepted into processing)
//   - Message (string) — optional status / error text
//   - OrderUUID (string) — id for order status / history
//   - CartUUID (string) — cart that was submitted (may be empty)
type SubmitOrderResult struct {
	Success   bool   `json:"success"`
	Message   string `json:"message,omitempty"`
	OrderUUID string `json:"order_uuid"`
	CartUUID  string `json:"cart_uuid,omitempty"`
}

// SubmitOrder places an order and charges the consumer's default payment method.
//
// Params:
//   - ctx (context.Context) — cancel / timeout for the CLI call
//   - intentText (string) — required DoorDash intent blob; use BuildIntent() to build it
//   - options (SubmitOrderOptions) — cart, tip, optional fulfillment / schedule flags
//
// Returns:
//   - *SubmitOrderResult — order_uuid plus success flag
//   - error — missing args, CLI failure, bad JSON, or success:false from CLI
//
// Notes: runs `dd-cli --json-output order submit … --yes`. Always passes --yes because
// this Go process is non-interactive to the CLI — callers must gate confirmation themselves
// before calling this method. Also requires ALLOW_SUBMIT_ORDER=true (see .env.example).
// DESTRUCTIVE: charges money. Treat the cart as spent after success; re-submitting the same
// cart_uuid can duplicate the order. Poll GetOrderStatus until not pending.
func (client *CLIClient) SubmitOrder(ctx context.Context, intentText string, options SubmitOrderOptions) (*SubmitOrderResult, error) {
	if !AllowSubmit() {
		return nil, fmt.Errorf("ddcli: submit blocked — set %s=true in .env (see .env.example)", EnvAllowSubmitOrder)
	}
	if intentText == "" {
		return nil, fmt.Errorf("ddcli: intent is required")
	}
	if options.CartUUID == "" {
		return nil, fmt.Errorf("ddcli: cartUUID is required")
	}
	if options.TipCents < 0 {
		return nil, fmt.Errorf("ddcli: tipCents must be >= 0 (cents, not dollars)")
	}

	cliArgs := []string{
		"order", "submit",
		"--cart-uuid", options.CartUUID,
		"--tip-cents", strconv.Itoa(options.TipCents),
		"--yes",
		"--intent", intentText,
	}
	if options.Fulfillment != "" {
		cliArgs = append(cliArgs, "--fulfillment", options.Fulfillment)
	}
	if options.ScheduledTime != "" {
		cliArgs = append(cliArgs, "--scheduled-time", options.ScheduledTime)
	}
	if options.Priority {
		cliArgs = append(cliArgs, "--priority")
	}
	if options.NoApplyCredits {
		cliArgs = append(cliArgs, "--no-apply-credits")
	}

	cliStdout, err := client.RunCLICommand(ctx, cliArgs...)
	if err != nil {
		return nil, err
	}

	var submitResult SubmitOrderResult
	if err := decodeStructuredContent(cliStdout, &submitResult); err != nil {
		return nil, err
	}

	if !submitResult.Success {
		failureMessage := submitResult.Message
		if failureMessage == "" {
			failureMessage = "order submit failed"
		}
		return &submitResult, fmt.Errorf("ddcli: %s", failureMessage)
	}

	return &submitResult, nil
}

// GetOrderStatusResult is the structuredContent payload for order status.
//
// Fields:
//   - Success (bool) — CLI reported success
//   - Message (string) — optional status / error text
//   - OrderUUID (string) — order being checked
//   - Status (string) — successful | pending | action_required | failed | not_found
//   - ErrorMessage (string) — present on failed / some degraded outcomes
type GetOrderStatusResult struct {
	Success      bool   `json:"success"`
	Message      string `json:"message,omitempty"`
	OrderUUID    string `json:"order_uuid,omitempty"`
	Status       string `json:"status"`
	ErrorMessage string `json:"error_message,omitempty"`
}

// GetOrderStatus checks whether a submitted order cleared processing.
//
// Params:
//   - ctx (context.Context) — cancel / timeout for the CLI call
//   - intentText (string) — required DoorDash intent blob; use BuildIntent() to build it
//   - orderUUID (string) — from SubmitOrder (or order history)
//
// Returns:
//   - *GetOrderStatusResult — status string plus success flag
//   - error — missing args, CLI failure, bad JSON, or success:false from CLI
//
// Notes: runs `dd-cli --json-output order status --order-uuid …`. Read-only.
// One call is one check — callers own any retry loop while status is pending.
// Report "order placed" only when Status == "successful".
func (client *CLIClient) GetOrderStatus(ctx context.Context, intentText string, orderUUID string) (*GetOrderStatusResult, error) {
	if intentText == "" {
		return nil, fmt.Errorf("ddcli: intent is required")
	}
	if orderUUID == "" {
		return nil, fmt.Errorf("ddcli: orderUUID is required")
	}

	cliStdout, err := client.RunCLICommand(
		ctx,
		"order", "status",
		"--order-uuid", orderUUID,
		"--intent", intentText,
	)
	if err != nil {
		return nil, err
	}

	var statusResult GetOrderStatusResult
	if err := decodeStructuredContent(cliStdout, &statusResult); err != nil {
		return nil, err
	}

	if !statusResult.Success {
		failureMessage := statusResult.Message
		if failureMessage == "" {
			failureMessage = "order status failed"
		}
		return &statusResult, fmt.Errorf("ddcli: %s", failureMessage)
	}

	return &statusResult, nil
}
