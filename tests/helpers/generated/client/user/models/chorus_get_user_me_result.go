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

// ChorusGetUserMeResult Get User (me)
//
// swagger:model chorusGetUserMeResult
type ChorusGetUserMeResult struct {

	// me
	Me *ChorusUser `json:"me,omitempty"`
}

// Validate validates this chorus get user me result
func (m *ChorusGetUserMeResult) Validate(formats strfmt.Registry) error {
	var res []error

	if err := m.validateMe(formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *ChorusGetUserMeResult) validateMe(formats strfmt.Registry) error {
	if swag.IsZero(m.Me) { // not required
		return nil
	}

	if m.Me != nil {
		if err := m.Me.Validate(formats); err != nil {
			if ve, ok := err.(*errors.Validation); ok {
				return ve.ValidateName("me")
			} else if ce, ok := err.(*errors.CompositeError); ok {
				return ce.ValidateName("me")
			}
			return err
		}
	}

	return nil
}

// ContextValidate validate this chorus get user me result based on the context it is used
func (m *ChorusGetUserMeResult) ContextValidate(ctx context.Context, formats strfmt.Registry) error {
	var res []error

	if err := m.contextValidateMe(ctx, formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *ChorusGetUserMeResult) contextValidateMe(ctx context.Context, formats strfmt.Registry) error {

	if m.Me != nil {

		if swag.IsZero(m.Me) { // not required
			return nil
		}

		if err := m.Me.ContextValidate(ctx, formats); err != nil {
			if ve, ok := err.(*errors.Validation); ok {
				return ve.ValidateName("me")
			} else if ce, ok := err.(*errors.CompositeError); ok {
				return ce.ValidateName("me")
			}
			return err
		}
	}

	return nil
}

// MarshalBinary interface implementation
func (m *ChorusGetUserMeResult) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *ChorusGetUserMeResult) UnmarshalBinary(b []byte) error {
	var res ChorusGetUserMeResult
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}