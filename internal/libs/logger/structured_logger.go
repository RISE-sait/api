package logger

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"
	"time"

	contextUtils "api/utils/context"
)

// LogLevel represents the severity of a log entry
type LogLevel string

const (
	LevelDebug LogLevel = "DEBUG"
	LevelInfo  LogLevel = "INFO"
	LevelWarn  LogLevel = "WARN"
	LevelError LogLevel = "ERROR"
	LevelFatal LogLevel = "FATAL"
)

// LogEntry represents a structured log entry
type LogEntry struct {
	Timestamp   time.Time              `json:"timestamp"`
	Level       LogLevel               `json:"level"`
	Message     string                 `json:"message"`
	Component   string                 `json:"component,omitempty"`
	UserID      string                 `json:"user_id,omitempty"`
	RequestID   string                 `json:"request_id,omitempty"`
	TraceID     string                 `json:"trace_id,omitempty"`
	File        string                 `json:"file,omitempty"`
	Line        int                    `json:"line,omitempty"`
	Function    string                 `json:"function,omitempty"`
	Fields      map[string]interface{} `json:"fields,omitempty"`
	Error       string                 `json:"error,omitempty"`
	StackTrace  string                 `json:"stack_trace,omitempty"`
}

// StructuredLogger provides structured logging functionality
type StructuredLogger struct {
	component string
	fields    map[string]interface{}
}

var (
	// Global logger instance
	globalLogger = &StructuredLogger{
		component: "api",
		fields:    make(map[string]interface{}),
	}
)

// NewLogger creates a new structured logger for a specific component
func NewLogger(component string) *StructuredLogger {
	return &StructuredLogger{
		component: component,
		fields:    make(map[string]interface{}),
	}
}

// WithField adds a field to the logger context
func (l *StructuredLogger) WithField(key string, value interface{}) *StructuredLogger {
	newLogger := &StructuredLogger{
		component: l.component,
		fields:    make(map[string]interface{}),
	}
	
	// Copy existing fields
	for k, v := range l.fields {
		newLogger.fields[k] = v
	}
	
	// Add new field
	newLogger.fields[key] = value
	return newLogger
}

// WithFields adds multiple fields to the logger context
func (l *StructuredLogger) WithFields(fields map[string]interface{}) *StructuredLogger {
	newLogger := &StructuredLogger{
		component: l.component,
		fields:    make(map[string]interface{}),
	}
	
	// Copy existing fields
	for k, v := range l.fields {
		newLogger.fields[k] = v
	}
	
	// Add new fields
	for k, v := range fields {
		newLogger.fields[k] = v
	}
	
	return newLogger
}

// WithContext adds context information (user ID, request ID) to the logger
func (l *StructuredLogger) WithContext(ctx context.Context) *StructuredLogger {
	fields := make(map[string]interface{})
	
	// Extract user ID from context if available
	if userID, err := contextUtils.GetUserID(ctx); err == nil {
		fields["user_id"] = userID.String()
	}
	
	// Extract request ID from context if available
	if requestID, ok := ctx.Value("request_id").(string); ok {
		fields["request_id"] = requestID
	}
	
	// Extract trace ID from context if available
	if traceID, ok := ctx.Value("trace_id").(string); ok {
		fields["trace_id"] = traceID
	}
	
	return l.WithFields(fields)
}

// Debug logs a debug message
func (l *StructuredLogger) Debug(message string) {
	l.log(LevelDebug, message, nil, false)
}

// Debugf logs a formatted debug message
func (l *StructuredLogger) Debugf(format string, args ...interface{}) {
	l.log(LevelDebug, fmt.Sprintf(format, args...), nil, false)
}

// Info logs an info message
func (l *StructuredLogger) Info(message string) {
	l.log(LevelInfo, message, nil, false)
}

// Infof logs a formatted info message
func (l *StructuredLogger) Infof(format string, args ...interface{}) {
	l.log(LevelInfo, fmt.Sprintf(format, args...), nil, false)
}

// Warn logs a warning message
func (l *StructuredLogger) Warn(message string) {
	l.log(LevelWarn, message, nil, false)
}

// Warnf logs a formatted warning message
func (l *StructuredLogger) Warnf(format string, args ...interface{}) {
	l.log(LevelWarn, fmt.Sprintf(format, args...), nil, false)
}

// Error logs an error message
func (l *StructuredLogger) Error(message string, err error) {
	l.log(LevelError, message, err, true)
	
	// Send critical payment errors to Slack
	if IsPaymentCritical(LevelError, message, l.component) {
		alertType := "PAYMENT_ERROR"
		useSync := false
		
		if strings.Contains(strings.ToLower(message), "webhook") {
			alertType = "WEBHOOK_FAILED"
			useSync = true // Webhook failures are critical for revenue
		} else if strings.Contains(strings.ToLower(message), "database") {
			alertType = "DB_CONNECTION"
			useSync = true // Database issues are critical
		} else if strings.Contains(strings.ToLower(message), "payment processing failed") {
			useSync = true // Core payment failures are critical
		}
		
		fields := map[string]interface{}{
			"Component": l.component,
			"Error": func() string {
				if err != nil {
					return err.Error()
				}
				return "N/A"
			}(),
		}
		
		// Add custom fields if available
		for k, v := range l.fields {
			fields[k] = v
		}
		
		// Use sync for critical alerts to avoid delays
		if useSync {
			SendSlackAlertSync(alertType, message, fields)
		} else {
			SendSlackAlert(alertType, message, fields)
		}
	}
}

