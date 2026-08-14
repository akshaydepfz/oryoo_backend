package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// MarketingSpendDate stores a calendar date and accepts "YYYY-MM-DD" or RFC3339 JSON.
type MarketingSpendDate struct {
	time.Time
}

func (d MarketingSpendDate) MarshalJSON() ([]byte, error) {
	if d.Time.IsZero() {
		return []byte("null"), nil
	}
	return json.Marshal(d.Time.Format("2006-01-02"))
}

func (d *MarketingSpendDate) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		d.Time = time.Time{}
		return nil
	}
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	s = strings.TrimSpace(s)
	if s == "" {
		d.Time = time.Time{}
		return nil
	}
	layouts := []string{
		"2006-01-02",
		time.RFC3339,
		"2006-01-02T15:04:05Z07:00",
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05",
	}
	for _, layout := range layouts {
		if t, err := time.Parse(layout, s); err == nil {
			d.Time = t
			return nil
		}
	}
	return fmt.Errorf("invalid date")
}

func (d MarketingSpendDate) Value() (driver.Value, error) {
	if d.Time.IsZero() {
		return nil, nil
	}
	return d.Time, nil
}

func (d *MarketingSpendDate) Scan(value interface{}) error {
	if value == nil {
		d.Time = time.Time{}
		return nil
	}
	t, ok := value.(time.Time)
	if !ok {
		return fmt.Errorf("cannot scan %T into MarketingSpendDate", value)
	}
	d.Time = t
	return nil
}

func (d MarketingSpendDate) IsZero() bool {
	return d.Time.IsZero()
}

type MarketingSpend struct {
	ID               string              `json:"id"`
	Platform         string              `json:"platform"`
	CampaignName     string              `json:"campaign_name"`
	AmountSpent      float64             `json:"amount_spent"`
	StartDate        MarketingSpendDate  `json:"start_date"`
	EndDate          *MarketingSpendDate `json:"end_date"`
	OrdersGenerated  int                 `json:"orders_generated"`
	RevenueGenerated float64             `json:"revenue_generated"`
	Notes            *string             `json:"notes"`
	CreatedBy        string              `json:"created_by"`
	CreatedAt        *time.Time          `json:"created_at"`
	UpdatedAt        *time.Time          `json:"updated_at"`
	CostPerOrder     float64             `json:"cost_per_order"`
	ROAS             float64             `json:"roas"`
}

func (m *MarketingSpend) ApplyMetrics() {
	if m == nil {
		return
	}
	m.CostPerOrder = 0
	m.ROAS = 0
	if m.OrdersGenerated > 0 && m.AmountSpent > 0 {
		m.CostPerOrder = m.AmountSpent / float64(m.OrdersGenerated)
	}
	if m.RevenueGenerated > 0 && m.AmountSpent > 0 {
		m.ROAS = m.RevenueGenerated / m.AmountSpent
	}
}

type CreateMarketingSpendRequest struct {
	Platform         string              `json:"platform"`
	CampaignName     string              `json:"campaign_name"`
	AmountSpent      float64             `json:"amount_spent"`
	StartDate        MarketingSpendDate  `json:"start_date"`
	EndDate          *MarketingSpendDate `json:"end_date"`
	OrdersGenerated  *int                `json:"orders_generated"`
	RevenueGenerated *float64            `json:"revenue_generated"`
	Notes            *string             `json:"notes"`
	CreatedBy        string              `json:"created_by"`
}

type UpdateMarketingSpendRequest struct {
	Platform         *string             `json:"platform"`
	CampaignName     *string             `json:"campaign_name"`
	AmountSpent      *float64            `json:"amount_spent"`
	StartDate        *MarketingSpendDate `json:"start_date"`
	EndDate          *MarketingSpendDate `json:"end_date"`
	OrdersGenerated  *int                `json:"orders_generated"`
	RevenueGenerated *float64            `json:"revenue_generated"`
	Notes            *string             `json:"notes"`
}

type MarketingSpendPlatformBreakdown struct {
	Platform    string  `json:"platform"`
	AmountSpent float64 `json:"amount_spent"`
}

type MarketingSpendMonthly struct {
	Month       string  `json:"month"`
	AmountSpent float64 `json:"amount_spent"`
}

type MarketingSpendSummary struct {
	TotalSpend        float64                           `json:"total_spend"`
	ThisMonthSpend    float64                           `json:"this_month_spend"`
	CampaignCount     int                               `json:"campaign_count"`
	TotalOrders       int                               `json:"total_orders"`
	TotalRevenue      float64                           `json:"total_revenue"`
	CostPerOrder      float64                           `json:"cost_per_order"`
	ROAS              float64                           `json:"roas"`
	PlatformBreakdown []MarketingSpendPlatformBreakdown `json:"platform_breakdown"`
	MonthlySpend      []MarketingSpendMonthly           `json:"monthly_spend"`
}

func EmptyMarketingSpendSummary() MarketingSpendSummary {
	return MarketingSpendSummary{
		PlatformBreakdown: []MarketingSpendPlatformBreakdown{},
		MonthlySpend:      []MarketingSpendMonthly{},
	}
}

