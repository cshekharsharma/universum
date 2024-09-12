// Package logger implements a thread-safe logging system that outputs messages
// to both the console and a log file. It supports different log levels such as
// DEBUG, INFO, WARN, ERROR, and FATAL, with console output being color-coded for
// each level for easy differentiation.
//
// The logger ensures thread safety through the use of mutexes, guaranteeing that
// concurrent log messages are processed in an orderly manner. Log messages are
// timestamped and formatted consistently across both output destinations.
package logger

import (
	"fmt"
	"log"
	"os"
	"sync"
	"universum/config"
	"universum/utils"
)

// ANSI color codes for console logger
const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorPurple = "\033[35m"
	colorCyan   = "\033[36m"
	colorWhite  = "\033[37m"

	levelDebug string = "DEBUG"
	levelInfo  string = "INFO"
	levelWarn  string = "WARN"
	levelError string = "ERROR"
	levelFatal string = "FATAL"

	levelCodeDebug int = 0
	levelCodeInfo  int = 1
	levelCodeWarn  int = 2
	levelCodeError int = 3
	levelCodeFatal int = 4
)

var (
	loggerInstance *Logger
	once           sync.Once
)

type Logger struct {
	consoleLogger *log.Logger
	fileLogger    *log.Logger
	mutex         sync.Mutex
}

// Get returns a singleton instance of Logger. It initializes the logger with
// output directed to both the console and a log file specified by the application
// configuration. The function ensures that only one instance of Logger is created
// and used throughout the application, providing a centralized logging solution.
func Get() *Logger {
	once.Do(func() {
		logFilePath := config.GetServerLogFilePath()
		file, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			log.Fatalf("Failed to open log file: %v\n", err)
		}

		loggerInstance = &Logger{
			consoleLogger: log.New(os.Stdout, "", 0),
			fileLogger:    log.New(file, "", 0),
		}
	})

	return loggerInstance
}

// log is an internal method that formats and logs messages. It applies color coding for console output,
// handles timestamping, and ensures thread safety. Used by Debug, Info, Warn, Error, and Fatal methods.
func (l *Logger) log(level string, color, format string, v ...interface{}) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	if getLevelIdFromName(level) < getLevelIdFromName(config.GetMinimumLogLevel()) {
		return
	}

	currTime := utils.GetCurrentReadableTime()
	message := fmt.Sprintf(format, v...)

	levelStr := fmt.Sprintf("%5s", level)

	l.consoleLogger.Printf("%s %s%s%s :: %s\n", currTime, color, levelStr, colorReset, message)
	l.fileLogger.Printf("%s %s :: %s\n", currTime, levelStr, message)
}

// Debug logs a message at the DEBUG level. Messages are displayed in cyan in the
// console and are also written to the log file with a timestamp and the DEBUG label.
func (l *Logger) Debug(format string, v ...interface{}) {
	l.log(levelDebug, colorCyan, format, v...)
}

// Info logs a message at the INFO level, using green for console output.
func (l *Logger) Info(format string, v ...interface{}) {
	l.log(levelInfo, colorGreen, format, v...)
}

// Warn logs a message at the WARN level, with yellow coloring in the console.
func (l *Logger) Warn(format string, v ...interface{}) {
	l.log(levelWarn, colorYellow, format, v...)
}

// Error logs a message at the ERROR level. These messages appear in red in the console.
func (l *Logger) Error(format string, v ...interface{}) {
	l.log(levelError, colorRed, format, v...)
}

// Fatal logs a message at the FATAL level, similar to Error, but intended for
// use with critical errors that will result in program termination.
func (l *Logger) Fatal(format string, v ...interface{}) {
	l.log(levelFatal, colorRed, format, v...)
}

// getLevelIdFromName is a helper function that returns the integer code for a given log level name.
func getLevelIdFromName(name string) int {
	switch name {
	case levelDebug:
		return levelCodeDebug
	case levelInfo:
		return levelCodeInfo
	case levelWarn:
		return levelCodeWarn
	case levelError:
		return levelCodeError
	case levelFatal:
		return levelCodeFatal
	default:
		return levelCodeInfo
	}
}
