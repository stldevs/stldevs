package aggregator

import (
	"log"
	"time"
)

// Interval is how often the aggregator runs.
const Interval = 24 * time.Hour

// Schedule runs the aggregator in a background goroutine. It runs once
// immediately if the last successful run is older than Interval (or unknown),
// then repeatedly on a ticker every Interval. The returned goroutine runs for
// the lifetime of the process.
//
// lastRun should return the timestamp of the most recent successful run, or
// the zero time if no run has ever happened.
func (a *Aggregator) Schedule(lastRun func() time.Time) {
	go func() {
		if since := time.Since(lastRun()); since >= Interval {
			log.Printf("aggregator: last run was %v ago, running now", since)
			a.Run()
		} else {
			wait := Interval - since
			log.Printf("aggregator: last run was %v ago, next run in %v", since, wait)
			time.Sleep(wait)
			a.Run()
		}
		ticker := time.NewTicker(Interval)
		defer ticker.Stop()
		for range ticker.C {
			a.Run()
		}
	}()
}

