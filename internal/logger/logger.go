package logger

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

var (
	logFile     *os.File
	logMutex    sync.Mutex
	initialized bool
)

// LogLevel represents the severity of a log message
type LogLevel string

const (
	LevelDebug LogLevel = "debug"
	LevelInfo  LogLevel = "info"
	LevelWarn  LogLevel = "warn"
	LevelError LogLevel = "error"
)

// LogEntry represents a structured log entry
type LogEntry struct {
	Timestamp string                 `json:"timestamp"`
	Level     LogLevel               `json:"level"`
	Message   string                 `json:"message"`
	Context   map[string]interface{} `json:"context,omitempty"`
}

// Init initializes the logger with a log file path
func Init(logPath string) error {
	logMutex.Lock()
	defer logMutex.Unlock()

	if initialized {
		return nil
	}

	// Create log directory if needed
	logDir := filepath.Dir(logPath)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	// Open log file (append mode)
	var err error
	logFile, err = os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}

	initialized = true
	return nil
}

// Close closes the log file
func Close() error {
	logMutex.Lock()
	defer logMutex.Unlock()

	if logFile != nil {
		err := logFile.Close()
		logFile = nil
		initialized = false
		return err
	}
	return nil
}

// log writes a log entry to both stderr and file
func log(level LogLevel, message string, context map[string]interface{}) {
	entry := LogEntry{
		Timestamp: time.Now().Format(time.RFC3339),
		Level:     level,
		Message:   message,
		Context:   context,
	}

	data, _ := json.Marshal(entry)

	// Always write to stderr (for real-time monitoring)
	fmt.Fprintln(os.Stderr, string(data))

	// Write to file if initialized
	logMutex.Lock()
	defer logMutex.Unlock()

	if logFile != nil {
		fmt.Fprintln(logFile, string(data))
		logFile.Sync() // Ensure it's written immediately
	}
}

// Debug logs a debug message
func Debug(message string, context ...map[string]interface{}) {
	ctx := mergeContext(context...)
	log(LevelDebug, message, ctx)
}

// Info logs an info message
func Info(message string, context ...map[string]interface{}) {
	ctx := mergeContext(context...)
	log(LevelInfo, message, ctx)
}

// Warn logs a warning message
func Warn(message string, context ...map[string]interface{}) {
	ctx := mergeContext(context...)
	log(LevelWarn, message, ctx)
}

// Error logs an error message
func Error(message string, context ...map[string]interface{}) {
	ctx := mergeContext(context...)
	log(LevelError, message, ctx)
}

// mergeContext merges multiple context maps
func mergeContext(contexts ...map[string]interface{}) map[string]interface{} {
	if len(contexts) == 0 {
		return nil
	}

	merged := make(map[string]interface{})
	for _, ctx := range contexts {
		for k, v := range ctx {
			merged[k] = v
		}
	}
	return merged
}
