package logger

import (
	"os"

	"github.com/sirupsen/logrus"
)

// LogLevel represents logging levels
type LogLevel int

const (
	// PanicLevel level, highest level of severity. Logs and then calls panic with the message passed in.
	PanicLevel LogLevel = iota
	// FatalLevel level. Logs and then calls `os.Exit(1)`. 
	FatalLevel
	// ErrorLevel level. Logs. Used for errors that should definitely be noted.
	ErrorLevel
	// WarnLevel level. Non-critical entries that deserve eyes.
	WarnLevel
	// InfoLevel level. General operational entries about what's happening.
	InfoLevel
	// DebugLevel level. Usually only enabled when debugging.
	DebugLevel
)

var log = logrus.New()

func init() {
	// Set default configuration
	log.SetOutput(os.Stdout)
	log.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
		ForceColors:   true,
	})
	log.SetLevel(logrus.InfoLevel)
}

// SetLevel sets the logging level
func SetLevel(level LogLevel) {
	switch level {
	case PanicLevel:
		log.SetLevel(logrus.PanicLevel)
	case FatalLevel:
		log.SetLevel(logrus.FatalLevel)
	case ErrorLevel:
		log.SetLevel(logrus.ErrorLevel)
	case WarnLevel:
		log.SetLevel(logrus.WarnLevel)
	case InfoLevel:
		log.SetLevel(logrus.InfoLevel)
	case DebugLevel:
		log.SetLevel(logrus.DebugLevel)
	}
}

// SetFormatter sets the log formatter
func SetFormatter(formatter logrus.Formatter) {
	log.SetFormatter(formatter)
}

// Debug logs a message at level Debug
func Debug(args ...interface{}) {
	log.Debug(args...)
}

// Debugf logs a message at level Debug
func Debugf(format string, args ...interface{}) {
	log.Debugf(format, args...)
}

// Info logs a message at level Info
func Info(args ...interface{}) {
	log.Info(args...)
}

// Infof logs a message at level Info
func Infof(format string, args ...interface{}) {
	log.Infof(format, args...)
}

// Warn logs a message at level Warn
func Warn(args ...interface{}) {
	log.Warn(args...)
}

// Warnf logs a message at level Warn
func Warnf(format string, args ...interface{}) {
	log.Warnf(format, args...)
}

// Error logs a message at level Error
func Error(args ...interface{}) {
	log.Error(args...)
}

// Errorf logs a message at level Error
func Errorf(format string, args ...interface{}) {
	log.Errorf(format, args...)
}

// Fatal logs a message at level Fatal then the process will exit with status set to 1
func Fatal(args ...interface{}) {
	log.Fatal(args...)
}

// Fatalf logs a message at level Fatal then the process will exit with status set to 1
func Fatalf(format string, args ...interface{}) {
	log.Fatalf(format, args...)
}

// WithField creates an entry with a single field
func WithField(key string, value interface{}) *logrus.Entry {
	return log.WithField(key, value)
}

// WithFields creates an entry with multiple fields
func WithFields(fields logrus.Fields) *logrus.Entry {
	return log.WithFields(fields)
}
