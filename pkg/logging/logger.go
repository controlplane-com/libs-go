package logging

import (
	"context"
	"errors"
	"fmt"
	"github.com/controlplane-com/libs-go/pkg/common"
	"reflect"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	ComponentName    = "audit"
	JSONEncoder      = "json"
	StdOutSinkName   = "stdout"
	StdErrSinkName   = "stderr"
	LoggerLevel      = "level"
	LoggerTimestamp  = "timestamp"
	LoggerLogger     = "logger"
	LoggerCaller     = "caller"
	LoggerMessage    = "message"
	LoggerStacktrace = "stacktrace"
)

var (
	logger          *zap.Logger
	loggerMu        sync.Mutex // guards logger (lazy init in Logger vs InitializeLogger)
	stdOutSink      = StdOutSinkName
	stdErrSink      = StdErrSinkName
	buildLoggerFunc = buildLoggerConfig
)

func SetOutputSinks(out string, err string) error {
	stdOutSink = out
	stdErrSink = err
	loggerMu.Lock()
	l := logger
	loggerMu.Unlock()
	if l != nil {
		return InitializeLogger(l.Level())
	}
	return nil
}

// InitializeLogger func
func InitializeLogger(logLevel zapcore.Level) error {
	config := zap.Config{
		Encoding:         JSONEncoder,
		OutputPaths:      []string{stdOutSink},
		ErrorOutputPaths: []string{stdErrSink},
		Level:            zap.NewAtomicLevelAt(logLevel),
		EncoderConfig: zapcore.EncoderConfig{
			NameKey:        LoggerLogger,
			MessageKey:     LoggerMessage,
			CallerKey:      LoggerCaller,
			StacktraceKey:  LoggerStacktrace,
			LineEnding:     zapcore.DefaultLineEnding,
			TimeKey:        LoggerTimestamp,
			LevelKey:       LoggerLevel,
			EncodeLevel:    zapcore.CapitalLevelEncoder,
			EncodeTime:     zapcore.ISO8601TimeEncoder,
			EncodeDuration: zapcore.SecondsDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		}}
	l, err := buildLoggerFunc(&config)
	if err != nil {
		return fmt.Errorf("failed to build logger: %s", err.Error())
	}
	loggerMu.Lock()
	logger = l
	loggerMu.Unlock()
	return nil
}

func buildLoggerConfig(config *zap.Config) (*zap.Logger, error) {
	return config.Build()
}

// LoggerWithContext func
func LoggerWithContext(ctx context.Context) *zap.SugaredLogger {
	newLogger := Logger().Sugar()
	fieldValue := ctx.Value(common.TraceIDKey)
	params := make([]interface{}, 0)
	if reflect.ValueOf(fieldValue).IsValid() {
		params = append(params, common.FieldTraceID)
		params = append(params, fieldValue)
	}
	return newLogger.With(params...)
}

// Logger func
func Logger() *zap.Logger {
	loggerMu.Lock()
	l := logger
	loggerMu.Unlock()
	if l == nil {
		InitializeLogger(zapcore.InfoLevel)
		loggerMu.Lock()
		l = logger
		loggerMu.Unlock()
	}
	return l
}

func ContextWithTraceID(value string) context.Context {
	return context.WithValue(context.Background(), common.TraceIDKey, value)
}

type ZapLogLevelMapper struct {
}

func (z ZapLogLevelMapper) Map(_ any, scannedValue any) (any, error) {
	switch v := scannedValue.(type) {
	case string:
		return zapcore.ParseLevel(v)
	case int8:
		return zapcore.Level(v), nil
	case uint8:
		return zapcore.Level(v), nil
	case zapcore.Level:
		return v, nil
	default:
		return nil, errors.New("unsupported type for ZapLogLevelMapper")
	}
}