func (s *MarketingSpendSummary) ApplyMetrics() {
	if s == nil {
		return
	}
	s.CostPerOrder = 0
	s.ROAS = 0
	if s.TotalOrders > 0 && s.TotalSpend > 0 {
		s.CostPerOrder = s.TotalSpend / float64(s.TotalOrders)
	}
	if s.TotalRevenue > 0 && s.TotalSpend > 0 {
		s.ROAS = s.TotalRevenue / s.TotalSpend
	}
	if s.PlatformBreakdown == nil {
		s.PlatformBreakdown = []MarketingSpendPlatformBreakdown{}
	}
	if s.MonthlySpend == nil {
		s.MonthlySpend = []MarketingSpendMonthly{}
	}
}

func NormalizeOptionalString(v *string) *string {
	if v == nil {
		return nil
	}
	s := strings.TrimSpace(*v)
	if s == "" {
		return nil
	}
	return &s
}

func ValidateMarketingSpendFields(platform, campaignName string, amountSpent float64, startDate MarketingSpendDate, endDate *MarketingSpendDate, ordersGenerated int, revenueGenerated float64) error {
	if strings.TrimSpace(platform) == "" {
		return fmt.Errorf("platform cannot be empty")
	}
	if strings.TrimSpace(campaignName) == "" {
		return fmt.Errorf("campaign_name cannot be empty")
	}
	if amountSpent <= 0 {
		return fmt.Errorf("amount_spent must be greater than 0")
	}
	if startDate.IsZero() {
		return fmt.Errorf("start_date is required")
	}
	if ordersGenerated < 0 {
		return fmt.Errorf("orders_generated cannot be negative")
	}
	if revenueGenerated < 0 {
		return fmt.Errorf("revenue_generated cannot be negative")
	}
	if endDate != nil && !endDate.IsZero() {
		start := startDate.Time.Truncate(24 * time.Hour)
		end := endDate.Time.Truncate(24 * time.Hour)
		if end.Before(start) {
			return fmt.Errorf("end_date cannot be before start_date")
		}
	}
	return nil
}

func ResolveCreateMarketingSpend(req *CreateMarketingSpendRequest) (*MarketingSpend, error) {
	if req == nil {
		return nil, fmt.Errorf("invalid request")
	}
	orders := 0
	if req.OrdersGenerated != nil {
		orders = *req.OrdersGenerated
	}
	revenue := 0.0
	if req.RevenueGenerated != nil {
		revenue = *req.RevenueGenerated
	}
	platform := strings.TrimSpace(req.Platform)
	campaign := strings.TrimSpace(req.CampaignName)
	if err := ValidateMarketingSpendFields(platform, campaign, req.AmountSpent, req.StartDate, req.EndDate, orders, revenue); err != nil {
		return nil, err
	}
	m := &MarketingSpend{
		Platform:         platform,
		CampaignName:     campaign,
		AmountSpent:      req.AmountSpent,
		StartDate:        req.StartDate,
		EndDate:          req.EndDate,
		OrdersGenerated:  orders,
		RevenueGenerated: revenue,
		Notes:            NormalizeOptionalString(req.Notes),
		CreatedBy:        strings.TrimSpace(req.CreatedBy),
	}
	if m.EndDate != nil && m.EndDate.IsZero() {
		m.EndDate = nil
	}
	if m.CreatedBy == "" {
		return nil, fmt.Errorf("created_by is required")
	}
	m.ApplyMetrics()
	return m, nil
}

func ApplyMarketingSpendUpdate(existing *MarketingSpend, req *UpdateMarketingSpendRequest, present map[string]bool) error {
	if existing == nil {
		return fmt.Errorf("marketing spend not found")
	}
	if req == nil {
		return nil
	}

	platform := existing.Platform
	campaign := existing.CampaignName
	amount := existing.AmountSpent
	start := existing.StartDate
	end := existing.EndDate
	orders := existing.OrdersGenerated
	revenue := existing.RevenueGenerated

	if present["platform"] {
		if req.Platform == nil || strings.TrimSpace(*req.Platform) == "" {
			return fmt.Errorf("platform cannot be empty")
		}
		platform = strings.TrimSpace(*req.Platform)
	}
	if present["campaign_name"] {
		if req.CampaignName == nil || strings.TrimSpace(*req.CampaignName) == "" {
			return fmt.Errorf("campaign_name cannot be empty")
		}
		campaign = strings.TrimSpace(*req.CampaignName)
	}
	if present["amount_spent"] {
		if req.AmountSpent == nil {
			return fmt.Errorf("amount_spent must be greater than 0")
		}
		amount = *req.AmountSpent
	}
	if present["start_date"] {
		if req.StartDate == nil || req.StartDate.IsZero() {
			return fmt.Errorf("start_date is required")
		}
		start = *req.StartDate
	}
	if present["end_date"] {
		end = req.EndDate
		if end != nil && end.IsZero() {
			end = nil
		}
	}
	if present["orders_generated"] {
		if req.OrdersGenerated == nil {
			orders = 0
		} else {
			orders = *req.OrdersGenerated
		}
	}
	if present["revenue_generated"] {
		if req.RevenueGenerated == nil {
			revenue = 0
		} else {
			revenue = *req.RevenueGenerated
		}
	}

	if err := ValidateMarketingSpendFields(platform, campaign, amount, start, end, orders, revenue); err != nil {
		return err
	}

	existing.Platform = platform
	existing.CampaignName = campaign
	existing.AmountSpent = amount
	existing.StartDate = start
	existing.EndDate = end
	existing.OrdersGenerated = orders
	existing.RevenueGenerated = revenue
	if present["notes"] {
		existing.Notes = NormalizeOptionalString(req.Notes)
	}
	existing.ApplyMetrics()
	return nil
}
