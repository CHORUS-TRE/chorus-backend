// Code generated by go-swagger; DO NOT EDIT.

package user_service

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/go-openapi/runtime"
	"github.com/go-openapi/strfmt"

	"github.com/CHORUS-TRE/chorus-backend/tests/helpers/generated/client/user/models"
)

// UserServiceEnableTotpReader is a Reader for the UserServiceEnableTotp structure.
type UserServiceEnableTotpReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *UserServiceEnableTotpReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 200:
		result := NewUserServiceEnableTotpOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	default:
		result := NewUserServiceEnableTotpDefault(response.Code())
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		if response.Code()/100 == 2 {
			return result, nil
		}
		return nil, result
	}
}

// NewUserServiceEnableTotpOK creates a UserServiceEnableTotpOK with default headers values
func NewUserServiceEnableTotpOK() *UserServiceEnableTotpOK {
	return &UserServiceEnableTotpOK{}
}

/*
UserServiceEnableTotpOK describes a response with status code 200, with default header values.

A successful response.
*/
type UserServiceEnableTotpOK struct {
	Payload *models.ChorusEnableTotpReply
}

// IsSuccess returns true when this user service enable totp o k response has a 2xx status code
func (o *UserServiceEnableTotpOK) IsSuccess() bool {
	return true
}

// IsRedirect returns true when this user service enable totp o k response has a 3xx status code
func (o *UserServiceEnableTotpOK) IsRedirect() bool {
	return false
}

// IsClientError returns true when this user service enable totp o k response has a 4xx status code
func (o *UserServiceEnableTotpOK) IsClientError() bool {
	return false
}

// IsServerError returns true when this user service enable totp o k response has a 5xx status code
func (o *UserServiceEnableTotpOK) IsServerError() bool {
	return false
}

// IsCode returns true when this user service enable totp o k response a status code equal to that given
func (o *UserServiceEnableTotpOK) IsCode(code int) bool {
	return code == 200
}

// Code gets the status code for the user service enable totp o k response
func (o *UserServiceEnableTotpOK) Code() int {
	return 200
}

func (o *UserServiceEnableTotpOK) Error() string {
	payload, _ := json.Marshal(o.Payload)
	return fmt.Sprintf("[POST /api/rest/v1/users/me/totp/enable][%d] userServiceEnableTotpOK %s", 200, payload)
}

func (o *UserServiceEnableTotpOK) String() string {
	payload, _ := json.Marshal(o.Payload)
	return fmt.Sprintf("[POST /api/rest/v1/users/me/totp/enable][%d] userServiceEnableTotpOK %s", 200, payload)
}

func (o *UserServiceEnableTotpOK) GetPayload() *models.ChorusEnableTotpReply {
	return o.Payload
}

func (o *UserServiceEnableTotpOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.ChorusEnableTotpReply)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewUserServiceEnableTotpDefault creates a UserServiceEnableTotpDefault with default headers values
func NewUserServiceEnableTotpDefault(code int) *UserServiceEnableTotpDefault {
	return &UserServiceEnableTotpDefault{
		_statusCode: code,
	}
}

/*
UserServiceEnableTotpDefault describes a response with status code -1, with default header values.

An unexpected error response.
*/
type UserServiceEnableTotpDefault struct {
	_statusCode int

	Payload *models.RPCStatus
}

// IsSuccess returns true when this user service enable totp default response has a 2xx status code
func (o *UserServiceEnableTotpDefault) IsSuccess() bool {
	return o._statusCode/100 == 2
}

// IsRedirect returns true when this user service enable totp default response has a 3xx status code
func (o *UserServiceEnableTotpDefault) IsRedirect() bool {
	return o._statusCode/100 == 3
}

// IsClientError returns true when this user service enable totp default response has a 4xx status code
func (o *UserServiceEnableTotpDefault) IsClientError() bool {
	return o._statusCode/100 == 4
}

// IsServerError returns true when this user service enable totp default response has a 5xx status code
func (o *UserServiceEnableTotpDefault) IsServerError() bool {
	return o._statusCode/100 == 5
}

// IsCode returns true when this user service enable totp default response a status code equal to that given
func (o *UserServiceEnableTotpDefault) IsCode(code int) bool {
	return o._statusCode == code
}

// Code gets the status code for the user service enable totp default response
func (o *UserServiceEnableTotpDefault) Code() int {
	return o._statusCode
}

func (o *UserServiceEnableTotpDefault) Error() string {
	payload, _ := json.Marshal(o.Payload)
	return fmt.Sprintf("[POST /api/rest/v1/users/me/totp/enable][%d] UserService_EnableTotp default %s", o._statusCode, payload)
}

func (o *UserServiceEnableTotpDefault) String() string {
	payload, _ := json.Marshal(o.Payload)
	return fmt.Sprintf("[POST /api/rest/v1/users/me/totp/enable][%d] UserService_EnableTotp default %s", o._statusCode, payload)
}

func (o *UserServiceEnableTotpDefault) GetPayload() *models.RPCStatus {
	return o.Payload
}

func (o *UserServiceEnableTotpDefault) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.RPCStatus)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}