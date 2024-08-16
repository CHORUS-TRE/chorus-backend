package logger

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	jack "gopkg.in/natefinch/lumberjack.v2"

	"github.com/CHORUS-TRE/chorus-backend/internal/component"
	"github.com/CHORUS-TRE/chorus-backend/internal/config"
)

// Identification strings of each logger category. These are used to
// fetch the corresponding settings from the configuration structure
// and to identify each logging output string.
const (
	TechnicalCategory = "technical"
	BusinessCategory  = "business"
	SecurityCategory  = "security"

	categoryKey           = "category" // Category key field in the logging output.
	runtimeEnvironmentKey = "runtime_environment"
)

// All loggers are globally accessible objects.
var (
	TechLog *ContextLogger
	BizLog  *ContextLogger
	SecLog  *ContextLogger

	onceInit sync.Once
)

type options struct {
	componentName    string
	componentID      string
	componentVersion string
	signalCh         chan<- os.Signal
}

// Option is used to pass options to the Logger.
type Option func(*options)

// WithComponentName is the option used to set the component name field in the logger.
func WithComponentName(name string) Option {
	return func(o *options) {
		o.componentName = name
	}
}

// WithComponentID is the option used to set the component ID field in the logger.
func WithComponentID(id string) Option {
	return func(o *options) {
		o.componentID = id
	}
}

// WithComponentVersion is the option used to set the component version field in the logger.
func WithComponentVersion(version string) Option {
	return func(o *options) {
		o.componentVersion = version
	}
}

func WithSignal(c chan<- os.Signal) Option {
	return func(o *options) {
		o.signalCh = c
	}
}

func (o *options) appendToZapFields(zapFields ...zap.Field) []zap.Field {
	fields := []zap.Field{}
	fields = append(fields, zapFields...)

	if o.componentName != "" {
		fields = append(fields, zap.String("cmp_name", o.componentName))
	}
	if o.componentID != "" {
		fields = append(fields, zap.String("cmp_id", o.componentID))
	}
	if o.componentVersion != "" {
		fields = append(fields, zap.String("cmp_version", o.componentVersion))
	}
	if component.RuntimeEnvironment != "" {
		fields = append(fields, zap.String(runtimeEnvironmentKey, component.RuntimeEnvironment))
	}

	return fields
}

type Stoppable interface {
	Stop() error
}

var stopCores = func() { /* uninitialized */ }
var stdout zapcore.WriteSyncer

// InitLoggers configurates and instanstiates the global logger objects
// with parameters specified in the configuration structure. The
// library 'go.uber.org/zap' is used as the underlying engine.
func InitLoggers(cfg config.Config, opts ...Option) (func(), error) {
	var err error

	onceInit.Do(func() {
		o := &options{}
		// Apply options.
		for _, opt := range opts {
			opt(o)
		}

		stdout = zapcore.Lock(os.Stdout)

		encoderConfig := getEncoderConfig()
		jsonEncoder := zapcore.NewJSONEncoder(encoderConfig)

		technicalCores, err := initLoggerCores(TechnicalCategory, &cfg, jsonEncoder, stdout, o.signalCh)
		if err != nil {
			return
		}
		if len(technicalCores) == 0 {
			fmt.Println("No technical logger could be configured. You may want to fix that...")
		}
		techLog := zap.New(
			zapcore.NewTee(technicalCores...),
			zap.AddCaller(),
			zap.Fields(o.appendToZapFields(zap.String(categoryKey, TechnicalCategory))...))
		TechLog = NewContextLogger(techLog)
		zap.RedirectStdLog(TechLog.Logger)
		zap.RedirectStdLog(TechLog.loggerCallerSkip)

		businessCores, err := initLoggerCores(BusinessCategory, &cfg, jsonEncoder, stdout, o.signalCh)
		if err != nil {
			return
		}
		if len(businessCores) == 0 {
			fmt.Println("No business logger could be configured. You may want to fix that...")
		}
		bizLog := zap.New(
			zapcore.NewTee(businessCores...),
			zap.AddCaller(),
			zap.Fields(o.appendToZapFields(zap.String(categoryKey, BusinessCategory))...))
		BizLog = NewContextLogger(bizLog)

		securityCores, err := initLoggerCores(SecurityCategory, &cfg, jsonEncoder, stdout, o.signalCh)
		if err != nil {
			return
		}
		if len(securityCores) == 0 {
			fmt.Println("No security logger could be configured. You may want to fix that...")
		}
		secLog := zap.New(
			zapcore.NewTee(securityCores...),
			zap.AddCaller(),
			zap.Fields(o.appendToZapFields(zap.String(categoryKey, SecurityCategory))...))
		SecLog = NewContextLogger(secLog)

		stopCores = func() {
			doStopCores(technicalCores)
			doStopCores(businessCores)
			doStopCores(securityCores)
		}
	})
	return stopCores, err
}

