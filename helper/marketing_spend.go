package helper

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"oryoo.com/models"
)

func InsertMarketingSpend(m *models.MarketingSpend) error {
	now := time.Now()
	m.CreatedAt = &now
	m.UpdatedAt = &now

	query := `
		INSERT INTO marketing_spend (
			id,
			platform,
			campaign_name,
			amount_spent,
			start_date,
			end_date,
			orders_generated,
			revenue_generated,
			notes,
			created_by,
			created_at,
			updated_at
		)
		VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12
		)
	`

	_, err := DB.ExecContext(
		context.Background(),
		query,
		m.ID,
		m.Platform,
		m.CampaignName,
		m.AmountSpent,
		m.StartDate,
		endDateValue(m.EndDate),
		m.OrdersGenerated,
		m.RevenueGenerated,
		m.Notes,
		m.CreatedBy,
		m.CreatedAt,
		m.UpdatedAt,
	)
	return err
}

func endDateValue(d *models.MarketingSpendDate) interface{} {
	if d == nil || d.IsZero() {
		return nil
	}
	return d.Time
}

func scanMarketingSpend(scanner interface{ Scan(dest ...any) error }, m *models.MarketingSpend) error {
	var end sql.NullTime
	var notes sql.NullString
	err := scanner.Scan(
		&m.ID,
		&m.Platform,
		&m.CampaignName,
		&m.AmountSpent,
		&m.StartDate,
		&end,
		&m.OrdersGenerated,
		&m.RevenueGenerated,
		&notes,
		&m.CreatedBy,
		&m.CreatedAt,
		&m.UpdatedAt,
	)
	if err != nil {
		return err
	}
	if end.Valid {
		d := models.MarketingSpendDate{Time: end.Time}
		m.EndDate = &d
	} else {
		m.EndDate = nil
	}
	if notes.Valid {
		s := notes.String
		m.Notes = &s
	} else {
		m.Notes = nil
	}
	m.ApplyMetrics()
	return nil
}

const marketingSpendSelectColumns = `
			id,
			platform,
			campaign_name,
			amount_spent,
			start_date,
			end_date,
			orders_generated,
			revenue_generated,
			notes,
			created_by,
			created_at,
			updated_at`

func GetMarketingSpendByID(id string) (*models.MarketingSpend, error) {
	query := `SELECT ` + marketingSpendSelectColumns + ` FROM marketing_spend WHERE id = $1`
	var m models.MarketingSpend
	err := scanMarketingSpend(DB.QueryRowContext(context.Background(), query, id), &m)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("marketing spend not found")
	}
	if err != nil {
		return nil, err
	}
	return &m, nil
}

type MarketingSpendListFilter struct {
	CreatedBy string
	Platform  string
	StartDate string
	EndDate   string
	Month     string
}

func ListMarketingSpend(filter MarketingSpendListFilter) ([]models.MarketingSpend, error) {
	query := `SELECT ` + marketingSpendSelectColumns + `
		FROM marketing_spend
		WHERE created_by = $1`
	args := []interface{}{filter.CreatedBy}
	n := 2

	if p := strings.TrimSpace(filter.Platform); p != "" {
		query += fmt.Sprintf(" AND LOWER(platform) = LOWER($%d)", n)
		args = append(args, p)
		n++
	}
	if s := strings.TrimSpace(filter.StartDate); s != "" {
		query += fmt.Sprintf(" AND start_date >= $%d::date", n)
		args = append(args, s)
		n++
	}
	if e := strings.TrimSpace(filter.EndDate); e != "" {
		query += fmt.Sprintf(" AND start_date <= $%d::date", n)
		args = append(args, e)
		n++
	}
	if month := strings.TrimSpace(filter.Month); month != "" {
		query += fmt.Sprintf(" AND to_char(start_date, 'YYYY-MM') = $%d", n)
		args = append(args, month)
		n++
	}

	query += " ORDER BY start_date DESC, created_at DESC"

	rows, err := DB.QueryContext(context.Background(), query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]models.MarketingSpend, 0)
	for rows.Next() {
		var m models.MarketingSpend
		if err := scanMarketingSpend(rows, &m); err != nil {
			return nil, err
		}
		items = append(items, m)
	}
	return items, rows.Err()
}

