package logger

import (
	"fmt"
	"time"
)

/*
	0 TRACE / ALL
	1 DEBUG
	2 INFO
	3 WARNING
	4 ERROR
 */
type Logger struct {
	date_format string
	level int
}

func NewDefaultLogger() (*Logger) {
	return NewLogger("INFO")
}

func NewLogger(level string) (*Logger) {
	return &Logger{
		date_format: "2006-01-02 15:04:05",
		level: LevelToInt(level),
	}
}

func GetDefaultLogLevel() (int) {
	return 2
}

func LevelToString(level int) (string) {
	switch level {
	case 0: return "TRACE"
	case 1: return "DEBUG"
	case 2: return "INFO"
	case 3: return "WARN"
	case 4: return "ERROR"
	}
	return "UNKNOWN"
}

func LevelToInt(level string) (int) {
	switch level {
	case "TRACE": return 0
	case "DEBUG": return 1
	case "INFO": return 2
	case "WARN": return 3
	case "ERROR": return 4
	}
	//l.Warn("Unknown log level specified: %s, falling back to default: %s", level, LevelToString(GetDefaultLogLevel()))
	return 2
}

func (l *Logger) SetLevel(level int) {
	l.level = level
}



func (l *Logger) GetLevel() (int) {
	return l.level
}

func (l *Logger) GetLevelAsString() (string) {
	return LevelToString(l.level)
}

func (l *Logger) Info(_fmt string, args ...interface{}) {
	l.message(2, _fmt, args...)
}

func (l *Logger) Warn(_fmt string, args ...interface{}) {
	l.message(3, _fmt, args...)
}

func (l *Logger) Error(_fmt string, args ...interface{}) {
	l.message(4, _fmt, args...)
}

func (l *Logger) Debug(_fmt string, args ...interface{}) {
	l.message(1, _fmt, args...)
}

func (l *Logger) Trace(_fmt string, args ...interface{}) {
	l.message(0, _fmt, args...)
}

func (l *Logger) message(level int, _fmt string, args ...interface{}) {
	if level >= l.level {
		lblString := ""
		switch level {
		case 0: lblString = "TRACE"
		case 1: lblString = "DEBUG"
		case 2: lblString = "INFO"
		case 3: lblString = "WARN"
		case 4: lblString = "ERROR"
		}
		msg := fmt.Sprintf(_fmt, args...)
		fmt.Printf("%s [%5s] %s\n", time.Now().Format(l.date_format), lblString, msg)
	}
}
