package job

import (
	"context"
	"time"
)

type Status int

const (
	StatusSuccess Status = iota
	StatusFailure
)

func (s Status) String() string {
	switch s {
	case StatusSuccess:
		return "success"
	case StatusFailure:
		return "failure"
	default:
		return "unknown"
	}
}

type Job interface {
	Do(ctx context.Context) (msg string, err error)
}

type Registration struct {
	Job      Job
	Interval time.Duration
	Timeout  time.Duration // Optional timeout for the job execution. 0 means no timeout.
}
