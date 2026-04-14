package repo

import (
	"context"
	"database/sql"
	"log"
	"time"
)

const (
	dailyKeepAliveHour   = 0
	dailyKeepAliveMinute = 0
	dailyKeepAliveSecond = 0
)

// StartDailyDBKeepAlive runs a lightweight DB query every midnight to keep the
// connection pool active and surface connectivity issues in logs.
func StartDailyDBKeepAlive(db *sql.DB) {
	go func() {
		for {
			nextRun := nextDailyRun(time.Now())
			wait := time.Until(nextRun)
			timer := time.NewTimer(wait)
			<-timer.C

			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			log.Printf("daily DB keep-alive running scheduled_at=%s", nextRun.Format(time.RFC3339))
			if err := db.QueryRowContext(ctx, "SELECT 1 FROM dual").Scan(new(int)); err != nil {
				log.Printf("daily DB keep-alive failed: %v", err)
			} else {
				log.Printf("daily DB keep-alive success at %s", time.Now().Format(time.RFC3339))
			}
			cancel()
		}
	}()
}

func nextDailyRun(now time.Time) time.Time {
	next := time.Date(
		now.Year(),
		now.Month(),
		now.Day(),
		dailyKeepAliveHour,
		dailyKeepAliveMinute,
		dailyKeepAliveSecond,
		0,
		now.Location(),
	)
	if !next.After(now) {
		next = next.Add(24 * time.Hour)
	}
	return next
}
