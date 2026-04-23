package repositories

import (
	"context"

	"github.com/lib/pq"
	"oryoo.com/helper"
)

type OrderRepo struct{}

type PendingOrderStats struct {
	Count       int
	Amount      float64
	ClientCount int
}

func NewOrderRepo() *OrderRepo {
	return &OrderRepo{}
}

// CountPendingOrdersByUser returns pending order counts keyed by firebase_uid.
// It maps order ownership via created_by fallback added_by.
func (r *OrderRepo) CountPendingOrdersByUser(ctx context.Context, userIDs []string) (map[string]int, error) {
	counts := make(map[string]int, len(userIDs))
	if len(userIDs) == 0 {
		return counts, nil
	}

	rows, err := helper.DB.QueryContext(ctx, `
		SELECT COALESCE(NULLIF(created_by, ''), NULLIF(added_by, '')) AS user_id, COUNT(*)
		FROM orders
		WHERE COALESCE(NULLIF(LOWER(btrim(payment_status)), ''), 'pending') <> 'paid'
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

func (r *OrderRepo) PendingOrderStatsByUser(ctx context.Context, userIDs []string) (map[string]PendingOrderStats, error) {
	stats := make(map[string]PendingOrderStats, len(userIDs))
	if len(userIDs) == 0 {
		return stats, nil
	}

	rows, err := helper.DB.QueryContext(ctx, `
		SELECT
			COALESCE(NULLIF(created_by, ''), NULLIF(added_by, '')) AS user_id,
			COUNT(*) AS pending_count,
			COALESCE(SUM(total_amount), 0) AS pending_amount,
			COUNT(DISTINCT client_id) AS client_count
		FROM orders
		WHERE COALESCE(NULLIF(LOWER(btrim(payment_status)), ''), 'pending') <> 'paid'
		  AND COALESCE(NULLIF(created_by, ''), NULLIF(added_by, '')) = ANY($1)
		GROUP BY COALESCE(NULLIF(created_by, ''), NULLIF(added_by, ''))
	`, pq.Array(userIDs))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var userID string
		var s PendingOrderStats
		if err := rows.Scan(&userID, &s.Count, &s.Amount, &s.ClientCount); err != nil {
			return nil, err
		}
		stats[userID] = s
	}

	return stats, rows.Err()
}
