// Code generated by go-swagger; DO NOT EDIT.

package index

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/go-openapi/runtime"
	"github.com/go-openapi/strfmt"

	"github.com/CHORUS-TRE/chorus-backend/tests/helpers/generated/client/index/models"
)

// IndexServiceGetHelloReader is a Reader for the IndexServiceGetHello structure.
type IndexServiceGetHelloReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *IndexServiceGetHelloReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 200:
		result := NewIndexServiceGetHelloOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	default:
		result := NewIndexServiceGetHelloDefault(response.Code())
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		if response.Code()/100 == 2 {
			return result, nil
		}
		return nil, result
	}
}

// NewIndexServiceGetHelloOK creates a IndexServiceGetHelloOK with default headers values
func NewIndexServiceGetHelloOK() *IndexServiceGetHelloOK {
	return &IndexServiceGetHelloOK{}
}

/*
IndexServiceGetHelloOK describes a response with status code 200, with default header values.

A successful response.
*/
type IndexServiceGetHelloOK struct {
	Payload *models.ChorusGetHelloReply
}

// IsSuccess returns true when this index service get hello o k response has a 2xx status code
func (o *IndexServiceGetHelloOK) IsSuccess() bool {
	return true
}

// IsRedirect returns true when this index service get hello o k response has a 3xx status code
func (o *IndexServiceGetHelloOK) IsRedirect() bool {
	return false
}

// IsClientError returns true when this index service get hello o k response has a 4xx status code
func (o *IndexServiceGetHelloOK) IsClientError() bool {
	return false
}

// IsServerError returns true when this index service get hello o k response has a 5xx status code
func (o *IndexServiceGetHelloOK) IsServerError() bool {
	return false
}

// IsCode returns true when this index service get hello o k response a status code equal to that given
func (o *IndexServiceGetHelloOK) IsCode(code int) bool {
	return code == 200
}

// Code gets the status code for the index service get hello o k response
func (o *IndexServiceGetHelloOK) Code() int {
	return 200
}

func (o *IndexServiceGetHelloOK) Error() string {
	payload, _ := json.Marshal(o.Payload)
	return fmt.Sprintf("[GET /api/v1/hello][%d] indexServiceGetHelloOK %s", 200, payload)
}

func (o *IndexServiceGetHelloOK) String() string {
	payload, _ := json.Marshal(o.Payload)
	return fmt.Sprintf("[GET /api/v1/hello][%d] indexServiceGetHelloOK %s", 200, payload)
}

func (o *IndexServiceGetHelloOK) GetPayload() *models.ChorusGetHelloReply {
	return o.Payload
}

func (o *IndexServiceGetHelloOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.ChorusGetHelloReply)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewIndexServiceGetHelloDefault creates a IndexServiceGetHelloDefault with default headers values
func NewIndexServiceGetHelloDefault(code int) *IndexServiceGetHelloDefault {
	return &IndexServiceGetHelloDefault{
		_statusCode: code,
	}
}

/*
IndexServiceGetHelloDefault describes a response with status code -1, with default header values.

An unexpected error response.
*/
type IndexServiceGetHelloDefault struct {
	_statusCode int

	Payload *models.RPCStatus
}

// IsSuccess returns true when this index service get hello default response has a 2xx status code
func (o *IndexServiceGetHelloDefault) IsSuccess() bool {
	return o._statusCode/100 == 2
}

// IsRedirect returns true when this index service get hello default response has a 3xx status code
func (o *IndexServiceGetHelloDefault) IsRedirect() bool {
	return o._statusCode/100 == 3
}

// IsClientError returns true when this index service get hello default response has a 4xx status code
func (o *IndexServiceGetHelloDefault) IsClientError() bool {
	return o._statusCode/100 == 4
}

// IsServerError returns true when this index service get hello default response has a 5xx status code
func (o *IndexServiceGetHelloDefault) IsServerError() bool {
	return o._statusCode/100 == 5
}

// IsCode returns true when this index service get hello default response a status code equal to that given
func (o *IndexServiceGetHelloDefault) IsCode(code int) bool {
	return o._statusCode == code
}

// Code gets the status code for the index service get hello default response
func (o *IndexServiceGetHelloDefault) Code() int {
	return o._statusCode
}

func (o *IndexServiceGetHelloDefault) Error() string {
	payload, _ := json.Marshal(o.Payload)
	return fmt.Sprintf("[GET /api/v1/hello][%d] IndexService_GetHello default %s", o._statusCode, payload)
}

func (o *IndexServiceGetHelloDefault) String() string {
	payload, _ := json.Marshal(o.Payload)
	return fmt.Sprintf("[GET /api/v1/hello][%d] IndexService_GetHello default %s", o._statusCode, payload)
}

func (o *IndexServiceGetHelloDefault) GetPayload() *models.RPCStatus {
	return o.Payload
}

func (o *IndexServiceGetHelloDefault) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.RPCStatus)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}