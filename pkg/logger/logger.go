// Package logger 提供企业级日志服务，基于 zap + lumberjack 实现。
//
// 核心设计理念：
//   - 全局单例：整个项目只初始化一次 logger，禁止到处 new 日志对象
//   - 配置驱动：通过独立的 YAML 配置文件统一控制日志行为
//   - 分层输出：支持同时输出到控制台和文件
//   - 自动切割：按大小/时间分割日志，自动清理过期日志
//   - 链路追踪：支持从 context 中提取 request_id/trace_id
//
// 使用流程：
//  1. 在 main.go 中调用 logger.LoadConfig() 加载配置
//  2. 调用 logger.InitLogger() 初始化日志实例
//  3. 在任意地方直接调用 logger.Info/logger.Infof 等方法
//
// 配置文件示例（config/logger.yaml）：
//
//	level: debug          # 日志级别: debug, info, warn, error, fatal
//	format: console       # 日志格式: console, json
//	output: both          # 输出方式: console, file, both
//	file_path: ./logs/app.log
//	max_size: 100         # 单文件最大大小(MB)
//	max_backups: 30       # 保留备份数
//	max_age: 7            # 保留天数
//	compress: true        # 是否压缩归档
package logger

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	log         *zap.Logger         // zap 核心日志实例，用于结构化日志
	sugar       *zap.SugaredLogger  // zap 格式化日志实例，用于 printf 风格日志
	logLevel    zapcore.Level       // 当前日志级别（只读）
	atom        zap.AtomicLevel     // 原子级别变量，支持运行时动态修改日志级别
	ServiceName = "talkabc"         // 服务名称，自动注入每条日志
	Environment = "development"     // 运行环境，自动注入每条日志
	Version     = "1.0.0"           // 版本号，自动注入每条日志
)

// Config 日志配置结构体
// 支持从 YAML/JSON 文件加载，也可在代码中直接构建
type Config struct {
	Level       string `json:"level" yaml:"level"`       // 日志级别：debug/info/warn/error/fatal
	Format      string `json:"format" yaml:"format"`     // 输出格式：console（人类可读）/json（机器可读）
	Output      string `json:"output" yaml:"output"`     // 输出方式：console/file/both
	FilePath    string `json:"file_path" yaml:"file_path"` // 日志文件路径
	MaxSize     int    `json:"max_size" yaml:"max_size"`  // 单文件最大大小（MB）
	MaxBackups  int    `json:"max_backups" yaml:"max_backups"` // 保留的最大备份文件数
	MaxAge      int    `json:"max_age" yaml:"max_age"`     // 日志文件保留天数
	Compress    bool   `json:"compress" yaml:"compress"`   // 是否压缩归档日志
}

// LoadConfig 从 YAML 文件加载日志配置
// filePath: 配置文件路径（如 "./config/logger.yaml"）
// 返回：配置对象指针，加载失败返回错误
//
// 使用示例：
//   cfg, err := logger.LoadConfig("./config/logger.yaml")
//   if err != nil {
//       // 使用默认配置
//       cfg = &logger.Config{Level: "info", Output: "console"}
//   }
//   logger.InitLogger(cfg)
func LoadConfig(filePath string) (*Config, error) {
	v := viper.New()
	v.SetConfigFile(filePath)
	v.SetConfigType("yaml")

	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	fmt.Printf("[Logger] Config loaded: Level=%s, Format=%s, Output=%s, FilePath=%s, MaxSize=%dMB, MaxBackups=%d, MaxAge=%dd, Compress=%v\n",
		cfg.Level, cfg.Format, cfg.Output, cfg.FilePath, cfg.MaxSize, cfg.MaxBackups, cfg.MaxAge, cfg.Compress)

	return &cfg, nil
}

type LoggerConfig struct {
	Level       string
	Format      string
	Output      string
	FilePath    string
	MaxSize     int
	MaxBackups  int
	MaxAge      int
	Compress    bool
}

