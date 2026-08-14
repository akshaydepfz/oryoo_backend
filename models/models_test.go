package models

import (
	"encoding/json"
	"testing"
	"time"
)

func TestUpdateUserPartialRequest_OMITFields(t *testing.T) {
	var req UpdateUserPartialRequest
	if err := json.Unmarshal([]byte(`{}`), &req); err != nil {
		t.Fatal(err)
	}
	if req.Name != nil || req.FCMToken != nil || req.LastOpen != nil {
		t.Fatalf("expected all nil, got name=%v fcm=%v last_open=%v", req.Name, req.FCMToken, req.LastOpen)
	}
}

func TestUpdateUserPartialRequest_PartialJSON(t *testing.T) {
	var req UpdateUserPartialRequest
	if err := json.Unmarshal([]byte(`{"name":"Alice","fcm_token":"tok"}`), &req); err != nil {
		t.Fatal(err)
	}
	if req.Name == nil || *req.Name != "Alice" {
		t.Fatalf("name: %v", req.Name)
	}
	if req.Email != nil {
		t.Fatalf("email should be omitted, got %v", *req.Email)
	}
	if req.FCMToken == nil || *req.FCMToken != "tok" {
		t.Fatalf("fcm_token: %v", req.FCMToken)
	}
}

func TestUpdateUserPartialRequest_EmptyFCMTokenString(t *testing.T) {
	var req UpdateUserPartialRequest
	if err := json.Unmarshal([]byte(`{"fcm_token":""}`), &req); err != nil {
		t.Fatal(err)
	}
	if req.FCMToken == nil || *req.FCMToken != "" {
		t.Fatalf("expected pointer to empty string for explicit empty, got %v", req.FCMToken)
	}
}

func TestUser_JSONNullOptionalFields(t *testing.T) {
	u := User{ID: 1, FirebaseUID: "x", Name: "n"}
	b, err := json.Marshal(u)
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]interface{}
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatal(err)
	}
	if m["fcm_token"] != nil {
		t.Fatalf("expected null/absent behavior, got %v", m["fcm_token"])
	}
}

func TestUpdateUserActivityRequest_OptionalFCM(t *testing.T) {
	var req UpdateUserActivityRequest
	if err := json.Unmarshal([]byte(`{"user_id":"uid1"}`), &req); err != nil {
		t.Fatal(err)
	}
	if req.UserID != "uid1" || req.FCMToken != nil {
		t.Fatalf("%+v", req)
	}
	if err := json.Unmarshal([]byte(`{"user_id":"uid1","fcm_token":"t"}`), &req); err != nil {
		t.Fatal(err)
	}
	if req.FCMToken == nil || *req.FCMToken != "t" {
		t.Fatalf("%+v", req)
	}
}

func TestUpdateUserPartialRequest_LastOpenRFC3339(t *testing.T) {
	var req UpdateUserPartialRequest
	payload := `{"last_open":"2026-03-31T12:00:00Z"}`
	if err := json.Unmarshal([]byte(payload), &req); err != nil {
		t.Fatal(err)
	}
	if req.LastOpen == nil {
		t.Fatal("expected last_open")
	}
	want := time.Date(2026, 3, 31, 12, 0, 0, 0, time.UTC)
	if !req.LastOpen.Equal(want) {
		t.Fatalf("got %v want %v", req.LastOpen, want)
	}
}

