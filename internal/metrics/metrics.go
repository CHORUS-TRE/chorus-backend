package metrics

import (
	"context"
	"strings"
	"time"

	prom "github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

//------------------
// Custom ServerMetrics (replace histogram with summary)
// (Based on from grpc-ecosystem/go-grpc-prometheus)
//------------------

var (
	allCodes = []codes.Code{
		codes.OK, codes.Canceled, codes.Unknown, codes.InvalidArgument, codes.DeadlineExceeded, codes.NotFound,
		codes.AlreadyExists, codes.PermissionDenied, codes.Unauthenticated, codes.ResourceExhausted,
		codes.FailedPrecondition, codes.Aborted, codes.OutOfRange, codes.Unimplemented, codes.Internal,
		codes.Unavailable, codes.DataLoss,
	}
)

func NewMetricsUnaryServerInterceptors(s *ServerMetrics) []grpc.UnaryServerInterceptor {
	return []grpc.UnaryServerInterceptor{s.UnaryServerInterceptor()}
}

func NewMetricsStreamServerInterceptors(s *ServerMetrics) []grpc.StreamServerInterceptor {
	return []grpc.StreamServerInterceptor{s.StreamServerInterceptor()}
}

type grpcType string

const (
	Unary        grpcType = "unary"
	ClientStream grpcType = "client_stream"
	ServerStream grpcType = "server_stream"
	BidiStream   grpcType = "bidi_stream"
)

// ServerMetrics represents a collection of metrics to be registered on a
// Prometheus metrics registry for a gRPC server.
type ServerMetrics struct {
	serverStartedCounter    *prom.CounterVec
	serverHandledCounter    *prom.CounterVec
	serverStreamMsgReceived *prom.CounterVec
	serverStreamMsgSent     *prom.CounterVec
	serverHandledSummary    *prom.SummaryVec
}

// NewServerMetrics returns a ServerMetrics object. Use a new instance of
// ServerMetrics when not using the default Prometheus metrics registry, for
// example when wanting to control which metrics are added to a registry as
// opposed to automatically adding metrics via init functions.
func NewServerMetrics() *ServerMetrics {
	s := &ServerMetrics{
		serverStartedCounter: prom.NewCounterVec(
			prom.CounterOpts{
				Name: "grpc_server_started_total",
				Help: "Total number of RPCs started on the server.",
			}, []string{"grpc_type", "grpc_service", "grpc_method"}),
		serverHandledCounter: prom.NewCounterVec(
			prom.CounterOpts{
				Name: "grpc_server_handled_total",
				Help: "Total number of RPCs completed on the server, regardless of success or failure.",
			}, []string{"grpc_type", "grpc_service", "grpc_method", "grpc_code"}),
		serverStreamMsgReceived: prom.NewCounterVec(
			prom.CounterOpts{
				Name: "grpc_server_msg_received_total",
				Help: "Total number of RPC stream messages received on the server.",
			}, []string{"grpc_type", "grpc_service", "grpc_method"}),
		serverStreamMsgSent: prom.NewCounterVec(
			prom.CounterOpts{
				Name: "grpc_server_msg_sent_total",
				Help: "Total number of gRPC stream messages sent by the server.",
			}, []string{"grpc_type", "grpc_service", "grpc_method"}),
		serverHandledSummary: prom.NewSummaryVec(prom.SummaryOpts{
			Name:       "grpc_server_handling_seconds",
			Help:       "Summary of response latency (seconds) of gRPC that had been application-level handled by the server.",
			Objectives: map[float64]float64{0.5: 0.05, 0.99: 0.001, 0.999: 0.0001},
		}, []string{"grpc_type", "grpc_service", "grpc_method"}),
	}

	prom.MustRegister(s.serverStartedCounter)
	prom.MustRegister(s.serverHandledCounter)
	prom.MustRegister(s.serverStreamMsgReceived)
	prom.MustRegister(s.serverStreamMsgSent)
	prom.MustRegister(s.serverHandledSummary)

	return s
}

// Describe sends the super-set of all possible descriptors of metrics
// collected by this Collector to the provided channel and returns once
// the last descriptor has been sent.
func (m *ServerMetrics) Describe(ch chan<- *prom.Desc) {
	m.serverStartedCounter.Describe(ch)
	m.serverHandledCounter.Describe(ch)
	m.serverStreamMsgReceived.Describe(ch)
	m.serverStreamMsgSent.Describe(ch)
	m.serverHandledSummary.Describe(ch)
}

// Collect is called by the Prometheus registry when collecting
// metrics. The implementation sends each collected metric via the
// provided channel and returns once the last metric has been sent.
func (m *ServerMetrics) Collect(ch chan<- prom.Metric) {
	m.serverStartedCounter.Collect(ch)
	m.serverHandledCounter.Collect(ch)
	m.serverStreamMsgReceived.Collect(ch)
	m.serverStreamMsgSent.Collect(ch)
	m.serverHandledSummary.Collect(ch)
}

// unaryServerInterceptor is a gRPC server-side interceptor that provides Prometheus monitoring for Unary RPCs.
func (m *ServerMetrics) UnaryServerInterceptor() func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		monitor := newServerReporter(m, Unary, info.FullMethod)
		monitor.ReceivedMessage()
		resp, err := handler(ctx, req)
		st, _ := status.FromError(err)
		monitor.Handled(st.Code())
		if err == nil {
			monitor.SentMessage()
		}
		return resp, err
	}
}

