package timeapi

import "time"

// Clock provides the current time. Inject a fixed implementation in tests.
type Clock interface {
	Now() time.Time
}

// RealClock uses the system clock.
type RealClock struct{}

func (RealClock) Now() time.Time {
	return time.Now()
}
