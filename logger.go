package snowman

import (
	"fmt"
	"log"
)

// Logger is responsible for providing logging facilities to bot instance.
type Logger interface {
	Debugf(msg string, args ...interface{})
	Infof(msg string, args ...interface{})
	Warnf(msg string, args ...interface{})
	Errorf(msg string, args ...interface{})
}

type stdLogger struct{}

func (s stdLogger) Debugf(msg string, args ...interface{}) { stdLog("DEBUG", msg, args...) }
func (s stdLogger) Infof(msg string, args ...interface{})  { stdLog("INFO ", msg, args...) }
func (s stdLogger) Warnf(msg string, args ...interface{})  { stdLog("WARN ", msg, args...) }
func (s stdLogger) Errorf(msg string, args ...interface{}) { stdLog("ERR  ", msg, args...) }

type noOpLogger struct{}

func (n noOpLogger) Debugf(string, ...interface{}) {}
func (n noOpLogger) Infof(string, ...interface{})  {}
func (n noOpLogger) Warnf(string, ...interface{})  {}
func (n noOpLogger) Errorf(string, ...interface{}) {}

func stdLog(level, msg string, args ...interface{}) {
	log.Printf("[" + level + "]" + fmt.Sprintf(msg, args...))
}
