package logger

import (
	"encoding/json"
	"fmt"
	"sync"
)

// BaseLog is the default log structure.
type BaseLog struct {
	Operation string `json:"operation"`
	Data      any    `json:"data"`
}

// Data is an example structure used in BaseLog.
type Data struct {
	FieldA string  `json:"fieldA"`
	FieldB float64 `json:"fieldB"`
	FieldC int     `json:"fieldC"`
}

type AsyncFormattedLogger struct {
	logger  Logger
	logChan chan logMessage
	wg      sync.WaitGroup
}

type logMessage struct {
	level string
	log   BaseLog
}

// NewAsyncFormattedLogger creates a new instance of AsyncFormattedLogger.
// Takes a logger instance and a buffer size for the log channel.
func NewAsyncFormattedLogger(logger Logger, config Config) *AsyncFormattedLogger {
	afl := &AsyncFormattedLogger{
		logger:  logger,
		logChan: make(chan logMessage, config.BufferSize),
	}

	afl.wg.Add(1)
	go afl.processLogs()

	return afl
}

// processLogs processes the log messages asynchronously.
// It reads messages from the channel and sends them to the appropriate logging method.
func (l *AsyncFormattedLogger) processLogs() {
	defer l.wg.Done()

	for msg := range l.logChan {
		formattedMsg := l.beautify(msg.log)
		switch msg.level {
		case "info":
			l.logger.Infow(formattedMsg)
		case "debug":
			l.logger.Debugw(formattedMsg)
		case "error":
			l.logger.Errorw(formattedMsg)
		case "panic":
			l.logger.Panic(formattedMsg)
		}
	}
}

// beautify formats the BaseLog into a nicely indented JSON string.
func (l *AsyncFormattedLogger) beautify(log BaseLog) string {
	// Serialize BaseLog into JSON with indentation
	jsonData, err := json.MarshalIndent(log, "", "  ")
	if err != nil {
		return fmt.Sprintf("Failed to marshal log: %v", err)
	}

	return string(jsonData)
}

// Info logs an informational message.
func (l *AsyncFormattedLogger) Info(log BaseLog) {
	l.logChan <- logMessage{level: "info", log: log}
}

// Debug logs a debug message.
func (l *AsyncFormattedLogger) Debug(log BaseLog) {
	l.logChan <- logMessage{level: "debug", log: log}
}

// Error logs an error message.
func (l *AsyncFormattedLogger) Error(log BaseLog) {
	l.logChan <- logMessage{level: "error", log: log}
}

// Panic logs a message and triggers a panic.
func (l *AsyncFormattedLogger) Panic(log BaseLog) {
	l.logChan <- logMessage{level: "panic", log: log}
}

// Shutdown shuts down the logger, waiting for all log messages to be processed.
func (l *AsyncFormattedLogger) Shutdown() {
	close(l.logChan)
	l.wg.Wait()
}
