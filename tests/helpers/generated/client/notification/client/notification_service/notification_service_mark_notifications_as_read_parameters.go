// Code generated by go-swagger; DO NOT EDIT.

package notification_service

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"context"
	"net/http"
	"time"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime"
	cr "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"

	"github.com/CHORUS-TRE/chorus-backend/tests/helpers/generated/client/notification/models"
)

// NewNotificationServiceMarkNotificationsAsReadParams creates a new NotificationServiceMarkNotificationsAsReadParams object,
// with the default timeout for this client.
//
// Default values are not hydrated, since defaults are normally applied by the API server side.
//
// To enforce default values in parameter, use SetDefaults or WithDefaults.
func NewNotificationServiceMarkNotificationsAsReadParams() *NotificationServiceMarkNotificationsAsReadParams {
	return &NotificationServiceMarkNotificationsAsReadParams{
		timeout: cr.DefaultTimeout,
	}
}

// NewNotificationServiceMarkNotificationsAsReadParamsWithTimeout creates a new NotificationServiceMarkNotificationsAsReadParams object
// with the ability to set a timeout on a request.
func NewNotificationServiceMarkNotificationsAsReadParamsWithTimeout(timeout time.Duration) *NotificationServiceMarkNotificationsAsReadParams {
	return &NotificationServiceMarkNotificationsAsReadParams{
		timeout: timeout,
	}
}

// NewNotificationServiceMarkNotificationsAsReadParamsWithContext creates a new NotificationServiceMarkNotificationsAsReadParams object
// with the ability to set a context for a request.
func NewNotificationServiceMarkNotificationsAsReadParamsWithContext(ctx context.Context) *NotificationServiceMarkNotificationsAsReadParams {
	return &NotificationServiceMarkNotificationsAsReadParams{
		Context: ctx,
	}
}

// NewNotificationServiceMarkNotificationsAsReadParamsWithHTTPClient creates a new NotificationServiceMarkNotificationsAsReadParams object
// with the ability to set a custom HTTPClient for a request.
func NewNotificationServiceMarkNotificationsAsReadParamsWithHTTPClient(client *http.Client) *NotificationServiceMarkNotificationsAsReadParams {
	return &NotificationServiceMarkNotificationsAsReadParams{
		HTTPClient: client,
	}
}

/*
NotificationServiceMarkNotificationsAsReadParams contains all the parameters to send to the API endpoint

	for the notification service mark notifications as read operation.

	Typically these are written to a http.Request.
*/
type NotificationServiceMarkNotificationsAsReadParams struct {

	// Body.
	Body *models.ChorusMarkNotificationsAsReadRequest

	timeout    time.Duration
	Context    context.Context
	HTTPClient *http.Client
}

// WithDefaults hydrates default values in the notification service mark notifications as read params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *NotificationServiceMarkNotificationsAsReadParams) WithDefaults() *NotificationServiceMarkNotificationsAsReadParams {
	o.SetDefaults()
	return o
}

// SetDefaults hydrates default values in the notification service mark notifications as read params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *NotificationServiceMarkNotificationsAsReadParams) SetDefaults() {
	// no default values defined for this parameter
}

// WithTimeout adds the timeout to the notification service mark notifications as read params
func (o *NotificationServiceMarkNotificationsAsReadParams) WithTimeout(timeout time.Duration) *NotificationServiceMarkNotificationsAsReadParams {
	o.SetTimeout(timeout)
	return o
}

// SetTimeout adds the timeout to the notification service mark notifications as read params
func (o *NotificationServiceMarkNotificationsAsReadParams) SetTimeout(timeout time.Duration) {
	o.timeout = timeout
}

// WithContext adds the context to the notification service mark notifications as read params
func (o *NotificationServiceMarkNotificationsAsReadParams) WithContext(ctx context.Context) *NotificationServiceMarkNotificationsAsReadParams {
	o.SetContext(ctx)
	return o
}

// SetContext adds the context to the notification service mark notifications as read params
func (o *NotificationServiceMarkNotificationsAsReadParams) SetContext(ctx context.Context) {
	o.Context = ctx
}

// WithHTTPClient adds the HTTPClient to the notification service mark notifications as read params
func (o *NotificationServiceMarkNotificationsAsReadParams) WithHTTPClient(client *http.Client) *NotificationServiceMarkNotificationsAsReadParams {
	o.SetHTTPClient(client)
	return o
}

// SetHTTPClient adds the HTTPClient to the notification service mark notifications as read params
func (o *NotificationServiceMarkNotificationsAsReadParams) SetHTTPClient(client *http.Client) {
	o.HTTPClient = client
}

// WithBody adds the body to the notification service mark notifications as read params
func (o *NotificationServiceMarkNotificationsAsReadParams) WithBody(body *models.ChorusMarkNotificationsAsReadRequest) *NotificationServiceMarkNotificationsAsReadParams {
	o.SetBody(body)
	return o
}

// SetBody adds the body to the notification service mark notifications as read params
func (o *NotificationServiceMarkNotificationsAsReadParams) SetBody(body *models.ChorusMarkNotificationsAsReadRequest) {
	o.Body = body
}

// WriteToRequest writes these params to a swagger request
func (o *NotificationServiceMarkNotificationsAsReadParams) WriteToRequest(r runtime.ClientRequest, reg strfmt.Registry) error {

	if err := r.SetTimeout(o.timeout); err != nil {
		return err
	}
	var res []error
	if o.Body != nil {
		if err := r.SetBodyParam(o.Body); err != nil {
			return err
		}
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}