func TestCreateOrderRequest_OmitsNewBillFields(t *testing.T) {
	payload := `{
		"client_id":"c1",
		"client_name":"Alice",
		"total_amount":118,
		"status":"Pending",
		"payment_status":"Pending",
		"delivery_address":"123 St",
		"items":[{"name":"Widget","price":100,"quantity":1}]
	}`
	var req CreateOrderRequest
	if err := json.Unmarshal([]byte(payload), &req); err != nil {
		t.Fatal(err)
	}
	if req.GstNumber != nil || req.Discount != nil || req.DeliveryFee != nil {
		t.Fatalf("expected omitted bill fields, got gst=%v discount=%v delivery=%v", req.GstNumber, req.Discount, req.DeliveryFee)
	}
	if err := ResolveCreateBillFields(&req); err != nil {
		t.Fatal(err)
	}
	if req.GstNumber != nil {
		t.Fatalf("gst_number should stay null, got %v", *req.GstNumber)
	}
	if req.Discount == nil || *req.Discount != 0 {
		t.Fatalf("discount default: %v", req.Discount)
	}
	if req.DeliveryFee == nil || *req.DeliveryFee != 0 {
		t.Fatalf("delivery_fee default: %v", req.DeliveryFee)
	}
	if req.TotalAmount != 118 {
		t.Fatalf("existing total_amount must be unchanged, got %v", req.TotalAmount)
	}
}

func TestCreateOrderRequest_OptionalBillFields(t *testing.T) {
	payload := `{
		"client_id":"c1",
		"client_name":"Alice",
		"total_amount":118,
		"gst_number":"22AAAAA0000A1Z5",
		"discount":10,
		"delivery_fee":20,
		"items":[{"name":"Widget","price":100,"quantity":1}]
	}`
	var req CreateOrderRequest
	if err := json.Unmarshal([]byte(payload), &req); err != nil {
		t.Fatal(err)
	}
	if err := ResolveCreateBillFields(&req); err != nil {
		t.Fatal(err)
	}
	if req.GstNumber == nil || *req.GstNumber != "22AAAAA0000A1Z5" {
		t.Fatalf("gst_number: %v", req.GstNumber)
	}
	if req.TotalAmount != 128 {
		t.Fatalf("final amount: got %v want 128", req.TotalAmount)
	}
}

func TestResolveCreateBillFields_RejectsNegativeAmounts(t *testing.T) {
	neg := -1.0
	req := CreateOrderRequest{TotalAmount: 100, Discount: &neg}
	if err := ResolveCreateBillFields(&req); err == nil {
		t.Fatal("expected discount validation error")
	}
	req = CreateOrderRequest{TotalAmount: 100, DeliveryFee: &neg}
	if err := ResolveCreateBillFields(&req); err == nil {
		t.Fatal("expected delivery_fee validation error")
	}
}

func TestCalculateFinalBillAmount_PreservesExistingTax(t *testing.T) {
	items := []OrderItemModel{{Price: 100, Quantity: 1}}
	got := CalculateFinalBillAmount(118, 10, 20, items)
	if got != 128 {
		t.Fatalf("got %v want 128 (100 - 10 + 20 + 18)", got)
	}
}

func TestCalculateFinalBillAmount_NoItemsUsesTotalAsBase(t *testing.T) {
	got := CalculateFinalBillAmount(500, 50, 30, nil)
	if got != 480 {
		t.Fatalf("got %v want 480", got)
	}
}

func TestOrderModel_GetResponseIncludesBillFields(t *testing.T) {
	gst := "GSTIN"
	o := OrderModel{ID: "1", TotalAmount: 90, GstNumber: &gst, Discount: 10, DeliveryFee: 5}
	b, err := json.Marshal(o)
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]interface{}
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatal(err)
	}
	if m["gst_number"] != "GSTIN" {
		t.Fatalf("gst_number: %v", m["gst_number"])
	}
	if m["discount"] != 10.0 {
		t.Fatalf("discount: %v", m["discount"])
	}
	if m["delivery_fee"] != 5.0 {
		t.Fatalf("delivery_fee: %v", m["delivery_fee"])
	}

	o.GstNumber = nil
	b, err = json.Marshal(o)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatal(err)
	}
	if m["gst_number"] != nil {
		t.Fatalf("expected null gst_number, got %v", m["gst_number"])
	}
}

