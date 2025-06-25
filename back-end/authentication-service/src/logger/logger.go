package logger

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/fatih/color"
)

type LogLevel string

const (
	INFO  LogLevel = "INFO"
	ERROR LogLevel = "ERROR"
	DEBUG LogLevel = "DEBUG"
	WARN  LogLevel = "WARN"
)

type Logger struct {
	level LogLevel
}

// NewLogger initializes a new Logger with the desired log level
func NewLogger(level LogLevel) *Logger {
	// Disable the default log timestamp from the log package
	log.SetOutput(os.Stdout)
	log.SetFlags(0) // No date/time prefix added automatically

	l := &Logger{level: level}
	l.Info("✅ Logger initialized successfully with level: " + string(level))
	return l
}

// logMessage prints the log with timestamp, level, and colored output
func (l *Logger) logMessage(level LogLevel, message string) {
	// Create your own timestamp
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	// Get the colored log level
	coloredLevel := getColoredLevel(level)
	// Print the log message with formatted timestamp and level
	log.Printf("[%s] [%s] %s\n", timestamp, coloredLevel, message)
}

// Color mapping for different log levels
func getColoredLevel(level LogLevel) string {
	switch level {
	case INFO:
		return color.New(color.FgBlue).Sprint(string(INFO))
	case ERROR:
		return color.New(color.FgRed).Sprint(string(ERROR))
	case DEBUG:
		return color.New(color.FgCyan).Sprint(string(DEBUG))
	case WARN:
		return color.New(color.FgYellow).Sprint(string(WARN))
	default:
		return string(level)
	}
}

// Info logs an informational message
func (l *Logger) Info(msg string) {
	l.logMessage(INFO, msg)
}

// Error logs an error message
func (l *Logger) Error(msg string) {
	l.logMessage(ERROR, msg)
}

// Debug logs a debug message only if the level is DEBUG
func (l *Logger) Debug(msg string) {
	if l.level == DEBUG {
		l.logMessage(DEBUG, msg)
	}
}

// Warn logs a warning message
func (l *Logger) Warn(msg string) {
	l.logMessage(WARN, msg)
}

// Fatalf logs a fatal error message and exits the program
func (l *Logger) Fatalf(msg string, args ...interface{}) {
	// Format the message with any arguments
	message := fmt.Sprintf(msg, args...)
	// Log the fatal error
	l.logMessage(ERROR, message)
	// Exit the program after logging
	os.Exit(1)
}
