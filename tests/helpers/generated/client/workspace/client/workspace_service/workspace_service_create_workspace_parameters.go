// Code generated by go-swagger; DO NOT EDIT.

package workspace_service

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

	"github.com/CHORUS-TRE/chorus-backend/tests/helpers/generated/client/workspace/models"
)

// NewWorkspaceServiceCreateWorkspaceParams creates a new WorkspaceServiceCreateWorkspaceParams object,
// with the default timeout for this client.
//
// Default values are not hydrated, since defaults are normally applied by the API server side.
//
// To enforce default values in parameter, use SetDefaults or WithDefaults.
func NewWorkspaceServiceCreateWorkspaceParams() *WorkspaceServiceCreateWorkspaceParams {
	return &WorkspaceServiceCreateWorkspaceParams{
		timeout: cr.DefaultTimeout,
	}
}

// NewWorkspaceServiceCreateWorkspaceParamsWithTimeout creates a new WorkspaceServiceCreateWorkspaceParams object
// with the ability to set a timeout on a request.
func NewWorkspaceServiceCreateWorkspaceParamsWithTimeout(timeout time.Duration) *WorkspaceServiceCreateWorkspaceParams {
	return &WorkspaceServiceCreateWorkspaceParams{
		timeout: timeout,
	}
}

// NewWorkspaceServiceCreateWorkspaceParamsWithContext creates a new WorkspaceServiceCreateWorkspaceParams object
// with the ability to set a context for a request.
func NewWorkspaceServiceCreateWorkspaceParamsWithContext(ctx context.Context) *WorkspaceServiceCreateWorkspaceParams {
	return &WorkspaceServiceCreateWorkspaceParams{
		Context: ctx,
	}
}

// NewWorkspaceServiceCreateWorkspaceParamsWithHTTPClient creates a new WorkspaceServiceCreateWorkspaceParams object
// with the ability to set a custom HTTPClient for a request.
func NewWorkspaceServiceCreateWorkspaceParamsWithHTTPClient(client *http.Client) *WorkspaceServiceCreateWorkspaceParams {
	return &WorkspaceServiceCreateWorkspaceParams{
		HTTPClient: client,
	}
}

/*
WorkspaceServiceCreateWorkspaceParams contains all the parameters to send to the API endpoint

	for the workspace service create workspace operation.

	Typically these are written to a http.Request.
*/
type WorkspaceServiceCreateWorkspaceParams struct {

	// Body.
	Body *models.ChorusWorkspace

	timeout    time.Duration
	Context    context.Context
	HTTPClient *http.Client
}

// WithDefaults hydrates default values in the workspace service create workspace params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *WorkspaceServiceCreateWorkspaceParams) WithDefaults() *WorkspaceServiceCreateWorkspaceParams {
	o.SetDefaults()
	return o
}

// SetDefaults hydrates default values in the workspace service create workspace params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *WorkspaceServiceCreateWorkspaceParams) SetDefaults() {
	// no default values defined for this parameter
}

// WithTimeout adds the timeout to the workspace service create workspace params
func (o *WorkspaceServiceCreateWorkspaceParams) WithTimeout(timeout time.Duration) *WorkspaceServiceCreateWorkspaceParams {
	o.SetTimeout(timeout)
	return o
}

// SetTimeout adds the timeout to the workspace service create workspace params
func (o *WorkspaceServiceCreateWorkspaceParams) SetTimeout(timeout time.Duration) {
	o.timeout = timeout
}

// WithContext adds the context to the workspace service create workspace params
func (o *WorkspaceServiceCreateWorkspaceParams) WithContext(ctx context.Context) *WorkspaceServiceCreateWorkspaceParams {
	o.SetContext(ctx)
	return o
}

// SetContext adds the context to the workspace service create workspace params
func (o *WorkspaceServiceCreateWorkspaceParams) SetContext(ctx context.Context) {
	o.Context = ctx
}

// WithHTTPClient adds the HTTPClient to the workspace service create workspace params
func (o *WorkspaceServiceCreateWorkspaceParams) WithHTTPClient(client *http.Client) *WorkspaceServiceCreateWorkspaceParams {
	o.SetHTTPClient(client)
	return o
}

// SetHTTPClient adds the HTTPClient to the workspace service create workspace params
func (o *WorkspaceServiceCreateWorkspaceParams) SetHTTPClient(client *http.Client) {
	o.HTTPClient = client
}

// WithBody adds the body to the workspace service create workspace params
func (o *WorkspaceServiceCreateWorkspaceParams) WithBody(body *models.ChorusWorkspace) *WorkspaceServiceCreateWorkspaceParams {
	o.SetBody(body)
	return o
}

// SetBody adds the body to the workspace service create workspace params
func (o *WorkspaceServiceCreateWorkspaceParams) SetBody(body *models.ChorusWorkspace) {
	o.Body = body
}

// WriteToRequest writes these params to a swagger request
func (o *WorkspaceServiceCreateWorkspaceParams) WriteToRequest(r runtime.ClientRequest, reg strfmt.Registry) error {

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
