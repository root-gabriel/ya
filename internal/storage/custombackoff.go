package storage

import (
	"time"

	"github.com/sethvargo/go-retry"
)

type linearBackoff struct {
	attempt int
	timeout time.Duration
}

func NewlinearBackoff(timeout time.Duration) retry.Backoff {
	if timeout <= 0 {
		timeout = 1
	}

	return &linearBackoff{
		attempt: 0,
		timeout: timeout,
	}
}

// provides a linear sequence in 2 sec steps (1,3,5)
func (b *linearBackoff) Next() (time.Duration, bool) {
	next := b.timeout + (time.Second * 2 * time.Duration(b.attempt))
	b.attempt++
	return next, false
}
