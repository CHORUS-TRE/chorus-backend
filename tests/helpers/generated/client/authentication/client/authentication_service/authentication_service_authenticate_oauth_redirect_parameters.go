// Code generated by go-swagger; DO NOT EDIT.

package authentication_service

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
)

// NewAuthenticationServiceAuthenticateOauthRedirectParams creates a new AuthenticationServiceAuthenticateOauthRedirectParams object,
// with the default timeout for this client.
//
// Default values are not hydrated, since defaults are normally applied by the API server side.
//
// To enforce default values in parameter, use SetDefaults or WithDefaults.
func NewAuthenticationServiceAuthenticateOauthRedirectParams() *AuthenticationServiceAuthenticateOauthRedirectParams {
	return &AuthenticationServiceAuthenticateOauthRedirectParams{
		timeout: cr.DefaultTimeout,
	}
}

// NewAuthenticationServiceAuthenticateOauthRedirectParamsWithTimeout creates a new AuthenticationServiceAuthenticateOauthRedirectParams object
// with the ability to set a timeout on a request.
func NewAuthenticationServiceAuthenticateOauthRedirectParamsWithTimeout(timeout time.Duration) *AuthenticationServiceAuthenticateOauthRedirectParams {
	return &AuthenticationServiceAuthenticateOauthRedirectParams{
		timeout: timeout,
	}
}

// NewAuthenticationServiceAuthenticateOauthRedirectParamsWithContext creates a new AuthenticationServiceAuthenticateOauthRedirectParams object
// with the ability to set a context for a request.
func NewAuthenticationServiceAuthenticateOauthRedirectParamsWithContext(ctx context.Context) *AuthenticationServiceAuthenticateOauthRedirectParams {
	return &AuthenticationServiceAuthenticateOauthRedirectParams{
		Context: ctx,
	}
}

// NewAuthenticationServiceAuthenticateOauthRedirectParamsWithHTTPClient creates a new AuthenticationServiceAuthenticateOauthRedirectParams object
// with the ability to set a custom HTTPClient for a request.
func NewAuthenticationServiceAuthenticateOauthRedirectParamsWithHTTPClient(client *http.Client) *AuthenticationServiceAuthenticateOauthRedirectParams {
	return &AuthenticationServiceAuthenticateOauthRedirectParams{
		HTTPClient: client,
	}
}

/*
AuthenticationServiceAuthenticateOauthRedirectParams contains all the parameters to send to the API endpoint

	for the authentication service authenticate oauth redirect operation.

	Typically these are written to a http.Request.
*/
type AuthenticationServiceAuthenticateOauthRedirectParams struct {

	// Code.
	Code *string

	// ID.
	ID string

	// SessionState.
	SessionState *string

	// State.
	State *string

	timeout    time.Duration
	Context    context.Context
	HTTPClient *http.Client
}

// WithDefaults hydrates default values in the authentication service authenticate oauth redirect params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *AuthenticationServiceAuthenticateOauthRedirectParams) WithDefaults() *AuthenticationServiceAuthenticateOauthRedirectParams {
	o.SetDefaults()
	return o
}

// SetDefaults hydrates default values in the authentication service authenticate oauth redirect params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *AuthenticationServiceAuthenticateOauthRedirectParams) SetDefaults() {
	// no default values defined for this parameter
}

// WithTimeout adds the timeout to the authentication service authenticate oauth redirect params
func (o *AuthenticationServiceAuthenticateOauthRedirectParams) WithTimeout(timeout time.Duration) *AuthenticationServiceAuthenticateOauthRedirectParams {
	o.SetTimeout(timeout)
	return o
}

// SetTimeout adds the timeout to the authentication service authenticate oauth redirect params
func (o *AuthenticationServiceAuthenticateOauthRedirectParams) SetTimeout(timeout time.Duration) {
	o.timeout = timeout
}

// WithContext adds the context to the authentication service authenticate oauth redirect params
func (o *AuthenticationServiceAuthenticateOauthRedirectParams) WithContext(ctx context.Context) *AuthenticationServiceAuthenticateOauthRedirectParams {
	o.SetContext(ctx)
	return o
}

