package provider

import (
	"sync"

	"github.com/CHORUS-TRE/chorus-backend/internal/job"
	jobmemory "github.com/CHORUS-TRE/chorus-backend/internal/job/memory"
	jobpostgres "github.com/CHORUS-TRE/chorus-backend/internal/job/postgres"
	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
	"github.com/CHORUS-TRE/chorus-backend/internal/migration"
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
