package logger

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
	"universum/config"
)

func setupLoggerTests(t *testing.T) {
	resetLoggerForTesting()
	config.Store = config.GetSkeleton()
	config.Store.Logging.MinimumLogLevel = config.LogLevelInfo

	tempDir := t.TempDir()
	config.Store.Logging.LogFileDirectory = tempDir
}

func redirectOutput() (*os.File, *os.File) {
	originalStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	return originalStdout, r
}

func restoreOutput(originalStdout *os.File, r *os.File) string {
	w := os.Stdout
	w.Close()
	os.Stdout = originalStdout
	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String()
}

func TestLoggerInitialization(t *testing.T) {
	setupLoggerTests(t)

	logger := Get()

	if logger == nil {
		t.Fatalf("Expected logger instance, got nil")
	}

	if logger.logLevel != levelCodeInfo {
		t.Errorf("Expected log level %d, got %d", levelCodeInfo, logger.logLevel)
	}

	if logger.consoleLogger == nil || logger.fileLogger == nil {
		t.Errorf("Expected console and file loggers to be initialized")
	}
}

func TestLoggerLogLevels(t *testing.T) {
	setupLoggerTests(t)
	config.Store.Logging.MinimumLogLevel = config.LogLevelWarn

	originalStdout, stdoutReader := redirectOutput()
	defer func() { os.Stdout = originalStdout }()

	logger := Get()

	logger.messageQueue = make(chan *logMessage, 1000)

	logger.Debug("This is a DEBUG message")
	logger.Info("This is an INFO message")
	logger.Warn("This is a WARN message")
	logger.Error("This is an ERROR message")
	logger.Fatal("This is a FATAL message")

	time.Sleep(logger.batchInterval + 50*time.Millisecond)

	consoleOutput := restoreOutput(originalStdout, stdoutReader)

	if strings.Contains(consoleOutput, "DEBUG") || strings.Contains(consoleOutput, "INFO") {
		t.Errorf("Expected DEBUG and INFO messages to be filtered out")
	}

	if !strings.Contains(consoleOutput, "WARN") ||
		!strings.Contains(consoleOutput, "ERROR") ||
		!strings.Contains(consoleOutput, "FATAL") {
		t.Errorf("Expected WARN, ERROR, and FATAL messages to be present")
	}
}

func TestLoggerBatchWriting(t *testing.T) {
	setupLoggerTests(t)

	logger := Get()
	logger.batchSize = 5

	var wg sync.WaitGroup

	numMessages := 15
	wg.Add(numMessages)

	for i := 0; i < numMessages; i++ {
		go func(i int) {
			defer wg.Done()
			logger.Info("Message %d", i)
		}(i)
	}

	wg.Wait()

	time.Sleep(logger.batchInterval + 50*time.Millisecond)

	logFilePath := filepath.Join(config.Store.Logging.LogFileDirectory, config.DefaultServerLogFile)
	data, err := os.ReadFile(logFilePath)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	logContents := string(data)

	for i := 0; i < numMessages; i++ {
		expectedMessage := fmt.Sprintf("Message %d", i)
		if !strings.Contains(logContents, expectedMessage) {
			t.Errorf("Expected log file to contain '%s'", expectedMessage)
		}
	}
}

func TestLoggerConsoleOutput(t *testing.T) {
	setupLoggerTests(t)

	originalStdout, stdoutReader := redirectOutput()
	defer func() { os.Stdout = originalStdout }()

	logger := Get()
	logger.Info("Test console output")

	time.Sleep(logger.batchInterval + 50*time.Millisecond)
	consoleOutput := restoreOutput(originalStdout, stdoutReader)

	if !strings.Contains(consoleOutput, "Test console output") {
		t.Errorf("Expected console output to contain 'Test console output'")
	}

	if !strings.Contains(consoleOutput, colorGreen) {
		t.Errorf("Expected console output to contain green color code")
	}
}

func TestLoggerFileOutput(t *testing.T) {
	setupLoggerTests(t)

	logger := Get()

	logger.Info("Test file output")
	time.Sleep(logger.batchInterval + 50*time.Millisecond)

	logFilePath := filepath.Join(config.Store.Logging.LogFileDirectory, config.DefaultServerLogFile)
	data, err := os.ReadFile(logFilePath)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	logContents := string(data)

	if !strings.Contains(logContents, "Test file output") {
		t.Errorf("Expected log file to contain 'Test file output'")
	}

	if strings.Contains(logContents, colorGreen) {
		t.Errorf("Expected log file output to not contain color codes")
	}
}

func TestLoggerShutdown(t *testing.T) {
	setupLoggerTests(t)

	logger := Get()
	logger.Info("Message before shutdown")
	logger.Info("Message before shutdown")
	logger.Info("Message before shutdown")

	logger.Close()

	logFilePath := filepath.Join(config.Store.Logging.LogFileDirectory, config.DefaultServerLogFile)
	data, err := os.ReadFile(logFilePath)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	logContents := string(data)

	if !strings.Contains(logContents, "Message before shutdown") {
		t.Errorf("Expected log file to contain 'Message before shutdown' after shutdown")
	}
}

func TestLoggerQueueFull(t *testing.T) {
	setupLoggerTests(t)

	logger := Get()
	logger.messageQueue = make(chan *logMessage, 5)

	for i := 0; i < 10; i++ {
		logger.Info("Message %d", i)
	}

	time.Sleep(logger.batchInterval + 50*time.Millisecond)

	logFilePath := filepath.Join(config.Store.Logging.LogFileDirectory, config.DefaultServerLogFile)
	data, err := os.ReadFile(logFilePath)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	logContents := string(data)

	messageCount := 0
	for i := 0; i < 10; i++ {
		expectedMessage := fmt.Sprintf("Message %d", i)
		if strings.Contains(logContents, expectedMessage) {
			messageCount++
		}
	}

	if messageCount < 5 {
		t.Errorf("Expected at least 5 messages in the log file, got %d", messageCount)
	}

	if messageCount > 5 {
		t.Logf("Some messages were dropped due to full queue")
	}
}

func TestLoggerBatchInterval(t *testing.T) {
	setupLoggerTests(t)

	logger := Get()
	logger.batchSize = 100
	logger.batchInterval = 200 * time.Millisecond

	logger.Info("Batch interval test message 1")
	logger.Info("Batch interval test message 2")

	time.Sleep(logger.batchInterval + 50*time.Millisecond)

	logFilePath := filepath.Join(config.Store.Logging.LogFileDirectory, config.DefaultServerLogFile)
	data, err := os.ReadFile(logFilePath)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	logContents := string(data)

	if !strings.Contains(logContents, "Batch interval test message 1") ||
		!strings.Contains(logContents, "Batch interval test message 2") {
		t.Errorf("Expected messages to be flushed after batch interval")
	}
}
