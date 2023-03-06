package logging

import (
	"time"

	"github.com/sirupsen/logrus"
)

// jsonFormatter is like the standard logrus JSONFormatter, except that it outputs timestamps in
// Unix timestamps with millisecond precision.
type jsonFormatter struct {
	formatter *logrus.JSONFormatter
}

var _ logrus.Formatter = &jsonFormatter{}

func newJSONFormatter() *jsonFormatter {
	return &jsonFormatter{
		&logrus.JSONFormatter{
			DisableTimestamp: true,
			FieldMap: logrus.FieldMap{
				logrus.FieldKeyTime:  "@timestamp",
				logrus.FieldKeyLevel: "@level",
			},
		},
	}
}

// Format the fields to match our common ElasticSearch formats and index
func (j *jsonFormatter) Format(e *logrus.Entry) ([]byte, error) {
	e.Data["time"] = e.Time.UnixNano() / int64(time.Millisecond)
	e.Data["level"] = intLevel(e.Level)
	return j.formatter.Format(e)
}

func intLevel(lvl logrus.Level) uint {
	switch lvl {
	case logrus.TraceLevel:
		return 0
	case logrus.DebugLevel:
		return 10
	case logrus.InfoLevel:
		return 20
	case logrus.WarnLevel:
		return 30
	case logrus.ErrorLevel:
		return 40
	default:
		// should only be fatal
		return 50
	}
}
