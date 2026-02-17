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
		audit.Record(ctx, c.auditWriter,
			model.AuditActionWorkbenchList,
			audit.WithDescription("Failed to list workbenches."),
			audit.WithError(err),
			audit.WithDetail("filter", req.Filter),
		)
	}
	//  else {
	// 	audit.Record(ctx, c.auditWriter,
	// 		model.AuditActionWorkbenchList,
	// 		audit.WithDescription("Listed workbenches."),
	// 		audit.WithDetail("filter", req.Filter),
	// 		audit.WithDetail("result_count", len(res.Result.Workbenches)),
	// 	)
	// }

	return res, err
}

func (c workbenchControllerAudit) CreateWorkbench(ctx context.Context, req *chorus.Workbench) (*chorus.CreateWorkbenchReply, error) {
	res, err := c.next.CreateWorkbench(ctx, req)
	if err != nil {
		audit.Record(ctx, c.auditWriter,
			model.AuditActionWorkbenchCreate,
			audit.WithDescription("Failed to create workbench."),
			audit.WithWorkspaceID(req.WorkspaceId),
			audit.WithError(err),
			audit.WithDetail("workbench_name", req.Name),
		)
	} else {
		audit.Record(ctx, c.auditWriter,
			model.AuditActionWorkbenchCreate,
			audit.WithDescription(fmt.Sprintf("Created workbench with ID %d.", res.Result.Workbench.Id)),
			audit.WithWorkspaceID(res.Result.Workbench.WorkspaceId),
			audit.WithWorkbenchID(res.Result.Workbench.Id),
			audit.WithDetail("workbench_id", res.Result.Workbench.Id),
			audit.WithDetail("workbench_name", req.Name),
		)
	}

	return res, err
}

func (c workbenchControllerAudit) GetWorkbench(ctx context.Context, req *chorus.GetWorkbenchRequest) (*chorus.GetWorkbenchReply, error) {
	res, err := c.next.GetWorkbench(ctx, req)
	if err != nil {
		audit.Record(ctx, c.auditWriter,
			model.AuditActionWorkbenchRead,
			audit.WithDescription("Failed to get workbench."),
			// TODO: audit.WithWorkspaceID(req.WorkspaceId),
			audit.WithWorkbenchID(req.Id),
			audit.WithError(err),
			audit.WithDetail("workbench_id", req.Id),
		)
	}
	//  else {
	// 	audit.Record(ctx, c.auditWriter,
	// 		model.AuditActionWorkbenchRead,
	// 		audit.WithDescription(fmt.Sprintf("Retrieved workbench with ID %d.", req.Id)),
	// 		audit.WithWorkspaceID(res.Result.Workbench.WorkspaceId),
	// 		audit.WithWorkbenchID(req.Id),
	// 		audit.WithDetail("workbench_id", req.Id),
	// 		audit.WithDetail("workbench_name", res.Result.Workbench.Name),
	// 	)
	// }

	return res, err
}

func (c workbenchControllerAudit) UpdateWorkbench(ctx context.Context, req *chorus.Workbench) (*chorus.UpdateWorkbenchReply, error) {
	res, err := c.next.UpdateWorkbench(ctx, req)
	if err != nil {
		audit.Record(ctx, c.auditWriter,
			model.AuditActionWorkbenchUpdate,
			audit.WithDescription("Failed to update workbench."),
			audit.WithWorkspaceID(req.WorkspaceId),
			audit.WithWorkbenchID(req.Id),
			audit.WithError(err),
			audit.WithDetail("workbench_id", req.Id),
		)
	} else {
		audit.Record(ctx, c.auditWriter,
			model.AuditActionWorkbenchUpdate,
			audit.WithDescription(fmt.Sprintf("Updated workbench with ID %d.", req.Id)),
			audit.WithWorkspaceID(res.Result.Workbench.WorkspaceId),
			audit.WithWorkbenchID(req.Id),
			audit.WithDetail("workbench_id", req.Id),
			audit.WithDetail("workbench_name", req.Name),
		)
	}

	return res, err
}

