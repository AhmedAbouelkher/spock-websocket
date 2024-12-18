package main

import (
	log "github.com/sirupsen/logrus"
)

var (
	AppLogger = log.StandardLogger()
)

func InitLogger() error {
	AppLogger.SetLevel(log.WarnLevel | log.InfoLevel | log.DebugLevel | log.ErrorLevel)
	// AppLogger.SetFormatter(&log.JSONFormatter{})
	AppLogger.SetFormatter(&log.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
	})
	return nil
}
