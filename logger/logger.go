package logger

import (
	"log"
	"os"
)

type CustomLogger struct {
	infoLogger  *log.Logger
	errorLogger *log.Logger
}

var logger *CustomLogger

func init() {
	logger, _ = NewCustomLogger()
}

func NewCustomLogger() (*CustomLogger, error) {
	file, err := os.OpenFile("app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}

	return &CustomLogger{
		infoLogger:  log.New(file, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile),
		errorLogger: log.New(file, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile),
	}, nil
}

func (l *CustomLogger) Info(message string) {
	l.infoLogger.Println(message)
}

func (l *CustomLogger) Error(err error) {
	l.errorLogger.Println(err)
}

func Info(message string) {
	logger.infoLogger.Println(message)
}

func Error(err error) {
	logger.errorLogger.Println(err)
}
