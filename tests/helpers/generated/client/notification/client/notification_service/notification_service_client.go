// Code generated by go-swagger; DO NOT EDIT.

package notification_service

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"github.com/go-openapi/runtime"
	httptransport "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
)

// New creates a new notification service API client.
func New(transport runtime.ClientTransport, formats strfmt.Registry) ClientService {
	return &Client{transport: transport, formats: formats}
}

// New creates a new notification service API client with basic auth credentials.
// It takes the following parameters:
// - host: http host (github.com).
// - basePath: any base path for the API client ("/v1", "/v3").
// - scheme: http scheme ("http", "https").
// - user: user for basic authentication header.
// - password: password for basic authentication header.
func NewClientWithBasicAuth(host, basePath, scheme, user, password string) ClientService {
	transport := httptransport.New(host, basePath, []string{scheme})
	transport.DefaultAuthentication = httptransport.BasicAuth(user, password)
	return &Client{transport: transport, formats: strfmt.Default}
}

// New creates a new notification service API client with a bearer token for authentication.
// It takes the following parameters:
// - host: http host (github.com).
// - basePath: any base path for the API client ("/v1", "/v3").
// - scheme: http scheme ("http", "https").
// - bearerToken: bearer token for Bearer authentication header.
func NewClientWithBearerToken(host, basePath, scheme, bearerToken string) ClientService {
	transport := httptransport.New(host, basePath, []string{scheme})
	transport.DefaultAuthentication = httptransport.BearerToken(bearerToken)
	return &Client{transport: transport, formats: strfmt.Default}
}

/*
Client for notification service API
*/
type Client struct {
	transport runtime.ClientTransport
	formats   strfmt.Registry
}

// ClientOption may be used to customize the behavior of Client methods.
type ClientOption func(*runtime.ClientOperation)

// ClientService is the interface for Client methods
type ClientService interface {
	NotificationServiceCountUnreadNotifications(params *NotificationServiceCountUnreadNotificationsParams, opts ...ClientOption) (*NotificationServiceCountUnreadNotificationsOK, error)

	NotificationServiceGetNotifications(params *NotificationServiceGetNotificationsParams, opts ...ClientOption) (*NotificationServiceGetNotificationsOK, error)

	NotificationServiceMarkNotificationsAsRead(params *NotificationServiceMarkNotificationsAsReadParams, opts ...ClientOption) (*NotificationServiceMarkNotificationsAsReadOK, error)

	SetTransport(transport runtime.ClientTransport)
}

/*
NotificationServiceCountUnreadNotifications counts unread notifications

This endpoint returns the amount of unread notifications
*/
func (a *Client) NotificationServiceCountUnreadNotifications(params *NotificationServiceCountUnreadNotificationsParams, opts ...ClientOption) (*NotificationServiceCountUnreadNotificationsOK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewNotificationServiceCountUnreadNotificationsParams()
	}
	op := &runtime.ClientOperation{
		ID:                 "NotificationService_CountUnreadNotifications",
		Method:             "GET",
		PathPattern:        "/api/rest/v1/notifications/count",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http"},
		Params:             params,
		Reader:             &NotificationServiceCountUnreadNotificationsReader{formats: a.formats},
		Context:            params.Context,
		Client:             params.HTTPClient,
	}
	for _, opt := range opts {
		opt(op)
	}

	result, err := a.transport.Submit(op)
	if err != nil {
		return nil, err
	}
	success, ok := result.(*NotificationServiceCountUnreadNotificationsOK)
	if ok {
		return success, nil
	}
	// unexpected success response
	unexpectedSuccess := result.(*NotificationServiceCountUnreadNotificationsDefault)
	return nil, runtime.NewAPIError("unexpected success response: content available as default response in error", unexpectedSuccess, unexpectedSuccess.Code())
}

/*
NotificationServiceGetNotifications lists notifications

This endpoint returns a list of notifications
*/
func (a *Client) NotificationServiceGetNotifications(params *NotificationServiceGetNotificationsParams, opts ...ClientOption) (*NotificationServiceGetNotificationsOK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewNotificationServiceGetNotificationsParams()
	}
	op := &runtime.ClientOperation{
		ID:                 "NotificationService_GetNotifications",
		Method:             "GET",
		PathPattern:        "/api/rest/v1/notifications",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http"},
		Params:             params,
		Reader:             &NotificationServiceGetNotificationsReader{formats: a.formats},
		Context:            params.Context,
		Client:             params.HTTPClient,
	}
	for _, opt := range opts {
		opt(op)
	}

	result, err := a.transport.Submit(op)
	if err != nil {
		return nil, err
	}
	success, ok := result.(*NotificationServiceGetNotificationsOK)
	if ok {
		return success, nil
	}
	// unexpected success response
	unexpectedSuccess := result.(*NotificationServiceGetNotificationsDefault)
	return nil, runtime.NewAPIError("unexpected success response: content available as default response in error", unexpectedSuccess, unexpectedSuccess.Code())
}

/*
NotificationServiceMarkNotificationsAsRead marks a notification as read

This endpoint marks a notification as read
*/
func (a *Client) NotificationServiceMarkNotificationsAsRead(params *NotificationServiceMarkNotificationsAsReadParams, opts ...ClientOption) (*NotificationServiceMarkNotificationsAsReadOK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewNotificationServiceMarkNotificationsAsReadParams()
	}
	op := &runtime.ClientOperation{
		ID:                 "NotificationService_MarkNotificationsAsRead",
		Method:             "POST",
		PathPattern:        "/api/rest/v1/notifications/read",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http"},
		Params:             params,
		Reader:             &NotificationServiceMarkNotificationsAsReadReader{formats: a.formats},
		Context:            params.Context,
		Client:             params.HTTPClient,
	}
	for _, opt := range opts {
		opt(op)
	}

	result, err := a.transport.Submit(op)
	if err != nil {
		return nil, err
	}
	success, ok := result.(*NotificationServiceMarkNotificationsAsReadOK)
	if ok {
		return success, nil
	}
	// unexpected success response
	unexpectedSuccess := result.(*NotificationServiceMarkNotificationsAsReadDefault)
	return nil, runtime.NewAPIError("unexpected success response: content available as default response in error", unexpectedSuccess, unexpectedSuccess.Code())
}

// SetTransport changes the transport on the client
func (a *Client) SetTransport(transport runtime.ClientTransport) {
	a.transport = transport
}