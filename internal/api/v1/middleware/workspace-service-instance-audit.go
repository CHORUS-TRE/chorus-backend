package middleware

import (
	"context"
	"fmt"

	"github.com/CHORUS-TRE/chorus-backend/internal/api/v1/chorus"
	"github.com/CHORUS-TRE/chorus-backend/internal/audit"
	"github.com/CHORUS-TRE/chorus-backend/pkg/audit/model"
	"github.com/CHORUS-TRE/chorus-backend/pkg/audit/service"
)

var _ chorus.WorkspaceServiceInstanceServiceServer = (*workspaceServiceInstanceControllerAudit)(nil)

type workspaceServiceInstanceControllerAudit struct {
	next        chorus.WorkspaceServiceInstanceServiceServer
	auditWriter service.AuditWriter
}

func NewWorkspaceServiceInstanceAuditMiddleware(auditWriter service.AuditWriter) func(chorus.WorkspaceServiceInstanceServiceServer) chorus.WorkspaceServiceInstanceServiceServer {
	return func(next chorus.WorkspaceServiceInstanceServiceServer) chorus.WorkspaceServiceInstanceServiceServer {
		return &workspaceServiceInstanceControllerAudit{
			next:        next,
			auditWriter: auditWriter,
		}
	}
}

func (c workspaceServiceInstanceControllerAudit) GetWorkspaceServiceInstance(ctx context.Context, req *chorus.GetWorkspaceServiceInstanceRequest) (*chorus.GetWorkspaceServiceInstanceReply, error) {
	res, err := c.next.GetWorkspaceServiceInstance(ctx, req)

	if err != nil {
		audit.Record(ctx, c.auditWriter, model.AuditActionServiceInstanceRead,
			audit.WithDetail("service_instance_id", req.Id),
			audit.WithDescription(fmt.Sprintf("Failed to get service instance %d.", req.Id)),
			audit.WithError(err),
		)
	}

	return res, err
}

func (c workspaceServiceInstanceControllerAudit) ListWorkspaceServiceInstances(ctx context.Context, req *chorus.ListWorkspaceServiceInstancesRequest) (*chorus.ListWorkspaceServiceInstancesReply, error) {
	res, err := c.next.ListWorkspaceServiceInstances(ctx, req)

	if err != nil {
		audit.Record(ctx, c.auditWriter, model.AuditActionServiceInstanceList,
			audit.WithDescription("Failed to list service instances."),
			audit.WithError(err),
		)
	}

	return res, err
}

func (c workspaceServiceInstanceControllerAudit) CreateWorkspaceServiceInstance(ctx context.Context, req *chorus.WorkspaceServiceInstance) (*chorus.CreateWorkspaceServiceInstanceReply, error) {
	res, err := c.next.CreateWorkspaceServiceInstance(ctx, req)

	opts := []audit.Option{
		audit.WithWorkspaceID(req.WorkspaceId),
		audit.WithDetail("workspace_id", req.WorkspaceId),
		audit.WithDetail("service_name", req.Name),
	}

	if err != nil {
		opts = append(opts,
			audit.WithDescription(fmt.Sprintf("Failed to create service instance '%s' in workspace %d.", req.Name, req.WorkspaceId)),
			audit.WithError(err),
		)
	} else {
		opts = append(opts,
			audit.WithDescription(fmt.Sprintf("Created service instance '%s' (ID %d) in workspace %d.", req.Name, res.Result.WorkspaceServiceInstance.Id, req.WorkspaceId)),
			audit.WithDetail("service_instance_id", res.Result.WorkspaceServiceInstance.Id),
		)
	}

	audit.Record(ctx, c.auditWriter, model.AuditActionServiceInstanceCreate, opts...)

	return res, err
}

func (c workspaceServiceInstanceControllerAudit) UpdateWorkspaceServiceInstance(ctx context.Context, req *chorus.WorkspaceServiceInstance) (*chorus.UpdateWorkspaceServiceInstanceReply, error) {
	res, err := c.next.UpdateWorkspaceServiceInstance(ctx, req)

	opts := []audit.Option{
		audit.WithWorkspaceID(req.WorkspaceId),
		audit.WithDetail("workspace_id", req.WorkspaceId),
		audit.WithDetail("service_instance_id", req.Id),
		audit.WithDetail("service_name", req.Name),
	}

	if err != nil {
		opts = append(opts,
			audit.WithDescription(fmt.Sprintf("Failed to update service instance '%s' (ID %d) in workspace %d.", req.Name, req.Id, req.WorkspaceId)),
			audit.WithError(err),
		)
	} else {
		opts = append(opts,
			audit.WithDescription(fmt.Sprintf("Updated service instance '%s' (ID %d) in workspace %d.", req.Name, req.Id, req.WorkspaceId)),
		)
	}

	audit.Record(ctx, c.auditWriter, model.AuditActionServiceInstanceUpdate, opts...)

	return res, err
}

func (c workspaceServiceInstanceControllerAudit) DeleteWorkspaceServiceInstance(ctx context.Context, req *chorus.DeleteWorkspaceServiceInstanceRequest) (*chorus.DeleteWorkspaceServiceInstanceReply, error) {
	res, err := c.next.DeleteWorkspaceServiceInstance(ctx, req)

	opts := []audit.Option{
		audit.WithDetail("service_instance_id", req.Id),
	}

	if err != nil {
		opts = append(opts,
			audit.WithDescription(fmt.Sprintf("Failed to delete service instance %d.", req.Id)),
			audit.WithError(err),
		)
	} else {
		opts = append(opts,
			audit.WithDescription(fmt.Sprintf("Deleted service instance %d.", req.Id)),
		)
	}

	audit.Record(ctx, c.auditWriter, model.AuditActionServiceInstanceDelete, opts...)

	return res, err
}
