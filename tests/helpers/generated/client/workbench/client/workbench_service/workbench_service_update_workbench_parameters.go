// Code generated by go-swagger; DO NOT EDIT.

package workbench_service

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

	"github.com/CHORUS-TRE/chorus-backend/tests/helpers/generated/client/workbench/models"
)

// NewWorkbenchServiceUpdateWorkbenchParams creates a new WorkbenchServiceUpdateWorkbenchParams object,
// with the default timeout for this client.
//
// Default values are not hydrated, since defaults are normally applied by the API server side.
//
// To enforce default values in parameter, use SetDefaults or WithDefaults.
func NewWorkbenchServiceUpdateWorkbenchParams() *WorkbenchServiceUpdateWorkbenchParams {
	return &WorkbenchServiceUpdateWorkbenchParams{
		timeout: cr.DefaultTimeout,
	}
}

// NewWorkbenchServiceUpdateWorkbenchParamsWithTimeout creates a new WorkbenchServiceUpdateWorkbenchParams object
// with the ability to set a timeout on a request.
func NewWorkbenchServiceUpdateWorkbenchParamsWithTimeout(timeout time.Duration) *WorkbenchServiceUpdateWorkbenchParams {
	return &WorkbenchServiceUpdateWorkbenchParams{
		timeout: timeout,
	}
}

// NewWorkbenchServiceUpdateWorkbenchParamsWithContext creates a new WorkbenchServiceUpdateWorkbenchParams object
// with the ability to set a context for a request.
func NewWorkbenchServiceUpdateWorkbenchParamsWithContext(ctx context.Context) *WorkbenchServiceUpdateWorkbenchParams {
	return &WorkbenchServiceUpdateWorkbenchParams{
		Context: ctx,
	}
}

// NewWorkbenchServiceUpdateWorkbenchParamsWithHTTPClient creates a new WorkbenchServiceUpdateWorkbenchParams object
// with the ability to set a custom HTTPClient for a request.
func NewWorkbenchServiceUpdateWorkbenchParamsWithHTTPClient(client *http.Client) *WorkbenchServiceUpdateWorkbenchParams {
	return &WorkbenchServiceUpdateWorkbenchParams{
		HTTPClient: client,
	}
}

/*
WorkbenchServiceUpdateWorkbenchParams contains all the parameters to send to the API endpoint

	for the workbench service update workbench operation.

	Typically these are written to a http.Request.
*/
type WorkbenchServiceUpdateWorkbenchParams struct {

	// Body.
	Body *models.ChorusUpdateWorkbenchRequest

	timeout    time.Duration
	Context    context.Context
	HTTPClient *http.Client
}

// WithDefaults hydrates default values in the workbench service update workbench params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *WorkbenchServiceUpdateWorkbenchParams) WithDefaults() *WorkbenchServiceUpdateWorkbenchParams {
	o.SetDefaults()
	return o
}

// SetDefaults hydrates default values in the workbench service update workbench params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *WorkbenchServiceUpdateWorkbenchParams) SetDefaults() {
	// no default values defined for this parameter
}

// WithTimeout adds the timeout to the workbench service update workbench params
func (o *WorkbenchServiceUpdateWorkbenchParams) WithTimeout(timeout time.Duration) *WorkbenchServiceUpdateWorkbenchParams {
	o.SetTimeout(timeout)
	return o
}

// SetTimeout adds the timeout to the workbench service update workbench params
func (o *WorkbenchServiceUpdateWorkbenchParams) SetTimeout(timeout time.Duration) {
	o.timeout = timeout
}

// WithContext adds the context to the workbench service update workbench params
func (o *WorkbenchServiceUpdateWorkbenchParams) WithContext(ctx context.Context) *WorkbenchServiceUpdateWorkbenchParams {
	o.SetContext(ctx)
	return o
}

// SetContext adds the context to the workbench service update workbench params
func (o *WorkbenchServiceUpdateWorkbenchParams) SetContext(ctx context.Context) {
	o.Context = ctx
}

// WithHTTPClient adds the HTTPClient to the workbench service update workbench params
func (o *WorkbenchServiceUpdateWorkbenchParams) WithHTTPClient(client *http.Client) *WorkbenchServiceUpdateWorkbenchParams {
	o.SetHTTPClient(client)
	return o
}

// SetHTTPClient adds the HTTPClient to the workbench service update workbench params
func (o *WorkbenchServiceUpdateWorkbenchParams) SetHTTPClient(client *http.Client) {
	o.HTTPClient = client
}

// WithBody adds the body to the workbench service update workbench params
func (o *WorkbenchServiceUpdateWorkbenchParams) WithBody(body *models.ChorusUpdateWorkbenchRequest) *WorkbenchServiceUpdateWorkbenchParams {
	o.SetBody(body)
	return o
}

// SetBody adds the body to the workbench service update workbench params
func (o *WorkbenchServiceUpdateWorkbenchParams) SetBody(body *models.ChorusUpdateWorkbenchRequest) {
	o.Body = body
}

// WriteToRequest writes these params to a swagger request
func (o *WorkbenchServiceUpdateWorkbenchParams) WriteToRequest(r runtime.ClientRequest, reg strfmt.Registry) error {

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
