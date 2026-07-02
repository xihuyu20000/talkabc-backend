package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
	FATAL
)

var (
	debugLogger *log.Logger
	infoLogger  *log.Logger
	warnLogger  *log.Logger
	errorLogger *log.Logger
	fatalLogger *log.Logger
	logLevel    LogLevel
)

func getLevelString(level LogLevel) string {
	switch level {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARN:
		return "WARN"
	case ERROR:
		return "ERROR"
	case FATAL:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}

func getCallerInfo() string {
	_, file, line, ok := runtime.Caller(3)
	if !ok {
		return "unknown:0"
	}
	parts := strings.Split(file, "/")
	if len(parts) > 2 {
		file = filepath.Join(parts[len(parts)-2:]...)
	}
	return fmt.Sprintf("%s:%d", file, line)
}

func createLogFile() (io.Writer, error) {
	logDir := "./logs"
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, err
	}

	today := time.Now().Format("2006-01-02")
	filePath := filepath.Join(logDir, fmt.Sprintf("app_%s.log", today))

	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}

	return file, nil
}

func InitLogger(logLevelStr string) {
	logFile, err := createLogFile()
	if err != nil {
		log.Printf("Failed to create log file: %v", err)
		logFile = os.Stderr
	}

	multiWriter := io.MultiWriter(os.Stdout, logFile)

	debugLogger = log.New(multiWriter, "", 0)
	infoLogger = log.New(multiWriter, "", 0)
	warnLogger = log.New(multiWriter, "", 0)
	errorLogger = log.New(multiWriter, "", 0)
	fatalLogger = log.New(multiWriter, "", 0)

	switch strings.ToLower(logLevelStr) {
	case "debug":
		logLevel = DEBUG
	case "info":
		logLevel = INFO
	case "warn":
		logLevel = WARN
	case "error":
		logLevel = ERROR
	case "fatal":
		logLevel = FATAL
	default:
		logLevel = INFO
	}
}

func formatMessage(level LogLevel, message string) string {
	timestamp := time.Now().Format("2006-01-02 15:04:05.000")
	caller := getCallerInfo()
	return fmt.Sprintf("[%s] [%s] [%s] %s", timestamp, getLevelString(level), caller, message)
}

func Debug(format string, args ...interface{}) {
	if logLevel <= DEBUG {
		message := fmt.Sprintf(format, args...)
		debugLogger.Println(formatMessage(DEBUG, message))
	}
}

func Info(format string, args ...interface{}) {
	if logLevel <= INFO {
		message := fmt.Sprintf(format, args...)
		infoLogger.Println(formatMessage(INFO, message))
	}
}

func Warn(format string, args ...interface{}) {
	if logLevel <= WARN {
		message := fmt.Sprintf(format, args...)
		warnLogger.Println(formatMessage(WARN, message))
	}
}

func Error(format string, args ...interface{}) {
	if logLevel <= ERROR {
		message := fmt.Sprintf(format, args...)
		errorLogger.Println(formatMessage(ERROR, message))
	}
}

func Fatal(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	fatalLogger.Println(formatMessage(FATAL, message))
	os.Exit(1)
}
