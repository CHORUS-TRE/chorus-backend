//go:build integration

package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"

	approval_request_model "github.com/CHORUS-TRE/chorus-backend/pkg/approval-request/model"
	approval_request_service "github.com/CHORUS-TRE/chorus-backend/pkg/approval-request/service"
	integration "github.com/CHORUS-TRE/chorus-backend/tests/integration/postgres"
)

const (
	approvalTestTenantID      = uint64(88100)
	approvalTestTenant2ID     = uint64(88101)
	approvalTestAliceID       = uint64(88200)
	approvalTestBobID         = uint64(88201)
	approvalTestCharlieID     = uint64(88202)
	approvalTestOtherTenantID = uint64(88203)
)

type approvalRequestFixtures struct {
	tenantID  uint64
	userIDs   map[string]uint64
	otherUser uint64
}

func setupApprovalRequestFixtures(t *testing.T, db *sqlx.DB) approvalRequestFixtures {
	t.Helper()
	ctx := context.Background()

	_, err := db.ExecContext(ctx, `INSERT INTO tenants (id, name, createdat, updatedat) VALUES ($1, 'approval_test_tenant', NOW(), NOW())`, approvalTestTenantID)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, `INSERT INTO tenants (id, name, createdat, updatedat) VALUES ($1, 'approval_test_tenant_2', NOW(), NOW())`, approvalTestTenant2ID)
	require.NoError(t, err)

	users := []struct {
		id       uint64
		tenantID uint64
		name     string
	}{
		{id: approvalTestAliceID, tenantID: approvalTestTenantID, name: "alice"},
		{id: approvalTestBobID, tenantID: approvalTestTenantID, name: "bob"},
		{id: approvalTestCharlieID, tenantID: approvalTestTenantID, name: "charlie"},
		{id: approvalTestOtherTenantID, tenantID: approvalTestTenant2ID, name: "other"},
	}

	for _, user := range users {
		_, err = db.ExecContext(ctx, `
			INSERT INTO users (id, tenantid, firstname, lastname, username, status, createdat, updatedat)
			VALUES ($1, $2, $3, $4, $5, 'active', NOW(), NOW())
		`, user.id, user.tenantID, user.name, user.name, user.name+"@test.com")
		require.NoError(t, err)
	}

	return approvalRequestFixtures{
		tenantID: approvalTestTenantID,
		userIDs: map[string]uint64{
			"alice":   approvalTestAliceID,
			"bob":     approvalTestBobID,
			"charlie": approvalTestCharlieID,
		},
		otherUser: approvalTestOtherTenantID,
	}
}

func newExtractionRequest(requesterID uint64, approverIDs []uint64, status approval_request_model.ApprovalRequestStatus, sourceWorkspaceID uint64, title string) *approval_request_model.ApprovalRequest {
	return &approval_request_model.ApprovalRequest{
		RequesterID:     requesterID,
		Type:            approval_request_model.ApprovalRequestTypeDataExtraction,
		Status:          status,
		Title:           title,
		Description:     title + " description",
		ApproverIDs:     approverIDs,
		CreatedAt:       time.Now().UTC(),
		UpdatedAt:       time.Now().UTC(),
		ApprovalMessage: "",
		Details: approval_request_model.ApprovalRequestDetails{
			DataExtractionDetails: &approval_request_model.DataExtractionDetails{
				SourceWorkspaceID: sourceWorkspaceID,
				Files: []approval_request_model.ApprovalRequestFile{{
					SourcePath:      "source/a.txt",
					DestinationPath: "approval-request/a.txt",
					Size:            10,
				}},
			},
		},
	}
}

func newTransferRequest(requesterID uint64, approverIDs []uint64, status approval_request_model.ApprovalRequestStatus, sourceWorkspaceID, destinationWorkspaceID uint64, title string) *approval_request_model.ApprovalRequest {
	return &approval_request_model.ApprovalRequest{
		RequesterID:     requesterID,
		Type:            approval_request_model.ApprovalRequestTypeDataTransfer,
		Status:          status,
		Title:           title,
		Description:     title + " description",
		ApproverIDs:     approverIDs,
		CreatedAt:       time.Now().UTC(),
		UpdatedAt:       time.Now().UTC(),
		ApprovalMessage: "",
		Details: approval_request_model.ApprovalRequestDetails{
			DataTransferDetails: &approval_request_model.DataTransferDetails{
				SourceWorkspaceID:      sourceWorkspaceID,
				DestinationWorkspaceID: destinationWorkspaceID,
				Files: []approval_request_model.ApprovalRequestFile{{
					SourcePath:      "source/b.txt",
					DestinationPath: "approval-request/b.txt",
					Size:            20,
				}},
			},
		},
	}
}