// InitLogger 初始化全局日志实例
// 必须在应用启动时调用，且只能调用一次
// cfg: 日志配置，可以是 Config 或 LoggerConfig 类型
//
// 初始化流程：
//   1. 解析日志级别配置
//   2. 创建编码器（JSON/Console）
//   3. 创建文件输出（如果配置了 file 或 both）
//   4. 创建控制台输出（如果配置了 console 或 both）
//   5. 使用 zapcore.NewTee 合并多个输出
//   6. 创建 zap.Logger 和 zap.SugaredLogger 实例
//
// 每条日志自动携带的字段：
//   - service: 服务名称
//   - env: 运行环境
//   - version: 版本号
//   - caller: 调用位置（文件名:行号）
//   - stacktrace: 错误级别以上自动附加堆栈信息
func InitLogger(cfg interface{}) {
	var config Config
	switch c := cfg.(type) {
	case *Config:
		config = *c
	case *LoggerConfig:
		config = Config{
			Level:       c.Level,
			Format:      c.Format,
			Output:      c.Output,
			FilePath:    c.FilePath,
			MaxSize:     c.MaxSize,
			MaxBackups:  c.MaxBackups,
			MaxAge:      c.MaxAge,
			Compress:    c.Compress,
		}
	case Config:
		config = c
	case LoggerConfig:
		config = Config{
			Level:       c.Level,
			Format:      c.Format,
			Output:      c.Output,
			FilePath:    c.FilePath,
			MaxSize:     c.MaxSize,
			MaxBackups:  c.MaxBackups,
			MaxAge:      c.MaxAge,
			Compress:    c.Compress,
		}
	default:
		config = Config{Level: "info", Output: "console"}
	}

	var level zapcore.Level
	switch config.Level {
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
	atom = zap.NewAtomicLevelAt(level)

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
	if config.Format == "json" {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	}

	if config.Output == "file" || config.Output == "both" {
		logDir := filepath.Dir(config.FilePath)
		if err := os.MkdirAll(logDir, 0755); err != nil {
			fmt.Printf("Failed to create log directory %s: %v\n", logDir, err)
		}
		fileWriter := zapcore.AddSync(&lumberjack.Logger{
			Filename:   config.FilePath,
			MaxSize:    config.MaxSize,
			MaxBackups: config.MaxBackups,
			MaxAge:     config.MaxAge,
			Compress:   config.Compress,
		})
		fileCore := zapcore.NewCore(encoder, fileWriter, atom)
		cores = append(cores, fileCore)
	}

	if config.Output == "console" || config.Output == "both" {
		consoleWriter := zapcore.AddSync(os.Stdout)
		consoleCore := zapcore.NewCore(encoder, consoleWriter, atom)
		cores = append(cores, consoleCore)
	}

	core := zapcore.NewTee(cores...)

	log = zap.New(core,
		zap.AddCaller(),
		zap.AddCallerSkip(1),
		zap.AddStacktrace(zapcore.ErrorLevel),
		// zap.Fields(
		// 	zap.String("service", ServiceName),
		// 	zap.String("env", Environment),
		// 	zap.String("version", Version),
		// ),
	)

	sugar = log.Sugar()
}

// SetLogLevel 运行时动态修改日志级别
// level: 目标级别字符串（debug/info/warn/error/fatal）
// 返回：级别解析错误
//
// 使用场景：
//   - 通过 API 接口动态调整日志级别，无需重启服务
//   - 在生产环境遇到问题时临时调高级别减少日志量
//   - 在排查问题时临时调低级别获取更详细日志
//
// 示例：
//   logger.SetLogLevel("debug")  // 调整为调试级别
//   logger.SetLogLevel("error")  // 调整为错误级别
func SetLogLevel(level string) error {
	var zapLevel zapcore.Level
	if err := zapLevel.UnmarshalText([]byte(level)); err != nil {
		return err
	}
	atom.SetLevel(zapLevel)
	logLevel = zapLevel
	Info("Log level changed", zap.String("level", level))
	return nil
}

// Sync 刷新日志缓冲区
// 应在应用退出时调用（通常通过 defer logger.Sync()）
// 确保所有缓冲的日志都被写入文件
func Sync() {
	if log != nil {
		_ = log.Sync()
	}
}

// Debug 输出调试级别日志（结构化方式）
// msg: 日志消息
// fields: 结构化字段（如 zap.String("key", value)）
//
// 使用场景：详细的调试信息，仅在开发/调试阶段使用
// 示例：
//   logger.Debug("User login", zap.String("phone", phone))
func Debug(msg string, fields ...zap.Field) {
	if log != nil {
		log.Debug(msg, fields...)
	}
}

// Debugf 输出调试级别日志（printf 风格）
// format: 格式化字符串
// args: 格式化参数
//
// 使用场景：简单的调试信息，需要灵活的字符串拼接
// 示例：
//   logger.Debugf("User %s logged in from %s", username, ip)
func Debugf(format string, args ...interface{}) {
	if sugar != nil {
		sugar.Debugf(format, args...)
	}
}

// Info 输出信息级别日志（结构化方式）
// msg: 日志消息
// fields: 结构化字段
//
// 使用场景：正常的业务流程记录，如请求处理、操作完成等
// 示例：
//   logger.Info("Request processed", zap.String("path", path), zap.Int("status", status))
func Info(msg string, fields ...zap.Field) {
	if log != nil {
		log.Info(msg, fields...)
	}
}

// Infof 输出信息级别日志（printf 风格）
// format: 格式化字符串
// args: 格式化参数
//
// 使用场景：简单的业务流程记录
// 示例：
//   logger.Infof("[Handler] Register - PhoneNum: %s", phone)
func Infof(format string, args ...interface{}) {
	if sugar != nil {
		sugar.Infof(format, args...)
	}
}

// Warn 输出警告级别日志（结构化方式）
// msg: 日志消息
// fields: 结构化字段
//
// 使用场景：需要关注但不影响业务的异常情况，如配置缺失、降级处理等
// 示例：
//   logger.Warn("SMS gateway unavailable", zap.Error(err))
func Warn(msg string, fields ...zap.Field) {
	if log != nil {
		log.Warn(msg, fields...)
	}
}

// Warnf 输出警告级别日志（printf 风格）
// format: 格式化字符串
// args: 格式化参数
func Warnf(format string, args ...interface{}) {
	if sugar != nil {
		sugar.Warnf(format, args...)
	}
}

// Error 输出错误级别日志（结构化方式）
// msg: 日志消息
// fields: 结构化字段（通常包含 zap.Error(err)）
//
// 使用场景：业务错误，会影响功能正常执行，如数据库操作失败、API调用失败等
// 示例：
//   logger.Error("Failed to save user", zap.Error(err), zap.String("phone", phone))
func Error(msg string, fields ...zap.Field) {
	if log != nil {
		log.Error(msg, fields...)
	}
}

// Errorf 输出错误级别日志（printf 风格）
// format: 格式化字符串
// args: 格式化参数
func Errorf(format string, args ...interface{}) {
	if sugar != nil {
		sugar.Errorf(format, args...)
	}
}

// Panic 输出恐慌级别日志并触发 panic（结构化方式）
// msg: 日志消息
// fields: 结构化字段
//
// 使用场景：不可恢复的严重错误，会导致程序崩溃
func Panic(msg string, fields ...zap.Field) {
	if log != nil {
		log.Panic(msg, fields...)
	}
}

// Panicf 输出恐慌级别日志并触发 panic（printf 风格）
// format: 格式化字符串
// args: 格式化参数
func Panicf(format string, args ...interface{}) {
	if sugar != nil {
		sugar.Panicf(format, args...)
	}
}

// Fatal 输出致命级别日志并终止程序（结构化方式）
// msg: 日志消息
// fields: 结构化字段
//
// 使用场景：程序无法继续运行的严重错误，会调用 os.Exit(1)
func Fatal(msg string, fields ...zap.Field) {
	if log != nil {
		log.Fatal(msg, fields...)
	}
}

// Fatalf 输出致命级别日志并终止程序（printf 风格）
// format: 格式化字符串
// args: 格式化参数
func Fatalf(format string, args ...interface{}) {
	if sugar != nil {
		sugar.Fatalf(format, args...)
	}
}

// ctxKey context 键类型，用于从 context 中提取日志相关字段
type ctxKey string

const (
	RequestIDKey ctxKey = "request_id" // 请求ID键名，用于链路追踪
	TraceIDKey   ctxKey = "trace_id"   // 追踪ID键名，用于分布式链路追踪
)

// WithContext 从 context 中提取日志字段并返回带字段的 SugaredLogger
// ctx: Go 标准 context.Context
// 返回：带 request_id/trace_id 字段的 SugaredLogger
//
// 使用场景：在 HTTP 请求处理中，自动携带请求ID，便于链路追踪
// 示例：
//   reqLogger := logger.WithContext(c.Request.Context())
//   reqLogger.Infof("User %s logged in", uid)
func WithContext(ctx context.Context) *zap.SugaredLogger {
	if sugar == nil {
		return sugar
	}
	s := sugar
	if ctx != nil {
		if requestID := ctx.Value(RequestIDKey); requestID != nil {
			s = s.With("request_id", requestID)
		}
		if traceID := ctx.Value(TraceIDKey); traceID != nil {
			s = s.With("trace_id", traceID)
		}
	}
	return s
}

// WithFields 从 context 中提取字段并附加自定义字段
// ctx: Go 标准 context.Context
// fields: 自定义字段映射
// 返回：带所有字段的 SugaredLogger
//
// 使用场景：需要同时携带上下文和自定义字段的日志记录
// 示例：
//   logger.WithFields(ctx, map[string]interface{}{
//       "user_id": uid,
//       "action":  "login",
//   }).Infof("Operation completed")
func WithFields(ctx context.Context, fields map[string]interface{}) *zap.SugaredLogger {
	if sugar == nil {
		return sugar
	}
	s := WithContext(ctx)
	for k, v := range fields {
		s = s.With(k, v)
	}
	return s
}

// GetLogLevel 获取当前日志级别
// 返回：当前日志级别（zapcore.Level 类型）
func GetLogLevel() zapcore.Level {
	return logLevel
}

// MaskToken 对敏感 Token 进行脱敏处理
// token: 原始 Token
// 返回：脱敏后的 Token（如 "abc1***456"）
//
// 使用场景：记录 Token 时避免明文泄露，保护安全
// 示例：
//   logger.Infof("Token: %s", logger.MaskToken(token))
func MaskToken(token string) string {
	if len(token) <= 8 {
		return "***"
	}
	return token[:4] + "***" + token[len(token)-4:]
}

// MaskSensitive 对敏感信息进行脱敏处理
// value: 原始敏感值
// 返回：脱敏后的值
//
// 使用场景：记录手机号、邮箱、身份证号等敏感信息时进行脱敏
// 脱敏规则：
//   - 长度 <= 4: 全部隐藏（***）
//   - 长度 <= 8: 保留前后各2位（如 "ab***cd"）
//   - 长度 > 8: 保留前后各4位（如 "abcd***wxyz"）
func MaskSensitive(value string) string {
	if len(value) <= 4 {
		return "***"
	}
	if len(value) <= 8 {
		return value[:2] + "***" + value[len(value)-2:]
	}
	return value[:4] + "***" + value[len(value)-4:]
}

// GetLogger 获取底层 zap.Logger 实例
// 返回：zap.Logger 指针
//
// 使用场景：需要直接使用 zap 原生 API 的高级场景
func GetLogger() *zap.Logger {
	return log
}

// GetSugarLogger 获取底层 zap.SugaredLogger 实例
// 返回：zap.SugaredLogger 指针
//
// 使用场景：需要直接使用 sugared logger 的高级场景
func GetSugarLogger() *zap.SugaredLogger {
	return sugar
}