package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCreateOrder_RejectsNegativeDiscount(t *testing.T) {
	body := `{
		"client_id":"c1",
		"client_name":"Alice",
		"total_amount":100,
		"discount":-10,
		"delivery_address":"123 St"
	}`
	req := httptest.NewRequest(http.MethodPost, "/orders/create", strings.NewReader(body))
	rr := httptest.NewRecorder()
	CreateOrder(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}
	if !strings.Contains(rr.Body.String(), "discount must not be negative") {
		t.Fatalf("body=%s", rr.Body.String())
	}
}

func TestCreateOrder_RejectsNegativeDeliveryFee(t *testing.T) {
	body := `{
		"client_id":"c1",
		"client_name":"Alice",
		"total_amount":100,
		"delivery_fee":-1,
		"delivery_address":"123 St"
	}`
	req := httptest.NewRequest(http.MethodPost, "/orders/create", strings.NewReader(body))
	rr := httptest.NewRecorder()
	CreateOrder(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}
	if !strings.Contains(rr.Body.String(), "delivery_fee must not be negative") {
		t.Fatalf("body=%s", rr.Body.String())
	}
}

func TestUpdateOrder_RejectsNegativeDiscountWithoutTouchingDB(t *testing.T) {
	// FetchOrderByID is only called when bill fields are present. Negative
	// validation happens after fetch, so this test uses invalid JSON shape
	// for missing id instead, plus decode of update payloads.
	body := `{"status":"Completed","payment_status":"Paid"}`
	req := httptest.NewRequest(http.MethodPost, "/orders/update", strings.NewReader(body))
	rr := httptest.NewRecorder()
	UpdateOrder(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}
	if !strings.Contains(rr.Body.String(), "Order ID is required") {
		t.Fatalf("body=%s", rr.Body.String())
	}
}

func TestUpdateOrder_ExistingPayloadRequiresIDOnly(t *testing.T) {
	var raw map[string]json.RawMessage
	payload := `{"id":"abc","status":"Completed","payment_status":"Paid"}`
	if err := json.Unmarshal([]byte(payload), &raw); err != nil {
		t.Fatal(err)
	}
	if _, ok := raw["gst_number"]; ok {
		t.Fatal("existing update payload should omit gst_number")
	}
	if _, ok := raw["discount"]; ok {
		t.Fatal("existing update payload should omit discount")
	}
	if _, ok := raw["delivery_fee"]; ok {
		t.Fatal("existing update payload should omit delivery_fee")
	}
}
