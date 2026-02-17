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

	opts := []audit.Option{
		audit.WithDetail("app_instance_id", req.Id),
	}

	if err != nil {
		opts = append(opts,
			audit.WithDescription("Failed to get app instance."),
			audit.WithError(err),
		)
	}
	//  else {
	// 		opts = append(opts,
	// 			audit.WithDescription(fmt.Sprintf("Retrieved app instance with ID %d.", req.Id)),
	// 		)
	// }

	audit.Record(ctx, c.auditWriter, model.AuditActionAppInstanceRead, opts...)

	return res, err
}

func (c appInstanceControllerAudit) ListAppInstances(ctx context.Context, req *chorus.ListAppInstancesRequest) (*chorus.ListAppInstancesReply, error) {
	res, err := c.next.ListAppInstances(ctx, req)

	var opts []audit.Option

	if err != nil {
		opts = append(opts,
			audit.WithDescription("Failed to list app instances."),
			audit.WithError(err),
		)
	}
	//  else {
	// 		opts = append(opts,
	// 			audit.WithDescription("Listed app instances."),
	// 			audit.WithDetail("result_count", len(res.Result.AppInstances)),
	// 		)
	// }

	audit.Record(ctx, c.auditWriter, model.AuditActionAppInstanceList, opts...)

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
			audit.WithDescription("Failed to create app instance."),
			audit.WithError(err),
		)
	} else {
		opts = append(opts,
			audit.WithDescription(fmt.Sprintf("Created app instance with ID %d.", res.Result.AppInstance.Id)),
			audit.WithDetail("app_instance_id", res.Result.AppInstance.Id),
		)
	}

	audit.Record(ctx, c.auditWriter, model.AuditActionAppInstanceCreate, opts...)

	return res, err
}

func (c appInstanceControllerAudit) UpdateAppInstance(ctx context.Context, req *chorus.AppInstance) (*chorus.UpdateAppInstanceReply, error) {
	res, err := c.next.UpdateAppInstance(ctx, req)

	opts := []audit.Option{
		audit.WithWorkspaceID(req.WorkspaceId),
		audit.WithWorkbenchID(req.WorkbenchId),
		audit.WithDetail("app_instance_id", req.Id),
	}

	if err != nil {
		opts = append(opts,
			audit.WithDescription("Failed to update app instance."),
			audit.WithError(err),
		)
	} else {
		opts = append(opts,
			audit.WithDescription(fmt.Sprintf("Updated app instance with ID %d.", req.Id)),
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
			audit.WithDescription("Failed to delete app instance."),
			audit.WithError(err),
		)
	} else {
		opts = append(opts,
			audit.WithDescription(fmt.Sprintf("Deleted app instance with ID %d.", req.Id)),
			// TODO: audit.WithWorkspaceID(res.Result.AppInstance.WorkspaceId),
			// TODO: audit.WithWorkbenchID(res.Result.AppInstance.WorkbenchId),
		)
	}

	audit.Record(ctx, c.auditWriter, model.AuditActionAppInstanceDelete, opts...)

	return res, err
}
