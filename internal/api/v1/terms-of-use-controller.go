package v1

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/CHORUS-TRE/chorus-backend/internal/api/v1/chorus"
	"github.com/CHORUS-TRE/chorus-backend/internal/api/v1/converter"
	cerr "github.com/CHORUS-TRE/chorus-backend/internal/errors"
	jwt_model "github.com/CHORUS-TRE/chorus-backend/internal/jwt/model"
	tou_service "github.com/CHORUS-TRE/chorus-backend/pkg/terms-of-use/service"
)

var _ chorus.TermsOfUseServiceServer = (*TermsOfUseController)(nil)

type TermsOfUseController struct {
	termsOfUse tou_service.TermsOfUseer
}

func NewTermsOfUseController(termsOfUse tou_service.TermsOfUseer) TermsOfUseController {
	return TermsOfUseController{termsOfUse: termsOfUse}
}

func (c TermsOfUseController) CreateTermsOfUseVersion(ctx context.Context, req *chorus.CreateTermsOfUseVersionRequest) (*chorus.CreateTermsOfUseVersionReply, error) {
	tenantID, err := jwt_model.ExtractTenantID(ctx)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "could not extract tenant id from jwt-token")
	}

	version, err := c.termsOfUse.CreateTermsOfUseVersion(ctx, tenantID, req.Content)
	if err != nil {
		return nil, err
	}

	proto, err := converter.TermsOfUseVersionFromBusiness(version)
	if err != nil {
		return nil, cerr.ErrConversion.Wrap(err, "failed to convert terms of use version")
	}

	return &chorus.CreateTermsOfUseVersionReply{
		Result: &chorus.CreateTermsOfUseVersionResult{TermsOfUseVersion: proto},
	}, nil
}

func (c TermsOfUseController) UpdateTermsOfUseVersion(ctx context.Context, req *chorus.UpdateTermsOfUseVersionRequest) (*chorus.UpdateTermsOfUseVersionReply, error) {
	tenantID, err := jwt_model.ExtractTenantID(ctx)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "could not extract tenant id from jwt-token")
	}

	version, err := c.termsOfUse.UpdateTermsOfUseVersion(ctx, tenantID, req.Id, req.Content)
	if err != nil {
		return nil, err
	}

	proto, err := converter.TermsOfUseVersionFromBusiness(version)
	if err != nil {
		return nil, cerr.ErrConversion.Wrap(err, "failed to convert terms of use version")
	}

	return &chorus.UpdateTermsOfUseVersionReply{
		Result: &chorus.UpdateTermsOfUseVersionResult{TermsOfUseVersion: proto},
	}, nil
}

func (c TermsOfUseController) PublishTermsOfUseVersion(ctx context.Context, req *chorus.PublishTermsOfUseVersionRequest) (*chorus.PublishTermsOfUseVersionReply, error) {
	tenantID, err := jwt_model.ExtractTenantID(ctx)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "could not extract tenant id from jwt-token")
	}

	version, err := c.termsOfUse.PublishTermsOfUseVersion(ctx, tenantID, req.Id)
	if err != nil {
		return nil, err
	}

	proto, err := converter.TermsOfUseVersionFromBusiness(version)
	if err != nil {
		return nil, cerr.ErrConversion.Wrap(err, "failed to convert terms of use version")
	}

	return &chorus.PublishTermsOfUseVersionReply{
		Result: &chorus.PublishTermsOfUseVersionResult{TermsOfUseVersion: proto},
	}, nil
}

func (c TermsOfUseController) GetTermsOfUseVersion(ctx context.Context, req *chorus.GetTermsOfUseVersionRequest) (*chorus.GetTermsOfUseVersionReply, error) {
	tenantID, _ := jwt_model.ExtractTenantID(ctx)

	version, err := c.termsOfUse.GetTermsOfUseVersion(ctx, tenantID, req.Id)
	if err != nil {
		return nil, err
	}

	proto, err := converter.TermsOfUseVersionFromBusiness(version)
	if err != nil {
		return nil, cerr.ErrConversion.Wrap(err, "failed to convert terms of use version")
	}

	return &chorus.GetTermsOfUseVersionReply{
		Result: &chorus.GetTermsOfUseVersionResult{TermsOfUseVersion: proto},
	}, nil
}

