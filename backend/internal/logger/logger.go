package logger

import (
	"os"
	"strings"

	"github.com/rs/zerolog"
)

// Logger wraps zerolog to provide structured logging with shared fields.
type Logger struct {
	logger zerolog.Logger
}

// New creates a logger with level/env/service defaults.
func New(level string) *Logger {
	logLevel, err := zerolog.ParseLevel(level)
	if err != nil {
		logLevel = zerolog.InfoLevel
	}

	zerolog.SetGlobalLevel(logLevel)
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	env := strings.TrimSpace(os.Getenv("ENV"))
	if env == "" {
		env = strings.TrimSpace(os.Getenv("APP_ENV"))
	}
	if env == "" {
		env = "development"
	}

	service := strings.TrimSpace(os.Getenv("SERVICE_NAME"))
	if service == "" {
		service = "backend"
	}

	logger := zerolog.New(os.Stdout).
		With().
		Timestamp().
		Str("service", service).
		Str("env", env).
		Logger()

	return &Logger{logger: logger}
}

// Debug writes a debug-level message.
func (l *Logger) Debug(msg string, fields ...map[string]interface{}) {
	l.logEvent(l.logger.Debug(), msg, fields...)
}

// Info writes an info-level message.
func (l *Logger) Info(msg string, fields ...map[string]interface{}) {
	l.logEvent(l.logger.Info(), msg, fields...)
}

// Warn writes a warn-level message.
func (l *Logger) Warn(msg string, fields ...map[string]interface{}) {
	l.logEvent(l.logger.Warn(), msg, fields...)
}

// Error writes an error-level message.
func (l *Logger) Error(msg string, fields ...map[string]interface{}) {
	l.logEvent(l.logger.Error(), msg, fields...)
}

// Fatal writes a fatal-level message and exits.
func (l *Logger) Fatal(msg string, fields ...map[string]interface{}) {
	l.logEvent(l.logger.Fatal(), msg, fields...)
}

// logEvent attaches structured fields and sends the event.
func (l *Logger) logEvent(event *zerolog.Event, msg string, fields ...map[string]interface{}) {
	if len(fields) > 0 {
		for key, value := range fields[0] {
			event = event.Interface(key, value)
		}
	}
	event.Msg(msg)
}

// With returns a child logger with additional fields.
func (l *Logger) With(fields map[string]interface{}) *Logger {
	logger := l.logger.With().Fields(fields).Logger()
	return &Logger{logger: logger}
}
