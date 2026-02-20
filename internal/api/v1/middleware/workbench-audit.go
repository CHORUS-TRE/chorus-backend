package middleware

import (
	"context"
	"fmt"

	"github.com/CHORUS-TRE/chorus-backend/internal/api/v1/chorus"
	"github.com/CHORUS-TRE/chorus-backend/internal/audit"
	"github.com/CHORUS-TRE/chorus-backend/pkg/audit/model"
	"github.com/CHORUS-TRE/chorus-backend/pkg/audit/service"
)

var _ chorus.WorkbenchServiceServer = (*workbenchControllerAudit)(nil)

type workbenchControllerAudit struct {
	next        chorus.WorkbenchServiceServer
	auditWriter service.AuditWriter
}

func NewWorkbenchAuditMiddleware(auditWriter service.AuditWriter) func(chorus.WorkbenchServiceServer) chorus.WorkbenchServiceServer {
	return func(next chorus.WorkbenchServiceServer) chorus.WorkbenchServiceServer {
		return &workbenchControllerAudit{
			next:        next,
			auditWriter: auditWriter,
		}
	}
}

func (c workbenchControllerAudit) ListWorkbenches(ctx context.Context, req *chorus.ListWorkbenchesRequest) (*chorus.ListWorkbenchesReply, error) {
	res, err := c.next.ListWorkbenches(ctx, req)

	if err != nil {
		audit.Record(ctx, c.auditWriter, model.AuditActionWorkbenchList,
			audit.WithDetail("filter", req.Filter),
			audit.WithDescription("Failed to list workbenches."),
			audit.WithError(err),
		)
	}
	//  else {
	// 		audit.Record(ctx, c.auditWriter, model.AuditActionWorkbenchList,
	// 			audit.WithDetail("filter", req.Filter),
	// 			audit.WithDescription("Listed workbenches."),
	// 			audit.WithDetail("result_count", len(res.Result.Workbenches)),
	// 		)
	// }

	return res, err
}

func (c workbenchControllerAudit) CreateWorkbench(ctx context.Context, req *chorus.Workbench) (*chorus.CreateWorkbenchReply, error) {
	res, err := c.next.CreateWorkbench(ctx, req)

	opts := []audit.Option{
		audit.WithWorkspaceID(req.WorkspaceId),
		audit.WithDetail("workbench_name", req.Name),
	}

	if err != nil {
		opts = append(opts,
			audit.WithDescription("Failed to create workbench."),
			audit.WithError(err),
		)
	} else {
		opts = append(opts,
			audit.WithDescription(fmt.Sprintf("Created workbench with ID %d.", res.Result.Workbench.Id)),
			audit.WithWorkbenchID(res.Result.Workbench.Id),
			audit.WithDetail("workbench_id", res.Result.Workbench.Id),
		)
	}

	audit.Record(ctx, c.auditWriter, model.AuditActionWorkbenchCreate, opts...)

	return res, err
}

func (c workbenchControllerAudit) GetWorkbench(ctx context.Context, req *chorus.GetWorkbenchRequest) (*chorus.GetWorkbenchReply, error) {
	res, err := c.next.GetWorkbench(ctx, req)

	if err != nil {
		audit.Record(ctx, c.auditWriter, model.AuditActionWorkbenchRead,
			// TODO: audit.WithWorkspaceID(req.WorkspaceId),
			audit.WithWorkbenchID(req.Id),
			audit.WithDetail("workbench_id", req.Id),
			audit.WithDescription("Failed to get workbench."),
			audit.WithError(err),
		)
	}
	//  else {
	// 		audit.Record(ctx, c.auditWriter, model.AuditActionWorkbenchRead,
	// 			// TODO: audit.WithWorkspaceID(req.WorkspaceId),
	// 			audit.WithWorkbenchID(req.Id),
	// 			audit.WithDetail("workbench_id", req.Id),
	// 			audit.WithDescription(fmt.Sprintf("Retrieved workbench with ID %d.", req.Id)),
	// 			audit.WithDetail("workbench_name", res.Result.Workbench.Name),
	// 		)
	// }

	return res, err
}

