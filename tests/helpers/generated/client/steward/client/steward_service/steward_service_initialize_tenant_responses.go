// Code generated by go-swagger; DO NOT EDIT.

package steward_service

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/go-openapi/runtime"
	"github.com/go-openapi/strfmt"

	"github.com/CHORUS-TRE/chorus-backend/tests/helpers/generated/client/steward/models"
)

// StewardServiceInitializeTenantReader is a Reader for the StewardServiceInitializeTenant structure.
type StewardServiceInitializeTenantReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *StewardServiceInitializeTenantReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 200:
		result := NewStewardServiceInitializeTenantOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	default:
		result := NewStewardServiceInitializeTenantDefault(response.Code())
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		if response.Code()/100 == 2 {
			return result, nil
		}
		return nil, result
	}
}

// NewStewardServiceInitializeTenantOK creates a StewardServiceInitializeTenantOK with default headers values
func NewStewardServiceInitializeTenantOK() *StewardServiceInitializeTenantOK {
	return &StewardServiceInitializeTenantOK{}
}

/*
StewardServiceInitializeTenantOK describes a response with status code 200, with default header values.

A successful response.
*/
type StewardServiceInitializeTenantOK struct {
	Payload interface{}
}

// IsSuccess returns true when this steward service initialize tenant o k response has a 2xx status code
func (o *StewardServiceInitializeTenantOK) IsSuccess() bool {
	return true
}

// IsRedirect returns true when this steward service initialize tenant o k response has a 3xx status code
func (o *StewardServiceInitializeTenantOK) IsRedirect() bool {
	return false
}

// IsClientError returns true when this steward service initialize tenant o k response has a 4xx status code
func (o *StewardServiceInitializeTenantOK) IsClientError() bool {
	return false
}

// IsServerError returns true when this steward service initialize tenant o k response has a 5xx status code
func (o *StewardServiceInitializeTenantOK) IsServerError() bool {
	return false
}

// IsCode returns true when this steward service initialize tenant o k response a status code equal to that given
func (o *StewardServiceInitializeTenantOK) IsCode(code int) bool {
	return code == 200
}

// Code gets the status code for the steward service initialize tenant o k response
func (o *StewardServiceInitializeTenantOK) Code() int {
	return 200
}

func (o *StewardServiceInitializeTenantOK) Error() string {
	payload, _ := json.Marshal(o.Payload)
	return fmt.Sprintf("[POST /api/rest/v1/steward/tenants/initialize][%d] stewardServiceInitializeTenantOK %s", 200, payload)
}

func (o *StewardServiceInitializeTenantOK) String() string {
	payload, _ := json.Marshal(o.Payload)
	return fmt.Sprintf("[POST /api/rest/v1/steward/tenants/initialize][%d] stewardServiceInitializeTenantOK %s", 200, payload)
}

func (o *StewardServiceInitializeTenantOK) GetPayload() interface{} {
	return o.Payload
}

func (o *StewardServiceInitializeTenantOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	// response payload
	if err := consumer.Consume(response.Body(), &o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewStewardServiceInitializeTenantDefault creates a StewardServiceInitializeTenantDefault with default headers values
func NewStewardServiceInitializeTenantDefault(code int) *StewardServiceInitializeTenantDefault {
	return &StewardServiceInitializeTenantDefault{
		_statusCode: code,
	}
}

/*
StewardServiceInitializeTenantDefault describes a response with status code -1, with default header values.

An unexpected error response.
*/
type StewardServiceInitializeTenantDefault struct {
	_statusCode int

	Payload *models.RPCStatus
}

// IsSuccess returns true when this steward service initialize tenant default response has a 2xx status code
func (o *StewardServiceInitializeTenantDefault) IsSuccess() bool {
	return o._statusCode/100 == 2
}

// IsRedirect returns true when this steward service initialize tenant default response has a 3xx status code
func (o *StewardServiceInitializeTenantDefault) IsRedirect() bool {
	return o._statusCode/100 == 3
}

// IsClientError returns true when this steward service initialize tenant default response has a 4xx status code
func (o *StewardServiceInitializeTenantDefault) IsClientError() bool {
	return o._statusCode/100 == 4
}

// IsServerError returns true when this steward service initialize tenant default response has a 5xx status code
func (o *StewardServiceInitializeTenantDefault) IsServerError() bool {
	return o._statusCode/100 == 5
}

// IsCode returns true when this steward service initialize tenant default response a status code equal to that given
func (o *StewardServiceInitializeTenantDefault) IsCode(code int) bool {
	return o._statusCode == code
}

// Code gets the status code for the steward service initialize tenant default response
func (o *StewardServiceInitializeTenantDefault) Code() int {
	return o._statusCode
}

func (o *StewardServiceInitializeTenantDefault) Error() string {
	payload, _ := json.Marshal(o.Payload)
	return fmt.Sprintf("[POST /api/rest/v1/steward/tenants/initialize][%d] StewardService_InitializeTenant default %s", o._statusCode, payload)
}

func (o *StewardServiceInitializeTenantDefault) String() string {
	payload, _ := json.Marshal(o.Payload)
	return fmt.Sprintf("[POST /api/rest/v1/steward/tenants/initialize][%d] StewardService_InitializeTenant default %s", o._statusCode, payload)
}

func (o *StewardServiceInitializeTenantDefault) GetPayload() *models.RPCStatus {
	return o.Payload
}

func (o *StewardServiceInitializeTenantDefault) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.RPCStatus)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}