func UpdateMarketingSpend(m *models.MarketingSpend) error {
	now := time.Now()
	m.UpdatedAt = &now
	query := `
		UPDATE marketing_spend SET
			platform = $1,
			campaign_name = $2,
			amount_spent = $3,
			start_date = $4,
			end_date = $5,
			orders_generated = $6,
			revenue_generated = $7,
			notes = $8,
			updated_at = $9
		WHERE id = $10 AND created_by = $11
	`
	res, err := DB.ExecContext(
		context.Background(),
		query,
		m.Platform,
		m.CampaignName,
		m.AmountSpent,
		m.StartDate,
		endDateValue(m.EndDate),
		m.OrdersGenerated,
		m.RevenueGenerated,
		m.Notes,
		m.UpdatedAt,
		m.ID,
		m.CreatedBy,
	)
	if err != nil {
		return err
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("marketing spend not found")
	}
	return nil
}

func DeleteMarketingSpend(id, createdBy string) error {
	res, err := DB.ExecContext(
		context.Background(),
		`DELETE FROM marketing_spend WHERE id = $1 AND created_by = $2`,
		id,
		createdBy,
	)
	if err != nil {
		return err
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("marketing spend not found")
	}
	return nil
}

func GetMarketingSpendSummary(createdBy string) (models.MarketingSpendSummary, error) {
	summary := models.EmptyMarketingSpendSummary()

	err := DB.QueryRowContext(context.Background(), `
		SELECT
			COALESCE(SUM(amount_spent), 0),
			COUNT(*),
			COALESCE(SUM(orders_generated), 0),
			COALESCE(SUM(revenue_generated), 0)
		FROM marketing_spend
		WHERE created_by = $1
	`, createdBy).Scan(
		&summary.TotalSpend,
		&summary.CampaignCount,
		&summary.TotalOrders,
		&summary.TotalRevenue,
	)
	if err != nil {
		return summary, err
	}

	err = DB.QueryRowContext(context.Background(), `
		SELECT COALESCE(SUM(amount_spent), 0)
		FROM marketing_spend
		WHERE created_by = $1
		  AND start_date >= date_trunc('month', CURRENT_DATE)::date
		  AND start_date < (date_trunc('month', CURRENT_DATE) + INTERVAL '1 month')::date
	`, createdBy).Scan(&summary.ThisMonthSpend)
	if err != nil {
		return summary, err
	}

	platformRows, err := DB.QueryContext(context.Background(), `
		SELECT platform, COALESCE(SUM(amount_spent), 0)
		FROM marketing_spend
		WHERE created_by = $1
		GROUP BY platform
		ORDER BY SUM(amount_spent) DESC, platform ASC
	`, createdBy)
	if err != nil {
		return summary, err
	}
	defer platformRows.Close()
	for platformRows.Next() {
		var item models.MarketingSpendPlatformBreakdown
		if err := platformRows.Scan(&item.Platform, &item.AmountSpent); err != nil {
			return summary, err
		}
		summary.PlatformBreakdown = append(summary.PlatformBreakdown, item)
	}
	if err := platformRows.Err(); err != nil {
		return summary, err
	}

	monthRows, err := DB.QueryContext(context.Background(), `
		SELECT to_char(date_trunc('month', start_date), 'YYYY-MM'), COALESCE(SUM(amount_spent), 0)
		FROM marketing_spend
		WHERE created_by = $1
		GROUP BY 1
		ORDER BY 1
	`, createdBy)
	if err != nil {
		return summary, err
	}
	defer monthRows.Close()
	for monthRows.Next() {
		var item models.MarketingSpendMonthly
		if err := monthRows.Scan(&item.Month, &item.AmountSpent); err != nil {
			return summary, err
		}
		summary.MonthlySpend = append(summary.MonthlySpend, item)
	}
	if err := monthRows.Err(); err != nil {
		return summary, err
	}

	summary.ApplyMetrics()
	return summary, nil
}
