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
			audit.WithDescription(fmt.Sprintf("Failed to get app with ID %d.", req.Id)),
			audit.WithDetail("app_id", req.Id),
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
			audit.WithDescription(fmt.Sprintf("Failed to create app %s.", req.Name)),
			audit.WithError(err),
		)
	} else {
		opts = append(opts,
			audit.WithDescription(fmt.Sprintf("Created app %s (ID %d).", res.Result.App.Name, res.Result.App.Id)),
			audit.WithDetail("app_id", res.Result.App.Id),
			audit.WithDetail("app_name", res.Result.App.Name),
			audit.WithDetail("app_description", res.Result.App.Description),
			audit.WithDetail("app_registry", res.Result.App.DockerImageRegistry),
			audit.WithDetail("app_image_name", res.Result.App.DockerImageName),
			audit.WithDetail("app_image_tag", res.Result.App.DockerImageTag),
			audit.WithDetail("app_shm_size", res.Result.App.ShmSize),
			audit.WithDetail("app_max_cpu", res.Result.App.MaxCPU),
			audit.WithDetail("app_min_cpu", res.Result.App.MinCPU),
			audit.WithDetail("app_max_memory", res.Result.App.MaxMemory),
			audit.WithDetail("app_min_memory", res.Result.App.MinMemory),
			audit.WithDetail("app_max_ephemeral_storage", res.Result.App.MaxEphemeralStorage),
			audit.WithDetail("app_min_ephemeral_storage", res.Result.App.MinEphemeralStorage),
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
			audit.WithDescription(fmt.Sprintf("Failed to update app %s (ID %d).", req.Name, req.Id)),
			audit.WithError(err),
		)
	} else {
		opts = append(opts,
			audit.WithDescription(fmt.Sprintf("Updated app %s (ID %d).", req.Name, req.Id)),
			audit.WithDetail("app_id", res.Result.App.Id),
			audit.WithDetail("app_name", res.Result.App.Name),
			audit.WithDetail("app_description", res.Result.App.Description),
			audit.WithDetail("app_registry", res.Result.App.DockerImageRegistry),
			audit.WithDetail("app_image_name", res.Result.App.DockerImageName),
			audit.WithDetail("app_image_tag", res.Result.App.DockerImageTag),
			audit.WithDetail("app_shm_size", res.Result.App.ShmSize),
			audit.WithDetail("app_max_cpu", res.Result.App.MaxCPU),
			audit.WithDetail("app_min_cpu", res.Result.App.MinCPU),
			audit.WithDetail("app_max_memory", res.Result.App.MaxMemory),
			audit.WithDetail("app_min_memory", res.Result.App.MinMemory),
			audit.WithDetail("app_max_ephemeral_storage", res.Result.App.MaxEphemeralStorage),
			audit.WithDetail("app_min_ephemeral_storage", res.Result.App.MinEphemeralStorage),
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
			audit.WithDescription(fmt.Sprintf("Failed to delete app with ID %d.", req.Id)),
			audit.WithError(err),
		)
	} else {
		opts = append(opts,
			audit.WithDescription(fmt.Sprintf("Deleted app with ID %d.", req.Id)),
			audit.WithDetail("app_id", req.Id),
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
			audit.WithDescription(fmt.Sprintf("Failed to bulk create %d apps.", len(req.Apps))),
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
