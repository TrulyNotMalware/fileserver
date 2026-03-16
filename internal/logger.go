package internal

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

type Level int

const (
	LevelDebug Level = iota
	LevelInfo
	LevelWarn
	LevelError
)

func (l Level) String() string {
	switch l {
	case LevelDebug:
		return "DEBUG"
	case LevelInfo:
		return "INFO"
	case LevelWarn:
		return "WARN"
	case LevelError:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

type Logger struct {
	level  Level
	output io.Writer
}

func New(level Level) *Logger {
	return &Logger{
		level:  level,
		output: os.Stdout,
	}
}

func (l *Logger) log(level Level, msg string) {
	if level < l.level {
		return
	}
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	_, err := fmt.Fprintf(l.output, "%s [%s] %s\n", timestamp, level, msg)
	if err != nil {
		return
	}
}

func (l *Logger) Debug(msg string) { l.log(LevelDebug, msg) }
func (l *Logger) Info(msg string)  { l.log(LevelInfo, msg) }
func (l *Logger) Warn(msg string)  { l.log(LevelWarn, msg) }
func (l *Logger) Error(msg string) { l.log(LevelError, msg) }

func (l *Logger) Debugf(format string, args ...any) { l.log(LevelDebug, fmt.Sprintf(format, args...)) }
func (l *Logger) Infof(format string, args ...any)  { l.log(LevelInfo, fmt.Sprintf(format, args...)) }
func (l *Logger) Warnf(format string, args ...any)  { l.log(LevelWarn, fmt.Sprintf(format, args...)) }
func (l *Logger) Errorf(format string, args ...any) { l.log(LevelError, fmt.Sprintf(format, args...)) }

func ParseLevel(s string) (Level, error) {
	switch strings.ToUpper(s) {
	case "DEBUG":
		return LevelDebug, nil
	case "INFO":
		return LevelInfo, nil
	case "WARN":
		return LevelWarn, nil
	case "ERROR":
		return LevelError, nil
	default:
		return LevelInfo, fmt.Errorf("unknown log level %q: must be one of [DEBUG, INFO, WARN, ERROR]", s)
	}
}

// global logger instance
var std = New(LevelInfo)

// Init initializes the global logger with the given level.
func Init(level Level) {
	std = New(level)
}

func Debug(msg string) { std.Debug(msg) }
func Info(msg string)  { std.Info(msg) }
func Warn(msg string)  { std.Warn(msg) }
func Error(msg string) { std.Error(msg) }

func Debugf(format string, args ...any) { std.Debugf(format, args...) }
func Infof(format string, args ...any)  { std.Infof(format, args...) }
func Warnf(format string, args ...any)  { std.Warnf(format, args...) }
func Errorf(format string, args ...any) { std.Errorf(format, args...) }
