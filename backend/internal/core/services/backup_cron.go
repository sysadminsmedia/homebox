package services

import (
	"errors"
	"fmt"
	"time"

	"github.com/robfig/cron/v3"
)

// minCronGap is the shortest allowed time between scheduled backups. A backup
// zips the whole collection, so a schedule that fires every minute would only
// pile load on the server.
const minCronGap = 5 * time.Minute

var cronParser = cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor)

// parseCron parses a standard 5-field expression (or a descriptor such as
//
//	@daily)	and rejects schedules that run more often than minCronGap. A
//
// CRON_TZ=Zone prefix selects another time zone; the default is the server's.
func parseCron(expr string) (cron.Schedule, error) {
	if expr == "" {
		return nil, errors.New("a cron expression is required")
	}
	sched, err := cronParser.Parse(expr)
	if err != nil {
		return nil, fmt.Errorf("invalid cron expression: %w", err)
	}
	// Check a stretch of consecutive runs, not just the first pair, so patterns
	// like "0,1 * * * *" (a 1-minute gap, then 59) are caught.
	t := time.Now()
	prev := sched.Next(t)
	if prev.IsZero() {
		return nil, errors.New("the expression never matches a date")
	}
	for i := 0; i < 24; i++ {
		next := sched.Next(prev)
		if next.IsZero() {
			break
		}
		if next.Sub(prev) < minCronGap {
			return nil, fmt.Errorf("backups can run at most every %d minutes", int(minCronGap/time.Minute))
		}
		prev = next
	}
	return sched, nil
}

// cronPeriod is the longest gap between consecutive runs over the next stretch of
// a schedule. It sizes the no-recent-backup alert: a cron that runs weekdays only
// has a three-day gap every weekend, which must not look like a problem.
func cronPeriod(sched cron.Schedule) time.Duration {
	prev := sched.Next(time.Now())
	var longest time.Duration
	for i := 0; i < 64 && !prev.IsZero(); i++ {
		next := sched.Next(prev)
		if next.IsZero() {
			break
		}
		longest = max(longest, next.Sub(prev))
		prev = next
	}
	if longest == 0 {
		return 24 * time.Hour
	}
	return longest
}
