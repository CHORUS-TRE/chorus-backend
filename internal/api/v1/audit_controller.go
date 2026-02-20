package v1

import (
	"context"
	"fmt"

	"github.com/CHORUS-TRE/chorus-backend/internal/api/v1/chorus"
	"github.com/CHORUS-TRE/chorus-backend/internal/api/v1/converter"
	jwt_model "github.com/CHORUS-TRE/chorus-backend/internal/jwt/model"
	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
	"github.com/CHORUS-TRE/chorus-backend/internal/utils/grpc"
	audit_model "github.com/CHORUS-TRE/chorus-backend/pkg/audit/model"
	"github.com/CHORUS-TRE/chorus-backend/pkg/audit/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ chorus.AuditServiceServer = (*AuditController)(nil)

// AuditController is the audit service controller handler.
type AuditController struct {
	auditService service.AuditReader
}

func NewAuditController(auditService service.AuditReader) *AuditController {
	return &AuditController{
		auditService: auditService,
	}
}

func (c *AuditController) ListPlatformAudit(ctx context.Context, req *chorus.ListPlatformAuditRequest) (*chorus.ListPlatformAuditReply, error) {
	if req == nil {
		logger.TechLog.Error(ctx, "empty request")
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	tenantID, err := jwt_model.ExtractTenantID(ctx)
	if err != nil {
		logger.TechLog.Error(ctx, fmt.Sprintf("unable to extract tenant id from context: %v", err))
		return nil, status.Error(codes.InvalidArgument, "could not extract tenant id from jwt-token")
	}

	pagination := converter.PaginationToBusiness(req.Pagination)

	filter, err := converter.AuditFilterToBusiness(req.Filter)
	if err != nil {
		logger.TechLog.Error(ctx, fmt.Sprintf("invalid audit filter: %v", err))
		return nil, status.Errorf(codes.InvalidArgument, "invalid filter: %v", err)
	}

	res, paginationRes, err := c.auditService.List(ctx, tenantID, &pagination, filter)
	if err != nil {
		logger.TechLog.Error(ctx, fmt.Sprintf("unable to call 'ListAuditEntries': %v", err.Error()))
		return nil, status.Errorf(grpc.ErrorCode(err), "unable to call 'ListAuditEntries': %v", err)
	}

	var entries []*chorus.AuditEntry
	for _, r := range res {
		entry, err := converter.AuditEntryFromBusiness(r)
		if err != nil {
			logger.TechLog.Error(ctx, fmt.Sprintf("unable to convert audit entry from business model: %v", err))
			return nil, status.Errorf(codes.Internal, "conversion error: %v", err)
		}
		entries = append(entries, entry)
	}

	paginationResult := converter.PaginationResultFromBusiness(paginationRes)

	return &chorus.ListPlatformAuditReply{Result: &chorus.ListPlatformAuditResult{Entries: entries}, Pagination: paginationResult}, nil
}

func (c *AuditController) ListWorkspaceAudit(ctx context.Context, req *chorus.ListEntityAuditRequest) (*chorus.ListPlatformAuditReply, error) {
	if req == nil {
		logger.TechLog.Error(ctx, "empty request")
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	tenantID, err := jwt_model.ExtractTenantID(ctx)
	if err != nil {
		logger.TechLog.Error(ctx, fmt.Sprintf("unable to extract tenant id from context: %v", err))
		return nil, status.Error(codes.InvalidArgument, "could not extract tenant id from jwt-token")
	}

	pagination := converter.PaginationToBusiness(req.Pagination)

	filter, err := converter.AuditFilterToBusiness(req.Filter)
	if err != nil {
		logger.TechLog.Error(ctx, fmt.Sprintf("invalid audit filter: %v", err))
		return nil, status.Errorf(codes.InvalidArgument, "invalid filter: %v", err)
	}
	if filter == nil {
		filter = &audit_model.AuditFilter{}
	}
	filter.WorkspaceID = req.Id

	res, paginationRes, err := c.auditService.List(ctx, tenantID, &pagination, filter)
	if err != nil {
		logger.TechLog.Error(ctx, fmt.Sprintf("unable to call 'ListWorkspaceAudit': %v", err))
		return nil, status.Errorf(grpc.ErrorCode(err), "unable to call 'ListWorkspaceAudit': %v", err)
	}

	var entries []*chorus.AuditEntry
	for _, r := range res {
		entry, err := converter.AuditEntryFromBusiness(r)
		if err != nil {
			logger.TechLog.Error(ctx, fmt.Sprintf("unable to convert audit entry from business model: %v", err))
			return nil, status.Errorf(codes.Internal, "conversion error: %v", err)
		}
		entries = append(entries, entry)
	}

	return &chorus.ListPlatformAuditReply{Result: &chorus.ListPlatformAuditResult{Entries: entries}, Pagination: converter.PaginationResultFromBusiness(paginationRes)}, nil
}

func (c *AuditController) ListWorkbenchAudit(ctx context.Context, req *chorus.ListEntityAuditRequest) (*chorus.ListPlatformAuditReply, error) {
	if req == nil {
		logger.TechLog.Error(ctx, "empty request")
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	tenantID, err := jwt_model.ExtractTenantID(ctx)
	if err != nil {
		logger.TechLog.Error(ctx, fmt.Sprintf("unable to extract tenant id from context: %v", err))
		return nil, status.Error(codes.InvalidArgument, "could not extract tenant id from jwt-token")
	}

	pagination := converter.PaginationToBusiness(req.Pagination)

	filter, err := converter.AuditFilterToBusiness(req.Filter)
	if err != nil {
		logger.TechLog.Error(ctx, fmt.Sprintf("invalid audit filter: %v", err))
		return nil, status.Errorf(codes.InvalidArgument, "invalid filter: %v", err)
	}
	if filter == nil {
		filter = &audit_model.AuditFilter{}
	}
	filter.WorkbenchID = req.Id

	res, paginationRes, err := c.auditService.List(ctx, tenantID, &pagination, filter)
	if err != nil {
		logger.TechLog.Error(ctx, fmt.Sprintf("unable to call 'ListWorkbenchAudit': %v", err))
		return nil, status.Errorf(grpc.ErrorCode(err), "unable to call 'ListWorkbenchAudit': %v", err)
	}

	var entries []*chorus.AuditEntry
	for _, r := range res {
		entry, err := converter.AuditEntryFromBusiness(r)
		if err != nil {
			logger.TechLog.Error(ctx, fmt.Sprintf("unable to convert audit entry from business model: %v", err))
			return nil, status.Errorf(codes.Internal, "conversion error: %v", err)
		}
		entries = append(entries, entry)
	}

	return &chorus.ListPlatformAuditReply{Result: &chorus.ListPlatformAuditResult{Entries: entries}, Pagination: converter.PaginationResultFromBusiness(paginationRes)}, nil
}

func (c *AuditController) ListUserAudit(ctx context.Context, req *chorus.ListEntityAuditRequest) (*chorus.ListPlatformAuditReply, error) {
	if req == nil {
		logger.TechLog.Error(ctx, "empty request")
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	tenantID, err := jwt_model.ExtractTenantID(ctx)
	if err != nil {
		logger.TechLog.Error(ctx, fmt.Sprintf("unable to extract tenant id from context: %v", err))
		return nil, status.Error(codes.InvalidArgument, "could not extract tenant id from jwt-token")
	}

	pagination := converter.PaginationToBusiness(req.Pagination)

	filter, err := converter.AuditFilterToBusiness(req.Filter)
	if err != nil {
		logger.TechLog.Error(ctx, fmt.Sprintf("invalid audit filter: %v", err))
		return nil, status.Errorf(codes.InvalidArgument, "invalid filter: %v", err)
	}
	if filter == nil {
		filter = &audit_model.AuditFilter{}
	}
	filter.UserID = req.Id

	res, paginationRes, err := c.auditService.List(ctx, tenantID, &pagination, filter)
	if err != nil {
		logger.TechLog.Error(ctx, fmt.Sprintf("unable to call 'ListUserAudit': %v", err))
		return nil, status.Errorf(grpc.ErrorCode(err), "unable to call 'ListUserAudit': %v", err)
	}

	var entries []*chorus.AuditEntry
	for _, r := range res {
		entry, err := converter.AuditEntryFromBusiness(r)
		if err != nil {
			logger.TechLog.Error(ctx, fmt.Sprintf("unable to convert audit entry from business model: %v", err))
			return nil, status.Errorf(codes.Internal, "conversion error: %v", err)
		}
		entries = append(entries, entry)
	}

	return &chorus.ListPlatformAuditReply{Result: &chorus.ListPlatformAuditResult{Entries: entries}, Pagination: converter.PaginationResultFromBusiness(paginationRes)}, nil
}
