package ddcli

import (
	"encoding/json"
	"testing"
)

func TestSubmitOrderDecode(t *testing.T) {
	sampleEnvelope := cliJSONOutputEnvelope{IsError: false}
	sampleEnvelope.StructuredContent, _ = json.Marshal(SubmitOrderResult{
		Success:   true,
		OrderUUID: "ord-abc-123",
		CartUUID:  "cart-xyz",
		Message:   "accepted",
	})
	envelopeJSON, err := json.Marshal(sampleEnvelope)
	if err != nil {
		t.Fatal(err)
	}

	var decoded SubmitOrderResult
	if err := decodeStructuredContent(envelopeJSON, &decoded); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if !decoded.Success || decoded.OrderUUID != "ord-abc-123" {
		t.Fatalf("decoded: %+v", decoded)
	}
}

func TestGetOrderStatusDecode(t *testing.T) {
	sampleEnvelope := cliJSONOutputEnvelope{IsError: false}
	sampleEnvelope.StructuredContent, _ = json.Marshal(GetOrderStatusResult{
		Success:   true,
		OrderUUID: "ord-abc-123",
		Status:    "successful",
	})
	envelopeJSON, err := json.Marshal(sampleEnvelope)
	if err != nil {
		t.Fatal(err)
	}

	var decoded GetOrderStatusResult
	if err := decodeStructuredContent(envelopeJSON, &decoded); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if decoded.Status != "successful" {
		t.Fatalf("status: %q", decoded.Status)
	}
}
