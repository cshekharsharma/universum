package logger

import (
	"fmt"
	"log"
	"os"
	"sync"
	"time"
	"universum/config"
)

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

func (l *Logger) log(level string, color, format string, v ...interface{}) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	currTime := time.Now().Format("2006-01-02 15:04:05.000")
	message := fmt.Sprintf(format, v...)

	levelStr := fmt.Sprintf("%5s", level)

	l.consoleLogger.Printf("%s %s%s%s :: %s\n", currTime, color, levelStr, colorReset, message)
	l.fileLogger.Printf("%s %s :: %s\n", currTime, levelStr, message)
}

func (l *Logger) Debug(format string, v ...interface{}) {
	l.log(levelDebug, colorCyan, format, v...)
}

func (l *Logger) Info(format string, v ...interface{}) {
	l.log(levelInfo, colorGreen, format, v...)
}

func (l *Logger) Warn(format string, v ...interface{}) {
	l.log(levelWarn, colorYellow, format, v...)
}

func (l *Logger) Error(format string, v ...interface{}) {
	l.log(levelError, colorRed, format, v...)
}

func (l *Logger) Fatal(format string, v ...interface{}) {
	l.log(levelFatal, colorRed, format, v...)
}
