package repositories

import (
	"context"

	"github.com/lib/pq"
	"oryoo.com/helper"
)

type PaymentRepo struct{}

type PendingPaymentStats struct {
	Count       int
	Amount      float64
	ClientCount int
}

func NewPaymentRepo() *PaymentRepo {
	return &PaymentRepo{}
}

// CountPendingPaymentsByUser returns pending payment counts keyed by firebase_uid.
// It maps payment ownership via created_by fallback added_by.
func (r *PaymentRepo) CountPendingPaymentsByUser(ctx context.Context, userIDs []string) (map[string]int, error) {
	counts := make(map[string]int, len(userIDs))
	if len(userIDs) == 0 {
		return counts, nil
	}

	rows, err := helper.DB.QueryContext(ctx, `
		SELECT COALESCE(NULLIF(created_by, ''), NULLIF(added_by, '')) AS user_id, COUNT(*)
		FROM payments
		WHERE COALESCE(NULLIF(LOWER(btrim(status)), ''), 'pending') <> 'paid'
		  AND COALESCE(NULLIF(created_by, ''), NULLIF(added_by, '')) = ANY($1)
		GROUP BY COALESCE(NULLIF(created_by, ''), NULLIF(added_by, ''))
	`, pq.Array(userIDs))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var userID string
		var count int
		if err := rows.Scan(&userID, &count); err != nil {
			return nil, err
		}
		counts[userID] = count
	}

	return counts, rows.Err()
}

func (r *PaymentRepo) PendingPaymentStatsByUser(ctx context.Context, userIDs []string) (map[string]PendingPaymentStats, error) {
	stats := make(map[string]PendingPaymentStats, len(userIDs))
	if len(userIDs) == 0 {
		return stats, nil
	}

	rows, err := helper.DB.QueryContext(ctx, `
		SELECT
			COALESCE(NULLIF(created_by, ''), NULLIF(added_by, '')) AS user_id,
			COUNT(*) AS pending_count,
			COALESCE(SUM(amount), 0) AS pending_amount,
			COUNT(DISTINCT client_id) AS client_count
		FROM payments
		WHERE COALESCE(NULLIF(LOWER(btrim(status)), ''), 'pending') <> 'paid'
		  AND COALESCE(NULLIF(created_by, ''), NULLIF(added_by, '')) = ANY($1)
		GROUP BY COALESCE(NULLIF(created_by, ''), NULLIF(added_by, ''))
	`, pq.Array(userIDs))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var userID string
		var s PendingPaymentStats
		if err := rows.Scan(&userID, &s.Count, &s.Amount, &s.ClientCount); err != nil {
			return nil, err
		}
		stats[userID] = s
	}

	return stats, rows.Err()
}
