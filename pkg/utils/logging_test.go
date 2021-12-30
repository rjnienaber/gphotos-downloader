package utils

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLogLevelIsSetWithSimpleNew(t *testing.T) {
	logger := NewLogger(Silent)
	assert.Equal(t, Silent, logger.level)
}

func runTests(logger Logger) {
	logger.Error.Print("error line")
	logger.Info.Print("info line")
	logger.Debug.Print("debug line")
	logger.Trace.Print("trace line")
}

func TestLogLevelError(t *testing.T) {
	builder := strings.Builder{}
	logger := NewLoggerWithOptions(WithLevel(Error), WithOut(&builder), WithFlags(0))

	runTests(logger)

	assert.Equal(t, "[ERROR] error line\n", builder.String())
}

func TestLogLevelInfo(t *testing.T) {
	builder := strings.Builder{}
	logger := NewLoggerWithOptions(WithLevel(Info), WithOut(&builder), WithFlags(0))

	runTests(logger)

	assert.Equal(t, "[ERROR] error line\n[INFO] info line\n", builder.String())
}

func TestLogLevelDebug(t *testing.T) {
	builder := strings.Builder{}
	logger := NewLoggerWithOptions(WithLevel(Debug), WithOut(&builder), WithFlags(0))

	runTests(logger)

	assert.Equal(t, "[ERROR] error line\n[INFO] info line\n[DEBUG] debug line\n", builder.String())
}

func TestLogLevelTrace(t *testing.T) {
	builder := strings.Builder{}
	logger := NewLoggerWithOptions(WithLevel(Trace), WithOut(&builder), WithFlags(0))

	runTests(logger)

	lines := "[ERROR] error line\n[INFO] info line\n[DEBUG] debug line\n[TRACE] trace line\n"
	assert.Equal(t, lines, builder.String())
}
