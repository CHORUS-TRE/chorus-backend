package middleware

import (
	"context"
	"fmt"

	"github.com/CHORUS-TRE/chorus-backend/internal/api/v1/chorus"
	"github.com/CHORUS-TRE/chorus-backend/internal/audit"
	"github.com/CHORUS-TRE/chorus-backend/pkg/audit/model"
	"github.com/CHORUS-TRE/chorus-backend/pkg/audit/service"
)

var _ chorus.AppInstanceServiceServer = (*appInstanceControllerAudit)(nil)

type appInstanceControllerAudit struct {
	next        chorus.AppInstanceServiceServer
	auditWriter service.AuditWriter
}

func NewAppInstanceAuditMiddleware(auditWriter service.AuditWriter) func(chorus.AppInstanceServiceServer) chorus.AppInstanceServiceServer {
	return func(next chorus.AppInstanceServiceServer) chorus.AppInstanceServiceServer {
		return &appInstanceControllerAudit{
			next:        next,
			auditWriter: auditWriter,
		}
	}
}

func (c appInstanceControllerAudit) GetAppInstance(ctx context.Context, req *chorus.GetAppInstanceRequest) (*chorus.GetAppInstanceReply, error) {
	res, err := c.next.GetAppInstance(ctx, req)

	if err != nil {
		audit.Record(ctx, c.auditWriter, model.AuditActionAppInstanceRead,
			audit.WithDetail("app_instance_id", req.Id),
			audit.WithDescription(fmt.Sprintf("Failed to get app instance with ID %d.", req.Id)),
			audit.WithError(err),
		)
	}
	//  else {
	// 		audit.Record(ctx, c.auditWriter, model.AuditActionAppInstanceRead,
	// 			audit.WithDetail("app_instance_id", req.Id),
	// 			audit.WithDescription(fmt.Sprintf("Retrieved app instance with ID %d.", req.Id)),
	// 		)
	// }

	return res, err
}

func (c appInstanceControllerAudit) ListAppInstances(ctx context.Context, req *chorus.ListAppInstancesRequest) (*chorus.ListAppInstancesReply, error) {
	res, err := c.next.ListAppInstances(ctx, req)

	if err != nil {
		audit.Record(ctx, c.auditWriter, model.AuditActionAppInstanceList,
			audit.WithDescription("Failed to list app instances."),
			audit.WithError(err),
		)
	}
	//  else {
	// 		audit.Record(ctx, c.auditWriter, model.AuditActionAppInstanceList,
	// 			audit.WithDescription("Listed app instances."),
	// 			audit.WithDetail("result_count", len(res.Result.AppInstances)),
	// 		)
	// }

	return res, err
}

func (c appInstanceControllerAudit) CreateAppInstance(ctx context.Context, req *chorus.AppInstance) (*chorus.CreateAppInstanceReply, error) {
	res, err := c.next.CreateAppInstance(ctx, req)

	opts := []audit.Option{
		audit.WithDetail("workbench_id", req.WorkbenchId),
		audit.WithWorkspaceID(req.WorkspaceId),
		audit.WithWorkbenchID(req.WorkbenchId),
	}

	if err != nil {
		opts = append(opts,
			audit.WithDescription(fmt.Sprintf("Failed to launch instance of app (ID %d).", req.AppId)),
			audit.WithDetail("app_id", req.AppId),
			audit.WithError(err),
		)
	} else {
		ai := res.Result.AppInstance
		opts = append(opts,
			audit.WithDescription(fmt.Sprintf("Launched instance of '%s' (version %s)", ai.AppName, ai.AppDockerImageTag)),
			audit.WithDetail("app_instance_id", ai.Id),
			audit.WithDetail("app_id", ai.AppId),
			audit.WithDetail("app_name", ai.AppName),
			audit.WithDetail("app_image_registry", ai.AppDockerImageRegistry),
			audit.WithDetail("app_image_name", ai.AppDockerImageName),
			audit.WithDetail("app_image_tag", ai.AppDockerImageTag),
		)
	}

	audit.Record(ctx, c.auditWriter, model.AuditActionAppInstanceCreate, opts...)

	return res, err
}

func (c appInstanceControllerAudit) UpdateAppInstance(ctx context.Context, req *chorus.AppInstance) (*chorus.UpdateAppInstanceReply, error) {
	res, err := c.next.UpdateAppInstance(ctx, req)

	opts := []audit.Option{
		audit.WithDetail("app_instance_id", req.Id),
		audit.WithWorkspaceID(req.WorkspaceId),
		audit.WithWorkbenchID(req.WorkbenchId),
	}

	if err != nil {
		opts = append(opts,
			audit.WithDescription(fmt.Sprintf("Failed to update instance (ID %d) of app (ID %d).", req.Id, req.AppId)),
			audit.WithDetail("app_id", req.AppId),
			audit.WithError(err),
		)
	} else {
		ai := res.Result.AppInstance
		opts = append(opts,
			audit.WithDescription(fmt.Sprintf("Updated instance (ID %d) of '%s' %s (app ID %d).", ai.Id, ai.AppName, ai.AppDockerImageTag, ai.AppId)),
			audit.WithDetail("app_id", ai.AppId),
			audit.WithDetail("app_name", ai.AppName),
			audit.WithDetail("app_image_registry", ai.AppDockerImageRegistry),
			audit.WithDetail("app_image_name", ai.AppDockerImageName),
			audit.WithDetail("app_image_tag", ai.AppDockerImageTag),
		)
	}

	audit.Record(ctx, c.auditWriter, model.AuditActionAppInstanceUpdate, opts...)

	return res, err
}

func (c appInstanceControllerAudit) DeleteAppInstance(ctx context.Context, req *chorus.DeleteAppInstanceRequest) (*chorus.DeleteAppInstanceReply, error) {
	res, err := c.next.DeleteAppInstance(ctx, req)

	opts := []audit.Option{
		audit.WithDetail("app_instance_id", req.Id),
	}

	if err != nil {
		opts = append(opts,
			audit.WithDescription(fmt.Sprintf("Failed to terminate app instance (ID %d).", req.Id)),
			audit.WithError(err),
		)
	} else {
		ai := res.Result.AppInstance
		opts = append(opts,
			audit.WithDescription(fmt.Sprintf("Terminated instance of '%s' (version %s).", ai.AppName, ai.AppDockerImageTag)),
			audit.WithWorkspaceID(ai.WorkspaceId),
			audit.WithWorkbenchID(ai.WorkbenchId),
			audit.WithDetail("app_id", ai.AppId),
			audit.WithDetail("app_name", ai.AppName),
			audit.WithDetail("app_image_registry", ai.AppDockerImageRegistry),
			audit.WithDetail("app_image_name", ai.AppDockerImageName),
			audit.WithDetail("app_image_tag", ai.AppDockerImageTag),
		)
	}

	audit.Record(ctx, c.auditWriter, model.AuditActionAppInstanceDelete, opts...)

	return res, err
}
