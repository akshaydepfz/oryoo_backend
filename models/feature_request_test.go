package models

import (
	"encoding/json"
	"testing"
)

func TestCreateFeatureRequest_RequiredFieldsAndPendingStatus(t *testing.T) {
	payload := `{
		"title":"Add WhatsApp integration",
		"description":"I want to send order updates directly to customers through WhatsApp.",
		"created_by":"uid-1",
		"status":"completed"
	}`
	var req CreateFeatureRequestRequest
	if err := json.Unmarshal([]byte(payload), &req); err != nil {
		t.Fatal(err)
	}
	if req.Category != nil {
		t.Fatal("category should be omitted")
	}
	record, err := ResolveCreateFeatureRequest(&req)
	if err != nil {
		t.Fatal(err)
	}
	if record.Status != FeatureRequestStatusPending {
		t.Fatalf("status=%q, client must not set status on create", record.Status)
	}
	if record.Category != nil {
		t.Fatalf("omitted category should be null, got %v", *record.Category)
	}
	if record.Title != "Add WhatsApp integration" {
		t.Fatalf("title=%q", record.Title)
	}
}

func TestCreateFeatureRequest_WithCategory(t *testing.T) {
	payload := `{
		"title":"Add WhatsApp integration",
		"description":"Send order updates on WhatsApp.",
		"category":"CRM",
		"created_by":"uid-1"
	}`
	var req CreateFeatureRequestRequest
	if err := json.Unmarshal([]byte(payload), &req); err != nil {
		t.Fatal(err)
	}
	record, err := ResolveCreateFeatureRequest(&req)
	if err != nil {
		t.Fatal(err)
	}
	if record.Category == nil || *record.Category != "CRM" {
		t.Fatalf("category=%v", record.Category)
	}
}

func TestCreateFeatureRequest_Validation(t *testing.T) {
	_, err := ResolveCreateFeatureRequest(&CreateFeatureRequestRequest{
		Title: "  ", Description: "desc", CreatedBy: "uid-1",
	})
	if err == nil || err.Error() != "title cannot be empty" {
		t.Fatalf("title: %v", err)
	}

	_, err = ResolveCreateFeatureRequest(&CreateFeatureRequestRequest{
		Title: "title", Description: "", CreatedBy: "uid-1",
	})
	if err == nil || err.Error() != "description cannot be empty" {
		t.Fatalf("description: %v", err)
	}

	_, err = ResolveCreateFeatureRequest(&CreateFeatureRequestRequest{
		Title: "title", Description: "desc", CreatedBy: "  ",
	})
	if err == nil || err.Error() != "created_by is required" {
		t.Fatalf("created_by: %v", err)
	}
}

func TestApplyFeatureRequestUpdate_IgnoresStatusAndCreatedBy(t *testing.T) {
	existing := &FeatureRequest{
		ID:          "fr-1",
		Title:       "Old title",
		Description: "Old description",
		CreatedBy:   "uid-1",
		Status:      FeatureRequestStatusPending,
	}
	status := "completed"
	createdBy := "hacker"
	title := "New title"
	req := UpdateFeatureRequestRequest{
		Title:     &title,
		Status:    &status,
		CreatedBy: &createdBy,
	}
	present := map[string]bool{"title": true, "status": true, "created_by": true}
	if err := ApplyFeatureRequestUpdate(existing, &req, present); err != nil {
		t.Fatal(err)
	}
	if existing.Title != "New title" {
		t.Fatalf("title=%q", existing.Title)
	}
	if existing.Status != FeatureRequestStatusPending {
		t.Fatalf("users must not change status, got %q", existing.Status)
	}
	if existing.CreatedBy != "uid-1" {
		t.Fatalf("users must not change created_by, got %q", existing.CreatedBy)
	}
}

func TestApplyFeatureRequestUpdate_CategoryAndDescription(t *testing.T) {
	existing := &FeatureRequest{
		ID:          "fr-1",
		Title:       "Title",
		Description: "Desc",
		CreatedBy:   "uid-1",
		Status:      FeatureRequestStatusPending,
		Category:    strPtr("CRM"),
	}
	desc := "Updated description"
	if err := ApplyFeatureRequestUpdate(existing, &UpdateFeatureRequestRequest{Description: &desc}, map[string]bool{"description": true}); err != nil {
		t.Fatal(err)
	}
	if existing.Description != "Updated description" || existing.Title != "Title" {
		t.Fatalf("%+v", existing)
	}

	if err := ApplyFeatureRequestUpdate(existing, &UpdateFeatureRequestRequest{Category: nil}, map[string]bool{"category": true}); err != nil {
		t.Fatal(err)
	}
	if existing.Category != nil {
		t.Fatal("category should be clearable")
	}

	cat := "Payments"
	if err := ApplyFeatureRequestUpdate(existing, &UpdateFeatureRequestRequest{Category: &cat}, map[string]bool{"category": true}); err != nil {
		t.Fatal(err)
	}
	if existing.Category == nil || *existing.Category != "Payments" {
		t.Fatalf("category=%v", existing.Category)
	}
}

func TestFeatureRequestListJSON_EmptySliceNotNull(t *testing.T) {
	items := make([]FeatureRequest, 0)
	b, err := json.Marshal(map[string]interface{}{
		"success":          true,
		"count":            len(items),
		"feature_requests": items,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !json.Valid(b) {
		t.Fatalf("invalid json: %s", b)
	}
	var m map[string]interface{}
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatal(err)
	}
	arr, ok := m["feature_requests"].([]interface{})
	if !ok || arr == nil {
		t.Fatalf("feature_requests should be an empty array, got %v", m["feature_requests"])
	}
	if len(arr) != 0 {
		t.Fatalf("expected empty array, got %v", arr)
	}
}
