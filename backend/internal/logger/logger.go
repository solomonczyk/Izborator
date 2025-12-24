package logger

import (
	"os"

	"github.com/rs/zerolog"
)

// Logger обёртка над zerolog для единообразного использования
type Logger struct {
	logger zerolog.Logger
}

// New создаёт новый логгер
func New(level string) *Logger {
	// Парсинг уровня логирования
	logLevel, err := zerolog.ParseLevel(level)
	if err != nil {
		logLevel = zerolog.InfoLevel
	}

	// Настройка zerolog
	zerolog.SetGlobalLevel(logLevel)
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	// Красивый вывод для development
	logger := zerolog.New(os.Stdout).
		With().
		Timestamp().
		Logger()

	return &Logger{logger: logger}
}

// Debug логирует сообщение уровня Debug
func (l *Logger) Debug(msg string, fields ...map[string]interface{}) {
	l.logEvent(l.logger.Debug(), msg, fields...)
}

// Info логирует сообщение уровня Info
func (l *Logger) Info(msg string, fields ...map[string]interface{}) {
	l.logEvent(l.logger.Info(), msg, fields...)
}

// Warn логирует сообщение уровня Warn
func (l *Logger) Warn(msg string, fields ...map[string]interface{}) {
	l.logEvent(l.logger.Warn(), msg, fields...)
}

// Error логирует сообщение уровня Error
func (l *Logger) Error(msg string, fields ...map[string]interface{}) {
	l.logEvent(l.logger.Error(), msg, fields...)
}

// Fatal логирует сообщение уровня Fatal и завершает программу
func (l *Logger) Fatal(msg string, fields ...map[string]interface{}) {
	l.logEvent(l.logger.Fatal(), msg, fields...)
}

// logEvent вспомогательная функция для логирования с полями
func (l *Logger) logEvent(event *zerolog.Event, msg string, fields ...map[string]interface{}) {
	if len(fields) > 0 {
		for key, value := range fields[0] {
			event = event.Interface(key, value)
		}
	}
	event.Msg(msg)
}

// With создаёт новый логгер с дополнительными полями
func (l *Logger) With(fields map[string]interface{}) *Logger {
	logger := l.logger.With().Fields(fields).Logger()
	return &Logger{logger: logger}
}
