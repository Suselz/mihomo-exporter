package logx

import (
	"fmt"
	"log"
	"os"
	"strings"
)

type Level int

const (
	LevelDebug Level = iota
	LevelInfo
	LevelWarn
	LevelError
)

type Logger struct {
	base  *log.Logger
	level Level
}

func ParseLevel(raw string) (Level, error) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "debug":
		return LevelDebug, nil
	case "info", "":
		return LevelInfo, nil
	case "warn", "warning":
		return LevelWarn, nil
	case "error":
		return LevelError, nil
	default:
		return LevelInfo, fmt.Errorf("unknown log level %q", raw)
	}
}

func New(levelRaw string) (*Logger, error) {
	lvl, err := ParseLevel(levelRaw)
	if err != nil {
		return nil, err
	}
	return &Logger{
		base:  log.New(os.Stdout, "", log.LstdFlags),
		level: lvl,
	}, nil
}

func (l *Logger) enabled(want Level) bool {
	return want >= l.level
}

func (l *Logger) Debugf(format string, args ...any) {
	if l.enabled(LevelDebug) {
		l.base.Printf("level=DEBUG "+format, args...)
	}
}

func (l *Logger) Infof(format string, args ...any) {
	if l.enabled(LevelInfo) {
		l.base.Printf("level=INFO "+format, args...)
	}
}

func (l *Logger) Warnf(format string, args ...any) {
	if l.enabled(LevelWarn) {
		l.base.Printf("level=WARN "+format, args...)
	}
}

func (l *Logger) Errorf(format string, args ...any) {
	if l.enabled(LevelError) {
		l.base.Printf("level=ERROR "+format, args...)
	}
}

func (l *Logger) LevelString() string {
	switch l.level {
	case LevelDebug:
		return "debug"
	case LevelInfo:
		return "info"
	case LevelWarn:
		return "warn"
	case LevelError:
		return "error"
	default:
		return "info"
	}
}
