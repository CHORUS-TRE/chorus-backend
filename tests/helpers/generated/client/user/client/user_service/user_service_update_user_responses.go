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

// UserServiceUpdateUserReader is a Reader for the UserServiceUpdateUser structure.
type UserServiceUpdateUserReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *UserServiceUpdateUserReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 200:
		result := NewUserServiceUpdateUserOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	default:
		result := NewUserServiceUpdateUserDefault(response.Code())
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		if response.Code()/100 == 2 {
			return result, nil
		}
		return nil, result
	}
}

// NewUserServiceUpdateUserOK creates a UserServiceUpdateUserOK with default headers values
func NewUserServiceUpdateUserOK() *UserServiceUpdateUserOK {
	return &UserServiceUpdateUserOK{}
}

/*
UserServiceUpdateUserOK describes a response with status code 200, with default header values.

A successful response.
*/
type UserServiceUpdateUserOK struct {
	Payload *models.ChorusUpdateUserReply
}

// IsSuccess returns true when this user service update user o k response has a 2xx status code
func (o *UserServiceUpdateUserOK) IsSuccess() bool {
	return true
}

// IsRedirect returns true when this user service update user o k response has a 3xx status code
func (o *UserServiceUpdateUserOK) IsRedirect() bool {
	return false
}

// IsClientError returns true when this user service update user o k response has a 4xx status code
func (o *UserServiceUpdateUserOK) IsClientError() bool {
	return false
}

// IsServerError returns true when this user service update user o k response has a 5xx status code
func (o *UserServiceUpdateUserOK) IsServerError() bool {
	return false
}

// IsCode returns true when this user service update user o k response a status code equal to that given
func (o *UserServiceUpdateUserOK) IsCode(code int) bool {
	return code == 200
}

// Code gets the status code for the user service update user o k response
func (o *UserServiceUpdateUserOK) Code() int {
	return 200
}

func (o *UserServiceUpdateUserOK) Error() string {
	payload, _ := json.Marshal(o.Payload)
	return fmt.Sprintf("[PUT /api/rest/v1/users][%d] userServiceUpdateUserOK %s", 200, payload)
}

func (o *UserServiceUpdateUserOK) String() string {
	payload, _ := json.Marshal(o.Payload)
	return fmt.Sprintf("[PUT /api/rest/v1/users][%d] userServiceUpdateUserOK %s", 200, payload)
}

func (o *UserServiceUpdateUserOK) GetPayload() *models.ChorusUpdateUserReply {
	return o.Payload
}

func (o *UserServiceUpdateUserOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.ChorusUpdateUserReply)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewUserServiceUpdateUserDefault creates a UserServiceUpdateUserDefault with default headers values
func NewUserServiceUpdateUserDefault(code int) *UserServiceUpdateUserDefault {
	return &UserServiceUpdateUserDefault{
		_statusCode: code,
	}
}

/*
UserServiceUpdateUserDefault describes a response with status code -1, with default header values.

An unexpected error response.
*/
type UserServiceUpdateUserDefault struct {
	_statusCode int

	Payload *models.RPCStatus
}

// IsSuccess returns true when this user service update user default response has a 2xx status code
func (o *UserServiceUpdateUserDefault) IsSuccess() bool {
	return o._statusCode/100 == 2
}

// IsRedirect returns true when this user service update user default response has a 3xx status code
func (o *UserServiceUpdateUserDefault) IsRedirect() bool {
	return o._statusCode/100 == 3
}

// IsClientError returns true when this user service update user default response has a 4xx status code
func (o *UserServiceUpdateUserDefault) IsClientError() bool {
	return o._statusCode/100 == 4
}

// IsServerError returns true when this user service update user default response has a 5xx status code
func (o *UserServiceUpdateUserDefault) IsServerError() bool {
	return o._statusCode/100 == 5
}

// IsCode returns true when this user service update user default response a status code equal to that given
func (o *UserServiceUpdateUserDefault) IsCode(code int) bool {
	return o._statusCode == code
}

// Code gets the status code for the user service update user default response
func (o *UserServiceUpdateUserDefault) Code() int {
	return o._statusCode
}

func (o *UserServiceUpdateUserDefault) Error() string {
	payload, _ := json.Marshal(o.Payload)
	return fmt.Sprintf("[PUT /api/rest/v1/users][%d] UserService_UpdateUser default %s", o._statusCode, payload)
}

func (o *UserServiceUpdateUserDefault) String() string {
	payload, _ := json.Marshal(o.Payload)
	return fmt.Sprintf("[PUT /api/rest/v1/users][%d] UserService_UpdateUser default %s", o._statusCode, payload)
}

func (o *UserServiceUpdateUserDefault) GetPayload() *models.RPCStatus {
	return o.Payload
}

func (o *UserServiceUpdateUserDefault) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.RPCStatus)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}