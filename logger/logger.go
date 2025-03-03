package logger

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/golang/glog"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"runtime/debug"
	"strings"
)

type Logger interface {
	Info(msg string)
	Infof(msg string, args ...interface{})
	Infow(msg string, keysAndValues ...interface{})
	Debug(msg string)
	Debugf(msg string, args ...interface{})
	Debugw(msg string, keysAndValues ...interface{})
	Error(msg string)
	Errorf(msg string, args ...interface{})
	Errorw(msg string, keysAndValues ...interface{})
	Panic(msg string)
}

// BaseLog is an example of default log structure.
type BaseLog struct {
	Operation string `json:"operation"`
	Data      any    `json:"data"`
}

// Data is an example structure used in BaseLog's Data field.
type Data struct {
	FieldA string  `json:"fieldA"`
	FieldB float64 `json:"fieldB"`
	FieldC int     `json:"fieldC"`
}

const (
	RequestIDField    = "_request_id"
	DebugField        = "_debug"
	VersionField      = "_version"
	DefaultCallerSkip = 2
)

type Log struct {
	logger    *zap.SugaredLogger
	Config    Config
	loggerStd *zap.Logger
	debug     bool
}

type Fld map[string]any
type SentryFld map[string]string

func Default() *Log {
	cfgDefault := zap.Config{
		Encoding:         "json",
		Level:            zap.NewAtomicLevelAt(zapcore.DebugLevel),
		OutputPaths:      []string{"stderr"},
		ErrorOutputPaths: []string{"stderr"},
		EncoderConfig: zapcore.EncoderConfig{
			MessageKey: "message",

			LevelKey:    "level",
			EncodeLevel: zapcore.CapitalColorLevelEncoder,

			TimeKey:    "time",
			EncodeTime: zapcore.ISO8601TimeEncoder,

			CallerKey:    "caller",
			EncodeCaller: zapcore.ShortCallerEncoder,
		},
	}
	logger, err := cfgDefault.Build()
	if err != nil {
		panic(err)
	}

	logger = zap.New(logger.Core(), zap.AddCaller(), zap.AddCallerSkip(DefaultCallerSkip))

	return &Log{
		logger:    logger.Sugar(),
		loggerStd: logger,
		Config: Config{
			ContextLogFields: []string{RequestIDField},
		},
	}
}

func New(cfg Config) *Log {
	l := Default()
	l.Config = cfg
	l.Config.ContextLogFields = addStr(l.Config.ContextLogFields, RequestIDField)

	l.logger = zap.New(l.logger.Desugar().Core(), zap.AddCaller(), zap.AddCallerSkip(cfg.CallerSkip)).Sugar()
	levelLog := l.logger.Level()
	err := levelLog.Set(cfg.LogLevel)
	if err != nil {
		return nil
	}

	return l
}

func (l *Log) ToAsync() *AsyncLogger {
	return NewAsyncLogger(*l)
}

// handleLogging handles the logging logic with optional formatting.
func (l *Log) handleLogging(logFunc func(string, ...any), msg string, args ...any) {
	if l.Config.IsFormatted {
		msg += l.beautify(args...)
		logFunc(msg)
		return
	}
	logFunc(msg, args...)
}

// beautify formats the arguments into a readable string.
func (l *Log) beautify(args ...any) string {
	const op = "logger.beautify"

	var output []string

	for _, arg := range args {
		switch v := arg.(type) {
		case string:
			output = append(output, v)
		case fmt.Stringer:
			output = append(output, v.String())
		default:
			jsonData, err := json.MarshalIndent(v, "", "  ")
			if err != nil {
				logErr := errors.Join(
					fmt.Errorf("%s: failed to marshal log", op),
					MarshalErr,
					err,
					fmt.Errorf("type: %T, value: %+v", arg, arg),
				).Error()
				output = append(output, logErr)
			} else {
				output = append(output, string(jsonData))
			}

		}
	}

	return strings.Join(output, " ")
}

// Info logs a message at the Info level.
func (l *Log) Info(msg string) {
	l.handleLogging(l.logger.Infof, msg)
}

// Infof logs a formatted message at the Info level.
func (l *Log) Infof(template string, args ...interface{}) {
	l.handleLogging(l.logger.Infof, template, args...)
}

// Infow logs a structured message at the Info level.
func (l *Log) Infow(msg string, keysAndValues ...interface{}) {
	if l.Config.IsFormatted {
		msg += l.beautify(keysAndValues...)
		l.logger.Info(msg)
	} else {
		l.logger.Infow(msg, keysAndValues...)
	}
}

