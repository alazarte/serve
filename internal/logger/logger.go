package logger

import (
	"fmt"
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
	logger := log.New(outInfo, "", log.LstdFlags)
	return func(t LogType, s string, a ...interface{}) {
		s = fmt.Sprintf("[%s] %s", string(t), s)
		logger.Printf(s, a...)
	}
}
