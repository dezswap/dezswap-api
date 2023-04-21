package logging

import (
	"github.com/evalphobia/logrus_sentry"
	"github.com/sirupsen/logrus"
)

func ConfigureReporter(logger Logger, dsn, env string, tag map[string]string) error {
	hook, err := logrus_sentry.NewSentryHook(dsn, []logrus.Level{
		logrus.WarnLevel,
		logrus.PanicLevel,
		logrus.FatalLevel,
		logrus.ErrorLevel,
	})
	if err != nil {
		panic(err)
	}
	hook.StacktraceConfiguration.Enable = true
	hook.SetEnvironment(env)
	hook.SetTagsContext(tag)
	AddHookToLogger(logger, hook)
	return nil
}