func TestApprovalRequestStorage_CreateGetAndUpdateApprovalRequest(t *testing.T) {
	db, err := integration.GetDB()
	require.NoError(t, err)
	t.Cleanup(func() {
		integration.CleanupTables(db)
	})

	fixtures := setupApprovalRequestFixtures(t, db)
	store := NewApprovalRequestStorage(db)
	ctx := context.Background()

	created, err := store.CreateApprovalRequest(ctx, fixtures.tenantID, newExtractionRequest(fixtures.userIDs["alice"], []uint64{fixtures.userIDs["bob"]}, approval_request_model.ApprovalRequestStatusPending, 101, "request-1"))
	require.NoError(t, err)
	require.NotZero(t, created.ID)
	require.Equal(t, fixtures.userIDs["alice"], created.RequesterID)
	require.Equal(t, []uint64{fixtures.userIDs["bob"]}, created.ApproverIDs)
	require.Equal(t, uint64(101), created.Details.DataExtractionDetails.SourceWorkspaceID)

	approvedByID := fixtures.userIDs["bob"]
	created.Status = approval_request_model.ApprovalRequestStatusApproved
	created.ApprovedByID = &approvedByID
	approvedAt := time.Now().UTC().Truncate(time.Second)
	created.ApprovedAt = &approvedAt
	created.ApprovalMessage = "approved"

	updated, err := store.UpdateApprovalRequest(ctx, fixtures.tenantID, created)
	require.NoError(t, err)
	require.Equal(t, approval_request_model.ApprovalRequestStatusApproved, updated.Status)
	require.NotNil(t, updated.ApprovedByID)
	require.Equal(t, approvedByID, *updated.ApprovedByID)
	require.NotNil(t, updated.ApprovedAt)
	require.Equal(t, "approved", updated.ApprovalMessage)

	fetched, err := store.GetApprovalRequest(ctx, fixtures.tenantID, created.ID)
	require.NoError(t, err)
	require.Equal(t, updated.ID, fetched.ID)
	require.Equal(t, approval_request_model.ApprovalRequestStatusApproved, fetched.Status)
	require.Equal(t, updated.Details.DataExtractionDetails.Files, fetched.Details.DataExtractionDetails.Files)
}

func TestApprovalRequestStorage_ListApprovalRequests_WithApproverAndRequesterFilters(t *testing.T) {
	db, err := integration.GetDB()
	require.NoError(t, err)
	t.Cleanup(func() {
		integration.CleanupTables(db)
	})

	fixtures := setupApprovalRequestFixtures(t, db)
	store := NewApprovalRequestStorage(db)
	ctx := context.Background()

	request1, err := store.CreateApprovalRequest(ctx, fixtures.tenantID, newExtractionRequest(fixtures.userIDs["alice"], []uint64{fixtures.userIDs["bob"]}, approval_request_model.ApprovalRequestStatusPending, 101, "alice-to-bob"))
	require.NoError(t, err)
	request2, err := store.CreateApprovalRequest(ctx, fixtures.tenantID, newTransferRequest(fixtures.userIDs["bob"], []uint64{fixtures.userIDs["alice"], fixtures.userIDs["charlie"]}, approval_request_model.ApprovalRequestStatusApproved, 202, 303, "bob-to-alice"))
	require.NoError(t, err)
	_, err = store.CreateApprovalRequest(ctx, fixtures.tenantID, newExtractionRequest(fixtures.userIDs["charlie"], []uint64{fixtures.userIDs["bob"]}, approval_request_model.ApprovalRequestStatusRejected, 404, "charlie-to-bob"))
	require.NoError(t, err)

	requesterID := fixtures.userIDs["bob"]
	requests, pagination, err := store.ListApprovalRequests(ctx, fixtures.tenantID, fixtures.userIDs["alice"], nil, approval_request_service.ApprovalRequestFilter{RequesterID: &requesterID})
	require.NoError(t, err)
	require.Len(t, requests, 1)
	require.NotNil(t, pagination)
	require.Equal(t, uint64(1), pagination.Total)
	require.Equal(t, request2.ID, requests[0].ID)

	approverID := fixtures.userIDs["bob"]
	pendingOnly := true
	requests, pagination, err = store.ListApprovalRequests(ctx, fixtures.tenantID, fixtures.userIDs["bob"], nil, approval_request_service.ApprovalRequestFilter{
		ApproverID:      &approverID,
		PendingApproval: &pendingOnly,
	})
	require.NoError(t, err)
	require.Len(t, requests, 1)
	require.Equal(t, uint64(1), pagination.Total)
	require.Equal(t, request1.ID, requests[0].ID)

	sourceWorkspaceID := uint64(202)
	requests, _, err = store.ListApprovalRequests(ctx, fixtures.tenantID, fixtures.userIDs["alice"], nil, approval_request_service.ApprovalRequestFilter{SourceWorkspaceID: &sourceWorkspaceID})
	require.NoError(t, err)
	require.Len(t, requests, 1)
	require.Equal(t, request2.ID, requests[0].ID)
}