// Errorf logs a formatted error message
func (l *StructuredLogger) Errorf(format string, err error, args ...interface{}) {
	l.log(LevelError, fmt.Sprintf(format, args...), err, true)
}

// Fatal logs a fatal message and exits the program
func (l *StructuredLogger) Fatal(message string, err error) {
	l.log(LevelFatal, message, err, true)
	
	// Always send fatal errors to Slack (critical system failure)
	fields := map[string]interface{}{
		"Component": l.component,
		"Error": func() string {
			if err != nil {
				return err.Error()
			}
			return "N/A"
		}(),
		"Severity": "FATAL - SYSTEM SHUTDOWN",
	}
	
	for k, v := range l.fields {
		fields[k] = v
	}
	
	SendSlackAlertSync("PAYMENT_DOWN", "FATAL: "+message, fields)
	
	os.Exit(1)
}

// Fatalf logs a formatted fatal message and exits the program
func (l *StructuredLogger) Fatalf(format string, err error, args ...interface{}) {
	l.log(LevelFatal, fmt.Sprintf(format, args...), err, true)
	os.Exit(1)
}

// log is the internal logging method
func (l *StructuredLogger) log(level LogLevel, message string, err error, includeStackTrace bool) {
	entry := LogEntry{
		Timestamp: time.Now().UTC(),
		Level:     level,
		Message:   message,
		Component: l.component,
		Fields:    l.fields,
	}
	
	// Add error information
	if err != nil {
		entry.Error = err.Error()
	}
	
	// Add caller information
	if pc, file, line, ok := runtime.Caller(2); ok {
		entry.File = getShortFilename(file)
		entry.Line = line
		entry.Function = getFunctionName(pc)
	}
	
	// Add stack trace for errors and fatals
	if includeStackTrace && (level == LevelError || level == LevelFatal) {
		entry.StackTrace = getStackTrace()
	}
	
	// Convert to JSON and output
	jsonBytes, jsonErr := json.Marshal(entry)
	if jsonErr != nil {
		// Fallback to simple logging if JSON marshaling fails
		log.Printf("[%s] %s: %s (JSON error: %v)", level, l.component, message, jsonErr)
		return
	}
	
	// Output the structured log
	log.Println(string(jsonBytes))
}

// getShortFilename extracts just the filename from a full path
func getShortFilename(fullPath string) string {
	parts := strings.Split(fullPath, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return fullPath
}

// getFunctionName extracts the function name from program counter
func getFunctionName(pc uintptr) string {
	fn := runtime.FuncForPC(pc)
	if fn == nil {
		return "unknown"
	}
	name := fn.Name()
	// Remove package path, keep only the function name
	if lastSlash := strings.LastIndex(name, "/"); lastSlash >= 0 {
		name = name[lastSlash+1:]
	}
	if lastDot := strings.LastIndex(name, "."); lastDot >= 0 {
		name = name[lastDot+1:]
	}
	return name
}

// getStackTrace captures the current stack trace
func getStackTrace() string {
	buf := make([]byte, 4096)
	n := runtime.Stack(buf, false)
	return string(buf[:n])
}

// Global logging functions for convenience

// Debug logs a debug message using the global logger
func Debug(message string) {
	globalLogger.Debug(message)
}

// Debugf logs a formatted debug message using the global logger
func Debugf(format string, args ...interface{}) {
	globalLogger.Debugf(format, args...)
}

// Info logs an info message using the global logger
func Info(message string) {
	globalLogger.Info(message)
}

// Infof logs a formatted info message using the global logger
func Infof(format string, args ...interface{}) {
	globalLogger.Infof(format, args...)
}

// Warn logs a warning message using the global logger
func Warn(message string) {
	globalLogger.Warn(message)
}

// Warnf logs a formatted warning message using the global logger
func Warnf(format string, args ...interface{}) {
	globalLogger.Warnf(format, args...)
}

// Error logs an error message using the global logger
func Error(message string, err error) {
	globalLogger.Error(message, err)
}

// Errorf logs a formatted error message using the global logger
func Errorf(format string, err error, args ...interface{}) {
	globalLogger.Errorf(format, err, args...)
}

// Fatal logs a fatal message using the global logger and exits
func Fatal(message string, err error) {
	globalLogger.Fatal(message, err)
}

// Fatalf logs a formatted fatal message using the global logger and exits
func Fatalf(format string, err error, args ...interface{}) {
	globalLogger.Fatalf(format, err, args...)
}

// WithContext creates a logger with context information
func WithContext(ctx context.Context) *StructuredLogger {
	return globalLogger.WithContext(ctx)
}

// WithField creates a logger with an additional field
func WithField(key string, value interface{}) *StructuredLogger {
	return globalLogger.WithField(key, value)
}

// WithFields creates a logger with additional fields
func WithFields(fields map[string]interface{}) *StructuredLogger {
	return globalLogger.WithFields(fields)
}

// WithComponent creates a logger for a specific component
func WithComponent(component string) *StructuredLogger {
	return NewLogger(component)
}