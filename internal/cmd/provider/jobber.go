package provider

import (
	"context"
	"sync"

	"github.com/CHORUS-TRE/chorus-backend/internal/job"
	jobmemory "github.com/CHORUS-TRE/chorus-backend/internal/job/memory"
	jobpostgres "github.com/CHORUS-TRE/chorus-backend/internal/job/postgres"
	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
	"github.com/CHORUS-TRE/chorus-backend/internal/migration"
	"go.uber.org/zap"
)

var jobberOnce sync.Once
var jobberInstance job.Jobber

func ProvideJobber() job.Jobber {
	jobberOnce.Do(func() {
		cfg := ProvideConfig()

		var lockStore job.LockStore

		switch cfg.Daemon.Jobber.LockStore {
		case "memory":
			lockStore = jobmemory.NewLockStore()
		default:
			db := ProvideMainDB(WithClient("jobber"), WithMigrations(migration.GetMigration))
			lockStore = jobpostgres.NewLockStore(db.GetSqlxDB())
		}

		jobberInstance = job.NewJobber(
			ProvideComponentInfo().ComponentID,
			cfg.Daemon.Jobber.Enabled,
			lockStore,
			cfg.Daemon.Jobber.CheckInterval,
			cfg.Daemon.Jobber.Jitter,
			logger.TechLog,
		)
	})
	return jobberInstance
}

func InitDaemonJobs() {
	cfg := ProvideConfig()

	for name, jobConfig := range cfg.Daemon.Jobs {
		var j job.Job
		switch name {
		default:
			logger.TechLog.Warn(context.Background(), "unknown job in config, skipping", zap.String("job", name))
			continue
		}

		if err := ProvideJobber().Register(name, j, jobConfig); err != nil {
			logger.TechLog.Error(context.Background(), "failed to register job", zap.String("job", name), zap.Error(err))
		}
	}
}
