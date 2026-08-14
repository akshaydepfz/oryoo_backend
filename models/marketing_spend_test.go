package models

import (
	"encoding/json"
	"testing"
)

func TestCreateMarketingSpend_OptionalFieldsOmitted(t *testing.T) {
	payload := `{
		"platform":"Instagram",
		"campaign_name":"Summer Offer",
		"amount_spent":5000,
		"start_date":"2026-08-01",
		"created_by":"uid-1"
	}`
	var req CreateMarketingSpendRequest
	if err := json.Unmarshal([]byte(payload), &req); err != nil {
		t.Fatal(err)
	}
	if req.EndDate != nil || req.OrdersGenerated != nil || req.RevenueGenerated != nil || req.Notes != nil {
		t.Fatalf("optional fields should be omitted: %+v", req)
	}
	record, err := ResolveCreateMarketingSpend(&req)
	if err != nil {
		t.Fatal(err)
	}
	if record.EndDate != nil || record.Notes != nil {
		t.Fatalf("end_date and notes should be null: %+v", record)
	}
	if record.OrdersGenerated != 0 || record.RevenueGenerated != 0 {
		t.Fatalf("defaults: orders=%d revenue=%v", record.OrdersGenerated, record.RevenueGenerated)
	}
	if record.CostPerOrder != 0 || record.ROAS != 0 {
		t.Fatalf("metrics without orders/revenue should be 0: %+v", record)
	}
}

func TestCreateMarketingSpend_FullPayloadAndMetrics(t *testing.T) {
	payload := `{
		"platform":"Instagram",
		"campaign_name":"Summer Offer",
		"amount_spent":5000,
		"start_date":"2026-08-01",
		"end_date":"2026-08-10",
		"orders_generated":32,
		"revenue_generated":18500,
		"notes":"Instagram promotion",
		"created_by":"uid-1"
	}`
	var req CreateMarketingSpendRequest
	if err := json.Unmarshal([]byte(payload), &req); err != nil {
		t.Fatal(err)
	}
	record, err := ResolveCreateMarketingSpend(&req)
	if err != nil {
		t.Fatal(err)
	}
	if record.Platform != "Instagram" || record.CampaignName != "Summer Offer" {
		t.Fatalf("%+v", record)
	}
	if record.EndDate == nil || record.EndDate.Format("2006-01-02") != "2026-08-10" {
		t.Fatalf("end_date: %v", record.EndDate)
	}
	if record.CostPerOrder != 5000.0/32.0 {
		t.Fatalf("cost_per_order=%v", record.CostPerOrder)
	}
	if record.ROAS != 18500.0/5000.0 {
		t.Fatalf("roas=%v", record.ROAS)
	}
}

func TestCreateMarketingSpend_Validation(t *testing.T) {
	base := func() CreateMarketingSpendRequest {
		return CreateMarketingSpendRequest{
			Platform:     "Instagram",
			CampaignName: "Summer",
			AmountSpent:  100,
			StartDate:    mustDate("2026-08-01"),
			CreatedBy:    "uid-1",
		}
	}

	req := base()
	req.Platform = "  "
	if _, err := ResolveCreateMarketingSpend(&req); err == nil {
		t.Fatal("expected empty platform error")
	}

	req = base()
	req.CampaignName = ""
	if _, err := ResolveCreateMarketingSpend(&req); err == nil {
		t.Fatal("expected empty campaign_name error")
	}

	req = base()
	req.AmountSpent = 0
	if _, err := ResolveCreateMarketingSpend(&req); err == nil {
		t.Fatal("expected amount_spent error")
	}

	neg := -1
	req = base()
	req.OrdersGenerated = &neg
	if _, err := ResolveCreateMarketingSpend(&req); err == nil {
		t.Fatal("expected negative orders error")
	}

	negRev := -10.0
	req = base()
	req.RevenueGenerated = &negRev
	if _, err := ResolveCreateMarketingSpend(&req); err == nil {
		t.Fatal("expected negative revenue error")
	}

	end := mustDate("2026-07-01")
	req = base()
	req.EndDate = &end
	if _, err := ResolveCreateMarketingSpend(&req); err == nil {
		t.Fatal("expected end_date before start_date error")
	}

	req = base()
	req.CreatedBy = ""
	if _, err := ResolveCreateMarketingSpend(&req); err == nil {
		t.Fatal("expected created_by error")
	}
}

func TestApplyMarketingSpendUpdate_IndependentAndClearable(t *testing.T) {
	existing := &MarketingSpend{
		ID:               "ms-1",
		Platform:         "Instagram",
		CampaignName:     "Summer",
		AmountSpent:      5000,
		StartDate:        mustDate("2026-08-01"),
		OrdersGenerated:  10,
		RevenueGenerated: 20000,
		CreatedBy:        "uid-1",
		Notes:            strPtr("keep"),
	}
	existing.ApplyMetrics()

	notesNull := UpdateMarketingSpendRequest{Notes: nil}
	if err := ApplyMarketingSpendUpdate(existing, &notesNull, map[string]bool{"notes": true}); err != nil {
		t.Fatal(err)
	}
	if existing.Notes != nil {
		t.Fatalf("notes should be cleared, got %v", *existing.Notes)
	}

	end := mustDate("2026-08-20")
	if err := ApplyMarketingSpendUpdate(existing, &UpdateMarketingSpendRequest{EndDate: &end}, map[string]bool{"end_date": true}); err != nil {
		t.Fatal(err)
	}
	if existing.EndDate == nil || existing.EndDate.Format("2006-01-02") != "2026-08-20" {
		t.Fatalf("end_date: %v", existing.EndDate)
	}
	if err := ApplyMarketingSpendUpdate(existing, &UpdateMarketingSpendRequest{EndDate: nil}, map[string]bool{"end_date": true}); err != nil {
		t.Fatal(err)
	}
	if existing.EndDate != nil {
		t.Fatal("end_date should be clearable")
	}
	if existing.Platform != "Instagram" || existing.AmountSpent != 5000 {
		t.Fatal("independent update must not change other fields")
	}
}

func TestEmptyMarketingSpendSummary(t *testing.T) {
	s := EmptyMarketingSpendSummary()
	s.ApplyMetrics()
	if s.TotalSpend != 0 || s.CampaignCount != 0 || s.CostPerOrder != 0 || s.ROAS != 0 {
		t.Fatalf("%+v", s)
	}
	if s.PlatformBreakdown == nil || s.MonthlySpend == nil {
		t.Fatal("empty slices must not be null")
	}
	b, err := json.Marshal(s)
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]interface{}
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatal(err)
	}
	if _, ok := m["platform_breakdown"].([]interface{}); !ok {
		t.Fatalf("platform_breakdown should be an array: %v", m["platform_breakdown"])
	}
}

func TestMarketingSpendDate_JSON(t *testing.T) {
	var d MarketingSpendDate
	if err := json.Unmarshal([]byte(`"2026-08-01"`), &d); err != nil {
		t.Fatal(err)
	}
	if d.Format("2006-01-02") != "2026-08-01" {
		t.Fatalf("%v", d)
	}
	b, err := json.Marshal(d)
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != `"2026-08-01"` {
		t.Fatalf("marshal=%s", b)
	}
}

func mustDate(s string) MarketingSpendDate {
	var d MarketingSpendDate
	if err := json.Unmarshal([]byte(`"`+s+`"`), &d); err != nil {
		panic(err)
	}
	return d
}

func strPtr(s string) *string { return &s }
