package postgres

import (
	"context"
	"fmt"
	"strings"

	"github.com/CHORUS-TRE/chorus-backend/pkg/common/storage"

	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
	common "github.com/CHORUS-TRE/chorus-backend/pkg/common/model"
	"github.com/CHORUS-TRE/chorus-backend/pkg/notification/model"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"go.uber.org/zap"
)

type NotificationStorage struct {
	db *sqlx.DB
}

func NewNotificationStorage(db *sqlx.DB) *NotificationStorage {
	return &NotificationStorage{db: db}
}

func (s *NotificationStorage) CreateNotification(ctx context.Context, notification *model.Notification, userIDs []uint64) error {
	const query = `INSERT INTO notifications (id, tenantid, message) VALUES ($1, $2, $3)`

	if _, err := s.db.ExecContext(ctx, query, notification.ID, notification.TenantID, notification.Message); err != nil {
		// on duplicate key return no error
		const DuplicateKeyErrorCode = "23505"
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == DuplicateKeyErrorCode {
			return nil
		}
		return fmt.Errorf("unable to create notification: %w", err)
	}
	for _, userID := range userIDs {
		const query = `INSERT INTO notifications_read_by (tenantid, notificationid, userid) VALUES ($1,$2,$3)`
		if _, err := s.db.ExecContext(ctx, query, notification.TenantID, notification.ID, userID); err != nil {
			logger.TechLog.Error(ctx, "unable to insert notifications_read_by", zap.Uint64("user-id", userID))
		}
	}
	return nil
}

func (s *NotificationStorage) CountUnreadNotifications(ctx context.Context, tenantID, userID uint64) (uint32, error) {
	const query = `
SELECT count(*) as count FROM notifications_read_by nrb
WHERE nrb.tenantid = $1 AND nrb.userid = $2
AND nrb.readat IS NULL
	`
	var count uint32
	if err := s.db.GetContext(ctx, &count, query, tenantID, userID); err != nil {
		return 0, err
	}

	return count, nil
}

func (s *NotificationStorage) MarkNotificationsAsRead(ctx context.Context, tenantID, userID uint64, notificationIDs []string, markAll bool) error {
	if markAll {
		const query = `UPDATE notifications_read_by SET readat=now() WHERE tenantid=$1 AND userid=$2 AND readat IS null`
		if _, err := s.db.ExecContext(ctx, query, tenantID, userID); err != nil {
			return fmt.Errorf("unable to mark notifications as read: %w", err)
		}
	} else {
		for _, notificationID := range notificationIDs {
			const query = `UPDATE notifications_read_by SET readat=now() WHERE tenantid=$1 AND notificationid=$2 AND userid=$3 AND readat IS null`
			if _, err := s.db.ExecContext(ctx, query, tenantID, notificationID, userID); err != nil {
				return fmt.Errorf("unable to mark notification as read: %w", err)
			}
		}
	}
	return nil
}

func (s *NotificationStorage) GetNotifications(ctx context.Context, tenantID, userID uint64, query string, isRead *bool, offset, limit uint64, sort common.Sort) ([]*model.Notification, uint32, error) {
	args, whereClauses := buildWhereClauses(tenantID, userID, query, isRead)
	selectArgs, sortClause := buildSortClause(args, sort, offset, limit)

	notifications, err := s.getNotifications(ctx, whereClauses, sortClause, selectArgs)
	if err != nil {
		return nil, 0, err
	}

	count, err := s.countNotifications(ctx, whereClauses, args)
	if err != nil {
		return nil, 0, err
	}

	return notifications, count, nil
}

func buildWhereClauses(tenantID, userID uint64, query string, isRead *bool) ([]interface{}, string) {
	var args []interface{}
	args = append(args, tenantID, userID)
	whereClauses := "WHERE n.tenantid = ? AND nrb.userid = ? "

	if query != "" {
		likeQuery := "%" + query + "%"
		args = append(args, likeQuery, likeQuery)
		whereClauses += " AND (n.message ilike ? OR n.id ilike ?)"
	}

	if isRead != nil {
		if *isRead {
			whereClauses += " AND nrb.readat IS NOT NULL"
		} else {
			whereClauses += " AND nrb.readat IS NULL"
		}
	}

	return args, whereClauses
}

func buildSortClause(args []interface{}, sort common.Sort, offset uint64, limit uint64) ([]interface{}, string) {
	columnName := model.NotificationSortTypeToString[strings.ToUpper(sort.SortType)]
	sortOrder := storage.SortOrderToString(strings.ToUpper(sort.SortOrder))
	args = append(args, offset, limit)
	sortClause := fmt.Sprintf(` ORDER BY %s %s offset ? limit ?`, columnName, sortOrder)
	return args, sortClause
}

func (s *NotificationStorage) getNotifications(ctx context.Context, whereClause, sortClause string, args []interface{}) ([]*model.Notification, error) {
	selectQuery := `
SELECT n.id, n.tenantid, n.message, n.createdat, nrb.readat FROM notifications n
left join notifications_read_by nrb on n.id = nrb.notificationid
` + whereClause + sortClause
	query, args, err := sqlx.In(selectQuery, args...)
	if err != nil {
		return nil, err
	}
	query = sqlx.Rebind(sqlx.DOLLAR, query)
	var tokens []*model.Notification
	if err := s.db.SelectContext(ctx, &tokens, query, args...); err != nil {
		return nil, err
	}
	return tokens, nil
}

func (s *NotificationStorage) countNotifications(ctx context.Context, whereClauses string, args []interface{}) (uint32, error) {
	countQuery := `
SELECT COUNT(n.id) FROM notifications n
left join notifications_read_by nrb on n.id = nrb.notificationid
` + whereClauses
	query, args, err := sqlx.In(countQuery, args...)
	if err != nil {
		return 0, err
	}
	query = sqlx.Rebind(sqlx.DOLLAR, query)
	var count uint32
	if err := s.db.GetContext(ctx, &count, query, args...); err != nil {
		return 0, err
	}
	return count, nil
}
