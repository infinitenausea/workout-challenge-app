package workers

import (
	"context"
	"log"

	"github.com/robfig/cron/v3"
	"workout-challenge-app/internal/database"
)

// StartFailedChallengeWorker starts a cron job that checks for expired active challenges
// and marks them as 'failed' if they haven't met their target progress.
func StartFailedChallengeWorker(db *database.DBWrapper) *cron.Cron {
	c := cron.New()
	
	// Run every hour. For testing purposes, you can change this to "* * * * *" (every minute)
	_, err := c.AddFunc("@hourly", func() {
		ctx := context.Background()
		
		query := `
			UPDATE challenges 
			SET status = 'failed' 
			WHERE status = 'active' 
			  AND end_date < CURRENT_DATE 
			  AND current_progress < target_value;
		`
		
		tag, err := db.Pool.Exec(ctx, query)
		if err != nil {
			log.Printf("[CronWorker] Failed to update expired challenges: %v\n", err)
			return
		}
		
		if tag.RowsAffected() > 0 {
			log.Printf("[CronWorker] Updated %d expired challenges to 'failed' status\n", tag.RowsAffected())
		}
	})
	
	if err != nil {
		log.Fatalf("Failed to initialize cron worker: %v", err)
	}

	c.Start()
	log.Println("Cron worker started: checking for expired challenges")
	return c
}
