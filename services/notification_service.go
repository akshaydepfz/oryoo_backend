package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"google.golang.org/api/option"
	"oryoo.com/models"
	"oryoo.com/repositories"
)

const (
	defaultBatchSize            = 500
	defaultCooldown             = 2 * time.Hour
	orderOnlyCooldown           = 6 * time.Hour
	ignoredCooldown             = 12 * time.Hour
	defaultHighPriorityCooldown = 90 * time.Minute
)

type NotificationService struct {
	userRepo    *repositories.UserRepo
	orderRepo   *repositories.OrderRepo
	paymentRepo *repositories.PaymentRepo
	fcmClient   *messaging.Client
}

func NewNotificationService(ctx context.Context) (*NotificationService, error) {
	client, err := newFCMClient(ctx)
	if err != nil {
		return nil, err
	}

	return &NotificationService{
		userRepo:    repositories.NewUserRepo(),
		orderRepo:   repositories.NewOrderRepo(),
		paymentRepo: repositories.NewPaymentRepo(),
		fcmClient:   client,
	}, nil
}

func newFCMClient(ctx context.Context) (*messaging.Client, error) {
	opt, source, err := initFirebaseOption()
	if err != nil {
		return nil, err
	}
	log.Printf("firebase credentials source: %s", source)

	app, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		return nil, fmt.Errorf("firebase init failed: %w", err)
	}

	client, err := app.Messaging(ctx)
	if err != nil {
		return nil, fmt.Errorf("firebase messaging init failed: %w", err)
	}
	return client, nil
}

// initFirebaseOption supports cloud env JSON (Koyeb) and local file fallback.
// Priority:
// 1) FIREBASE_CREDENTIALS (raw service-account JSON in env)
// 2) GOOGLE_APPLICATION_CREDENTIALS (path to service-account JSON file)
func initFirebaseOption() (option.ClientOption, string, error) {
	credsJSON := strings.TrimSpace(os.Getenv("FIREBASE_CREDENTIALS"))
	if credsJSON != "" {
		var parsed map[string]any
		if err := json.Unmarshal([]byte(credsJSON), &parsed); err != nil {
			return nil, "", fmt.Errorf("FIREBASE_CREDENTIALS is not valid JSON: %w", err)
		}

		// Minimal validation to fail fast on malformed/partial env payloads.
		required := []string{"type", "private_key", "client_email", "project_id"}
		for _, k := range required {
			v, ok := parsed[k]
			if !ok || strings.TrimSpace(fmt.Sprintf("%v", v)) == "" {
				return nil, "", fmt.Errorf("FIREBASE_CREDENTIALS missing required field: %s", k)
			}
		}

		return option.WithCredentialsJSON([]byte(credsJSON)), "FIREBASE_CREDENTIALS", nil
	}

	credPath := strings.TrimSpace(os.Getenv("GOOGLE_APPLICATION_CREDENTIALS"))
	if credPath != "" {
		return option.WithCredentialsFile(credPath), "GOOGLE_APPLICATION_CREDENTIALS", nil
	}

	return nil, "", fmt.Errorf("firebase credentials not found: set FIREBASE_CREDENTIALS or GOOGLE_APPLICATION_CREDENTIALS")
}

func (s *NotificationService) RunPendingReminderJob(ctx context.Context) error {
	batchSize := getEnvInt("NOTIFICATION_BATCH_SIZE", defaultBatchSize)
	offset := 0
	for {
		users, err := s.userRepo.ListUsersWithFCMToken(ctx, batchSize, offset)
		if err != nil {
			return err
		}
		if len(users) == 0 {
			break
		}

		if err := s.processBatch(ctx, users); err != nil {
			log.Printf("notification batch failed offset=%d size=%d err=%v", offset, len(users), err)
		}

		if len(users) < batchSize {
			break
		}
		offset += batchSize
	}
	return nil
}

