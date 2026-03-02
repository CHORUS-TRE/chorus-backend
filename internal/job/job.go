package job

import (
	"context"
	"time"
)

type Status int

const (
	StatusSuccess Status = iota
	StatusFailure
	StatusSkipped
)

func (s Status) String() string {
	switch s {
	case StatusSuccess:
		return "success"
	case StatusFailure:
		return "failure"
	case StatusSkipped:
		return "skipped"
	default:
		return "unknown"
	}
}

type Job interface {
	Do(ctx context.Context) Status
}

type Registration struct {
	Job      Job
	Interval time.Duration
	Timeout  time.Duration
}
