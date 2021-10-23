package logger

import (
	"io"
	"log"
)

type LogType string

var (
	Info  LogType = "info"
	Error LogType = "error"
	Debug LogType = "debug"
)

type Logger func(LogType, string, ...interface{})

func (l Logger) Errf(s string, a ...interface{}) {
	l(Error, s, a...)
}

func (l Logger) Infof(s string, a ...interface{}) {
	l(Info, s, a...)
}

func (l Logger) Debugf(s string, a ...interface{}) {
	l(Debug, s, a...)
}

func New(outInfo, outErr, outDebug io.Writer) Logger {
	infoLogger := log.New(outInfo, "[info] ", log.LstdFlags)
	errLogger := log.New(outErr, "[error] ", log.LstdFlags)
	debugLogger := log.New(outDebug, "[debug] ", log.LstdFlags)
	return func(t LogType, s string, a ...interface{}) {
		switch t {
		case Info:
			infoLogger.Printf(s, a...)
		case Error:
			errLogger.Printf(s, a...)
		default:
			debugLogger.Printf(s, a...)
		}
	}
}
