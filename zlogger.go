package Logger

import (
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

func New(level slog.Level, skipFrameCount *int, disableShowCaller bool) *Logger {
	zeroContext := zerolog.New(os.Stdout).With().Timestamp()
	var logger zerolog.Logger
	if !disableShowCaller {
		sfc := setupLogger(level, skipFrameCount)
		logger = zeroContext.CallerWithSkipFrameCount(zerolog.CallerSkipFrameCount + sfc).Logger()
	} else {
		logger = zeroContext.Logger()
	}

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
		return zerolog.New(output).With().Timestamp()
	})
	return logger
}

func (l *Logger) Err(err error) *Logger {
	subLogger := Logger{logger: l.logger}
	lg := subLogger.logger.With().Err(err).Logger()
	subLogger.logger = &lg
	return &subLogger
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
