package utils

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
}
