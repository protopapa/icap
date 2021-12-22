package icap

import (
	"log"
	"os"
)

type loggerType int

const (
	standard loggerType = iota
	logfile
)

func newLogger(lType loggerType) *log.Logger {
	if lType == logfile {
		file, err := os.OpenFile("/tmp/icap-lib-record-requests.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
		if err != nil {
			log.Fatal(err)
		}

		logger := log.New(file, "DEBUG ", log.LstdFlags)
		logger.Println("WRITING TO FILE ENABLED")
		return logger
	}

	return log.New(os.Stderr, "", log.LstdFlags)
}

var (
	Std     = newLogger(standard)
	Logfile = newLogger(logfile)
)