func TestResolveUpdateBillFields_IndependentAndClearable(t *testing.T) {
	existing := &OrderModel{ID: "ord-1", TotalAmount: 118, Discount: 0, DeliveryFee: 0, Status: "Pending"}
	gst := "22AAAAA0000A1Z5"
	bill, err := ResolveUpdateBillFields(existing, &gst, true, nil, false, nil, false)
	if err != nil {
		t.Fatal(err)
	}
	if !bill.GstNumberSet || bill.GstNumber == nil || *bill.GstNumber != gst {
		t.Fatalf("gst update: %+v", bill)
	}
	if bill.DiscountSet || bill.DeliveryFeeSet || bill.TotalAmountSet {
		t.Fatalf("gst-only update must not touch amounts: %+v", bill)
	}

	bill, err = ResolveUpdateBillFields(existing, nil, true, nil, false, nil, false)
	if err != nil {
		t.Fatal(err)
	}
	if !bill.GstNumberSet || bill.GstNumber != nil {
		t.Fatalf("gst should be clearable to null: %+v", bill)
	}

	d := 10.0
	bill, err = ResolveUpdateBillFields(existing, nil, false, &d, true, nil, false)
	if err != nil {
		t.Fatal(err)
	}
	if !bill.DiscountSet || bill.Discount != 10 || !bill.TotalAmountSet || bill.TotalAmount != 108 {
		t.Fatalf("discount update: %+v", bill)
	}
	if bill.DeliveryFeeSet || bill.GstNumberSet {
		t.Fatalf("discount-only update leaked other fields: %+v", bill)
	}
}

func TestResolveUpdateBillFields_OmittedKeepsExisting(t *testing.T) {
	existing := &OrderModel{ID: "ord-1", TotalAmount: 90, Discount: 10, DeliveryFee: 5}
	bill, err := ResolveUpdateBillFields(existing, nil, false, nil, false, nil, false)
	if err != nil {
		t.Fatal(err)
	}
	if bill != nil {
		t.Fatalf("status-only update must not produce bill fields, got %+v", bill)
	}
}

func TestResolveUpdateBillFields_RejectsNegative(t *testing.T) {
	existing := &OrderModel{ID: "ord-1", TotalAmount: 100}
	neg := -5.0
	if _, err := ResolveUpdateBillFields(existing, nil, false, &neg, true, nil, false); err == nil {
		t.Fatal("expected negative discount error")
	}
	if _, err := ResolveUpdateBillFields(existing, nil, false, nil, false, &neg, true); err == nil {
		t.Fatal("expected negative delivery_fee error")
	}
}

func TestUpdateOrderRequest_ExistingPayloadStillDecodes(t *testing.T) {
	payload := `{"id":"abc","status":"Completed","payment_status":"Paid","client_name":"ignored"}`
	var req UpdateOrderRequest
	if err := json.Unmarshal([]byte(payload), &req); err != nil {
		t.Fatal(err)
	}
	if req.ID != "abc" || req.Status != "Completed" || req.PaymentStatus != "Paid" {
		t.Fatalf("%+v", req)
	}
	if req.GstNumber != nil || req.Discount != nil || req.DeliveryFee != nil {
		t.Fatalf("new fields must stay omitted: %+v", req)
	}
}

func TestJSONFieldPresent_NullVsOmitted(t *testing.T) {
	var omitted map[string]json.RawMessage
	if err := json.Unmarshal([]byte(`{"id":"1"}`), &omitted); err != nil {
		t.Fatal(err)
	}
	if JSONFieldPresent(omitted, "gst_number") {
		t.Fatal("omitted gst_number should not be present")
	}

	var withNull map[string]json.RawMessage
	if err := json.Unmarshal([]byte(`{"id":"1","gst_number":null}`), &withNull); err != nil {
		t.Fatal(err)
	}
	if !JSONFieldPresent(withNull, "gst_number") {
		t.Fatal("explicit null gst_number should be present")
	}
}
