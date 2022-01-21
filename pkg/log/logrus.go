package log

import "github.com/sirupsen/logrus"

type logrusLog struct {
	log *logrus.Entry
	ops Options
}

func newLogrus(ops *Options) *logrusLog {
	ll := logrus.New()
	ll.SetLevel(loggerToLogrusLevel(ops.level))
	if ops.json {
		ll.SetFormatter(&logrus.JSONFormatter{})
	} else {
		ll.SetFormatter(&logrus.TextFormatter{
			FullTimestamp: true,
		})
	}
	if ops.output != nil {
		ll.SetOutput(ops.output)
	}
	l := logrusLog{
		log: logrus.NewEntry(ll),
		ops: *ops,
	}
	return &l
}

func (l *logrusLog) Options() Options {
	return l.ops
}

func (l *logrusLog) WithFields(fields map[string]interface{}) Interface {
	ll := &logrusLog{
		log: l.log.WithFields(fields),
		ops: l.ops,
	}
	return ll
}

func (l *logrusLog) Log(level Level, args ...interface{}) {
	l.log.Log(loggerToLogrusLevel(level), args...)
}

func (l *logrusLog) Logf(level Level, format string, args ...interface{}) {
	l.log.Logf(loggerToLogrusLevel(level), format, args...)
}

func loggerToLogrusLevel(level Level) logrus.Level {
	switch level {
	case TraceLevel:
		return logrus.TraceLevel
	case DebugLevel:
		return logrus.DebugLevel
	case InfoLevel:
		return logrus.InfoLevel
	case WarnLevel:
		return logrus.WarnLevel
	case ErrorLevel:
		return logrus.ErrorLevel
	case FatalLevel:
		return logrus.FatalLevel
	default:
		return logrus.InfoLevel
	}
}
