package log

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"gottb.io/goru/config"
	"gottb.io/goru/errors"
)

type LogLevel int

func (lv LogLevel) String() string {
	switch lv {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARNING:
		return "WARNING"
	case ERROR:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

const (
	DEBUG LogLevel = iota
	INFO
	WARNING
	ERROR
)

type logMeta int

const (
	metaLevel logMeta = iota
	metaDate
	metaTime
	metaShortfile
	metaLongfile
	metaMessage
	metaUnknown
)

type Logger struct {
	format string
	level  LogLevel
	trace  bool
	writer Writer
	metas  []logMeta
}

func NewLogger(format string, level LogLevel, trace bool, writer Writer) *Logger {
	l := &Logger{format, level, trace, writer, []logMeta{}}
	l.setMeta()
	return l
}

func (l *Logger) SetFormat(format string) {
	l.format = format
	l.setMeta()
}

func (l *Logger) SetLevel(level LogLevel) {
	l.level = level
}

func (l *Logger) SetTrace(trace bool) {
	l.trace = trace
}

func (l *Logger) SetWriter(writer Writer) {
	l.writer = writer
}

func (l *Logger) Output(callDepth int, v ...interface{}) {
	l.write(DEBUG, callDepth, fmt.Sprint(v...))
}

func (l *Logger) Debug(v ...interface{}) {
	l.write(DEBUG, 0, fmt.Sprint(v...))
}

func (l *Logger) Debugf(format string, v ...interface{}) {
	l.write(DEBUG, 0, fmt.Sprintf(format, v...))
}

func (l *Logger) Info(v ...interface{}) {
	l.write(INFO, 0, fmt.Sprint(v...))
}

func (l *Logger) Infof(format string, v ...interface{}) {
	l.write(INFO, 0, fmt.Sprintf(format, v...))
}

func (l *Logger) Warn(v ...interface{}) {
	l.write(WARNING, 0, fmt.Sprint(v...))
}

func (l *Logger) Warnf(format string, v ...interface{}) {
	l.write(WARNING, 0, fmt.Sprintf(format, v...))
}

func (l *Logger) Error(v ...interface{}) {
	l.write(ERROR, 0, fmt.Sprint(v...))
}

func (l *Logger) Errorf(format string, v ...interface{}) {
	l.write(ERROR, 0, fmt.Sprintf(format, v...))
}

func (l *Logger) Trace(err error) {
	if !l.trace {
		l.write(ERROR, 0, fmt.Sprint(err))
		return
	}
	if err, ok := err.(*errors.Error); ok {
		stack := err.Stack()
		if stack != "" {
			l.write(ERROR, 0, fmt.Sprintf("%s\n***STACK TRACE***\n%s", err.Error(), stack))
		} else {
			l.write(ERROR, 0, fmt.Sprint(err))
		}
	} else {
		l.write(ERROR, 0, fmt.Sprint(err))
	}
}

func (l *Logger) Close() {
	l.writer.Close()
}

func (l *Logger) write(level LogLevel, callDepth int, message string) {
	if level < l.level {
		return
	}
	data := []interface{}{}
	now := time.Now()
	date := now.Format("2006-01-02")
	hour := now.Format("15:04:05")
	_, file, line, ok := runtime.Caller(3 + callDepth)
	var longFile, shortFile string
	if ok {
		longFile = fmt.Sprintf("%s:%d", file, line)
		shortFile = fmt.Sprintf("%s:%d", filepath.Base(file), line)
	} else {
		longFile = "?:?"
		shortFile = "?:?"
	}
	for _, meta := range l.metas {
		switch meta {
		case metaLevel:
			data = append(data, level)
		case metaDate:
			data = append(data, date)
		case metaTime:
			data = append(data, hour)
		case metaShortfile:
			data = append(data, shortFile)
		case metaLongfile:
			data = append(data, longFile)
		case metaMessage:
			data = append(data, message)
		}
	}
	content := fmt.Sprintf(l.format, data...)
	l.writer.Write(content)
}

func (l *Logger) setMeta() {
	format := ""
	l.metas = []logMeta{}
	meta := false
	for _, r := range l.format {
		if !meta {
			format += string(r)
			if r == '%' {
				meta = true
			}
		} else {
			lMeta, ok := l.getMetaFromRune(r)
			if ok {
				format += "s"
				l.metas = append(l.metas, lMeta)
			} else {
				format += "%" + string(r)
			}
			meta = false
		}
	}
	l.format = format
}

func (l *Logger) getMetaFromRune(r rune) (logMeta, bool) {
	switch r {
	case 'l':
		return metaLevel, true
	case 'd':
		return metaDate, true
	case 't':
		return metaTime, true
	case 'f':
		return metaShortfile, true
	case 'F':
		return metaLongfile, true
	case 's':
		return metaMessage, true
	default:
		return metaUnknown, false
	}
}

var defaultLogger = NewLogger("%l\t%d %t\t%f\t%s", INFO, false, NewConsoleWriter())

func Output(callDepth int, v ...interface{}) {
	defaultLogger.Output(callDepth, v)
}

func Debug(v ...interface{}) {
	defaultLogger.Debug(v...)
}

func Debugf(format string, v ...interface{}) {
	defaultLogger.Debugf(format, v...)
}

func Info(v ...interface{}) {
	defaultLogger.Info(v...)
}

func Infof(format string, v ...interface{}) {
	defaultLogger.Infof(format, v...)
}

func Warn(v ...interface{}) {
	defaultLogger.Warn(v...)
}

func Warnf(format string, v ...interface{}) {
	defaultLogger.Warnf(format, v...)
}

func Error(v ...interface{}) {
	defaultLogger.Error(v...)
}

func Errorf(format string, v ...interface{}) {
	defaultLogger.Errorf(format, v...)
}

func Trace(err error) {
	defaultLogger.Trace(err)
}

func Close() {
	defaultLogger.Close()
}

func SetFormat(format string) {
	defaultLogger.SetFormat(format)
}

func SetLevel(level LogLevel) {
	defaultLogger.SetLevel(level)
}

func SetTrace(trace bool) {
	defaultLogger.SetTrace(trace)
}

func SetWriter(w Writer) {
	defaultLogger.SetWriter(w)
}

type Config struct {
	Type   string `config:"type"`
	Level  string `config:"level"`
	Format string `config:"format"`
	Trace  bool   `config:"trace"`
	File   string `config:"file"`
	Period string `config:"period"`
}

func Start(config *config.Config) error {
	if os.Getenv("RUNMODE") == "check" {
		return nil
	}
	logConfig := &Config{
		Type:   "console",
		Level:  "INFO",
		Format: "%l\t%d %t\t%f\t%s",
		File:   "logs/app.log",
	}
	subConfig, err := config.Get("log")
	if err == nil {
		subConfig.Unmarshal(logConfig)
	}

	return buildLogger(logConfig)
}

func Stop(config *config.Config) error {
	Close()
	return nil
}

func buildLogger(config *Config) error {
	switch config.Type {
	case "console":
		return buildConsoleLogger(config)
	case "rolling":
		return buildRollingLogger(config)
	default:
		return nil
	}
}

func buildConsoleLogger(config *Config) error {
	setLevelString(config.Level)
	SetFormat(config.Format)
	SetTrace(config.Trace)
	return nil
}

func buildRollingLogger(config *Config) error {
	setLevelString(config.Level)
	SetFormat(config.Format)
	SetTrace(config.Trace)
	period := RollDay
	switch strings.ToLower(config.Period) {
	case "hour":
		period = RollHour
	case "minute":
		period = RollMinute
	}
	w, err := NewRollingWriter(config.File, period)
	if err != nil {
		return err
	}
	SetWriter(w)
	return nil
}

func setLevelString(level string) {
	switch strings.ToUpper(level) {
	case "DEBUG":
		SetLevel(DEBUG)
	case "INFO":
		SetLevel(INFO)
	case "WARNING":
		SetLevel(WARNING)
	case "ERROR":
		SetLevel(ERROR)
	}
}