func TestApprovalRequestStorage_CountMyApprovalRequests_FillsAllMaps(t *testing.T) {
	db, err := integration.GetDB()
	require.NoError(t, err)
	t.Cleanup(func() {
		integration.CleanupTables(db)
	})

	fixtures := setupApprovalRequestFixtures(t, db)
	store := NewApprovalRequestStorage(db)
	ctx := context.Background()

	_, err = store.CreateApprovalRequest(ctx, fixtures.tenantID, newExtractionRequest(fixtures.userIDs["alice"], []uint64{fixtures.userIDs["bob"]}, approval_request_model.ApprovalRequestStatusPending, 1001, "pending-extraction"))
	require.NoError(t, err)
	_, err = store.CreateApprovalRequest(ctx, fixtures.tenantID, newTransferRequest(fixtures.userIDs["bob"], []uint64{fixtures.userIDs["alice"]}, approval_request_model.ApprovalRequestStatusApproved, 1002, 1003, "approved-transfer"))
	require.NoError(t, err)
	_, err = store.CreateApprovalRequest(ctx, fixtures.tenantID, newTransferRequest(fixtures.userIDs["alice"], []uint64{fixtures.userIDs["charlie"]}, approval_request_model.ApprovalRequestStatusCancelled, 1004, 1005, "cancelled-transfer"))
	require.NoError(t, err)
	rejected, err := store.CreateApprovalRequest(ctx, fixtures.tenantID, newExtractionRequest(fixtures.userIDs["charlie"], []uint64{fixtures.userIDs["alice"]}, approval_request_model.ApprovalRequestStatusRejected, 1006, "deleted-rejected"))
	require.NoError(t, err)
	require.NoError(t, store.DeleteApprovalRequest(ctx, fixtures.tenantID, rejected.ID))
	_, err = store.CreateApprovalRequest(ctx, approvalTestTenant2ID, newExtractionRequest(fixtures.otherUser, []uint64{fixtures.otherUser}, approval_request_model.ApprovalRequestStatusApproved, 2001, "other-tenant"))
	require.NoError(t, err)

	counts, err := store.CountMyApprovalRequests(ctx, fixtures.tenantID, fixtures.userIDs["alice"])
	require.NoError(t, err)
	require.Equal(t, uint64(3), counts.Total)
	require.Equal(t, uint64(1), counts.TotalApprover)
	require.Equal(t, uint64(2), counts.TotalRequester)
	require.Equal(t, map[string]uint64{
		string(approval_request_model.ApprovalRequestStatusPending):   1,
		string(approval_request_model.ApprovalRequestStatusApproved):  1,
		string(approval_request_model.ApprovalRequestStatusRejected):  0,
		string(approval_request_model.ApprovalRequestStatusCancelled): 1,
	}, counts.CountByStatus)
	require.Equal(t, map[string]uint64{
		string(approval_request_model.ApprovalRequestTypeDataExtraction): 1,
		string(approval_request_model.ApprovalRequestTypeDataTransfer):   2,
	}, counts.CountByType)
}

func TestApprovalRequestStorage_DeleteApprovalRequest_RemovesFromQueries(t *testing.T) {
	db, err := integration.GetDB()
	require.NoError(t, err)
	t.Cleanup(func() {
		integration.CleanupTables(db)
	})

	fixtures := setupApprovalRequestFixtures(t, db)
	store := NewApprovalRequestStorage(db)
	ctx := context.Background()

	created, err := store.CreateApprovalRequest(ctx, fixtures.tenantID, newExtractionRequest(fixtures.userIDs["alice"], []uint64{fixtures.userIDs["bob"]}, approval_request_model.ApprovalRequestStatusPending, 501, "to-delete"))
	require.NoError(t, err)

	require.NoError(t, store.DeleteApprovalRequest(ctx, fixtures.tenantID, created.ID))

	_, err = store.GetApprovalRequest(ctx, fixtures.tenantID, created.ID)
	require.Error(t, err)

	requests, pagination, err := store.ListApprovalRequests(ctx, fixtures.tenantID, fixtures.userIDs["alice"], nil, approval_request_service.ApprovalRequestFilter{})
	require.NoError(t, err)
	require.Empty(t, requests)
	require.NotNil(t, pagination)
	require.Equal(t, uint64(0), pagination.Total)

	counts, err := store.CountMyApprovalRequests(ctx, fixtures.tenantID, fixtures.userIDs["alice"])
	require.NoError(t, err)
	require.Equal(t, uint64(0), counts.Total)
	for _, status := range approval_request_model.ApprovalRequestStatuses() {
		require.Contains(t, counts.CountByStatus, string(status))
		require.Zero(t, counts.CountByStatus[string(status)])
	}
	for _, requestType := range approval_request_model.ApprovalRequestTypes() {
		require.Contains(t, counts.CountByType, string(requestType))
		require.Zero(t, counts.CountByType[string(requestType)])
	}
}
