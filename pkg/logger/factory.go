package logger

import "io"

// Factory is a utility to create Logger instances with predefined settings.
type Factory struct {
	target   io.Writer
	name     string
	logLevel LogLevel
}

// NewLoggerFactory creates a new LoggerFactory instance.
func NewLoggerFactory(target io.Writer, name string, logLevel LogLevel) *Factory {
	return &Factory{
		target:   target,
		name:     name,
		logLevel: logLevel,
	}
}

// NewLogger creates a new Logger instance with the specified module and ip.
func (lf *Factory) NewLogger(ip string) *Logger {
	return NewLogger(lf.target, lf.name, lf.logLevel, ip)
}