func (s *NotificationService) processBatch(ctx context.Context, users []models.User) error {
	userIDs := make([]string, 0, len(users))
	userByUID := make(map[string]models.User, len(users))
	for _, u := range users {
		if strings.TrimSpace(u.FirebaseUID) == "" || u.FCMToken == nil || strings.TrimSpace(*u.FCMToken) == "" {
			continue
		}
		if shouldSkipForInactivity(u) {
			continue
		}
		userIDs = append(userIDs, u.FirebaseUID)
		userByUID[u.FirebaseUID] = u
	}
	if len(userIDs) == 0 {
		return nil
	}

	paymentStats, err := s.paymentRepo.PendingPaymentStatsByUser(ctx, userIDs)
	if err != nil {
		return err
	}
	orderStats, err := s.orderRepo.PendingOrderStatsByUser(ctx, userIDs)
	if err != nil {
		return err
	}
	states, err := s.userRepo.GetNotificationStates(ctx, userIDs)
	if err != nil {
		return err
	}
	templatePerformance, err := s.userRepo.GetTemplatePerformance(ctx, time.Now().AddDate(0, 0, -30))
	if err != nil {
		log.Printf("template performance fetch failed: %v", err)
		templatePerformance = map[string]repositories.TemplatePerformance{}
	}

	now := time.Now()
	for _, userID := range userIDs {
		user := userByUID[userID]
		state := states[userID]
		pStats := paymentStats[userID]
		oStats := orderStats[userID]
		pendingPayments := pStats.Count
		pendingOrders := oStats.Count
		totalPending := pendingPayments + pendingOrders

		if totalPending == 0 {
			continue
		}
		if !isPreferredSlot(now, user) {
			continue
		}

		screen := "orders"
		cooldown := getEnvDurationHours("NOTIFICATION_ORDER_COOLDOWN_HOURS", orderOnlyCooldown)
		if pendingPayments > 0 {
			screen = "payments"
			cooldown = getEnvDurationHours("NOTIFICATION_PAYMENT_COOLDOWN_HOURS", defaultCooldown)
			if pendingPayments >= getEnvInt("NOTIFICATION_HIGH_PRIORITY_PAYMENT_COUNT", 3) {
				cooldown = minDuration(cooldown, getEnvDurationMinutes("NOTIFICATION_HIGH_PRIORITY_COOLDOWN_MINUTES", defaultHighPriorityCooldown))
			}
		}

		adjustedIgnore := effectiveIgnoreCount(state, user, now)
		if adjustedIgnore != state.IgnoreCount {
			if err := s.userRepo.UpdateIgnoreCount(ctx, userID, adjustedIgnore); err != nil {
				log.Printf("ignore_count update failed user=%s err=%v", userID, err)
			}
			state.IgnoreCount = adjustedIgnore
		}

		cooldown = applyAdaptiveCooldown(cooldown, state, user, pendingPayments, now)
		if state.IgnoreCount >= 3 {
			cooldown = maxDuration(cooldown, getEnvDurationHours("NOTIFICATION_IGNORED_COOLDOWN_HOURS", ignoredCooldown))
		}

		shouldSendNow, skipReason := shouldSend(state, totalPending, now, cooldown)
		if !shouldSendNow {
			if skipReason == "cooldown" {
				log.Printf("skipped بسبب cooldown user=%s total_pending=%d", userID, totalPending)
			} else if skipReason == "no_change" {
				log.Printf("skipped بسبب no change user=%s total_pending=%d", userID, totalPending)
			}
			continue
		}

		token := strings.TrimSpace(*user.FCMToken)
		templateKey, body := buildMessage(pStats, oStats, templatePerformance)
		if err := s.sendPendingReminder(ctx, token, body, screen); err != nil {
			log.Printf("fcm send failed user=%s err=%v", userID, err)
			if isInvalidFCMTokenError(err) {
				if clearErr := s.userRepo.ClearFCMToken(ctx, userID); clearErr != nil {
					log.Printf("failed to clear invalid fcm token user=%s err=%v", userID, clearErr)
				} else {
					log.Printf("invalid fcm token cleared user=%s", userID)
				}
			}
			continue
		}
		log.Printf("notification sent user=%s total_pending=%d", userID, totalPending)
		if err := s.userRepo.UpsertNotificationState(ctx, userID, now, totalPending); err != nil {
			log.Printf("notification state upsert failed user=%s err=%v", userID, err)
		} else {
			log.Printf("state updated user=%s total_pending=%d", userID, totalPending)
		}
		if err := s.userRepo.InsertNotificationLog(ctx, userID, templateKey, body, now); err != nil {
			log.Printf("notification log insert failed user=%s err=%v", userID, err)
		}
	}

	return nil
}

