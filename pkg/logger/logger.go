package logger

import (
	"fmt"
	"os"
	"path/filepath"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	log         *zap.Logger
	sugar       *zap.SugaredLogger
	logLevel    zapcore.Level
	ServiceName = "talkabc"
	Environment = "development"
	Version     = "1.0.0"
)

type Config struct {
	Level       string `json:"level"`
	Format      string `json:"format"`
	Output      string `json:"output"`
	FilePath    string `json:"file_path"`
	MaxSize     int    `json:"max_size"`
	MaxBackups  int    `json:"max_backups"`
	MaxAge      int    `json:"max_age"`
	Compress    bool   `json:"compress"`
}

func InitLogger(cfg *Config) {
	var level zapcore.Level
	switch cfg.Level {
	case "debug":
		level = zapcore.DebugLevel
	case "info":
		level = zapcore.InfoLevel
	case "warn":
		level = zapcore.WarnLevel
	case "error":
		level = zapcore.ErrorLevel
	case "dpanic":
		level = zapcore.DPanicLevel
	case "panic":
		level = zapcore.PanicLevel
	case "fatal":
		level = zapcore.FatalLevel
	default:
		level = zapcore.InfoLevel
	}
	logLevel = level

	var cores []zapcore.Core

	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "message",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	var encoder zapcore.Encoder
	if cfg.Format == "json" {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	}

	if cfg.Output == "file" || cfg.Output == "both" {
		logDir := filepath.Dir(cfg.FilePath)
		if err := os.MkdirAll(logDir, 0755); err != nil {
			fmt.Printf("Failed to create log directory %s: %v\n", logDir, err)
		}
		fileWriter := zapcore.AddSync(&lumberjack.Logger{
			Filename:   cfg.FilePath,
			MaxSize:    cfg.MaxSize,
			MaxBackups: cfg.MaxBackups,
			MaxAge:     cfg.MaxAge,
			Compress:   cfg.Compress,
		})
		fileCore := zapcore.NewCore(encoder, fileWriter, level)
		cores = append(cores, fileCore)
	}

	if cfg.Output == "console" || cfg.Output == "both" {
		consoleWriter := zapcore.AddSync(os.Stdout)
		consoleCore := zapcore.NewCore(encoder, consoleWriter, level)
		cores = append(cores, consoleCore)
	}

	core := zapcore.NewTee(cores...)

	log = zap.New(core,
		zap.AddCaller(),
		zap.AddCallerSkip(1),
		zap.AddStacktrace(zapcore.ErrorLevel),
		zap.Fields(
			zap.String("service", ServiceName),
			zap.String("env", Environment),
			zap.String("version", Version),
		),
	)

	sugar = log.Sugar()
}

func Sync() {
	if log != nil {
		_ = log.Sync()
	}
}

func Debug(msg string, fields ...zap.Field) {
	if log != nil {
		log.Debug(msg, fields...)
	}
}

func Debugf(format string, args ...interface{}) {
	if sugar != nil {
		sugar.Debugf(format, args...)
	}
}

func Info(msg string, fields ...zap.Field) {
	if log != nil {
		log.Info(msg, fields...)
	}
}

func Infof(format string, args ...interface{}) {
	if sugar != nil {
		sugar.Infof(format, args...)
	}
}

func Warn(msg string, fields ...zap.Field) {
	if log != nil {
		log.Warn(msg, fields...)
	}
}

func Warnf(format string, args ...interface{}) {
	if sugar != nil {
		sugar.Warnf(format, args...)
	}
}

func Error(msg string, fields ...zap.Field) {
	if log != nil {
		log.Error(msg, fields...)
	}
}

func Errorf(format string, args ...interface{}) {
	if sugar != nil {
		sugar.Errorf(format, args...)
	}
}

func Panic(msg string, fields ...zap.Field) {
	if log != nil {
		log.Panic(msg, fields...)
	}
}

func Panicf(format string, args ...interface{}) {
	if sugar != nil {
		sugar.Panicf(format, args...)
	}
}

func Fatal(msg string, fields ...zap.Field) {
	if log != nil {
		log.Fatal(msg, fields...)
	}
}

func Fatalf(format string, args ...interface{}) {
	if sugar != nil {
		sugar.Fatalf(format, args...)
	}
}

func WithContext(ctx map[string]interface{}) *zap.SugaredLogger {
	if sugar == nil {
		return sugar
	}
	s := sugar
	for k, v := range ctx {
		s = s.With(k, v)
	}
	return s
}

func GetLogLevel() zapcore.Level {
	return logLevel
}

func GetLogger() *zap.Logger {
	return log
}

func GetSugarLogger() *zap.SugaredLogger {
	return sugar
}