// Debug logs a message at the Debug level.
func (l *Log) Debug(msg string) {
	l.handleLogging(l.logger.Debugf, msg)
}

// Debugf logs a formatted message at the Debug level.
func (l *Log) Debugf(template string, args ...interface{}) {
	l.handleLogging(l.logger.Debugf, template, args...)
}

// Debugw logs a structured message at the Debug level.
func (l *Log) Debugw(msg string, keysAndValues ...interface{}) {
	if l.Config.IsFormatted {
		msg += l.beautify(keysAndValues...)
		l.logger.Debug(msg)
	} else {
		l.logger.Debugw(msg, keysAndValues...)
	}
}

// Error logs a message at the Error level.
func (l *Log) Error(msg string) {
	l.handleLogging(l.logger.Errorf, msg)
}

// Errorf logs a formatted message at the Error level.
func (l *Log) Errorf(template string, args ...interface{}) {
	l.handleLogging(l.logger.Errorf, template, args...)
}

// Errorw logs a structured message at the Error level.
func (l *Log) Errorw(msg string, keysAndValues ...interface{}) {
	if l.Config.IsFormatted {
		msg += l.beautify(keysAndValues...)
		l.logger.Error(msg)
	} else {
		l.logger.Errorw(msg, keysAndValues...)
	}
}

// Panic logs a message at the Panic level and then panics.
func (l *Log) Panic(msg string, args ...any) {
	l.handleLogging(l.logger.Panicf, msg, args...)
}

func (l *Log) Close() {
	_ = l.loggerStd.Sync()
}

func (l *Log) LogGRPC(ctx context.Context, lvl logging.Level, msg string, fields ...any) {
	switch lvl {
	case logging.LevelDebug:
		l.logger.Debugw(msg, fields...)
	case logging.LevelInfo:
		l.logger.Infow(msg, fields...)
	case logging.LevelWarn:
		l.logger.Warnw(msg, fields...)
	case logging.LevelError:
		l.logger.Errorw(msg, fields...)
	default:
		panic(fmt.Sprintf("unknown level %v", lvl))
	}
}

func (l *Log) With(fld Fld) *Log {
	fields := make([]interface{}, 0)
	for k, v := range fld {
		fields = append(fields, k, v)
	}
	return l.copyWithEntry(*l.logger.With(fields...))
}

func (l *Log) WithField(key string, val interface{}) *Log {
	return l.copyWithEntry(*l.logger).With(Fld{key: val})
}

func (l *Log) WithErr(err error) *Log {
	fields := Fld{}
	if e, ok := err.(errWithFields); ok {
		fields["error"] = e.Origin()
		for k, v := range e.Fields() {
			fields[k] = v
		}
	} else {
		fields["error"] = err
	}

	return l.copyWithEntry(*l.logger).With(fields)
}

func (l *Log) ErrWithError(ctx context.Context, err error, msg string) {
	l.WithCtx(ctx).WithErr(err).Error(msg)
}

func (l *Log) ErrWithErrorf(ctx context.Context, err error, msg string, args ...interface{}) {
	l.WithCtx(ctx).WithErr(err).Errorf(msg, args...)
}

func (l *Log) ErrWithErrorw(ctx context.Context, err error, msg string, keysAndValues ...interface{}) {
	l.WithCtx(ctx).WithErr(err).Errorw(msg, keysAndValues...)
}

func (l *Log) WithCtx(ctx context.Context) *Log {
	fields := Fld{}
	for _, key := range l.Config.ContextLogFields {
		v := ctx.Value(key)
		if v != nil {
			fields[key] = v
		}
	}

	copied := l.copyWithEntry(*l.logger).With(fields)
	if ctx.Value(DebugField) != nil {
		copied.debug = true
	}

	return copied
}

func (l *Log) copyWithEntry(entry zap.SugaredLogger) *Log {
	return &Log{
		logger: &entry,
		Config: l.Config,
	}
}

func LogPanic(recovered interface{}) { // nolint: revive
	if recovered == nil {
		return
	}
	glog.Errorf("Panic recovered: %s %s", recovered, string(debug.Stack()))
}

func (l *Log) LogPanic(recovered interface{}) {
	if recovered == nil {
		return
	}
	l.Errorf("Panic recovered: %s %s", recovered, string(debug.Stack()))
}

func (l *Log) Log(ctx context.Context, level int, msg string, fields ...interface{}) {
	l.WithCtx(ctx).logger.Logf(zapcore.Level(level), msg, fields...)
}

func addStr(ss []string, s string) []string {
	for i := range ss {
		if ss[i] == s {
			return ss
		}
	}

	return append(ss, s)
}
