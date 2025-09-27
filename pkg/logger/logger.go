// Package logger provides a simple logging utility with different log levels and color-coded output.
package logger

import (
	"fmt"
	"io"
	"sync/atomic"
	"time"
)

type termColors string

const (
	// Reset.
	reset termColors = "\033[0m"
	// Regular Colors.
	black     termColors = "\033[0;30m"
	red       termColors = "\033[0;31m"
	yellow    termColors = "\033[0;33m"
	blue      termColors = "\033[0;34m"
	green     termColors = "\033[0;32m"
	lightBlue termColors = "\033[0;36m"
)

// LogLevel represents the severity of the log message.
type LogLevel string

// Log levels.
const (
	SUCCESS LogLevel = "SUCCESS"
	DEBUG   LogLevel = "DEBUG"
	INFO    LogLevel = "INFO"
	WARNING LogLevel = "WARNING"
	ERROR   LogLevel = "ERROR"
)

// Logger is a simple logger that writes to an io.Writer.
type Logger struct {
	Target   io.Writer
	Module   atomic.Value
	logLevel LogLevel
	ip       string
}

// NewLogger creates a new Logger instance.
// The ip parameter can be substituted with System or similar for logs that are not tied to a specific client.
func NewLogger(target io.Writer, module string, logLevel LogLevel, ip string) *Logger {
	logger := &Logger{
		Target:   target,
		logLevel: logLevel,
		ip:       ip,
	}
	logger.Module.Store(module)
	return logger
}

func getLocalTime() string {
	return time.Now().Format("2006-01-02 - 15:04:05")
}

// Printf always prints
// The log will also be colored in green for easier identification.
func (l *Logger) Printf(format string, args ...any) {
	l.baseLogger(green, SUCCESS, format, args...)
}

// PrintfError always prints
// The log will also be colored in red for easier identification.
func (l *Logger) PrintfError(format string, args ...any) {
	l.baseLogger(red, ERROR, format, args...)
}

// PrintfWarning only prints if log level is set to WARNING, INFO or DEBUG
// The log will also be colored in yellow for easier identification.
func (l *Logger) PrintfWarning(format string, args ...any) {
	if l.logLevel == WARNING || l.logLevel == INFO ||
		l.logLevel == DEBUG {
		l.baseLogger(yellow, WARNING, format, args...)
	}
}

// PrintfInfo only prints if log level is set to INFO or DEBUG
// The log will also be colored in blue for easier identification.
func (l *Logger) PrintfInfo(format string, args ...any) {
	if l.logLevel == INFO || l.logLevel == DEBUG {
		l.baseLogger(blue, INFO, format, args...)
	}
}

// PrintfDebug only prints if log level is set to DEBUG
// The log will also be colored in light blue for easier identification.
func (l *Logger) PrintfDebug(format string, args ...any) {
	if l.logLevel == DEBUG {
		l.baseLogger(lightBlue, DEBUG, format, args...)
	}
}

func (l *Logger) baseLogger(color termColors, logType LogLevel, format string, args ...any) {
	moduleName := l.Module.Load().(string)

	formattedMessage := fmt.Sprintf(format, args...)
	//GENERAL SCHEMA:
	// {color}[KIND][TIME][IP][MODULE] MESSAGE{reset}
	_, _ = fmt.Fprintf(
		l.Target,
		"%s[%s][%s][%s][%s] %s%s\n",
		color,
		logType,
		getLocalTime(),
		l.ip,
		moduleName,
		formattedMessage,
		reset,
	)
}
