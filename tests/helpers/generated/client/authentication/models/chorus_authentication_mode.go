// Code generated by go-swagger; DO NOT EDIT.

package models

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"context"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
)

// ChorusAuthenticationMode chorus authentication mode
//
// swagger:model chorusAuthenticationMode
type ChorusAuthenticationMode struct {

	// internal
	Internal *ChorusInternal `json:"internal,omitempty"`

	// openid
	Openid *ChorusOpenID `json:"openid,omitempty"`

	// type
	Type string `json:"type,omitempty"`
}

// Validate validates this chorus authentication mode
func (m *ChorusAuthenticationMode) Validate(formats strfmt.Registry) error {
	var res []error

	if err := m.validateInternal(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateOpenid(formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *ChorusAuthenticationMode) validateInternal(formats strfmt.Registry) error {
	if swag.IsZero(m.Internal) { // not required
		return nil
	}

	if m.Internal != nil {
		if err := m.Internal.Validate(formats); err != nil {
			if ve, ok := err.(*errors.Validation); ok {
				return ve.ValidateName("internal")
			} else if ce, ok := err.(*errors.CompositeError); ok {
				return ce.ValidateName("internal")
			}
			return err
		}
	}

	return nil
}

func (m *ChorusAuthenticationMode) validateOpenid(formats strfmt.Registry) error {
	if swag.IsZero(m.Openid) { // not required
		return nil
	}

	if m.Openid != nil {
		if err := m.Openid.Validate(formats); err != nil {
			if ve, ok := err.(*errors.Validation); ok {
				return ve.ValidateName("openid")
			} else if ce, ok := err.(*errors.CompositeError); ok {
				return ce.ValidateName("openid")
			}
			return err
		}
	}

	return nil
}

// ContextValidate validate this chorus authentication mode based on the context it is used
func (m *ChorusAuthenticationMode) ContextValidate(ctx context.Context, formats strfmt.Registry) error {
	var res []error

	if err := m.contextValidateInternal(ctx, formats); err != nil {
		res = append(res, err)
	}

	if err := m.contextValidateOpenid(ctx, formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *ChorusAuthenticationMode) contextValidateInternal(ctx context.Context, formats strfmt.Registry) error {

	if m.Internal != nil {

		if swag.IsZero(m.Internal) { // not required
			return nil
		}

		if err := m.Internal.ContextValidate(ctx, formats); err != nil {
			if ve, ok := err.(*errors.Validation); ok {
				return ve.ValidateName("internal")
			} else if ce, ok := err.(*errors.CompositeError); ok {
				return ce.ValidateName("internal")
			}
			return err
		}
	}

	return nil
}

func (m *ChorusAuthenticationMode) contextValidateOpenid(ctx context.Context, formats strfmt.Registry) error {

	if m.Openid != nil {

		if swag.IsZero(m.Openid) { // not required
			return nil
		}

		if err := m.Openid.ContextValidate(ctx, formats); err != nil {
			if ve, ok := err.(*errors.Validation); ok {
				return ve.ValidateName("openid")
			} else if ce, ok := err.(*errors.CompositeError); ok {
				return ce.ValidateName("openid")
			}
			return err
		}
	}

	return nil
}

// MarshalBinary interface implementation
func (m *ChorusAuthenticationMode) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *ChorusAuthenticationMode) UnmarshalBinary(b []byte) error {
	var res ChorusAuthenticationMode
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}