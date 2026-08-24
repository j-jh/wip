package ddcli

import (
	"encoding/json"
	"testing"
)

func TestListPaymentMethodsDecode(t *testing.T) {
	sampleEnvelope := cliJSONOutputEnvelope{
		IsError: false,
	}
	sampleEnvelope.StructuredContent, _ = json.Marshal(ListPaymentMethodsResult{
		Success:                true,
		DefaultPaymentMethodID: "pm_abc123",
		Cards: []PaymentCard{
			{
				PaymentMethodID: "pm_abc123",
				Last4:           "4242",
				Brand:           "Visa",
				ExpMonth:        12,
				ExpYear:         2028,
			},
			{
				PaymentMethodID:         "pm_def456",
				Last4:                   "1111",
				Brand:                   "Mastercard",
				ExpMonth:                3,
				ExpYear:                 2027,
				ProviderPaymentMethodID: "prov_xyz",
			},
		},
	})

	envelopeJSON, err := json.Marshal(sampleEnvelope)
	if err != nil {
		t.Fatal(err)
	}

	var decoded ListPaymentMethodsResult
	if err := decodeStructuredContent(envelopeJSON, &decoded); err != nil {
		t.Fatalf("decodeStructuredContent: %v", err)
	}

	if !decoded.Success {
		t.Fatal("expected success true")
	}
	if decoded.DefaultPaymentMethodID != "pm_abc123" {
		t.Fatalf("default id: got %q", decoded.DefaultPaymentMethodID)
	}
	if len(decoded.Cards) != 2 {
		t.Fatalf("cards len: got %d", len(decoded.Cards))
	}
	if decoded.Cards[0].Brand != "Visa" || decoded.Cards[0].Last4 != "4242" {
		t.Fatalf("first card: %+v", decoded.Cards[0])
	}
}
