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

// NewWorkbenchServiceCreateWorkbenchParams creates a new WorkbenchServiceCreateWorkbenchParams object,
// with the default timeout for this client.
//
// Default values are not hydrated, since defaults are normally applied by the API server side.
//
// To enforce default values in parameter, use SetDefaults or WithDefaults.
func NewWorkbenchServiceCreateWorkbenchParams() *WorkbenchServiceCreateWorkbenchParams {
	return &WorkbenchServiceCreateWorkbenchParams{
		timeout: cr.DefaultTimeout,
	}
}

// NewWorkbenchServiceCreateWorkbenchParamsWithTimeout creates a new WorkbenchServiceCreateWorkbenchParams object
// with the ability to set a timeout on a request.
func NewWorkbenchServiceCreateWorkbenchParamsWithTimeout(timeout time.Duration) *WorkbenchServiceCreateWorkbenchParams {
	return &WorkbenchServiceCreateWorkbenchParams{
		timeout: timeout,
	}
}

// NewWorkbenchServiceCreateWorkbenchParamsWithContext creates a new WorkbenchServiceCreateWorkbenchParams object
// with the ability to set a context for a request.
func NewWorkbenchServiceCreateWorkbenchParamsWithContext(ctx context.Context) *WorkbenchServiceCreateWorkbenchParams {
	return &WorkbenchServiceCreateWorkbenchParams{
		Context: ctx,
	}
}

// NewWorkbenchServiceCreateWorkbenchParamsWithHTTPClient creates a new WorkbenchServiceCreateWorkbenchParams object
// with the ability to set a custom HTTPClient for a request.
func NewWorkbenchServiceCreateWorkbenchParamsWithHTTPClient(client *http.Client) *WorkbenchServiceCreateWorkbenchParams {
	return &WorkbenchServiceCreateWorkbenchParams{
		HTTPClient: client,
	}
}

/*
WorkbenchServiceCreateWorkbenchParams contains all the parameters to send to the API endpoint

	for the workbench service create workbench operation.

	Typically these are written to a http.Request.
*/
type WorkbenchServiceCreateWorkbenchParams struct {

	// Body.
	Body *models.ChorusWorkbench

	timeout    time.Duration
	Context    context.Context
	HTTPClient *http.Client
}

// WithDefaults hydrates default values in the workbench service create workbench params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *WorkbenchServiceCreateWorkbenchParams) WithDefaults() *WorkbenchServiceCreateWorkbenchParams {
	o.SetDefaults()
	return o
}

// SetDefaults hydrates default values in the workbench service create workbench params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *WorkbenchServiceCreateWorkbenchParams) SetDefaults() {
	// no default values defined for this parameter
}

// WithTimeout adds the timeout to the workbench service create workbench params
func (o *WorkbenchServiceCreateWorkbenchParams) WithTimeout(timeout time.Duration) *WorkbenchServiceCreateWorkbenchParams {
	o.SetTimeout(timeout)
	return o
}

// SetTimeout adds the timeout to the workbench service create workbench params
func (o *WorkbenchServiceCreateWorkbenchParams) SetTimeout(timeout time.Duration) {
	o.timeout = timeout
}

// WithContext adds the context to the workbench service create workbench params
func (o *WorkbenchServiceCreateWorkbenchParams) WithContext(ctx context.Context) *WorkbenchServiceCreateWorkbenchParams {
	o.SetContext(ctx)
	return o
}

// SetContext adds the context to the workbench service create workbench params
func (o *WorkbenchServiceCreateWorkbenchParams) SetContext(ctx context.Context) {
	o.Context = ctx
}

// WithHTTPClient adds the HTTPClient to the workbench service create workbench params
func (o *WorkbenchServiceCreateWorkbenchParams) WithHTTPClient(client *http.Client) *WorkbenchServiceCreateWorkbenchParams {
	o.SetHTTPClient(client)
	return o
}

// SetHTTPClient adds the HTTPClient to the workbench service create workbench params
func (o *WorkbenchServiceCreateWorkbenchParams) SetHTTPClient(client *http.Client) {
	o.HTTPClient = client
}

// WithBody adds the body to the workbench service create workbench params
func (o *WorkbenchServiceCreateWorkbenchParams) WithBody(body *models.ChorusWorkbench) *WorkbenchServiceCreateWorkbenchParams {
	o.SetBody(body)
	return o
}

// SetBody adds the body to the workbench service create workbench params
func (o *WorkbenchServiceCreateWorkbenchParams) SetBody(body *models.ChorusWorkbench) {
	o.Body = body
}

// WriteToRequest writes these params to a swagger request
func (o *WorkbenchServiceCreateWorkbenchParams) WriteToRequest(r runtime.ClientRequest, reg strfmt.Registry) error {

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