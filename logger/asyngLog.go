package logger

import (
	"context"
	"sync"
	"time"
)

const (
	InfoLogLevel   = "info"
	InfofLogLevel  = "infof"
	InfowLogLevel  = "infow"
	DebugLogLevel  = "debug"
	DebugfLogLevel = "debugf"
	DebugwLogLevel = "debugw"
	ErrorLogLevel  = "error"
	ErrorfLogLevel = "errorf"
	ErrorwLogLevel = "errorw"
	PanicLogLevel  = "panic"
)

type AsyncMsg struct {
	level   string
	message string
	args    []any
}

// AsyncLogger is a logger that processes messages asynchronously.
type AsyncLogger struct {
	Logger   Log
	logChan  chan AsyncMsg
	wg       sync.WaitGroup
	shutdown sync.Once
}

// NewAsyncLogger creates a new instance of AsyncLogger.
func NewAsyncLogger(logger Log) *AsyncLogger {
	afl := &AsyncLogger{
		Logger:  logger,
		logChan: make(chan AsyncMsg, logger.Config.BufferSize),
	}

	afl.wg.Add(1)
	go afl.processLogs()

	return afl
}

// processLogs processes log messages asynchronously from the logChan.
func (l *AsyncLogger) processLogs() {
	defer l.wg.Done()

	// Continuously process log messages from the channel
	for msg := range l.logChan {
		switch msg.level {
		case InfoLogLevel:
			l.Logger.Info(msg.message)
		case InfofLogLevel:
			l.Logger.Infof(msg.message, msg.args...)
		case InfowLogLevel:
			l.Logger.Infow(msg.message, msg.args...)
		case DebugLogLevel:
			l.Logger.Debug(msg.message)
		case DebugfLogLevel:
			l.Logger.Debugf(msg.message, msg.args...)
		case DebugwLogLevel:
			l.Logger.Debugw(msg.message, msg.args...)
		case ErrorLogLevel:
			l.Logger.Error(msg.message)
		case ErrorfLogLevel:
			l.Logger.Errorf(msg.message, msg.args...)
		case ErrorwLogLevel:
			l.Logger.Errorw(msg.message, msg.args...)
		case PanicLogLevel:
			l.Logger.Panic(msg.message)
		}
	}
}

// BufferOverflowStrategy defines how the logger should behave when the buffer overflows.
type BufferOverflowStrategy int

const (
	// Drop drops the message when the buffer is full.
	Drop BufferOverflowStrategy = iota
	// Block blocks until there is space in the buffer.
	Block
	// Retry retries a few times before dropping the message.
	Retry
)

// logAsync safely sends a message to the log channel based on the overflow strategy.
func (l *AsyncLogger) logAsync(msg AsyncMsg) {
	select {
	case l.logChan <- msg:
	default:
		switch l.Logger.Config.OverflowStrategy {
		case Block:
			// Block until space is available in the buffer
			l.logChan <- msg
		case Retry:
			// Try several times before dropping the message
			for i := 0; i < 3; i++ {
				select {
				case l.logChan <- msg:
					return // Successfully logged the message
				case <-time.After(10 * time.Millisecond):
					// Wait before retrying
				}
			}
			// Log an error if the message is dropped
			l.Logger.Errorf("Log buffer overflow, message dropped: %s", msg.message)
		case Drop:
			// Drop the message if the buffer is full
			l.Logger.Errorf("Log buffer overflow, message dropped: %s", msg.message)
		}
	}
}

// Info logs a message at the Info level.
func (l *AsyncLogger) Info(msg string) {
	l.logAsync(AsyncMsg{level: InfoLogLevel, message: msg})
}

// Infof logs a formatted message at the Info level.
func (l *AsyncLogger) Infof(template string, args ...interface{}) {
	l.logAsync(AsyncMsg{level: InfofLogLevel, message: template, args: args})
}

// Infow logs a message at the Info level with additional key-value pairs.
func (l *AsyncLogger) Infow(msg string, args ...any) {
	l.logAsync(AsyncMsg{level: InfowLogLevel, message: msg, args: args})
}

// Debug logs a message at the Debug level.
func (l *AsyncLogger) Debug(msg string) {
	l.logAsync(AsyncMsg{level: DebugLogLevel, message: msg})
}

// Debugf logs a formatted message at the Debug level.
func (l *AsyncLogger) Debugf(template string, args ...interface{}) {
	l.logAsync(AsyncMsg{level: DebugfLogLevel, message: template, args: args})
}

// Debugw logs a message at the Debug level with additional key-value pairs.
func (l *AsyncLogger) Debugw(msg string, args ...any) {
	l.logAsync(AsyncMsg{level: DebugwLogLevel, message: msg, args: args})
}

// Error logs a message at the Error level.
func (l *AsyncLogger) Error(msg string) {
	l.logAsync(AsyncMsg{level: ErrorLogLevel, message: msg})
}

// Errorf logs a formatted message at the Error level.
func (l *AsyncLogger) Errorf(template string, args ...interface{}) {
	l.logAsync(AsyncMsg{level: ErrorfLogLevel, message: template, args: args})
}

// Errorw logs a message at the Error level with additional key-value pairs.
func (l *AsyncLogger) Errorw(msg string, args ...any) {
	l.logAsync(AsyncMsg{level: ErrorwLogLevel, message: msg, args: args})
}

// Panic logs a message at the Panic level.
func (l *AsyncLogger) Panic(msg string) {
	l.logAsync(AsyncMsg{level: PanicLogLevel, message: msg})
}

// Shutdown waits for all log messages to be processed and gracefully shuts down the logger.
func (l *AsyncLogger) Shutdown(ctx context.Context) {
	l.shutdown.Do(func() {
		// Close the log channel to stop accepting new messages
		close(l.logChan)

		done := make(chan struct{})
		// Wait for all logs to be processed
		go func() {
			l.wg.Wait()
			close(done)
		}()

		select {
		case <-done:
		case <-ctx.Done():
		}
	})
}
