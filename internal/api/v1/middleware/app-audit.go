package middleware

import (
	"context"
	"fmt"

	"github.com/CHORUS-TRE/chorus-backend/internal/api/v1/chorus"
	"github.com/CHORUS-TRE/chorus-backend/internal/audit"
	"github.com/CHORUS-TRE/chorus-backend/pkg/audit/model"
	"github.com/CHORUS-TRE/chorus-backend/pkg/audit/service"
)

var _ chorus.AppServiceServer = (*appControllerAudit)(nil)

type appControllerAudit struct {
	next        chorus.AppServiceServer
	auditWriter service.AuditWriter
}

func NewAppAuditMiddleware(auditWriter service.AuditWriter) func(chorus.AppServiceServer) chorus.AppServiceServer {
	return func(next chorus.AppServiceServer) chorus.AppServiceServer {
		return &appControllerAudit{
			next:        next,
			auditWriter: auditWriter,
		}
	}
}

func (c appControllerAudit) GetApp(ctx context.Context, req *chorus.GetAppRequest) (*chorus.GetAppReply, error) {
	res, err := c.next.GetApp(ctx, req)

	if err != nil {
		audit.Record(ctx, c.auditWriter, model.AuditActionAppRead,
			audit.WithDetail("app_id", req.Id),
			audit.WithDescription("Failed to get app."),
			audit.WithError(err),
		)
	}
	//  else {
	// 		audit.Record(ctx, c.auditWriter, model.AuditActionAppRead,
	// 			audit.WithDetail("app_id", req.Id),
	// 			audit.WithDescription(fmt.Sprintf("Retrieved app with ID %d.", req.Id)),
	// 		)
	// }

	return res, err
}

func (c appControllerAudit) ListApps(ctx context.Context, req *chorus.ListAppsRequest) (*chorus.ListAppsReply, error) {
	res, err := c.next.ListApps(ctx, req)

	if err != nil {
		audit.Record(ctx, c.auditWriter, model.AuditActionAppList,
			audit.WithDescription("Failed to list apps."),
			audit.WithError(err),
		)
	}
	//  else {
	// 		audit.Record(ctx, c.auditWriter, model.AuditActionAppList,
	// 			audit.WithDescription("Listed apps."),
	// 			audit.WithDetail("result_count", len(res.Result.Apps)),
	// 		)
	// }

	return res, err
}

func (c appControllerAudit) CreateApp(ctx context.Context, req *chorus.App) (*chorus.CreateAppReply, error) {
	res, err := c.next.CreateApp(ctx, req)

	opts := []audit.Option{
		audit.WithDetail("app_name", req.Name),
	}

	if err != nil {
		opts = append(opts,
			audit.WithDescription("Failed to create app."),
			audit.WithError(err),
		)
	} else {
		opts = append(opts,
			audit.WithDescription(fmt.Sprintf("Created app with ID %d.", res.Result.App.Id)),
			audit.WithDetail("app_id", res.Result.App.Id),
		)
	}

	audit.Record(ctx, c.auditWriter, model.AuditActionAppCreate, opts...)

	return res, err
}

func (c appControllerAudit) UpdateApp(ctx context.Context, req *chorus.App) (*chorus.UpdateAppReply, error) {
	res, err := c.next.UpdateApp(ctx, req)

	opts := []audit.Option{
		audit.WithDetail("app_id", req.Id),
	}

	if err != nil {
		opts = append(opts,
			audit.WithDescription("Failed to update app."),
			audit.WithError(err),
		)
	} else {
		opts = append(opts,
			audit.WithDescription(fmt.Sprintf("Updated app with ID %d.", req.Id)),
			audit.WithDetail("app_name", req.Name),
		)
	}

	audit.Record(ctx, c.auditWriter, model.AuditActionAppUpdate, opts...)

	return res, err
}

func (c appControllerAudit) DeleteApp(ctx context.Context, req *chorus.DeleteAppRequest) (*chorus.DeleteAppReply, error) {
	res, err := c.next.DeleteApp(ctx, req)

	opts := []audit.Option{
		audit.WithDetail("app_id", req.Id),
	}

	if err != nil {
		opts = append(opts,
			audit.WithDescription("Failed to delete app."),
			audit.WithError(err),
		)
	} else {
		opts = append(opts,
			audit.WithDescription(fmt.Sprintf("Deleted app with ID %d.", req.Id)),
		)
	}

	audit.Record(ctx, c.auditWriter, model.AuditActionAppDelete, opts...)

	return res, err
}

func (c appControllerAudit) BulkCreateApps(ctx context.Context, req *chorus.BulkCreateAppsRequest) (*chorus.BulkCreateAppsReply, error) {
	res, err := c.next.BulkCreateApps(ctx, req)

	opts := []audit.Option{
		audit.WithDetail("app_count", len(req.Apps)),
	}

	if err != nil {
		opts = append(opts,
			audit.WithDescription("Failed to bulk create apps."),
			audit.WithError(err),
		)
	} else {
		opts = append(opts,
			audit.WithDescription(fmt.Sprintf("Bulk created %d apps.", len(req.Apps))),
		)
	}

	audit.Record(ctx, c.auditWriter, model.AuditActionAppBulkCreate, opts...)

	return res, err
}
