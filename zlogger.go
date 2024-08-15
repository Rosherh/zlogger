package Logger

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
)

const (
	defaultSkipFrameCount = 3
)

type Logger struct {
	logger *zerolog.Logger
}

type ZeroLogger func(int) zerolog.Context

type Interface interface {
	ApplyContext(ctx context.Context) *Logger
	ApplyRequest(ctx context.Context) *Logger
	ApplyResponse(ctx context.Context) *Logger
	Err(err error) *Logger
	Debugf(message string, args ...any)
	Infof(message string, args ...any)
	Warnf(message string, args ...any)
	Errorf(message string, args ...any)
	Fatalf(message string, args ...any)
}

func setupLogger(level slog.Level, skipFrameCount *int) int {
	sfc := defaultSkipFrameCount
	if skipFrameCount != nil {
		if *skipFrameCount < 0 {
			sfc = defaultSkipFrameCount
		} else {
			sfc = *skipFrameCount
		}
	}
	var l zerolog.Level

	switch level {
	case slog.LevelError:
		l = zerolog.ErrorLevel
	case slog.LevelWarn:
		l = zerolog.WarnLevel
	case slog.LevelInfo:
		l = zerolog.InfoLevel
	case slog.LevelDebug:
		l = zerolog.DebugLevel
	default:
		l = zerolog.InfoLevel
	}

	zerolog.SetGlobalLevel(l)
	return sfc
}

func New(level slog.Level, skipFrameCount *int) *Logger {
	sfc := setupLogger(level, skipFrameCount)
	logger := zerolog.New(os.Stdout).With().Timestamp().CallerWithSkipFrameCount(zerolog.CallerSkipFrameCount + sfc).Logger()

	return &Logger{
		logger: &logger,
	}
}

func NewWithCustomLogger(level slog.Level, skipFrameCount *int, fn ZeroLogger) *Logger {
	sfc := setupLogger(level, skipFrameCount)

	logger := fn(sfc).Logger()
	return &Logger{
		logger: &logger,
	}
}

func NewPrettyLogger(level slog.Level, skipFrameCount *int) *Logger {
	output := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}
	output.FormatLevel = func(i interface{}) string {
		return strings.ToUpper(fmt.Sprintf("| %-6s|", i))
	}
	output.FormatMessage = func(i interface{}) string {
		return fmt.Sprintf("%s", i)
	}
	output.FormatFieldName = func(i interface{}) string {
		return fmt.Sprintf("%s:", i)
	}
	output.FormatFieldValue = func(i interface{}) string {
		return strings.ToUpper(fmt.Sprintf("%s", i))
	}

	logger := NewWithCustomLogger(level, skipFrameCount, func(sfc int) zerolog.Context {
		return zerolog.New(output).With().Timestamp().CallerWithSkipFrameCount(zerolog.CallerSkipFrameCount + sfc)
	})
	return logger
}

func (l *Logger) Err(err error) *Logger {
	subLogger := Logger{logger: l.logger}
	lg := subLogger.logger.With().Err(err).Logger()
	subLogger.logger = &lg
	return &subLogger
}

func (l *Logger) ApplyRequest(ctx context.Context) *Logger {
	subLogger := Logger{logger: l.logger}
	logContext := subLogger.logger.With()

	if val, ok := getValStr(ctx, "Method"); ok {
		logContext = logContext.Str("method", val)
	}
	if val, ok := getValStr(ctx, "RequestURI"); ok {
		logContext = logContext.Str("req_uri", val)
	}
	if val, ok := getValStr(ctx, "ReqHeader"); ok {
		logContext = logContext.Str("req_header", val)
	}
	if val, ok := getValStr(ctx, "ReqBody"); ok {
		logContext = logContext.Str("req_body", val)
	}
	if val, ok := getValStr(ctx, "Time"); ok {
		logContext = logContext.Str("time", val)
	}

	lg := logContext.Logger()
	subLogger.logger = &lg
	return &subLogger
}

func (l *Logger) ApplyResponse(ctx context.Context) *Logger {
	subLogger := Logger{logger: l.logger}
	logContext := subLogger.logger.With()

	if val, ok := getValStr(ctx, "RespHeader"); ok {
		logContext = logContext.Str("resp_header", val)
	}
	if val, ok := getValStr(ctx, "RespBody"); ok {
		logContext = logContext.Str("resp_body", val)
	}
	if val, ok := getValInt(ctx, "StatusCode"); ok {
		logContext = logContext.Int("status_code", val)
	}
	lg := logContext.Logger()
	subLogger.logger = &lg
	return &subLogger
}

func (l *Logger) ApplyContext(ctx context.Context) *Logger {
	subLogger := Logger{logger: l.logger}
	logContext := subLogger.logger.With()

	if val, ok := getValStr(ctx, "Tag"); ok {
		logContext = logContext.Str("tag", val)
	}
	if val, ok := getValStr(ctx, "DocumentId"); ok {
		logContext = logContext.Str("document_id", val)
	}
	if val, ok := getValStr(ctx, "ReqId"); ok {
		logContext = logContext.Str("req_id", val)
	}
	if val, ok := getValStr(ctx, "XReqId"); ok {
		logContext = logContext.Str("x_req_id", val)
	}

	lg := logContext.Logger()
	subLogger.logger = &lg
	return &subLogger
}

func getValStr(c context.Context, key string) (string, bool) {
	if valStr, ok := c.Value(key).(string); ok && valStr != "" {
		return valStr, true
	}
	return "", false
}

func getValInt(c context.Context, key string) (int, bool) {
	if valInt, ok := c.Value(key).(int); ok {
		return valInt, true
	}
	return 0, false
}

func (l *Logger) Debugf(message string, args ...any) {
	l.logger.Debug().Msgf(message, args...)
}

func (l *Logger) Infof(message string, args ...any) {
	l.logger.Info().Msgf(message, args...)
}

func (l *Logger) Warnf(message string, args ...any) {
	l.logger.Warn().Int("severity", 400).Msgf(message, args...)
}

func (l *Logger) Errorf(message string, args ...any) {
	l.logger.Error().Int("severity", 500).Msgf(message, args...)
}

func (l *Logger) Fatalf(message string, args ...any) {
	l.logger.Fatal().Int("severity", 800).Msgf(message, args...)

	os.Exit(1)
}
