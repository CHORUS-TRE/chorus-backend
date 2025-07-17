package openapi

import (
	"context"
	"net/http"
	"net/http/httputil"
	"strings"

	"github.com/CHORUS-TRE/chorus-backend/internal/logger"

	"github.com/go-openapi/runtime"
	"github.com/go-openapi/strfmt"
	"go.uber.org/zap"
)

type ClientTransport struct {
	Host      string
	BasePath  string
	Consumer  runtime.Consumer
	Producers map[string]runtime.Producer
	schemes   []string
	Formats   strfmt.Registry
	client    *http.Client
	logger    *logger.ContextLogger
}

// This client transport enables client code to extract eventual non handled errors in the body of the response
// With the default transport implementation the only way to parse the body was in debug mode
// dumping the request and response messages in stderr
func NewNopCloserClientTransport(host, basePath string, schemes []string, logger *logger.ContextLogger) *ClientTransport {

	var tr ClientTransport

	tr.Host = host
	tr.BasePath = basePath
	tr.Consumer = runtime.JSONConsumer()
	tr.Producers = map[string]runtime.Producer{
		runtime.JSONMime: runtime.JSONProducer(),
	}
	if !strings.HasPrefix(tr.BasePath, "/") {
		tr.BasePath = "/" + tr.BasePath
	}

	if len(schemes) > 0 {
		tr.schemes = schemes
	}
	tr.client = &http.Client{}
	tr.logger = logger

	return &tr
}

// Submit a request and when there is a body on success it will turn that into the result
// all other things are turned into an api error for openapi which retains the status code
func (r *ClientTransport) Submit(operation *runtime.ClientOperation) (interface{}, error) {
	params, readResponse, auth := operation.Params, operation.Reader, operation.AuthInfo
	ctx := operation.Context
	if ctx == nil {
		ctx = context.Background()
	}

	request := newRequest(operation.Method, operation.PathPattern, params)

	var accept []string
	accept = append(accept, operation.ProducesMediaTypes...)
	if err := request.SetHeaderParam(runtime.HeaderAccept, accept...); err != nil {
		return nil, err
	}

	req, err := request.buildHTTP(runtime.JSONMime, r.BasePath, r.Producers, r.Formats, auth)
	if err != nil {
		return nil, err
	}
	req.URL.Scheme = r.pickScheme(operation.Schemes)
	req.URL.Host = r.Host
	req.Host = r.Host

	b, err := httputil.DumpRequestOut(req, true)
	if err != nil {
		return nil, err
	}
	r.logger.Debug(ctx, "request sent", zap.String("httpRequest", string(b)))

	ctx, cancel := context.WithTimeout(ctx, request.timeout)
	defer cancel()

	client := operation.Client
	if client == nil {
		client = r.client
	}
	req = req.WithContext(ctx)
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	b, err = httputil.DumpResponse(res, true)
	if err != nil {
		return nil, err
	}
	r.logger.Debug(ctx, "response received", zap.String("httpRequest", string(b)))

	return readResponse.ReadResponse(response{res}, r.Consumer)
}

func (r *ClientTransport) pickScheme(schemes []string) string {
	if v := r.selectScheme(r.schemes); v != "" {
		return v
	}
	if v := r.selectScheme(schemes); v != "" {
		return v
	}
	return "http"
}

func (r *ClientTransport) selectScheme(schemes []string) string {
	schLen := len(schemes)
	if schLen == 0 {
		return ""
	}

	scheme := schemes[0]
	// prefer https, but skip when not possible
	if scheme != "https" && schLen > 1 {
		for _, sch := range schemes {
			if sch == "https" {
				scheme = sch
				break
			}
		}
	}
	return scheme
}