func (c TermsOfUseController) ListTermsOfUseVersions(ctx context.Context, req *chorus.ListTermsOfUseVersionsRequest) (*chorus.ListTermsOfUseVersionsReply, error) {
	tenantID, err := jwt_model.ExtractTenantID(ctx)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "could not extract tenant id from jwt-token")
	}

	pagination := converter.PaginationToBusiness(req.Pagination)
	versions, paginationResult, err := c.termsOfUse.ListTermsOfUseVersions(ctx, tenantID, &pagination)
	if err != nil {
		return nil, err
	}

	var protos []*chorus.TermsOfUseVersion
	for _, v := range versions {
		proto, err := converter.TermsOfUseVersionFromBusiness(v)
		if err != nil {
			return nil, cerr.ErrConversion.Wrap(err, "failed to convert terms of use version")
		}
		protos = append(protos, proto)
	}

	return &chorus.ListTermsOfUseVersionsReply{
		Result:     &chorus.ListTermsOfUseVersionsResult{TermsOfUseVersions: protos},
		Pagination: converter.PaginationResultFromBusiness(paginationResult),
	}, nil
}

func (c TermsOfUseController) GetCurrentTermsOfUseVersion(ctx context.Context, _ *chorus.GetCurrentTermsOfUseVersionRequest) (*chorus.GetCurrentTermsOfUseVersionReply, error) {
	tenantID, err := jwt_model.ExtractTenantID(ctx)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "could not extract tenant id from jwt-token")
	}

	version, err := c.termsOfUse.GetCurrentTermsOfUseVersion(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	proto, err := converter.TermsOfUseVersionFromBusiness(version)
	if err != nil {
		return nil, cerr.ErrConversion.Wrap(err, "failed to convert terms of use version")
	}

	return &chorus.GetCurrentTermsOfUseVersionReply{
		Result: &chorus.GetCurrentTermsOfUseVersionResult{TermsOfUseVersion: proto},
	}, nil
}

func (c TermsOfUseController) ListTermsOfUseAcceptances(ctx context.Context, _ *chorus.ListTermsOfUseAcceptancesRequest) (*chorus.ListTermsOfUseAcceptancesReply, error) {
	tenantID, err := jwt_model.ExtractTenantID(ctx)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "could not extract tenant id from jwt-token")
	}

	userID, err := jwt_model.ExtractUserID(ctx)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "could not extract user id from jwt-token")
	}

	acceptances, err := c.termsOfUse.ListTermsOfUseAcceptances(ctx, tenantID, userID)
	if err != nil {
		return nil, err
	}

	var protos []*chorus.TermsOfUseAcceptance
	for _, a := range acceptances {
		proto, err := converter.TermsOfUseAcceptanceFromBusiness(a)
		if err != nil {
			return nil, cerr.ErrConversion.Wrap(err, "failed to convert terms of use acceptance")
		}
		protos = append(protos, proto)
	}

	return &chorus.ListTermsOfUseAcceptancesReply{
		Result: &chorus.ListTermsOfUseAcceptancesResult{TermsOfUseAcceptances: protos},
	}, nil
}

func (c TermsOfUseController) GetMyTermsOfUseStatus(ctx context.Context, _ *chorus.GetMyTermsOfUseStatusRequest) (*chorus.GetMyTermsOfUseStatusReply, error) {
	tenantID, err := jwt_model.ExtractTenantID(ctx)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "could not extract tenant id from jwt-token")
	}

	userID, err := jwt_model.ExtractUserID(ctx)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "could not extract user id from jwt-token")
	}

	accepted, err := c.termsOfUse.GetMyTermsOfUseStatus(ctx, tenantID, userID)
	if err != nil {
		return nil, err
	}

	return &chorus.GetMyTermsOfUseStatusReply{
		Result: &chorus.GetMyTermsOfUseStatusResult{
			Status: &chorus.TermsOfUseUserStatus{Accepted: accepted},
		},
	}, nil
}

func (c TermsOfUseController) AcceptTermsOfUse(ctx context.Context, _ *chorus.AcceptTermsOfUseRequest) (*chorus.AcceptTermsOfUseReply, error) {
	tenantID, err := jwt_model.ExtractTenantID(ctx)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "could not extract tenant id from jwt-token")
	}

	userID, err := jwt_model.ExtractUserID(ctx)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "could not extract user id from jwt-token")
	}

	acceptance, err := c.termsOfUse.AcceptTermsOfUse(ctx, tenantID, userID)
	if err != nil {
		return nil, err
	}

	proto, err := converter.TermsOfUseAcceptanceFromBusiness(acceptance)
	if err != nil {
		return nil, cerr.ErrConversion.Wrap(err, "failed to convert terms of use acceptance")
	}

	return &chorus.AcceptTermsOfUseReply{
		Result: &chorus.AcceptTermsOfUseResult{TermsOfUseAcceptance: proto},
	}, nil
}
