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
