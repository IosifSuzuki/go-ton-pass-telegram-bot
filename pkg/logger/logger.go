package logger

import (
	"errors"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Level int8

const (
	// LevelDebug logs are typically voluminous, and are usually disabled in
	// production.
	LevelDebug Level = iota - 1
	// LevelInfo is the default logging priority.
	LevelInfo
	// LevelWarn logs are more important than Info, but don't need individual
	// human review.
	LevelWarn
	// LevelError logs are high-priority. If an application is running smoothly,
	// it shouldn't generate any error-level logs.
	LevelError
	// LevelPanic logs a message, then panics.
	LevelPanic
	// LevelFatal logs a message, then calls os.Exit(1).
	LevelFatal
)

func (l *Level) FromString(lvl string) error {
	switch strings.TrimSpace(strings.ToLower(lvl)) {
	case "debug":
		*l = LevelDebug
	case "info":
		*l = LevelInfo
	case "warn":
		*l = LevelWarn
	case "error":
		*l = LevelError
	case "panic":
		*l = LevelPanic
	case "fatal":
		*l = LevelFatal
	default:
		return errors.New("invalid log level")
	}
	return nil
}

type ENV int8

const (
	DEV ENV = iota
	PROD
)

func (e *ENV) FromString(env string) error {
	switch strings.TrimSpace(strings.ToLower(env)) {
	case "prod":
		*e = PROD
	case "dev":
		*e = DEV
	default:
		return errors.New("invalid log environment")
	}
	return nil
}

type Logger interface {
	Debug(msg string, fields ...Field)
	Info(msg string, fields ...Field)
	Warn(msg string, fields ...Field)
	Error(msg string, fields ...Field)
	Panic(msg string, fields ...Field)
	Fatal(msg string, fields ...Field)
	Log(lvl Level, msg string, fields ...Field)
	Flush() error
}

type Field struct {
	Key string
	Val interface{}
}

func F(key string, val interface{}) Field {
	return Field{Key: key, Val: val}
}

func FError(e error) Field {
	return Field{Key: "error", Val: e}
}

type logger struct {
	lg *zap.Logger
}

func (l logger) Debug(msg string, fields ...Field) {
	l.Log(LevelDebug, msg, fields...)
}

func (l logger) Info(msg string, fields ...Field) {
	l.Log(LevelInfo, msg, fields...)
}

func (l logger) Warn(msg string, fields ...Field) {
	l.Log(LevelWarn, msg, fields...)
}

func (l logger) Error(msg string, fields ...Field) {
	l.Log(LevelError, msg, fields...)
}

func (l logger) Panic(msg string, fields ...Field) {
	l.Log(LevelPanic, msg, fields...)
}

func (l logger) Fatal(msg string, fields ...Field) {
	l.Log(LevelFatal, msg, fields...)
}

func (l logger) Flush() error {
	return l.lg.Sync()
}

func (l logger) Log(lvl Level, msg string, fields ...Field) {
	zapFields := make([]zap.Field, len(fields))
	for i, field := range fields {
		zapFields[i] = zap.Any(field.Key, field.Val)
	}
	l.lg.Log(zapcore.Level(lvl), msg, zapFields...)
}

func NewLogger(env ENV, level Level) Logger {
	lgCfg := zap.NewProductionConfig()
	if env == DEV {
		lgCfg = zap.NewDevelopmentConfig()
	}
	lgCfg.Level.SetLevel(zapcore.Level(level))
	lg, _ := lgCfg.Build()
	return logger{lg: lg}
}
