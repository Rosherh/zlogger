package Logger

import (
	"context"
	log "log/slog"
	"os"

	zl "github.com/rs/zerolog"
)

const (
	defaultSkipFrameCount = 3
)

type Logger struct {
	logger *zl.Logger
}

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

func New(level log.Level, skipFrameCount *int) *Logger {
	sfc := defaultSkipFrameCount
	if skipFrameCount != nil {
		if *skipFrameCount < 0 {
			sfc = defaultSkipFrameCount
		} else {
			sfc = *skipFrameCount
		}
	}
	var l zl.Level

	switch level {
	case log.LevelError:
		l = zl.ErrorLevel
	case log.LevelWarn:
		l = zl.WarnLevel
	case log.LevelInfo:
		l = zl.InfoLevel
	case log.LevelDebug:
		l = zl.DebugLevel
	default:
		l = zl.InfoLevel
	}

	zl.SetGlobalLevel(l)

	logger := zl.New(os.Stdout).With().Timestamp().CallerWithSkipFrameCount(zl.CallerSkipFrameCount + sfc).Logger()

	return &Logger{
		logger: &logger,
	}
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
