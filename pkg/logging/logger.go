package logging

import (
	"context"
	"errors"
	"fmt"
	"github.com/controlplane-com/libs-go/pkg/common"
	"reflect"

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
	stdOutSink      = StdOutSinkName
	stdErrSink      = StdErrSinkName
	buildLoggerFunc = buildLoggerConfig
)

func SetOutputSinks(out string, err string) error {
	stdOutSink = out
	stdErrSink = err
	if logger != nil {
		return InitializeLogger(logger.Level())
	}
	return nil
}

// InitializeLogger func
func InitializeLogger(logLevel zapcore.Level) error {
	var err error
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
	if logger, err = buildLoggerFunc(&config); err != nil {
		return fmt.Errorf("failed to build logger: %s", err.Error())
	}
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
	if logger == nil {
		InitializeLogger(zapcore.InfoLevel)
	}
	return logger
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
