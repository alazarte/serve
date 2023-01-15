/*
Some function that I rescued from main.go, not sure why but could be useful

func getLogOutput() io.Writer {
	var debugOut io.Writer = io.Discard

	switch config.Debug {
	case "":
		debugOut = io.Discard
	default:
		f, err := os.OpenFile(config.Debug, os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			panic(fmt.Sprintln("couldn't open file for debug logs:", config.Debug))
		}
		debugOut = f
	}

	return debugOut
}
*/

package logger

import (
	"fmt"
	"io"
	"log"
)

type LogLevel uint8

const (
	Debug LogLevel = iota
	Info
	Error
)

func logLevelText(l LogLevel) string {
	switch l {
	case Info:
		return "info"
	case Error:
		return "error"
	case Debug:
		return "debug"
	default:
		return "invalid"
	}
}

type Logger func(LogLevel, string, ...interface{})

func (l Logger) Errf(s string, a ...interface{}) {
	l(Error, s, a...)
}

func (l Logger) Infof(s string, a ...interface{}) {
	l(Info, s, a...)
}

func (l Logger) Debugf(s string, a ...interface{}) {
	l(Debug, s, a...)
}

func (a LogLevel) Bigger(b LogLevel) bool {
	return a > b
}

func New(output io.Writer, level LogLevel) Logger {
	logger := log.New(output, "", log.LstdFlags)
	return func(t LogLevel, s string, a ...interface{}) {
		if level.Bigger(t) {
			return
		}

		s = fmt.Sprintf("[%s] %s", logLevelText(t), s)
		logger.Printf(s, a...)
	}
}
