package logger

import (
	"github.com/sirupsen/logrus"
)

var Log *logrus.Logger

func InitLogger() {
	Log = logrus.New()
	Log.SetLevel(logrus.DebugLevel)
	Log.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
}
