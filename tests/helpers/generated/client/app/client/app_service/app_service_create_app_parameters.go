// Code generated by go-swagger; DO NOT EDIT.

package app_service

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

	"github.com/CHORUS-TRE/chorus-backend/tests/helpers/generated/client/app/models"
)

// NewAppServiceCreateAppParams creates a new AppServiceCreateAppParams object,
// with the default timeout for this client.
//
// Default values are not hydrated, since defaults are normally applied by the API server side.
//
// To enforce default values in parameter, use SetDefaults or WithDefaults.
func NewAppServiceCreateAppParams() *AppServiceCreateAppParams {
	return &AppServiceCreateAppParams{
		timeout: cr.DefaultTimeout,
	}
}

// NewAppServiceCreateAppParamsWithTimeout creates a new AppServiceCreateAppParams object
// with the ability to set a timeout on a request.
func NewAppServiceCreateAppParamsWithTimeout(timeout time.Duration) *AppServiceCreateAppParams {
	return &AppServiceCreateAppParams{
		timeout: timeout,
	}
}

// NewAppServiceCreateAppParamsWithContext creates a new AppServiceCreateAppParams object
// with the ability to set a context for a request.
func NewAppServiceCreateAppParamsWithContext(ctx context.Context) *AppServiceCreateAppParams {
	return &AppServiceCreateAppParams{
		Context: ctx,
	}
}

// NewAppServiceCreateAppParamsWithHTTPClient creates a new AppServiceCreateAppParams object
// with the ability to set a custom HTTPClient for a request.
func NewAppServiceCreateAppParamsWithHTTPClient(client *http.Client) *AppServiceCreateAppParams {
	return &AppServiceCreateAppParams{
		HTTPClient: client,
	}
}

/*
AppServiceCreateAppParams contains all the parameters to send to the API endpoint

	for the app service create app operation.

	Typically these are written to a http.Request.
*/
type AppServiceCreateAppParams struct {

	// Body.
	Body *models.ChorusApp

	timeout    time.Duration
	Context    context.Context
	HTTPClient *http.Client
}

// WithDefaults hydrates default values in the app service create app params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *AppServiceCreateAppParams) WithDefaults() *AppServiceCreateAppParams {
	o.SetDefaults()
	return o
}

// SetDefaults hydrates default values in the app service create app params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *AppServiceCreateAppParams) SetDefaults() {
	// no default values defined for this parameter
}

// WithTimeout adds the timeout to the app service create app params
func (o *AppServiceCreateAppParams) WithTimeout(timeout time.Duration) *AppServiceCreateAppParams {
	o.SetTimeout(timeout)
	return o
}

// SetTimeout adds the timeout to the app service create app params
func (o *AppServiceCreateAppParams) SetTimeout(timeout time.Duration) {
	o.timeout = timeout
}

// WithContext adds the context to the app service create app params
func (o *AppServiceCreateAppParams) WithContext(ctx context.Context) *AppServiceCreateAppParams {
	o.SetContext(ctx)
	return o
}

// SetContext adds the context to the app service create app params
func (o *AppServiceCreateAppParams) SetContext(ctx context.Context) {
	o.Context = ctx
}

// WithHTTPClient adds the HTTPClient to the app service create app params
func (o *AppServiceCreateAppParams) WithHTTPClient(client *http.Client) *AppServiceCreateAppParams {
	o.SetHTTPClient(client)
	return o
}

// SetHTTPClient adds the HTTPClient to the app service create app params
func (o *AppServiceCreateAppParams) SetHTTPClient(client *http.Client) {
	o.HTTPClient = client
}

// WithBody adds the body to the app service create app params
func (o *AppServiceCreateAppParams) WithBody(body *models.ChorusApp) *AppServiceCreateAppParams {
	o.SetBody(body)
	return o
}

// SetBody adds the body to the app service create app params
func (o *AppServiceCreateAppParams) SetBody(body *models.ChorusApp) {
	o.Body = body
}

// WriteToRequest writes these params to a swagger request
func (o *AppServiceCreateAppParams) WriteToRequest(r runtime.ClientRequest, reg strfmt.Registry) error {

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