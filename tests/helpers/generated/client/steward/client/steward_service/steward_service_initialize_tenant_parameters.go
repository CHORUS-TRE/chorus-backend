// Code generated by go-swagger; DO NOT EDIT.

package steward_service

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

	"github.com/CHORUS-TRE/chorus-backend/tests/helpers/generated/client/steward/models"
)

// NewStewardServiceInitializeTenantParams creates a new StewardServiceInitializeTenantParams object,
// with the default timeout for this client.
//
// Default values are not hydrated, since defaults are normally applied by the API server side.
//
// To enforce default values in parameter, use SetDefaults or WithDefaults.
func NewStewardServiceInitializeTenantParams() *StewardServiceInitializeTenantParams {
	return &StewardServiceInitializeTenantParams{
		timeout: cr.DefaultTimeout,
	}
}

// NewStewardServiceInitializeTenantParamsWithTimeout creates a new StewardServiceInitializeTenantParams object
// with the ability to set a timeout on a request.
func NewStewardServiceInitializeTenantParamsWithTimeout(timeout time.Duration) *StewardServiceInitializeTenantParams {
	return &StewardServiceInitializeTenantParams{
		timeout: timeout,
	}
}

// NewStewardServiceInitializeTenantParamsWithContext creates a new StewardServiceInitializeTenantParams object
// with the ability to set a context for a request.
func NewStewardServiceInitializeTenantParamsWithContext(ctx context.Context) *StewardServiceInitializeTenantParams {
	return &StewardServiceInitializeTenantParams{
		Context: ctx,
	}
}

// NewStewardServiceInitializeTenantParamsWithHTTPClient creates a new StewardServiceInitializeTenantParams object
// with the ability to set a custom HTTPClient for a request.
func NewStewardServiceInitializeTenantParamsWithHTTPClient(client *http.Client) *StewardServiceInitializeTenantParams {
	return &StewardServiceInitializeTenantParams{
		HTTPClient: client,
	}
}

/*
StewardServiceInitializeTenantParams contains all the parameters to send to the API endpoint

	for the steward service initialize tenant operation.

	Typically these are written to a http.Request.
*/
type StewardServiceInitializeTenantParams struct {

	// Body.
	Body *models.ChorusInitializeTenantRequest

	timeout    time.Duration
	Context    context.Context
	HTTPClient *http.Client
}

// WithDefaults hydrates default values in the steward service initialize tenant params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *StewardServiceInitializeTenantParams) WithDefaults() *StewardServiceInitializeTenantParams {
	o.SetDefaults()
	return o
}

// SetDefaults hydrates default values in the steward service initialize tenant params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *StewardServiceInitializeTenantParams) SetDefaults() {
	// no default values defined for this parameter
}

// WithTimeout adds the timeout to the steward service initialize tenant params
func (o *StewardServiceInitializeTenantParams) WithTimeout(timeout time.Duration) *StewardServiceInitializeTenantParams {
	o.SetTimeout(timeout)
	return o
}

// SetTimeout adds the timeout to the steward service initialize tenant params
func (o *StewardServiceInitializeTenantParams) SetTimeout(timeout time.Duration) {
	o.timeout = timeout
}

// WithContext adds the context to the steward service initialize tenant params
func (o *StewardServiceInitializeTenantParams) WithContext(ctx context.Context) *StewardServiceInitializeTenantParams {
	o.SetContext(ctx)
	return o
}

// SetContext adds the context to the steward service initialize tenant params
func (o *StewardServiceInitializeTenantParams) SetContext(ctx context.Context) {
	o.Context = ctx
}

// WithHTTPClient adds the HTTPClient to the steward service initialize tenant params
func (o *StewardServiceInitializeTenantParams) WithHTTPClient(client *http.Client) *StewardServiceInitializeTenantParams {
	o.SetHTTPClient(client)
	return o
}

// SetHTTPClient adds the HTTPClient to the steward service initialize tenant params
func (o *StewardServiceInitializeTenantParams) SetHTTPClient(client *http.Client) {
	o.HTTPClient = client
}

// WithBody adds the body to the steward service initialize tenant params
func (o *StewardServiceInitializeTenantParams) WithBody(body *models.ChorusInitializeTenantRequest) *StewardServiceInitializeTenantParams {
	o.SetBody(body)
	return o
}

// SetBody adds the body to the steward service initialize tenant params
func (o *StewardServiceInitializeTenantParams) SetBody(body *models.ChorusInitializeTenantRequest) {
	o.Body = body
}

// WriteToRequest writes these params to a swagger request
func (o *StewardServiceInitializeTenantParams) WriteToRequest(r runtime.ClientRequest, reg strfmt.Registry) error {

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