func shouldSend(state models.UserNotificationState, totalPending int, now time.Time, cooldown time.Duration) (bool, string) {
	if state.LastSentAt == nil {
		return true, ""
	}
	if totalPending == state.LastPendingCount {
		return false, "no_change"
	}
	if now.Sub(*state.LastSentAt) < cooldown {
		return false, "cooldown"
	}
	return true, ""
}

func (s *NotificationService) sendPendingReminder(ctx context.Context, fcmToken, body, screen string) error {
	msg := &messaging.Message{
		Token: fcmToken,
		Notification: &messaging.Notification{
			Title: "Pending Reminder",
			Body:  body,
		},
		Data: map[string]string{
			"type":   "pending_reminder",
			"screen": screen,
		},
	}
	_, err := s.fcmClient.Send(ctx, msg)
	return err
}

func (s *NotificationService) TrackNotificationClicked(ctx context.Context, userID string) error {
	now := time.Now()
	if err := s.userRepo.MarkNotificationClicked(ctx, userID, now); err != nil {
		return err
	}
	return s.userRepo.MarkLatestNotificationLogClicked(ctx, userID)
}

func isInAllowedSendWindow(now time.Time) bool {
	// Optional hard guardrail to avoid night sends.
	if strings.ToLower(strings.TrimSpace(os.Getenv("NOTIFICATION_TIME_WINDOW_ENABLED"))) != "true" {
		return true
	}
	hour := now.Hour()
	startHour := clampHour(getEnvInt("NOTIFICATION_ALLOWED_START_HOUR", 8))
	endHour := clampHour(getEnvInt("NOTIFICATION_ALLOWED_END_HOUR", 22))
	if startHour < endHour {
		return hour >= startHour && hour < endHour
	}
	return hour >= startHour || hour < endHour
}

func isPreferredSlot(now time.Time, user models.User) bool {
	if !isInAllowedSendWindow(now) {
		return false
	}
	preferredHour := preferredHourFromUser(user)
	window := getEnvInt("NOTIFICATION_FLEX_WINDOW_HOURS", 2)
	if window > 6 {
		window = 6
	}
	diff := absHourDistance(now.Hour(), preferredHour)
	return diff <= window
}

func buildMessage(payment repositories.PendingPaymentStats, orders repositories.PendingOrderStats, perf map[string]repositories.TemplatePerformance) (string, string) {
	pendingAmount := payment.Amount
	if pendingAmount <= 0 {
		pendingAmount = orders.Amount
	}
	amountText := formatINR(pendingAmount)

	if payment.Count > 0 {
		c := payment.ClientCount
		if c <= 0 {
			c = payment.Count
		}
		templates := []messageTemplate{
			{Key: "payments_amount_clients", Body: fmt.Sprintf("%s pending from %d clients \U0001F4B8", amountText, c)},
			{Key: "payments_clients_ready", Body: fmt.Sprintf("%d customers are ready to pay you today", c)},
			{Key: "payments_collect_today", Body: fmt.Sprintf("Collect %s today. %d payments waiting", amountText, payment.Count)},
		}
		t := pickTemplate(templates, perf)
		return t.Key, t.Body
	}

	if orders.Count > 0 {
		templates := []messageTemplate{
			{Key: "orders_unpaid_count", Body: fmt.Sprintf("%d unpaid orders waiting. Follow up now", orders.Count)},
			{Key: "orders_amount_nudge", Body: fmt.Sprintf("%s stuck in pending orders. Nudge customers today", amountText)},
		}
		t := pickTemplate(templates, perf)
		return t.Key, t.Body
	}
	return "default_pending_updates", "You have pending updates"
}

func formatINR(v float64) string {
	if v <= 0 {
		return "₹0"
	}
	return fmt.Sprintf("₹%.0f", math.Round(v))
}

func maxDuration(a, b time.Duration) time.Duration {
	if a > b {
		return a
	}
	return b
}

func minDuration(a, b time.Duration) time.Duration {
	if a < b {
		return a
	}
	return b
}

type messageTemplate struct {
	Key  string
	Body string
}

