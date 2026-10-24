package batch_mail

import (
	"context"
	"sync"
	"time"
)

// MaxSendIntervalSeconds caps the per-task delay between two emails.
const MaxSendIntervalSeconds = 3600

// intervalPacer enforces a fixed minimum delay between consecutive sends of a
// task, independent of the thread-based rate controller (which allows bursts
// and at least ~20 emails/second per thread). It lets users throttle a
// campaign to e.g. one email every 30 seconds (#362, #343).
type intervalPacer struct {
	interval time.Duration
	mu       sync.Mutex
	next     time.Time
}

func newIntervalPacer(interval time.Duration) *intervalPacer {
	if interval <= 0 {
		return nil
	}
	return &intervalPacer{interval: interval}
}

// Wait blocks until the next send slot. A nil pacer never waits.
func (p *intervalPacer) Wait(ctx context.Context) error {
	if p == nil {
		return nil
	}

	p.mu.Lock()
	now := time.Now()
	slot := p.next
	if slot.Before(now) {
		slot = now
	}
	p.next = slot.Add(p.interval)
	p.mu.Unlock()

	delay := time.Until(slot)
	if delay <= 0 {
		return nil
	}

	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

// normalizeSendInterval clamps a user supplied interval (seconds).
func normalizeSendInterval(seconds int) int {
	if seconds < 0 {
		return 0
	}
	if seconds > MaxSendIntervalSeconds {
		return MaxSendIntervalSeconds
	}
	return seconds
}

// NormalizeSendInterval is the exported form used by controllers.
func NormalizeSendInterval(seconds int) int {
	return normalizeSendInterval(seconds)
}
