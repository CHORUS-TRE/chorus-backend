package job

import (
	"context"
	"time"

	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
	"go.uber.org/zap"
)

// StatusStore is the interface of the status store
type StatusStore interface {
	GarbageCollect(ctx context.Context, n, m int) error
}

type JobStatusGarbageCollecter struct {
	StatusStore StatusStore
	Successes   int
	Failures    int
}

func (j *JobStatusGarbageCollecter) Do(ctx context.Context, meta interface{}, arg interface{}) (_ interface{}, _ map[string]string, err error) {
	// log := logger.With(logger.TechLog, zap.String("job_name", meta.JobName), zap.String("job_id", meta.JobID))
	log := logger.With(logger.TechLog, zap.String("job_name", ""), zap.String("job_id", "0"))

	log.Info(ctx, "job started", zap.Time("now", time.Now().UTC()))

	err = j.StatusStore.GarbageCollect(ctx, j.Successes, j.Failures)
	if err != nil {
		log.Error(ctx, "could not run job statuses garbage collection", zap.Error(err))
		return nil, map[string]string{"msg": "could not run job statuses garbage collection", "err": err.Error()}, err
	}

	log.Info(ctx, "successfully finished")
	return nil, map[string]string{"msg": "successfully finished"}, nil
}
