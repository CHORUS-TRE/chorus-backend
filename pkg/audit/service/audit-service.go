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

	IsEnabled() bool
}

type AuditWriter interface {
	Record(ctx context.Context, entry *model.AuditEntry) error
}

type AuditReader interface {
	List(ctx context.Context, pagination *common_model.Pagination, filter *model.AuditFilter) ([]*model.AuditEntry, *common_model.PaginationResult, error)
}

type AuditStore interface {
	Record(ctx context.Context, entry *model.AuditEntry) error
	RecordBatch(ctx context.Context, entries []*model.AuditEntry) error
	List(ctx context.Context, pagination *common_model.Pagination, filter *model.AuditFilter) ([]*model.AuditEntry, *common_model.PaginationResult, error)
	Count(ctx context.Context, filter *model.AuditFilter) (int64, error)
}

type AuditService struct {
	store   AuditStore
	enabled bool
}

func NewAuditService(store AuditStore, enabled bool) *AuditService {
	return &AuditService{
		enabled: enabled,
		store:   store,
	}
}

func (s *AuditService) IsEnabled() bool {
	return s.enabled
}

func (s *AuditService) Record(ctx context.Context, entry *model.AuditEntry) error {
	if !s.enabled {
		return nil
	}

	return fmt.Errorf("Not implemented")
}

func (s *AuditService) List(ctx context.Context, pagination *common_model.Pagination, filter *model.AuditFilter) ([]*model.AuditEntry, *common_model.PaginationResult, error) {
	if !s.enabled {
		return []*model.AuditEntry{}, nil, nil
	}

	return nil, nil, fmt.Errorf("Not implemented")
}