func pickTemplate(templates []messageTemplate, perf map[string]repositories.TemplatePerformance) messageTemplate {
	if len(templates) == 0 {
		return messageTemplate{Key: "default_pending_updates", Body: "You have pending updates"}
	}
	weights := make([]float64, len(templates))
	totalWeight := 0.0
	for i, t := range templates {
		p := perf[t.Key]
		// Bayesian-friendly smoothing: keeps exploration while preferring high CTR.
		ctr := float64(p.ClickedCount+1) / float64(p.SentCount+3)
		volumeBoost := math.Min(float64(p.SentCount)/30.0, 1.0)
		weight := 0.6 + ctr + volumeBoost
		weights[i] = weight
		totalWeight += weight
	}
	r := rand.Float64() * totalWeight
	running := 0.0
	for i := range templates {
		running += weights[i]
		if r <= running {
			return templates[i]
		}
	}
	return templates[len(templates)-1]
}

func preferredHourFromUser(user models.User) int {
	if user.LastOpen != nil {
		return user.LastOpen.Hour()
	}
	if !user.LastActive.IsZero() {
		return user.LastActive.Hour()
	}
	return 10
}

func absHourDistance(a, b int) int {
	diff := a - b
	if diff < 0 {
		diff = -diff
	}
	if diff > 12 {
		return 24 - diff
	}
	return diff
}

func clampHour(h int) int {
	if h < 0 {
		return 0
	}
	if h > 23 {
		return 23
	}
	return h
}

func effectiveIgnoreCount(state models.UserNotificationState, user models.User, now time.Time) int {
	ignore := state.IgnoreCount
	if ignore <= 0 {
		return 0
	}
	recentWindow := getEnvDurationHours("NOTIFICATION_RECENT_ACTIVITY_HOURS", 24*time.Hour)
	ref := user.LastOpen
	if ref == nil && !user.LastActive.IsZero() {
		t := user.LastActive
		ref = &t
	}
	if ref != nil && now.Sub(*ref) <= recentWindow {
		ignore--
	}

	decayEvery := getEnvDurationHours("NOTIFICATION_IGNORE_DECAY_HOURS", 24*time.Hour)
	if decayEvery > 0 {
		base := now
		if state.LastSentAt != nil {
			base = *state.LastSentAt
		}
		steps := int(now.Sub(base) / decayEvery)
		ignore -= steps
	}
	if ignore < 0 {
		return 0
	}
	return ignore
}

func applyAdaptiveCooldown(base time.Duration, state models.UserNotificationState, user models.User, pendingPayments int, now time.Time) time.Duration {
	cooldown := base
	recentWindow := getEnvDurationHours("NOTIFICATION_RECENT_ACTIVITY_HOURS", 24*time.Hour)
	ref := user.LastOpen
	if ref == nil && !user.LastActive.IsZero() {
		t := user.LastActive
		ref = &t
	}
	if ref != nil && now.Sub(*ref) <= recentWindow {
		cooldown = time.Duration(float64(cooldown) * 0.8)
	}
	if state.IgnoreCount >= 2 {
		scale := 1.0 + (0.2 * float64(state.IgnoreCount))
		if scale > 2.5 {
			scale = 2.5
		}
		cooldown = time.Duration(float64(cooldown) * scale)
	}
	if pendingPayments > 0 {
		cooldown = time.Duration(float64(cooldown) * 0.85)
	}
	if cooldown < 30*time.Minute {
		return 30 * time.Minute
	}
	return cooldown
}

func shouldSkipForInactivity(user models.User) bool {
	// Optional enhancement: skip users inactive for N days.
	days := getEnvInt("NOTIFICATION_IGNORE_INACTIVE_DAYS", 0)
	if days <= 0 || user.LastOpen == nil {
		return false
	}
	return time.Since(*user.LastOpen) > (time.Duration(days) * 24 * time.Hour)
}

func getEnvInt(key string, fallback int) int {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil || n <= 0 {
		return fallback
	}
	return n
}

func getEnvDurationHours(key string, fallback time.Duration) time.Duration {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil || n <= 0 {
		return fallback
	}
	return time.Duration(n) * time.Hour
}

func getEnvDurationMinutes(key string, fallback time.Duration) time.Duration {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil || n <= 0 {
		return fallback
	}
	return time.Duration(n) * time.Minute
}

func isInvalidFCMTokenError(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(strings.ToLower(err.Error()), strings.ToLower("Requested entity was not found"))
}