func doStopCores(cores []zapcore.Core) {
	for _, core := range cores {
		if stoppable, ok := core.(Stoppable); ok {
			_ = stoppable.Stop()
		}
	}
}

// General settings for all the loggers.
func getEncoderConfig() zapcore.EncoderConfig {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeCaller = zapcore.FullCallerEncoder
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeDuration = zapcore.StringDurationEncoder

	return encoderConfig
}

// initLoggerCores instantiates a 'go.uber.org/zap' logger core for a given
// category with the configuration file parameters.
func initLoggerCores(
	category string,
	cfg *config.Config,
	encoder zapcore.Encoder,
	stdout zapcore.WriteSyncer,
	signal chan<- os.Signal) ([]zapcore.Core, error) {

	cores := []zapcore.Core{}

	for _, logger := range cfg.Log.Loggers {
		if !logger.Enabled {
			continue
		}

		if strings.ToLower(logger.Category) == category {
			var level zapcore.Level
			if err := level.UnmarshalText([]byte(logger.Level)); err != nil {
				return nil, err
			}

			switch strings.ToLower(logger.Type) {
			case "stdout":
				cores = append(cores, zapcore.NewCore(encoder, stdout, level))
			case "file":
				core := zapcore.NewCore(encoder, zapcore.AddSync(&jack.Logger{
					Filename:   logger.Path,
					MaxSize:    getOrElse(logger.MaxSize, 100),    // megabytes
					MaxBackups: getOrElse(logger.MaxBackups, 100), // max 100*100 MB = 10 GB
					MaxAge:     getOrElse(logger.MaxAge, 30),      // days
				}), level)
				cores = append(cores, core)
			case "redis":
				core := NewRedisCore(getEncoderConfig(), logger, level, signal)
				cores = append(cores, core)
			case "opensearch":
				writer, err := NewOpenSearchWriteSyncer(logger, signal)
				if err != nil {
					return nil, err
				}
				core := zapcore.NewCore(encoder, writer, level)
				cores = append(cores, core)
			case "graylog":
				writer, err := NewGraylogWriteSyncer(logger, signal)
				if err != nil {
					return nil, err
				}
				core := zapcore.NewCore(encoder, writer, level)
				cores = append(cores, core)
			default:
				fmt.Printf("unrecognized logger type: %s", logger.Type)
			}
		}
	}
	return cores, nil
}

func getOrElse(value, orElse int) int {
	if value != 0 {
		return value
	}
	return orElse
}

// WrapExpectedError wraps an error such that LogError() will print it to the debug log.
func WrapExpectedError(err error) error {
	return ExpectedError{Err: err}
}

type ExpectedError struct {
	Err error
}

func (e ExpectedError) Unwrap() error {
	return e.Err
}

func (e ExpectedError) Error() string {
	return e.Err.Error()
}

// LogError write an error to the logger.
//
// # Dependending on the error, use the DEBUG or ERROR level
//
// Historically, all errors were logged in ERROR. This is not desired. Business errors are of interest
// to the users but not the developers/operators. We want to log in ERROR only the "technical" errors
// that need intervention.
// The best way to do that without a massive exhaustive analysis is to construct iteratively a list of
// `error` that we know we don't want log at ERROR level
func LogError(l *ContextLogger, err error, ctx context.Context, msg string, fields ...zapcore.Field) {
	logMethod := l.Error
	var expectedError ExpectedError
	// In the future we don't want to log all error in ERROR, most are business errors.
	// While we clean up, we just silence some known ones.

	// FIXME: Avoid cyclic imports.....
	// 		  Probably use global package/internal for errors
	if strings.Contains(err.Error(), "rules: not initialized") {
		logMethod = l.Debug
	} else if errors.As(err, &expectedError) {
		logMethod = l.Debug
	}

	logMethod(ctx, msg, fields...)
}

func NewBasicLogger() *ContextLogger {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeCaller = zapcore.FullCallerEncoder
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeDuration = zapcore.StringDurationEncoder

	jsonEncoder := zapcore.NewJSONEncoder(encoderConfig)
	stdoutCore := zapcore.NewCore(jsonEncoder, stdout, zap.InfoLevel)

	logger := zap.New(stdoutCore, zap.AddCaller())
	return NewContextLogger(logger)
}
