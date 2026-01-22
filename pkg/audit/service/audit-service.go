package service

import (
	"context"
	"fmt"

	"github.com/CHORUS-TRE/chorus-backend/pkg/audit/model"
	common_model "github.com/CHORUS-TRE/chorus-backend/pkg/common/model"
)

type Auditer interface {
	AuditWriter
	AuditReader
}

type AuditWriter interface {
	Record(ctx context.Context, entry *model.AuditEntry) (*model.AuditEntry, error)
}

type AuditReader interface {
	List(ctx context.Context, pagination *common_model.Pagination, filter *model.AuditFilter) ([]*model.AuditEntry, *common_model.PaginationResult, error)
}

type AuditStore interface {
	Record(ctx context.Context, entry *model.AuditEntry) (*model.AuditEntry, error)
	BulkRecord(ctx context.Context, entries []*model.AuditEntry) ([]*model.AuditEntry, error)
	List(ctx context.Context, pagination *common_model.Pagination, filter *model.AuditFilter) ([]*model.AuditEntry, *common_model.PaginationResult, error)
	Count(ctx context.Context, filter *model.AuditFilter) (int64, error)
}

type auditService struct {
	store AuditStore
}

func NewAuditService(store AuditStore) *auditService {
	return &auditService{
		store: store,
	}
}

func (s *auditService) Record(ctx context.Context, entry *model.AuditEntry) (*model.AuditEntry, error) {
	createdEntry, err := s.store.Record(ctx, entry)
	if err != nil {
		return nil, fmt.Errorf("unable to record audit entry: %w", err)
	}

	return createdEntry, nil
}

func (s *auditService) List(ctx context.Context, pagination *common_model.Pagination, filter *model.AuditFilter) ([]*model.AuditEntry, *common_model.PaginationResult, error) {
	return nil, nil, fmt.Errorf("Not implemented")
}

// noOpAuditer is a no-operation implementation of the Auditer interface when audit logging is disabled
type noOpAuditer struct{}

func NewNoOpAuditer() Auditer {
	return &noOpAuditer{}
}

func (n *noOpAuditer) Record(ctx context.Context, entry *model.AuditEntry) (*model.AuditEntry, error) {
	return nil, nil
}

func (n *noOpAuditer) List(ctx context.Context, pagination *common_model.Pagination, filter *model.AuditFilter) ([]*model.AuditEntry, *common_model.PaginationResult, error) {
	return []*model.AuditEntry{}, &common_model.PaginationResult{}, nil
}
