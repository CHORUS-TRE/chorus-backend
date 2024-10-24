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

// UserServiceGetUserMeReader is a Reader for the UserServiceGetUserMe structure.
type UserServiceGetUserMeReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *UserServiceGetUserMeReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 200:
		result := NewUserServiceGetUserMeOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	default:
		result := NewUserServiceGetUserMeDefault(response.Code())
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		if response.Code()/100 == 2 {
			return result, nil
		}
		return nil, result
	}
}

// NewUserServiceGetUserMeOK creates a UserServiceGetUserMeOK with default headers values
func NewUserServiceGetUserMeOK() *UserServiceGetUserMeOK {
	return &UserServiceGetUserMeOK{}
}

/*
UserServiceGetUserMeOK describes a response with status code 200, with default header values.

A successful response.
*/
type UserServiceGetUserMeOK struct {
	Payload *models.ChorusGetUserMeReply
}

// IsSuccess returns true when this user service get user me o k response has a 2xx status code
func (o *UserServiceGetUserMeOK) IsSuccess() bool {
	return true
}

// IsRedirect returns true when this user service get user me o k response has a 3xx status code
func (o *UserServiceGetUserMeOK) IsRedirect() bool {
	return false
}

// IsClientError returns true when this user service get user me o k response has a 4xx status code
func (o *UserServiceGetUserMeOK) IsClientError() bool {
	return false
}

// IsServerError returns true when this user service get user me o k response has a 5xx status code
func (o *UserServiceGetUserMeOK) IsServerError() bool {
	return false
}

// IsCode returns true when this user service get user me o k response a status code equal to that given
func (o *UserServiceGetUserMeOK) IsCode(code int) bool {
	return code == 200
}

// Code gets the status code for the user service get user me o k response
func (o *UserServiceGetUserMeOK) Code() int {
	return 200
}

func (o *UserServiceGetUserMeOK) Error() string {
	payload, _ := json.Marshal(o.Payload)
	return fmt.Sprintf("[GET /api/rest/v1/users/me][%d] userServiceGetUserMeOK %s", 200, payload)
}

func (o *UserServiceGetUserMeOK) String() string {
	payload, _ := json.Marshal(o.Payload)
	return fmt.Sprintf("[GET /api/rest/v1/users/me][%d] userServiceGetUserMeOK %s", 200, payload)
}

func (o *UserServiceGetUserMeOK) GetPayload() *models.ChorusGetUserMeReply {
	return o.Payload
}

func (o *UserServiceGetUserMeOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.ChorusGetUserMeReply)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewUserServiceGetUserMeDefault creates a UserServiceGetUserMeDefault with default headers values
func NewUserServiceGetUserMeDefault(code int) *UserServiceGetUserMeDefault {
	return &UserServiceGetUserMeDefault{
		_statusCode: code,
	}
}

/*
UserServiceGetUserMeDefault describes a response with status code -1, with default header values.

An unexpected error response.
*/
type UserServiceGetUserMeDefault struct {
	_statusCode int

	Payload *models.RPCStatus
}

// IsSuccess returns true when this user service get user me default response has a 2xx status code
func (o *UserServiceGetUserMeDefault) IsSuccess() bool {
	return o._statusCode/100 == 2
}

// IsRedirect returns true when this user service get user me default response has a 3xx status code
func (o *UserServiceGetUserMeDefault) IsRedirect() bool {
	return o._statusCode/100 == 3
}

// IsClientError returns true when this user service get user me default response has a 4xx status code
func (o *UserServiceGetUserMeDefault) IsClientError() bool {
	return o._statusCode/100 == 4
}

// IsServerError returns true when this user service get user me default response has a 5xx status code
func (o *UserServiceGetUserMeDefault) IsServerError() bool {
	return o._statusCode/100 == 5
}

// IsCode returns true when this user service get user me default response a status code equal to that given
func (o *UserServiceGetUserMeDefault) IsCode(code int) bool {
	return o._statusCode == code
}

// Code gets the status code for the user service get user me default response
func (o *UserServiceGetUserMeDefault) Code() int {
	return o._statusCode
}

func (o *UserServiceGetUserMeDefault) Error() string {
	payload, _ := json.Marshal(o.Payload)
	return fmt.Sprintf("[GET /api/rest/v1/users/me][%d] UserService_GetUserMe default %s", o._statusCode, payload)
}

func (o *UserServiceGetUserMeDefault) String() string {
	payload, _ := json.Marshal(o.Payload)
	return fmt.Sprintf("[GET /api/rest/v1/users/me][%d] UserService_GetUserMe default %s", o._statusCode, payload)
}

func (o *UserServiceGetUserMeDefault) GetPayload() *models.RPCStatus {
	return o.Payload
}

func (o *UserServiceGetUserMeDefault) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.RPCStatus)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}
