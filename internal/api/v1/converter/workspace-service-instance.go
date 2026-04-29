package converter

import (
	"encoding/json"
	"fmt"

	"github.com/CHORUS-TRE/chorus-backend/internal/api/v1/chorus"
	"github.com/CHORUS-TRE/chorus-backend/pkg/workspace/model"
)

func WorkspaceServiceInstanceToBusiness(pb *chorus.WorkspaceServiceInstance) (*model.WorkspaceServiceInstance, error) {
	ca, err := FromProtoTimestamp(pb.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("unable to convert createdAt timestamp: %w", err)
	}
	ua, err := FromProtoTimestamp(pb.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("unable to convert updatedAt timestamp: %w", err)
	}

	var values model.JSONMap[any]
	if pb.ValuesOverrideJson != "" {
		if err := json.Unmarshal([]byte(pb.ValuesOverrideJson), &values); err != nil {
			return nil, fmt.Errorf("unable to unmarshal valuesOverrideJson: %w", err)
		}
	}

	computedValues := model.JSONMap[string](pb.ComputedValues)

	return &model.WorkspaceServiceInstance{
		ID:          pb.Id,
		TenantID:    pb.TenantId,
		WorkspaceID: pb.WorkspaceId,
		Name:        pb.Name,

		State:                  model.ServiceInstanceState(pb.State),
		ChartRegistry:          pb.ChartRegistry,
		ChartRepository:        pb.ChartRepository,
		ChartTag:               pb.ChartTag,
		Values:                 values,
		CredentialsSecretName:  pb.CredentialsSecretName,
		CredentialsPaths:       model.StringSlice(pb.CredentialsPaths),
		ConnectionInfoTemplate: pb.ConnectionInfoTemplate,
		ComputedValues:         computedValues,

		Status:         model.ServiceInstanceStatus(pb.Status),
		StatusMessage:  pb.StatusMessage,
		ConnectionInfo: pb.ConnectionInfo,
		SecretName:     pb.SecretName,

		CreatedAt: ca,
		UpdatedAt: ua,
	}, nil
}

func WorkspaceServiceInstanceFromBusiness(svc *model.WorkspaceServiceInstance) (*chorus.WorkspaceServiceInstance, error) {
	ca, err := ToProtoTimestamp(svc.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("unable to convert createdAt timestamp: %w", err)
	}
	ua, err := ToProtoTimestamp(svc.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("unable to convert updatedAt timestamp: %w", err)
	}

	var valuesOverrideJson string
	if svc.Values != nil {
		b, err := json.Marshal(svc.Values)
		if err != nil {
			return nil, fmt.Errorf("unable to marshal values: %w", err)
		}
		valuesOverrideJson = string(b)
	}

	return &chorus.WorkspaceServiceInstance{
		Id:          svc.ID,
		TenantId:    svc.TenantID,
		WorkspaceId: svc.WorkspaceID,
		Name:        svc.Name,

		State:                  svc.State.String(),
		ChartRegistry:          svc.ChartRegistry,
		ChartRepository:        svc.ChartRepository,
		ChartTag:               svc.ChartTag,
		ValuesOverrideJson:     valuesOverrideJson,
		CredentialsSecretName:  svc.CredentialsSecretName,
		CredentialsPaths:       []string(svc.CredentialsPaths),
		ConnectionInfoTemplate: svc.ConnectionInfoTemplate,
		ComputedValues:         map[string]string(svc.ComputedValues),

		Status:         svc.Status.String(),
		StatusMessage:  svc.StatusMessage,
		ConnectionInfo: svc.ConnectionInfo,
		SecretName:     svc.SecretName,

		CreatedAt: ca,
		UpdatedAt: ua,
	}, nil
}