func (c workbenchControllerAudit) UpdateWorkbench(ctx context.Context, req *chorus.Workbench) (*chorus.UpdateWorkbenchReply, error) {
	res, err := c.next.UpdateWorkbench(ctx, req)

	opts := []audit.Option{
		audit.WithWorkspaceID(req.WorkspaceId),
		audit.WithWorkbenchID(req.Id),
		audit.WithDetail("workbench_id", req.Id),
	}

	if err != nil {
		opts = append(opts,
			audit.WithDescription("Failed to update workbench."),
			audit.WithError(err),
		)
	} else {
		opts = append(opts,
			audit.WithDescription(fmt.Sprintf("Updated workbench with ID %d.", req.Id)),
			audit.WithDetail("workbench_name", req.Name),
		)
	}

	audit.Record(ctx, c.auditWriter, model.AuditActionWorkbenchUpdate, opts...)

	return res, err
}

func (c workbenchControllerAudit) DeleteWorkbench(ctx context.Context, req *chorus.DeleteWorkbenchRequest) (*chorus.DeleteWorkbenchReply, error) {
	res, err := c.next.DeleteWorkbench(ctx, req)

	opts := []audit.Option{

		audit.WithWorkbenchID(req.Id),
		audit.WithDetail("workbench_id", req.Id),
	}

	if err != nil {
		opts = append(opts,
			audit.WithDescription("Failed to delete workbench."),
			audit.WithError(err),
		)
	} else {
		opts = append(opts,
			audit.WithDescription(fmt.Sprintf("Deleted workbench with ID %d.", req.Id)),
			// TODO: audit.WithWorkspaceID(res.Result.Workbench.WorkspaceId),
		)
	}

	audit.Record(ctx, c.auditWriter, model.AuditActionWorkbenchDelete, opts...)

	return res, err
}

func (c workbenchControllerAudit) ManageUserRoleInWorkbench(ctx context.Context, req *chorus.ManageUserRoleInWorkbenchRequest) (*chorus.ManageUserRoleInWorkbenchReply, error) {
	res, err := c.next.ManageUserRoleInWorkbench(ctx, req)

	opts := []audit.Option{
		// TODO: audit.WithWorkspaceID(req.WorkspaceId),
		audit.WithWorkbenchID(req.Id),
		audit.WithDetail("workbench_id", req.Id),
		audit.WithDetail("user_id", req.UserId),
		audit.WithDetail("role", req.Role),
	}

	if err != nil {
		opts = append(opts,
			audit.WithDescription(fmt.Sprintf("Failed to add user %d to workbench %d with role %s.", req.UserId, req.Id, req.Role)),
			audit.WithError(err),
		)
	} else {
		opts = append(opts,
			audit.WithDescription(fmt.Sprintf("Added user %d to workbench %d with role %s.", req.UserId, req.Id, req.Role)),
			audit.WithWorkspaceID(res.Result.Workbench.WorkspaceId),
		)
	}

	audit.Record(ctx, c.auditWriter, model.AuditActionWorkbenchMemberAdd, opts...)

	return res, err
}

func (c workbenchControllerAudit) RemoveUserFromWorkbench(ctx context.Context, req *chorus.RemoveUserFromWorkbenchRequest) (*chorus.RemoveUserFromWorkbenchReply, error) {
	res, err := c.next.RemoveUserFromWorkbench(ctx, req)

	opts := []audit.Option{
		// TODO: audit.WithWorkspaceID(req.WorkspaceId),
		audit.WithWorkbenchID(req.Id),
		audit.WithDetail("workbench_id", req.Id),
		audit.WithDetail("user_id", req.UserId),
	}

	if err != nil {
		opts = append(opts,
			audit.WithDescription(fmt.Sprintf("Failed to remove user %d from workbench %d.", req.UserId, req.Id)),
			audit.WithError(err),
		)
	} else {
		opts = append(opts,
			audit.WithDescription(fmt.Sprintf("Removed user %d from workbench %d.", req.UserId, req.Id)),
			audit.WithWorkspaceID(res.Result.Workbench.WorkspaceId),
		)
	}

	audit.Record(ctx, c.auditWriter, model.AuditActionWorkbenchMemberRemove, opts...)

	return res, err
}

