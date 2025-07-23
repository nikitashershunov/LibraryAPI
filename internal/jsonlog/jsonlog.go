package jsonlog

import (
	"encoding/json"
	"io"
	"os"
	"runtime/debug"
	"sync"
	"time"
)

// Level represents the severity level.
type Level int8

// Initialize constants which represent a specific severity level.
const (
	LevelInfo  Level = iota // 0
	LevelError              // 1
	LevelFatal              // 2
	LevelOff                // 3
)

// String returns a string for the severity level.
func (l Level) String() string {
	switch l {
	case LevelInfo:
		return "INFO"
	case LevelError:
		return "ERROR"
	case LevelFatal:
		return "FATAL"
	default:
		return ""
	}
}

// Logger is the custom logger. It holds the output destination that the log entries will be
// written to, the minimum severity level that log entries will be written for, a mutex.
type Logger struct {
	out      io.Writer
	minLevel Level
	mu       sync.Mutex
}

// NewLogger returns a new Logger instance which writes log entries at or above a minimum severity
// level to a specific output destination.
func NewLogger(out io.Writer, minLevel Level) *Logger {
	return &Logger{
		out:      out,
		minLevel: minLevel,
	}
}

// PrintInfo writes Info level log entries.
func (l *Logger) PrintInfo(message string, properties map[string]string) {
	l.print(LevelInfo, message, properties)
}

// PrintError writes Error level log entries.
func (l *Logger) PrintError(err error, properties map[string]string) {
	l.print(LevelError, err.Error(), properties)
}

// PrintFatal writes Fatal level log entries. It also terminates application.
func (l *Logger) PrintFatal(err error, properties map[string]string) {
	l.print(LevelFatal, err.Error(), properties)
	os.Exit(1)
}

// print is an internal method for writing a log entry.
func (l *Logger) print(level Level, message string, properties map[string]string) (int, error) {
	if level < l.minLevel {
		return 0, nil
	}

	aux := struct {
		Level      string            `json:"level"`
		Time       string            `json:"time"`
		Message    string            `json:"message"`
		Properties map[string]string `json:"properties,omitempty"`
		Trace      string            `json:"trace,omitempty"`
	}{
		Level:      level.String(),
		Time:       time.Now().UTC().Format(time.RFC3339),
		Message:    message,
		Properties: properties,
	}

	if level >= LevelError {
		aux.Trace = string(debug.Stack())
	}

	var line []byte

	line, err := json.Marshal(aux)
	if err != nil {
		line = []byte(LevelError.String() + ": unable to marshal log message:" + err.Error())
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	return l.out.Write(append(line, '\n'))
}

// Write satisfies the io.Writer interface. It writes a log entry at the ERROR level.
func (l *Logger) Write(message []byte) (n int, err error) {
	return l.print(LevelError, string(message), nil)
}