// streamServerInterceptor is a gRPC server-side interceptor that provides Prometheus monitoring for Streaming RPCs.
func (m *ServerMetrics) StreamServerInterceptor() func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		monitor := newServerReporter(m, streamRPCType(info), info.FullMethod)
		err := handler(srv, &monitoredServerStream{ss, monitor})
		st, _ := status.FromError(err)
		monitor.Handled(st.Code())
		return err
	}
}

// InitializeMetrics initializes all metrics, with their appropriate null
// value, for all gRPC methods registered on a gRPC server. This is useful, to
// ensure that all metrics exist when collecting and querying.
func (m *ServerMetrics) InitializeMetrics(server *grpc.Server) {
	serviceInfo := server.GetServiceInfo()
	for serviceName, info := range serviceInfo {
		for _, mInfo := range info.Methods {
			preRegisterMethod(m, serviceName, &mInfo)
		}
	}
}

func streamRPCType(info *grpc.StreamServerInfo) grpcType {
	if info.IsClientStream && !info.IsServerStream {
		return ClientStream
	} else if !info.IsClientStream && info.IsServerStream {
		return ServerStream
	}
	return BidiStream
}

// monitoredStream wraps grpc.ServerStream allowing each Sent/Recv of message to increment counters.
type monitoredServerStream struct {
	grpc.ServerStream
	monitor *serverReporter
}

func (s *monitoredServerStream) SendMsg(m interface{}) error {
	err := s.ServerStream.SendMsg(m)
	if err == nil {
		s.monitor.SentMessage()
	}
	return err
}

func (s *monitoredServerStream) RecvMsg(m interface{}) error {
	err := s.ServerStream.RecvMsg(m)
	if err == nil {
		s.monitor.ReceivedMessage()
	}
	return err
}

type serverReporter struct {
	metrics     *ServerMetrics
	rpcType     grpcType
	serviceName string
	methodName  string
	startTime   time.Time
}

func newServerReporter(m *ServerMetrics, rpcType grpcType, fullMethod string) *serverReporter {
	r := &serverReporter{
		metrics: m,
		rpcType: rpcType,
	}
	r.startTime = time.Now()
	r.serviceName, r.methodName = splitMethodName(fullMethod)
	r.metrics.serverStartedCounter.WithLabelValues(string(r.rpcType), r.serviceName, r.methodName).Inc()
	return r
}

func splitMethodName(fullMethodName string) (string, string) {
	fullMethodName = strings.TrimPrefix(fullMethodName, "/") // remove leading slash
	if i := strings.Index(fullMethodName, "/"); i >= 0 {
		return fullMethodName[:i], fullMethodName[i+1:]
	}
	return "unknown", "unknown"
}

func (r *serverReporter) ReceivedMessage() {
	r.metrics.serverStreamMsgReceived.WithLabelValues(string(r.rpcType), r.serviceName, r.methodName).Inc()
}

func (r *serverReporter) SentMessage() {
	r.metrics.serverStreamMsgSent.WithLabelValues(string(r.rpcType), r.serviceName, r.methodName).Inc()
}

func (r *serverReporter) Handled(code codes.Code) {
	r.metrics.serverHandledCounter.WithLabelValues(string(r.rpcType), r.serviceName, r.methodName, code.String()).Inc()
	r.metrics.serverHandledSummary.WithLabelValues(string(r.rpcType), r.serviceName, r.methodName).Observe(time.Since(r.startTime).Seconds())
}

// preRegisterMethod is invoked on Register of a Server, allowing all gRPC services labels to be pre-populated.
func preRegisterMethod(metrics *ServerMetrics, serviceName string, mInfo *grpc.MethodInfo) {
	methodName := mInfo.Name
	methodType := string(typeFromMethodInfo(mInfo))
	// These are just references (no increments), as just referencing will create the labels but not set values.
	_, _ = metrics.serverStartedCounter.GetMetricWithLabelValues(methodType, serviceName, methodName)
	_, _ = metrics.serverStreamMsgReceived.GetMetricWithLabelValues(methodType, serviceName, methodName)
	_, _ = metrics.serverStreamMsgSent.GetMetricWithLabelValues(methodType, serviceName, methodName)

	for _, code := range allCodes {
		_, _ = metrics.serverHandledCounter.GetMetricWithLabelValues(methodType, serviceName, methodName, code.String())
	}
}

func typeFromMethodInfo(mInfo *grpc.MethodInfo) grpcType {
	if !mInfo.IsClientStream && !mInfo.IsServerStream {
		return Unary
	}
	if mInfo.IsClientStream && !mInfo.IsServerStream {
		return ClientStream
	}
	if !mInfo.IsClientStream && mInfo.IsServerStream {
		return ServerStream
	}
	return BidiStream
}
