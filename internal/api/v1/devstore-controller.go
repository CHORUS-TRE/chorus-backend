package v1

import (
	"context"

	"github.com/CHORUS-TRE/chorus-backend/internal/api/v1/chorus"
	jwt_model "github.com/CHORUS-TRE/chorus-backend/internal/jwt/model"
	"github.com/CHORUS-TRE/chorus-backend/pkg/devstore/model"
	"github.com/CHORUS-TRE/chorus-backend/pkg/devstore/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ chorus.DevstoreServiceServer = (*DevstoreController)(nil)

type DevstoreController struct {
	devstore service.Devstorer
}

func NewDevstoreController(devstoreService service.Devstorer) DevstoreController {
	return DevstoreController{
		devstore: devstoreService,
	}
}

func (c DevstoreController) ListGlobalEntries(ctx context.Context, req *chorus.ListEntriesRequest) (*chorus.ListEntriesReply, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	tenantid := uint64(1)
	res, err := c.devstore.ListEntries(ctx, tenantid, model.DevStoreScopeGlobal, 0)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "unable to call 'ListGlobalEntries': %v", err.Error())
	}

	entries := map[string]string{}
	for _, entry := range res {
		entries[entry.Key] = entry.Value
	}

	return &chorus.ListEntriesReply{Result: &chorus.ListEntriesResult{Entries: entries}}, nil
}

func (c DevstoreController) GetGlobalEntry(ctx context.Context, req *chorus.GetEntryRequest) (*chorus.GetEntryReply, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	tenantid := uint64(1)

	entry, err := c.devstore.GetEntry(ctx, tenantid, model.DevStoreScopeGlobal, 0, req.Key)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "unable to call 'GetGlobalEntry': %v", err.Error())
	}

	return &chorus.GetEntryReply{Result: &chorus.GetEntryResult{Key: entry.Key, Value: entry.Value}}, nil
}

func (c DevstoreController) PutGlobalEntry(ctx context.Context, req *chorus.PutEntryRequest) (*chorus.PutEntryReply, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	tenantid := uint64(1)

	entry, err := c.devstore.PutEntry(ctx, tenantid, model.DevStoreScopeGlobal, 0, req.Key, req.Value)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "unable to call 'PutGlobalEntry': %v", err.Error())
	}

	return &chorus.PutEntryReply{Result: &chorus.PutEntryResult{Key: entry.Key, Value: entry.Value}}, nil
}

func (c DevstoreController) DeleteGlobalEntry(ctx context.Context, req *chorus.DeleteEntryRequest) (*chorus.DeleteEntryReply, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	tenantid := uint64(1)

	err := c.devstore.DeleteEntry(ctx, tenantid, model.DevStoreScopeGlobal, 0, req.Key)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "unable to call 'DeleteGlobalEntry': %v", err.Error())
	}

	return &chorus.DeleteEntryReply{Result: &chorus.DeleteEntryResult{}}, nil
}

func (c DevstoreController) ListUserEntries(ctx context.Context, req *chorus.ListEntriesRequest) (*chorus.ListEntriesReply, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	userID, err := jwt_model.ExtractUserID(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "unable to extract user ID from token: %v", err.Error())
	}

	tenantID, err := jwt_model.ExtractTenantID(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "unable to extract tenant ID from token: %v", err.Error())
	}

	res, err := c.devstore.ListEntries(ctx, tenantID, model.DevStoreScopeUser, userID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "unable to call 'ListUserEntries': %v", err.Error())
	}

	entries := map[string]string{}
	for _, entry := range res {
		entries[entry.Key] = entry.Value
	}

	return &chorus.ListEntriesReply{Result: &chorus.ListEntriesResult{Entries: entries}}, nil
}

func (c DevstoreController) GetUserEntry(ctx context.Context, req *chorus.GetEntryRequest) (*chorus.GetEntryReply, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	tenantID, err := jwt_model.ExtractTenantID(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "unable to extract tenant ID from token: %v", err.Error())
	}

	userID, err := jwt_model.ExtractUserID(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "unable to extract user ID from token: %v", err.Error())
	}

	entry, err := c.devstore.GetEntry(ctx, tenantID, model.DevStoreScopeUser, userID, req.Key)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "unable to call 'GetUserEntry': %v", err.Error())
	}

	return &chorus.GetEntryReply{Result: &chorus.GetEntryResult{Key: entry.Key, Value: entry.Value}}, nil
}

