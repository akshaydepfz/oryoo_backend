package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCreateMarketingSpend_ValidationErrors(t *testing.T) {
	tests := []struct {
		name string
		body string
		want string
	}{
		{
			name: "missing platform",
			body: `{"campaign_name":"Summer","amount_spent":5000,"start_date":"2026-08-01","created_by":"uid-1"}`,
			want: "platform cannot be empty",
		},
		{
			name: "amount zero",
			body: `{"platform":"Instagram","campaign_name":"Summer","amount_spent":0,"start_date":"2026-08-01","created_by":"uid-1"}`,
			want: "amount_spent must be greater than 0",
		},
		{
			name: "negative orders",
			body: `{"platform":"Instagram","campaign_name":"Summer","amount_spent":100,"start_date":"2026-08-01","orders_generated":-1,"created_by":"uid-1"}`,
			want: "orders_generated cannot be negative",
		},
		{
			name: "end before start",
			body: `{"platform":"Instagram","campaign_name":"Summer","amount_spent":100,"start_date":"2026-08-10","end_date":"2026-08-01","created_by":"uid-1"}`,
			want: "end_date cannot be before start_date",
		},
		{
			name: "missing created_by",
			body: `{"platform":"Instagram","campaign_name":"Summer","amount_spent":100,"start_date":"2026-08-01"}`,
			want: "created_by is required",
		},
		{
			name: "invalid json",
			body: `{`,
			want: "Invalid JSON",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/marketing-spend/create", strings.NewReader(tc.body))
			rr := httptest.NewRecorder()
			CreateMarketingSpend(rr, req)
			if rr.Code != http.StatusBadRequest {
				t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
			}
			if !strings.Contains(rr.Body.String(), tc.want) {
				t.Fatalf("body=%s want %q", rr.Body.String(), tc.want)
			}
		})
	}
}

func TestListMarketingSpend_RequiresOwner(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/marketing-spend", nil)
	rr := httptest.NewRecorder()
	ListMarketingSpend(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}
}

func TestSummaryMarketingSpend_RequiresOwner(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/marketing-spend/summary", nil)
	rr := httptest.NewRecorder()
	GetMarketingSpendSummary(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}
}

func TestGetMarketingSpendByID_RequiresOwner(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/marketing-spend/abc", nil)
	rr := httptest.NewRecorder()
	GetMarketingSpend(rr, req, "abc")
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}
}

func TestDeleteMarketingSpend_RequiresOwner(t *testing.T) {
	req := httptest.NewRequest(http.MethodDelete, "/marketing-spend/abc", nil)
	rr := httptest.NewRecorder()
	DeleteMarketingSpend(rr, req, "abc")
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}
}

func TestMarketingSpendOwnerFromHeader(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/marketing-spend", nil)
	req.Header.Set("X-Firebase-UID", "uid-header")
	if got := marketingSpendOwner(req, ""); got != "uid-header" {
		t.Fatalf("got %q", got)
	}

	req = httptest.NewRequest(http.MethodGet, "/marketing-spend?created_by=uid-query", nil)
	if got := marketingSpendOwner(req, ""); got != "uid-query" {
		t.Fatalf("got %q", got)
	}

	req = httptest.NewRequest(http.MethodPost, "/marketing-spend/create", nil)
	if got := marketingSpendOwner(req, "uid-body"); got != "uid-body" {
		t.Fatalf("got %q", got)
	}
}

func TestEmptySummaryJSONShape(t *testing.T) {
	// Ensures dashboard empty state serializes as zeros and arrays, not nulls.
	body, _ := json.Marshal(map[string]interface{}{
		"total_spend":        0,
		"this_month_spend":   0,
		"campaign_count":     0,
		"total_orders":       0,
		"total_revenue":      0,
		"platform_breakdown": []interface{}{},
		"monthly_spend":      []interface{}{},
	})
	if !strings.Contains(string(body), `"platform_breakdown":[]`) {
		t.Fatalf("%s", body)
	}
}

func TestMarketingSpendByIDHandler_RejectsReservedPaths(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/marketing-spend/summary", nil)
	rr := httptest.NewRecorder()
	MarketingSpendByIDHandler(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}
}
