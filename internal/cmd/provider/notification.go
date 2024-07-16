package provider

import (
	"context"
	"sync"

	v1 "github.com/CHORUS-TRE/chorus-backend/internal/api/v1"
	"github.com/CHORUS-TRE/chorus-backend/internal/api/v1/chorus"
	ctrl_mw "github.com/CHORUS-TRE/chorus-backend/internal/api/v1/middleware"
	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
	"github.com/CHORUS-TRE/chorus-backend/internal/migration"
	"github.com/CHORUS-TRE/chorus-backend/pkg/notification/service"
	service_mw "github.com/CHORUS-TRE/chorus-backend/pkg/notification/service/middleware"
	store_mw "github.com/CHORUS-TRE/chorus-backend/pkg/notification/store/middleware"
	"github.com/CHORUS-TRE/chorus-backend/pkg/notification/store/postgres"
	"github.com/CHORUS-TRE/chorus-backend/pkg/user/model"
)

var notificationOnce sync.Once
var notification service.Notificationer

func ProvideNotification() service.Notificationer {
	notificationOnce.Do(func() {
		notification = service.NewNotificationService(ProvideNotificationStore())
		notification = service_mw.Logging(logger.BizLog)(notification)
		notification = service_mw.Validation(ProvideValidator())(notification)
		notification = service_mw.NotificationCaching(logger.TechLog)(notification)
	})
	return notification
}

var notificationControllerOnce sync.Once
var notificationController chorus.NotificationServiceServer

func ProvideNotificationController() chorus.NotificationServiceServer {
	notificationControllerOnce.Do(func() {
		notificationController = v1.NewNotificationController(ProvideNotification())
		notificationController = ctrl_mw.NotificationAuthorizing(logger.SecLog, []string{model.RoleAuthenticated.String()})(notificationController)
	})
	return notificationController
}

var notificationStoreOnce sync.Once
var notificationStore service.NotificationStore

func ProvideNotificationStore() service.NotificationStore {
	notificationStoreOnce.Do(func() {
		db := ProvideMainDB(WithClient("notification-store"), WithMigrations(migration.GetMigration))
		switch db.Type {
		case POSTGRES:
			notificationStore = postgres.NewNotificationStorage(db.DB.GetSqlxDB())
		default:
			logger.TechLog.Fatal(context.Background(), "unsupported database type: "+db.Type)
		}
		notificationStore = store_mw.Logging(logger.TechLog)(notificationStore)
	})
	return notificationStore
}
