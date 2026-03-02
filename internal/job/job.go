package job

import (
	"context"

	"github.com/CHORUS-TRE/chorus-backend/internal/config"
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
	Do(ctx context.Context, options map[string]interface{}) (msg string, err error)
}

type registration struct {
	Job    Job
	Config config.Job
}