// SetContext adds the context to the authentication service authenticate oauth redirect params
func (o *AuthenticationServiceAuthenticateOauthRedirectParams) SetContext(ctx context.Context) {
	o.Context = ctx
}

// WithHTTPClient adds the HTTPClient to the authentication service authenticate oauth redirect params
func (o *AuthenticationServiceAuthenticateOauthRedirectParams) WithHTTPClient(client *http.Client) *AuthenticationServiceAuthenticateOauthRedirectParams {
	o.SetHTTPClient(client)
	return o
}

// SetHTTPClient adds the HTTPClient to the authentication service authenticate oauth redirect params
func (o *AuthenticationServiceAuthenticateOauthRedirectParams) SetHTTPClient(client *http.Client) {
	o.HTTPClient = client
}

// WithCode adds the code to the authentication service authenticate oauth redirect params
func (o *AuthenticationServiceAuthenticateOauthRedirectParams) WithCode(code *string) *AuthenticationServiceAuthenticateOauthRedirectParams {
	o.SetCode(code)
	return o
}

// SetCode adds the code to the authentication service authenticate oauth redirect params
func (o *AuthenticationServiceAuthenticateOauthRedirectParams) SetCode(code *string) {
	o.Code = code
}

// WithID adds the id to the authentication service authenticate oauth redirect params
func (o *AuthenticationServiceAuthenticateOauthRedirectParams) WithID(id string) *AuthenticationServiceAuthenticateOauthRedirectParams {
	o.SetID(id)
	return o
}

// SetID adds the id to the authentication service authenticate oauth redirect params
func (o *AuthenticationServiceAuthenticateOauthRedirectParams) SetID(id string) {
	o.ID = id
}

// WithSessionState adds the sessionState to the authentication service authenticate oauth redirect params
func (o *AuthenticationServiceAuthenticateOauthRedirectParams) WithSessionState(sessionState *string) *AuthenticationServiceAuthenticateOauthRedirectParams {
	o.SetSessionState(sessionState)
	return o
}

// SetSessionState adds the sessionState to the authentication service authenticate oauth redirect params
func (o *AuthenticationServiceAuthenticateOauthRedirectParams) SetSessionState(sessionState *string) {
	o.SessionState = sessionState
}

// WithState adds the state to the authentication service authenticate oauth redirect params
func (o *AuthenticationServiceAuthenticateOauthRedirectParams) WithState(state *string) *AuthenticationServiceAuthenticateOauthRedirectParams {
	o.SetState(state)
	return o
}

// SetState adds the state to the authentication service authenticate oauth redirect params
func (o *AuthenticationServiceAuthenticateOauthRedirectParams) SetState(state *string) {
	o.State = state
}

// WriteToRequest writes these params to a swagger request
func (o *AuthenticationServiceAuthenticateOauthRedirectParams) WriteToRequest(r runtime.ClientRequest, reg strfmt.Registry) error {

	if err := r.SetTimeout(o.timeout); err != nil {
		return err
	}
	var res []error

	if o.Code != nil {

		// query param code
		var qrCode string

		if o.Code != nil {
			qrCode = *o.Code
		}
		qCode := qrCode
		if qCode != "" {

			if err := r.SetQueryParam("code", qCode); err != nil {
				return err
			}
		}
	}

	// path param id
	if err := r.SetPathParam("id", o.ID); err != nil {
		return err
	}

	if o.SessionState != nil {

		// query param sessionState
		var qrSessionState string

		if o.SessionState != nil {
			qrSessionState = *o.SessionState
		}
		qSessionState := qrSessionState
		if qSessionState != "" {

			if err := r.SetQueryParam("sessionState", qSessionState); err != nil {
				return err
			}
		}
	}

	if o.State != nil {

		// query param state
		var qrState string

		if o.State != nil {
			qrState = *o.State
		}
		qState := qrState
		if qState != "" {

			if err := r.SetQueryParam("state", qState); err != nil {
				return err
			}
		}
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}
