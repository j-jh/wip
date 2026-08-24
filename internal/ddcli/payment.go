package ddcli

import (
	"context"
	"fmt"
)

// PaymentCard is one saved credit/debit card from payment-method list.
//
// Fields:
//   - PaymentMethodID (string) — id for future payment commands
//   - Last4 (string) — masked last 4 digits
//   - Brand (string) — Visa / Mastercard / etc. (best-effort)
//   - ExpMonth (string) — expiry month (best-effort; empty when unset)
//   - ExpYear (string) — expiry year (best-effort; empty when unset)
//   - ProviderPaymentMethodID (string) — external provider id (best-effort)
type PaymentCard struct {
	PaymentMethodID           string `json:"payment_method_id"`
	Last4                     string `json:"last4"`
	Brand                     string `json:"brand"`
	ExpMonth                  string `json:"exp_month"`
	ExpYear                   string `json:"exp_year"`
	ProviderPaymentMethodID   string `json:"provider_payment_method_id,omitempty"`
}

// ListPaymentMethodsResult is the structuredContent payload for payment-method list.
//
// Fields:
//   - Success (bool) — CLI reported success
//   - Message (string) — optional status / error text from CLI
//   - Cards ([]PaymentCard) — saved cards only (wallets not included)
//   - DefaultPaymentMethodID (string) — matches one card's payment_method_id when set
type ListPaymentMethodsResult struct {
	Success                 bool          `json:"success"`
	Message                 string        `json:"message,omitempty"`
	Cards                   []PaymentCard `json:"cards"`
	DefaultPaymentMethodID  string        `json:"default_payment_method_id,omitempty"`
}

// ListPaymentMethods lists saved credit/debit cards on the signed-in account.
//
// Params:
//   - ctx (context.Context) — cancel / timeout for the CLI call
//   - intentText (string) — required DoorDash intent blob; use BuildIntent() to build it
//
// Returns:
//   - *ListPaymentMethodsResult — cards plus default id and success flag
//   - error — missing intent, CLI failure, bad JSON, or success:false from CLI
//
// Notes: runs `dd-cli --json-output payment-method list …`. Needs prior `dd-cli login`.
// Response includes cards only — wallets (Apple Pay, etc.) are not in cards[].
// An empty cards[] does not mean the consumer has no payment methods on file.
func (client *CLIClient) ListPaymentMethods(ctx context.Context, intentText string) (*ListPaymentMethodsResult, error) {
	if intentText == "" {
		return nil, fmt.Errorf("ddcli: intent is required")
	}

	cliStdout, err := client.RunCLICommand(ctx, "payment-method", "list", "--intent", intentText)
	if err != nil {
		return nil, err
	}

	var paymentMethodsResult ListPaymentMethodsResult
	if err := decodeStructuredContent(cliStdout, &paymentMethodsResult); err != nil {
		return nil, err
	}

	if !paymentMethodsResult.Success {
		failureMessage := paymentMethodsResult.Message
		if failureMessage == "" {
			failureMessage = "payment-method list failed"
		}
		return &paymentMethodsResult, fmt.Errorf("ddcli: %s", failureMessage)
	}

	return &paymentMethodsResult, nil
}