func (c DevstoreController) PutUserEntry(ctx context.Context, req *chorus.PutEntryRequest) (*chorus.PutEntryReply, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	tenantID, err := jwt_model.ExtractTenantID(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "unable to extract tenant ID from token: %v", err.Error())
	}

	userID, err := jwt_model.ExtractUserID(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "unable to extract user ID from token: %v", err.Error())
	}

	entry, err := c.devstore.PutEntry(ctx, tenantID, model.DevStoreScopeUser, userID, req.Key, req.Value)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "unable to call 'PutUserEntry': %v", err.Error())
	}

	return &chorus.PutEntryReply{Result: &chorus.PutEntryResult{Key: entry.Key, Value: entry.Value}}, nil
}

func (c DevstoreController) DeleteUserEntry(ctx context.Context, req *chorus.DeleteEntryRequest) (*chorus.DeleteEntryReply, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	tenantID, err := jwt_model.ExtractTenantID(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "unable to extract tenant ID from token: %v", err.Error())
	}

	userID, err := jwt_model.ExtractUserID(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "unable to extract user ID from token: %v", err.Error())
	}

	err = c.devstore.DeleteEntry(ctx, tenantID, model.DevStoreScopeUser, userID, req.Key)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "unable to call 'DeleteUserEntry': %v", err.Error())
	}

	return &chorus.DeleteEntryReply{Result: &chorus.DeleteEntryResult{}}, nil
}

func (c DevstoreController) ListWorkspaceEntries(ctx context.Context, req *chorus.ListWorkspaceEntriesRequest) (*chorus.ListEntriesReply, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	tenantID, err := jwt_model.ExtractTenantID(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "unable to extract tenant ID from token: %v", err.Error())
	}

	res, err := c.devstore.ListEntries(ctx, tenantID, model.DevStoreScopeWorkspace, req.Id)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "unable to call 'ListWorkspaceEntries': %v", err.Error())
	}

	entries := map[string]string{}
	for _, entry := range res {
		entries[entry.Key] = entry.Value
	}

	return &chorus.ListEntriesReply{Result: &chorus.ListEntriesResult{Entries: entries}}, nil
}

func (c DevstoreController) GetWorkspaceEntry(ctx context.Context, req *chorus.GetWorkspaceEntryRequest) (*chorus.GetEntryReply, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	tenantID, err := jwt_model.ExtractTenantID(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "unable to extract tenant ID from token: %v", err.Error())
	}

	entry, err := c.devstore.GetEntry(ctx, tenantID, model.DevStoreScopeWorkspace, req.Id, req.Key)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "unable to call 'GetWorkspaceEntry': %v", err.Error())
	}

	return &chorus.GetEntryReply{Result: &chorus.GetEntryResult{Key: entry.Key, Value: entry.Value}}, nil
}

func (c DevstoreController) PutWorkspaceEntry(ctx context.Context, req *chorus.PutWorkspaceEntryRequest) (*chorus.PutEntryReply, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	tenantID, err := jwt_model.ExtractTenantID(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "unable to extract tenant ID from token: %v", err.Error())
	}

	entry, err := c.devstore.PutEntry(ctx, tenantID, model.DevStoreScopeWorkspace, req.Id, req.Key, req.Value)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "unable to call 'PutWorkspaceEntry': %v", err.Error())
	}

	return &chorus.PutEntryReply{Result: &chorus.PutEntryResult{Key: entry.Key, Value: entry.Value}}, nil
}

func (c DevstoreController) DeleteWorkspaceEntry(ctx context.Context, req *chorus.DeleteWorkspaceEntryRequest) (*chorus.DeleteEntryReply, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	tenantID, err := jwt_model.ExtractTenantID(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "unable to extract tenant ID from token: %v", err.Error())
	}

	err = c.devstore.DeleteEntry(ctx, tenantID, model.DevStoreScopeWorkspace, req.Id, req.Key)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "unable to call 'DeleteWorkspaceEntry': %v", err.Error())
	}

	return &chorus.DeleteEntryReply{Result: &chorus.DeleteEntryResult{}}, nil
}
