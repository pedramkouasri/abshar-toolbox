package logger

import (
	"log"
	"os"
)

type CustomLogger struct {
	infoLogger  *log.Logger
	errorLogger *log.Logger
}

func NewCustomLogger() *CustomLogger {
	return &CustomLogger{
		infoLogger:  log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime),
		errorLogger: log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile),
	}
}

func (l *CustomLogger) Info(message string) {
	l.infoLogger.Println(message)
}

func (l *CustomLogger) Error(err error) {
	l.errorLogger.Println(err)
}
