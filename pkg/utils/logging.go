package utils

import (
	"io"
	"log"
	"os"
)

type LogLevel int64

const (
	Silent LogLevel = iota
	Error
	Info
	Debug
	Trace
)

type Logger struct {
	out     io.Writer
	level   LogLevel
	flags   int
	Trace   *log.Logger
	Debug   *log.Logger
	Info    *log.Logger
	Error   *log.Logger
	Default *log.Logger
}

type Option func(svc *Logger)

func NewLogger(level LogLevel) Logger {
	return NewLoggerWithOptions(WithLevel(level))
}

func NewLoggerWithOptions(opts ...Option) Logger {
	defaultFlags := log.Ldate | log.Ltime | log.Lshortfile
	logger := Logger{out: os.Stdout, level: Silent, flags: defaultFlags}
	for _, opt := range opts {
		opt(&logger)
	}

	noOpLogger := log.New(io.Discard, "", 0)

	logger.Trace = noOpLogger
	logger.Debug = noOpLogger
	logger.Info = noOpLogger
	logger.Error = noOpLogger
	logger.Default = log.Default()

	if Error <= logger.level {
		logger.Error = log.New(logger.out, "[ERROR] ", logger.flags)
	}

	if Info <= logger.level {
		logger.Info = log.New(logger.out, "[INFO] ", logger.flags)
	}

	if Debug <= logger.level {
		logger.Debug = log.New(logger.out, "[DEBUG] ", logger.flags)
	}

	if Trace <= logger.level {
		logger.Trace = log.New(logger.out, "[TRACE] ", logger.flags)
	}

	return logger
}

func WithOut(out io.Writer) Option {
	return func(logger *Logger) {
		logger.out = out
	}
}

func WithLevel(level LogLevel) Option {
	return func(logger *Logger) {
		logger.level = level
	}
}

func WithFlags(flags int) Option {
	return func(logger *Logger) {
		logger.flags = flags
	}
}
