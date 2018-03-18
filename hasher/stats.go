package hasher

import "time"

// Stats is a simple tracker for basic performance information around
// the hashing computations.
type Stats struct {
	Total     uint64        `json:"total"`   // Total number of hash computations performed
	Avg       float64       `json:"average"` // The average time (in milliseconds) of each operation
	totalTime time.Duration // The total time for all operations.. needed for average
}

// update increments the totals and recalculates the average.
func (s *Stats) update(elapsed time.Duration) {
	s.Total++
	s.totalTime += elapsed
	s.Avg = float64(s.totalTime.Nanoseconds()) / float64(s.Total) / 1000000
}
