// Code generated by go-swagger; DO NOT EDIT.

package models

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"context"

	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
)

// ChorusEnableTotpReply chorus enable totp reply
//
// swagger:model chorusEnableTotpReply
type ChorusEnableTotpReply struct {

	// result
	Result ChorusEnableTotpResult `json:"result,omitempty"`
}

// Validate validates this chorus enable totp reply
func (m *ChorusEnableTotpReply) Validate(formats strfmt.Registry) error {
	return nil
}

// ContextValidate validates this chorus enable totp reply based on context it is used
func (m *ChorusEnableTotpReply) ContextValidate(ctx context.Context, formats strfmt.Registry) error {
	return nil
}

// MarshalBinary interface implementation
func (m *ChorusEnableTotpReply) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *ChorusEnableTotpReply) UnmarshalBinary(b []byte) error {
	var res ChorusEnableTotpReply
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}