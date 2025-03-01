package logger

import (
	"sync"
)

const (
	InfoLogLevel   = "info"
	InfowLogLevel  = "infow"
	DebugLogLevel  = "debug"
	DebugwLogLevel = "debugw"
	ErrorLogLevel  = "error"
	ErrorwLogLevel = "errorw"
	PanicLogLevel  = "panic"
)

type AsyncMsg struct {
	level   string
	message string
	args    []any
}

type AsyncLogger struct {
	Logger   Log
	logChan  chan AsyncMsg
	wg       sync.WaitGroup
	shutdown sync.Once
}

// NewAsyncLogger creates a new instance of AsyncLogger.
// Takes a logger instance and a buffer size for the log channel.
func NewAsyncLogger(logger Log) *AsyncLogger {
	afl := &AsyncLogger{
		Logger:  logger,
		logChan: make(chan AsyncMsg, logger.Config.BufferSize),
	}

	afl.wg.Add(1)
	go afl.processLogs()

	return afl
}

// processLogs processes log messages asynchronously.
// It reads messages from the channel and sends them to the appropriate logging method.
func (l *AsyncLogger) processLogs() {
	defer l.wg.Done()

	for msg := range l.logChan {
		switch msg.level {
		case InfoLogLevel:
			l.Logger.Info(msg.message)
		case InfowLogLevel:
			l.Logger.Infow(msg.message, msg.args...)
		case DebugLogLevel:
			l.Logger.Debug(msg.message)
		case DebugwLogLevel:
			l.Logger.Debugw(msg.message, msg.args...)
		case ErrorLogLevel:
			l.Logger.Error(msg.message)
		case ErrorwLogLevel:
			l.Logger.Errorw(msg.message, msg.args...)
		case PanicLogLevel:
			l.Logger.Panic(msg.message)
		}
	}
}

// logAsync safely sends a message to the log channel.
// If the channel is closed, it prevents panic.
func (l *AsyncLogger) logAsync(msg AsyncMsg) {
	select {
	case l.logChan <- msg:
	default:
		// Avoid blocking if the channel is full
	}
}

// Info logs a message at the Info level.
func (l *AsyncLogger) Info(msg string) {
	l.logAsync(AsyncMsg{level: InfoLogLevel, message: msg})
}

// Infow logs a message at the Info level with additional key-value pairs.
func (l *AsyncLogger) Infow(msg string, args ...any) {
	l.logAsync(AsyncMsg{level: InfowLogLevel, message: msg, args: args})
}

// Debug logs a message at the Debug level.
func (l *AsyncLogger) Debug(msg string) {
	l.logAsync(AsyncMsg{level: DebugLogLevel, message: msg})
}

// Debugw logs a message at the Debug level with additional key-value pairs.
func (l *AsyncLogger) Debugw(msg string, args ...any) {
	l.logAsync(AsyncMsg{level: DebugwLogLevel, message: msg, args: args})
}

// Error logs a message at the Error level.
func (l *AsyncLogger) Error(msg string) {
	l.logAsync(AsyncMsg{level: ErrorLogLevel, message: msg})
}

// Errorw logs a message at the Error level with additional key-value pairs.
func (l *AsyncLogger) Errorw(msg string, args ...any) {
	l.logAsync(AsyncMsg{level: ErrorwLogLevel, message: msg, args: args})
}

// Panic logs a message at the Panic level.
func (l *AsyncLogger) Panic(msg string) {
	l.logAsync(AsyncMsg{level: PanicLogLevel, message: msg})
}

// Shutdown shuts down the logger, waiting for all log messages to be processed.
func (l *AsyncLogger) Shutdown() {
	l.shutdown.Do(func() {
		close(l.logChan)
		l.wg.Wait()
	})
}
