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

// ChorusUpdateAppRequest chorus update app request
//
// swagger:model chorusUpdateAppRequest
type ChorusUpdateAppRequest struct {

	// app
	App *ChorusApp `json:"app,omitempty"`
}

// Validate validates this chorus update app request
func (m *ChorusUpdateAppRequest) Validate(formats strfmt.Registry) error {
	var res []error

	if err := m.validateApp(formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *ChorusUpdateAppRequest) validateApp(formats strfmt.Registry) error {
	if swag.IsZero(m.App) { // not required
		return nil
	}

	if m.App != nil {
		if err := m.App.Validate(formats); err != nil {
			if ve, ok := err.(*errors.Validation); ok {
				return ve.ValidateName("app")
			} else if ce, ok := err.(*errors.CompositeError); ok {
				return ce.ValidateName("app")
			}
			return err
		}
	}

	return nil
}

// ContextValidate validate this chorus update app request based on the context it is used
func (m *ChorusUpdateAppRequest) ContextValidate(ctx context.Context, formats strfmt.Registry) error {
	var res []error

	if err := m.contextValidateApp(ctx, formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *ChorusUpdateAppRequest) contextValidateApp(ctx context.Context, formats strfmt.Registry) error {

	if m.App != nil {

		if swag.IsZero(m.App) { // not required
			return nil
		}

		if err := m.App.ContextValidate(ctx, formats); err != nil {
			if ve, ok := err.(*errors.Validation); ok {
				return ve.ValidateName("app")
			} else if ce, ok := err.(*errors.CompositeError); ok {
				return ce.ValidateName("app")
			}
			return err
		}
	}

	return nil
}

// MarshalBinary interface implementation
func (m *ChorusUpdateAppRequest) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *ChorusUpdateAppRequest) UnmarshalBinary(b []byte) error {
	var res ChorusUpdateAppRequest
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}