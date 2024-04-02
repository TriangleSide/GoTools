package logger

import (
	"github.com/sirupsen/logrus"
)

// CustomFormatter extends the formatter from logrus. This wrap it so we can modify some properties.
type customFormatter struct {
	logrus.JSONFormatter
}

// Format an entry to set custom properties.
func (f *customFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	entry.Time = entry.Time.UTC()
	return f.JSONFormatter.Format(entry)
}
