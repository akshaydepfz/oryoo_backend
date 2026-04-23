package cron

import (
	"context"
	"log"
	"time"

	"oryoo.com/services"
)

const runInterval = time.Hour

func StartNotificationJob(ctx context.Context) {
	svc, err := services.NewNotificationService(ctx)
	if err != nil {
		log.Printf("notification job disabled: %v", err)
		return
	}

	// Run once on startup so users don't wait for the next hour boundary.
	go func() {
		runOnce(ctx, svc)
		ticker := time.NewTicker(runInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				log.Println("notification job stopped")
				return
			case <-ticker.C:
				runOnce(ctx, svc)
			}
		}
	}()
}

func runOnce(ctx context.Context, svc *services.NotificationService) {
	runCtx, cancel := context.WithTimeout(ctx, 20*time.Minute)
	defer cancel()

	if err := svc.RunPendingReminderJob(runCtx); err != nil {
		log.Printf("notification job failed: %v", err)
		return
	}
	log.Println("notification job completed")
}
