package logging

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/controlplane-com/libs-go/pkg/common"
	"go.uber.org/zap/zapcore"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
)

type TestSuite struct {
	suite.Suite
}

func TestAll(t *testing.T) {
	suite.Run(t, new(TestSuite))
}

func (s *TestSuite) SetupTest() {
	logger = nil
	buildLoggerFunc = buildLoggerConfig
}

func (s *TestSuite) TestLogger() {
	err := InitializeLogger(zapcore.InfoLevel)
	s.Nil(err)

	logger := Logger()
	s.NotNil(logger)
}

func (s *TestSuite) TestInitializerError() {
	errMsg := "log-init-error"
	buildLoggerFunc = func(*zap.Config) (*zap.Logger, error) {
		return nil, errors.New(errMsg)
	}
	err := InitializeLogger(zapcore.InfoLevel)
	s.NotNil(err)
	s.Equal(fmt.Sprintf("failed to build logger: %s", errMsg), err.Error())
}

func (s *TestSuite) TestLoggerWithContext() {
	// switch std input/output
	stdOut := os.Stdout
	stdIn := os.Stdin
	reader, writer, testerErr := os.Pipe()
	os.Stdout = writer
	if testerErr != nil {
		s.FailNow("pipe error")
		return
	}

	err := InitializeLogger(zapcore.InfoLevel)
	s.Nil(err)

	msg := "test"
	id := "id"
	ctx := context.WithValue(context.Background(), common.TraceIDKey, interface{}(id))
	ctxLogger := LoggerWithContext(ctx)

	// write to output
	ctxLogger.Infof("%s", msg)

	err = writer.Close()
	if err != nil {
		s.FailNow("writer error")
		return
	}
	loggedOutput, testerErr := io.ReadAll(reader)
	if testerErr != nil {
		s.FailNow("read error")
		return
	}
	os.Stdout = stdOut
	os.Stdin = stdIn
	result := make(map[string]interface{}, 0)
	if testerErr = json.Unmarshal(loggedOutput, &result); testerErr != nil {
		return
	}
	s.Equal(msg, result[LoggerMessage])
	s.Equal(id, result[common.FieldTraceID])
}
