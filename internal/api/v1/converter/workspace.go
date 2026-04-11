package converter

import (
	"encoding/json"
	"fmt"

	"github.com/CHORUS-TRE/chorus-backend/internal/api/v1/chorus"
	"github.com/CHORUS-TRE/chorus-backend/pkg/workspace/model"
	"google.golang.org/protobuf/types/known/structpb"
)

func WorkspaceToBusiness(workspace *chorus.Workspace) (*model.Workspace, error) {
	ca, err := FromProtoTimestamp(workspace.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("unable to convert createdAt timestamp: %w", err)
	}
	ua, err := FromProtoTimestamp(workspace.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("unable to convert updatedAt timestamp: %w", err)
	}

	services := model.JSONMap[model.WorkspaceServiceSpec]{}
	for name, svc := range workspace.GetServices() {
		var valuesRaw json.RawMessage
		if svc.Values != nil {
			b, err := json.Marshal(svc.Values.AsMap())
			if err != nil {
				return nil, fmt.Errorf("unable to marshal service values for %s: %w", name, err)
			}
			valuesRaw = b
		}
		var creds *model.WorkspaceServiceCredentials
		if svc.Credentials != nil {
			creds = &model.WorkspaceServiceCredentials{
				SecretName: svc.Credentials.SecretName,
				Paths:      svc.Credentials.Paths,
			}
		}
		var chart model.WorkspaceServiceChart
		if svc.Chart != nil {
			chart = model.WorkspaceServiceChart{
				Registry:   svc.Chart.Registry,
				Repository: svc.Chart.Repository,
				Tag:        svc.Chart.Tag,
			}
		}
		services[name] = model.WorkspaceServiceSpec{
			State:                  svc.State,
			Chart:                  chart,
			Values:                 valuesRaw,
			Credentials:            creds,
			ConnectionInfoTemplate: svc.ConnectionInfoTemplate,
			ComputedValues:         svc.ComputedValues,
		}
	}

	return &model.Workspace{
		ID: workspace.Id,

		TenantID: workspace.TenantId,
		UserID:   workspace.UserId,

		Name:        workspace.Name,
		ShortName:   workspace.ShortName,
		Description: workspace.Description,

		IsMain: workspace.IsMain,

		NetworkPolicy: workspace.NetworkPolicy,
		AllowedFQDNs:  model.StringSlice(workspace.AllowedFqdns),
		Clipboard:     workspace.Clipboard,
		Services:      services,

		CreatedAt: ca,
		UpdatedAt: ua,
	}, nil
}

func WorkspaceFromBusiness(workspace *model.Workspace) (*chorus.Workspace, error) {
	ca, err := ToProtoTimestamp(workspace.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("unable to convert createdAt timestamp: %w", err)
	}
	ua, err := ToProtoTimestamp(workspace.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("unable to convert updatedAt timestamp: %w", err)
	}

	services := map[string]*chorus.WorkspaceServiceSpec{}
	for name, svc := range workspace.Services {
		var values *structpb.Struct
		if len(svc.Values) > 0 {
			var m map[string]interface{}
			if err := json.Unmarshal(svc.Values, &m); err != nil {
				return nil, fmt.Errorf("unable to unmarshal service values for %s: %w", name, err)
			}
			values, err = structpb.NewStruct(m)
			if err != nil {
				return nil, fmt.Errorf("unable to convert service values for %s: %w", name, err)
			}
		}
		var creds *chorus.WorkspaceServiceCredentials
		if svc.Credentials != nil {
			creds = &chorus.WorkspaceServiceCredentials{
				SecretName: svc.Credentials.SecretName,
				Paths:      svc.Credentials.Paths,
			}
		}
		services[name] = &chorus.WorkspaceServiceSpec{
			State: svc.State,
			Chart: &chorus.WorkspaceServiceChart{
				Registry:   svc.Chart.Registry,
				Repository: svc.Chart.Repository,
				Tag:        svc.Chart.Tag,
			},
			Values:                 values,
			Credentials:            creds,
			ConnectionInfoTemplate: svc.ConnectionInfoTemplate,
			ComputedValues:         svc.ComputedValues,
		}
	}

	serviceStatuses := map[string]*chorus.WorkspaceServiceStatusInfo{}
	for name, ss := range workspace.ServiceStatuses {
		serviceStatuses[name] = &chorus.WorkspaceServiceStatusInfo{
			Status:         ss.Status,
			Message:        ss.Message,
			ConnectionInfo: ss.ConnectionInfo,
			SecretName:     ss.SecretName,
		}
	}

	return &chorus.Workspace{
		Id: workspace.ID,

		TenantId: workspace.TenantID,
		UserId:   workspace.UserID,

		Name:        workspace.Name,
		ShortName:   workspace.ShortName,
		Description: workspace.Description,

		Status: workspace.Status.String(),

		IsMain: workspace.IsMain,

		Namespace: workspace.GetClusterName(),

		NetworkPolicy:        workspace.NetworkPolicy,
		AllowedFqdns:         []string(workspace.AllowedFQDNs),
		NetworkPolicyStatus:  workspace.NetworkPolicyStatus,
		NetworkPolicyMessage: workspace.NetworkPolicyMessage,
		Clipboard:            workspace.Clipboard,
		Services:             services,
		ServiceStatuses:      serviceStatuses,

		CreatedAt: ca,
		UpdatedAt: ua,
	}, nil
}
