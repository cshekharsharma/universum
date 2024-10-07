// Package logger implements a high-performance, thread-safe logging system that outputs messages
// to both the console and a log file. It uses a buffered channel and a background goroutine to
// handle logging asynchronously, with batch writing to optimize I/O operations.
package logger

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
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

	levelDebug string = config.LogLevelDebug
	levelInfo  string = config.LogLevelInfo
	levelWarn  string = config.LogLevelWarn
	levelError string = config.LogLevelError
	levelFatal string = config.LogLevelFatal

	levelCodeDebug int = 0
	levelCodeInfo  int = 1
	levelCodeWarn  int = 2
	levelCodeError int = 3
	levelCodeFatal int = 4

	loggerQueueLength   int           = 10000
	loggerBatchSize     int           = 50
	loggerBatchInterval time.Duration = 100 * time.Millisecond
)

var (
	loggerInstance *Logger
	once           sync.Once
)

// logMessage represents a log message with its level and content.
type logMessage struct {
	level     string
	message   string
	timestamp string
}

// Logger represents the logger with batch writing capability.
type Logger struct {
	consoleLogger   *log.Logger
	fileLogger      *log.Logger
	logLevel        int
	messageQueue    chan *logMessage
	wg              sync.WaitGroup
	shutdown        chan struct{}
	colorMap        map[string]string
	batchSize       int
	batchInterval   time.Duration
	minimumLogLevel string
}

// Get returns a singleton instance of Logger.
func Get() *Logger {
	once.Do(func() {
		config.PopulateDefaultConfig()

		logFilePath := filepath.Clean(
			filepath.Join(config.Store.Logging.LogFileDirectory, config.DefaultServerLogFile))

		file, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			log.Fatalf("Failed to open log file: %v\n", err)
		}

		loggerInstance = &Logger{
			consoleLogger:   log.New(os.Stdout, "", 0),
			fileLogger:      log.New(file, "", 0),
			logLevel:        getLevelIdFromName(config.Store.Logging.MinimumLogLevel),
			messageQueue:    make(chan *logMessage, loggerQueueLength),
			shutdown:        make(chan struct{}),
			batchSize:       loggerBatchSize,     // Number of messages per batch
			batchInterval:   loggerBatchInterval, // Max wait time before flushing batch
			minimumLogLevel: config.Store.Logging.MinimumLogLevel,
			colorMap: map[string]string{
				levelDebug: colorCyan,
				levelInfo:  colorGreen,
				levelWarn:  colorYellow,
				levelError: colorRed,
				levelFatal: colorRed,
			},
		}

		// Start the background message processor
		loggerInstance.wg.Add(1)
		go loggerInstance.processLogQueue()
	})

	return loggerInstance
}

// processLogQueue processes log messages from the queue in batches.
func (l *Logger) processLogQueue() {
	defer l.wg.Done()

	batch := make([]*logMessage, 0, l.batchSize)
	timer := time.NewTimer(l.batchInterval)
	defer timer.Stop()

	for {
		select {
		case msg := <-l.messageQueue:
			batch = append(batch, msg)
			if len(batch) >= l.batchSize {
				l.flushBatch(batch)
				batch = batch[:0] // Reset batch
				timer.Reset(l.batchInterval)
			}
		case <-timer.C:
			if len(batch) > 0 {
				l.flushBatch(batch)
				batch = batch[:0]
			}
			timer.Reset(l.batchInterval)
		case <-l.shutdown:
			// Drain the messageQueue before exiting
			for len(l.messageQueue) > 0 {
				msg := <-l.messageQueue
				batch = append(batch, msg)
				if len(batch) >= l.batchSize {
					l.flushBatch(batch)
					batch = batch[:0]
				}
			}
			if len(batch) > 0 {
				l.flushBatch(batch)
			}
			return
		}
	}
}

// flushBatch writes a batch of log messages to the console and file.
func (l *Logger) flushBatch(batch []*logMessage) {
	var consoleBuilder strings.Builder
	var fileBuilder strings.Builder

	for _, msg := range batch {
		color := l.colorMap[msg.level]
		levelStr := fmt.Sprintf("%5s", strings.ToUpper(msg.level))

		consoleBuilder.WriteString(fmt.Sprintf("%s %s%s%s :: %s\n", msg.timestamp, color, levelStr, colorReset, msg.message))
		fileBuilder.WriteString(fmt.Sprintf("%s %s :: %s\n", msg.timestamp, levelStr, msg.message))
	}

	l.consoleLogger.Print(consoleBuilder.String())
	l.fileLogger.Print(fileBuilder.String())
}

// log enqueues a log message if the level is appropriate.
func (l *Logger) log(level string, format string, v ...interface{}) {
	levelId := getLevelIdFromName(level)
	if levelId < l.logLevel {
		return
	}

	msg := &logMessage{
		level:     level,
		message:   fmt.Sprintf(format, v...),
		timestamp: utils.GetCurrentReadableTime(),
	}

	select {
	case l.messageQueue <- msg:
	default:
		// If the queue is full, drop the message to avoid blocking
	}
}

// Debug logs a message at DEBUG level.
func (l *Logger) Debug(format string, v ...interface{}) {
	l.log(levelDebug, format, v...)
}

// Info logs a message at INFO level.
func (l *Logger) Info(format string, v ...interface{}) {
	l.log(levelInfo, format, v...)
}

// Warn logs a message at WARN level.
func (l *Logger) Warn(format string, v ...interface{}) {
	l.log(levelWarn, format, v...)
}

// Error logs a message at ERROR level.
func (l *Logger) Error(format string, v ...interface{}) {
	l.log(levelError, format, v...)
}

// Fatal logs a message at FATAL level.
func (l *Logger) Fatal(format string, v ...interface{}) {
	l.log(levelFatal, format, v...)
}

// Close gracefully shuts down the logger, ensuring all messages are processed.
func (l *Logger) Close() {
	close(l.shutdown)
	l.wg.Wait()

	if file, ok := l.fileLogger.Writer().(*os.File); ok {
		file.Close()
	}
}

// getLevelIdFromName returns the integer code for the given log level name.
func getLevelIdFromName(name string) int {
	switch strings.ToUpper(name) {
	case strings.ToUpper(levelDebug):
		return levelCodeDebug
	case strings.ToUpper(levelInfo):
		return levelCodeInfo
	case strings.ToUpper(levelWarn):
		return levelCodeWarn
	case strings.ToUpper(levelError):
		return levelCodeError
	case strings.ToUpper(levelFatal):
		return levelCodeFatal
	default:
		return levelCodeInfo // Default to INFO level
	}
}

// resetLoggerForTesting resets the logger instance for testing purposes.
func resetLoggerForTesting() {
	loggerInstance = nil
	once = sync.Once{}
}