func (c workbenchControllerAudit) DeleteWorkbench(ctx context.Context, req *chorus.DeleteWorkbenchRequest) (*chorus.DeleteWorkbenchReply, error) {
	res, err := c.next.DeleteWorkbench(ctx, req)
	if err != nil {
		audit.Record(ctx, c.auditWriter,
			model.AuditActionWorkbenchDelete,
			audit.WithDescription("Failed to delete workbench."),
			// TODO: audit.WithWorkspaceID(res.Result.Workbench.WorkspaceId),
			audit.WithWorkbenchID(req.Id),
			audit.WithError(err),
			audit.WithDetail("workbench_id", req.Id),
		)
	} else {
		audit.Record(ctx, c.auditWriter,
			model.AuditActionWorkbenchDelete,
			audit.WithDescription(fmt.Sprintf("Deleted workbench with ID %d.", req.Id)),
			// TODO: audit.WithWorkspaceID(res.Result.Workbench.WorkspaceId),
			audit.WithWorkbenchID(req.Id),
			audit.WithDetail("workbench_id", req.Id),
		)
	}

	return res, err
}

func (c workbenchControllerAudit) ManageUserRoleInWorkbench(ctx context.Context, req *chorus.ManageUserRoleInWorkbenchRequest) (*chorus.ManageUserRoleInWorkbenchReply, error) {
	res, err := c.next.ManageUserRoleInWorkbench(ctx, req)
	if err != nil {
		audit.Record(ctx, c.auditWriter,
			model.AuditActionWorkbenchMemberAdd,
			audit.WithDescription(fmt.Sprintf("Failed to add user %d to workbench %d with role %s.", req.UserId, req.Id, req.Role)),
			// TODO: audit.WithWorkspaceID(req.WorkspaceId),
			audit.WithWorkbenchID(req.Id),
			audit.WithError(err),
			audit.WithDetail("workbench_id", req.Id),
			audit.WithDetail("user_id", req.UserId),
			audit.WithDetail("role", req.Role),
		)
	} else {
		audit.Record(ctx, c.auditWriter,
			model.AuditActionWorkbenchMemberAdd,
			audit.WithDescription(fmt.Sprintf("Added user %d to workbench %d with role %s.", req.UserId, req.Id, req.Role)),
			audit.WithWorkspaceID(res.Result.Workbench.WorkspaceId),
			audit.WithWorkbenchID(req.Id),
			audit.WithDetail("workbench_id", req.Id),
			audit.WithDetail("user_id", req.UserId),
			audit.WithDetail("role", req.Role),
		)
	}

	return res, err
}

func (c workbenchControllerAudit) RemoveUserFromWorkbench(ctx context.Context, req *chorus.RemoveUserFromWorkbenchRequest) (*chorus.RemoveUserFromWorkbenchReply, error) {
	res, err := c.next.RemoveUserFromWorkbench(ctx, req)
	if err != nil {
		audit.Record(ctx, c.auditWriter,
			model.AuditActionWorkbenchMemberRemove,
			audit.WithDescription(fmt.Sprintf("Failed to remove user %d from workbench %d.", req.UserId, req.Id)),
			// TODO: audit.WithWorkspaceID(req.WorkspaceId),
			audit.WithWorkbenchID(req.Id),
			audit.WithError(err),
			audit.WithDetail("workbench_id", req.Id),
			audit.WithDetail("user_id", req.UserId),
		)
	} else {
		audit.Record(ctx, c.auditWriter,
			model.AuditActionWorkbenchMemberRemove,
			audit.WithDescription(fmt.Sprintf("Removed user %d from workbench %d.", req.UserId, req.Id)),
			audit.WithWorkspaceID(res.Result.Workbench.WorkspaceId),
			audit.WithWorkbenchID(req.Id),
			audit.WithDetail("workbench_id", req.Id),
			audit.WithDetail("user_id", req.UserId),
		)
	}

	return res, err
}
