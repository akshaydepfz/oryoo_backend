package repositories

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"oryoo.com/helper"
	"oryoo.com/models"
)

type UserRepo struct{}

type TemplatePerformance struct {
	SentCount    int
	ClickedCount int
}

func NewUserRepo() *UserRepo {
	return &UserRepo{}
}

func (r *UserRepo) ListUsersWithFCMToken(ctx context.Context, limit, offset int) ([]models.User, error) {
	rows, err := helper.DB.QueryContext(ctx, `
		SELECT id, firebase_uid, fcm_token, last_open, last_active
		FROM users
		WHERE fcm_token IS NOT NULL AND btrim(fcm_token) <> ''
		ORDER BY id
		LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := make([]models.User, 0, limit)
	for rows.Next() {
		var user models.User
		var token sql.NullString
		var lastOpen sql.NullTime
		var lastActive sql.NullTime
		if err := rows.Scan(&user.ID, &user.FirebaseUID, &token, &lastOpen, &lastActive); err != nil {
			return nil, err
		}
		if !token.Valid || strings.TrimSpace(token.String) == "" {
			continue
		}
		tok := strings.TrimSpace(token.String)
		user.FCMToken = &tok
		if lastOpen.Valid {
			t := lastOpen.Time
			user.LastOpen = &t
		}
		if lastActive.Valid {
			t := lastActive.Time
			user.LastActive = t
		}
		users = append(users, user)
	}

	return users, rows.Err()
}

func (r *UserRepo) GetNotificationStates(ctx context.Context, userIDs []string) (map[string]models.UserNotificationState, error) {
	if len(userIDs) == 0 {
		return map[string]models.UserNotificationState{}, nil
	}

	rows, err := helper.DB.QueryContext(ctx, `
		SELECT user_id, last_sent_at, last_pending_count, ignore_count, last_clicked_at
		FROM user_notification_state
		WHERE user_id = ANY($1)
	`, pq.Array(userIDs))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	states := make(map[string]models.UserNotificationState, len(userIDs))
	for rows.Next() {
		var s models.UserNotificationState
		var lastSent sql.NullTime
		var lastClicked sql.NullTime
		if err := rows.Scan(&s.UserID, &lastSent, &s.LastPendingCount, &s.IgnoreCount, &lastClicked); err != nil {
			return nil, err
		}
		if lastSent.Valid {
			t := lastSent.Time
			s.LastSentAt = &t
		}
		if lastClicked.Valid {
			t := lastClicked.Time
			s.LastClickedAt = &t
		}
		states[s.UserID] = s
	}

	return states, rows.Err()
}

func (r *UserRepo) UpsertNotificationState(ctx context.Context, userID string, sentAt time.Time, pendingCount int) error {
	_, err := helper.DB.ExecContext(ctx, `
		INSERT INTO user_notification_state (user_id, last_sent_at, last_pending_count, ignore_count)
		VALUES ($1, $2, $3, 1)
		ON CONFLICT (user_id)
		DO UPDATE SET
			last_sent_at = EXCLUDED.last_sent_at,
			last_pending_count = EXCLUDED.last_pending_count,
			ignore_count = user_notification_state.ignore_count + 1
	`, userID, sentAt, pendingCount)
	return err
}

func (r *UserRepo) ClearFCMToken(ctx context.Context, userID string) error {
	_, err := helper.DB.ExecContext(ctx, `
		UPDATE users
		SET fcm_token = NULL
		WHERE firebase_uid = $1
	`, userID)
	return err
}

func (r *UserRepo) MarkNotificationClicked(ctx context.Context, userID string, clickedAt time.Time) error {
	_, err := helper.DB.ExecContext(ctx, `
		INSERT INTO user_notification_state (user_id, last_clicked_at, ignore_count)
		VALUES ($1, $2, 0)
		ON CONFLICT (user_id)
		DO UPDATE SET
			last_clicked_at = EXCLUDED.last_clicked_at,
			ignore_count = 0
	`, userID, clickedAt)
	return err
}

func (r *UserRepo) InsertNotificationLog(ctx context.Context, userID, templateKey, message string, sentAt time.Time) error {
	_, err := helper.DB.ExecContext(ctx, `
		INSERT INTO notification_logs (id, user_id, template_key, message, sent_at, clicked)
		VALUES ($1, $2, $3, $4, $5, false)
	`, uuid.New().String(), userID, templateKey, message, sentAt)
	return err
}

func (r *UserRepo) MarkLatestNotificationLogClicked(ctx context.Context, userID string) error {
	_, err := helper.DB.ExecContext(ctx, `
		UPDATE notification_logs
		SET clicked = true
		WHERE id = (
			SELECT id
			FROM notification_logs
			WHERE user_id = $1 AND clicked = false
			ORDER BY sent_at DESC
			LIMIT 1
		)
	`, userID)
	return err
}

func (r *UserRepo) UpdateIgnoreCount(ctx context.Context, userID string, ignoreCount int) error {
	_, err := helper.DB.ExecContext(ctx, `
		INSERT INTO user_notification_state (user_id, ignore_count)
		VALUES ($1, $2)
		ON CONFLICT (user_id)
		DO UPDATE SET ignore_count = EXCLUDED.ignore_count
	`, userID, ignoreCount)
	return err
}

func (r *UserRepo) GetTemplatePerformance(ctx context.Context, since time.Time) (map[string]TemplatePerformance, error) {
	rows, err := helper.DB.QueryContext(ctx, `
		SELECT
			template_key,
			COUNT(*) AS sent_count,
			COUNT(*) FILTER (WHERE clicked = true) AS clicked_count
		FROM notification_logs
		WHERE template_key <> '' AND sent_at >= $1
		GROUP BY template_key
	`, since)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	perf := make(map[string]TemplatePerformance)
	for rows.Next() {
		var key string
		var p TemplatePerformance
		if err := rows.Scan(&key, &p.SentCount, &p.ClickedCount); err != nil {
			return nil, err
		}
		perf[key] = p
	}
	return perf, rows.Err()
}
