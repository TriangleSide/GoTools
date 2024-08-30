package logger

import (
	"github.com/sirupsen/logrus"
)

// UTCFormatter sets the timezone of the log to UTC.
type UTCFormatter struct {
	Next logrus.Formatter
}

// Format sets the timezone of the log to UTC then invokes the next formatter.
func (f *UTCFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	entry.Time = entry.Time.UTC()
	return f.Next.Format(entry)
}
