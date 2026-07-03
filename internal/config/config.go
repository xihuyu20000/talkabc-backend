package config

import (
	"log"

	"github.com/go-redis/redis/v8"
	"github.com/jinzhu/gorm"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type Config struct {
	System   SystemConfig   `json:"system"`
	Server   ServerConfig   `json:"server"`
	Database DatabaseConfig `json:"database"`
	JWT      JWTConfig     `json:"jwt"`
	Upload   UploadConfig   `json:"upload"`
	Redis    RedisConfig    `json:"redis"`
	CORS     CORSConfig     `json:"cors"`
}

type SystemConfig struct {
	Reset           int    `json:"reset"`
	LogLevel        string `json:"log_level"`
	SMSValidMinutes int    `json:"sms_valid_minutes"`
	SMSCooldownSeconds int `json:"sms_cooldown_seconds"`
	SMSHourlyLimit     int    `json:"sms_hourly_limit"`
	IPRegisterHourlyLimit int    `json:"ip_register_hourly_limit"`
	IPLoginMinuteLimit int    `json:"ip_login_minute_limit"`
	LoginFailureLockMinutes int    `json:"login_failure_lock_minutes"`
}

type ServerConfig struct {
	Port int `json:"port"`
}

type DatabaseConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	DBName   string `json:"dbname"`
	SSLMode  string `json:"sslmode"`
}

type JWTConfig struct {
	Secret      string `json:"secret"`
	ExpiresHour int    `json:"expires_hour"`
}

type UploadConfig struct {
	AvatarPath  string `json:"avatar_path"`
	MomentPath  string `json:"moment_path"`
	MessagePath string `json:"message_path"`
}

type RedisConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Password string `json:"password"`
	DB       int    `json:"db"`
}

type CORSConfig struct {
	Origins     []string `json:"origins"`
	Methods     []string `json:"methods"`
	Headers     []string `json:"headers"`
	Credentials bool     `json:"credentials"`
}

var AppConfig Config
var DB *gorm.DB
var RDB *redis.Client

func InitConfig() {
	// 1. 命令行显示
	pflag.Int("system-reset", 0, "system reset flag")
	pflag.String("system-log-level", "info", "log level")
	pflag.Int("system-sms-valid-minutes", 5, "sms valid minutes")
	pflag.Int("system-sms-cooldown-seconds", 60, "sms cooldown seconds")
	pflag.Int("system-sms-hourly-limit", 10, "sms-hourly limit")
	pflag.Int("system-ip-register-hourly-limit", 10, "ip-register-hourly-limit")
	pflag.Int("system-ip-login-minute-limit", 10, "ip-login-minute-limit")
	pflag.Int("system-login-failure-lock-minutes", 5, "login-failure-lock minutes")
	pflag.Int("server-port", 8080, "server port")
	pflag.String("database-host", "localhost", "database host")
	pflag.Int("database-port", 5432, "database port")
	pflag.String("database-user", "postgres", "database user")
	pflag.String("database-password", "", "database password")
	pflag.String("database-name", "talkabc", "database name")
	pflag.String("database-sslmode", "disable", "database sslmode")
	pflag.String("jwt-secret", "talkabc_secret_key", "jwt secret")
	pflag.Int("jwt-expireshour", 24, "jwt expires hour")
	pflag.String("redis-host", "localhost", "redis host")
	pflag.Int("redis-port", 6379, "redis port")
	pflag.String("redis-password", "", "redis password")
	pflag.Int("redis-db", 0, "redis db")
	pflag.String("upload-avatarpath", "./uploads/avatars", "avatar path")
	pflag.String("upload-momentpath", "./uploads/moments", "moment path")
	pflag.String("upload-messagepath", "./uploads/messages", "message path")
	// 2. 解析
	pflag.Parse()
	// 3. 绑定
	viper.BindPFlag("system.reset", pflag.CommandLine.Lookup("system-reset"))
	viper.BindPFlag("system.log_level", pflag.CommandLine.Lookup("system-log-level"))
	viper.BindPFlag("system.sms_valid_minutes", pflag.CommandLine.Lookup("system-sms-valid-minutes"))
	viper.BindPFlag("system.sms_cooldown_seconds", pflag.CommandLine.Lookup("system-sms-cooldown-seconds"))
	viper.BindPFlag("system.sms_hourly_limit", pflag.CommandLine.Lookup("system-sms-hourly-limit"))
	viper.BindPFlag("system.ip_register_hourly_limit", pflag.CommandLine.Lookup("system-ip-register-hourly-limit"))
	viper.BindPFlag("system.ip_login_minute_limit", pflag.CommandLine.Lookup("system-ip-login-minute-limit"))
	viper.BindPFlag("system.login_failure_lock_minutes", pflag.CommandLine.Lookup("system-login-failure-lock-minutes"))
	viper.BindPFlag("server.port", pflag.CommandLine.Lookup("server-port"))
	viper.BindPFlag("database.host", pflag.CommandLine.Lookup("database-host"))
	viper.BindPFlag("database.port", pflag.CommandLine.Lookup("database-port"))
	viper.BindPFlag("database.user", pflag.CommandLine.Lookup("database-user"))
	viper.BindPFlag("database.password", pflag.CommandLine.Lookup("database-password"))
	viper.BindPFlag("database.dbname", pflag.CommandLine.Lookup("database-name"))
	viper.BindPFlag("database.sslmode", pflag.CommandLine.Lookup("database-sslmode"))
	viper.BindPFlag("jwt.secret", pflag.CommandLine.Lookup("jwt-secret"))
	viper.BindPFlag("jwt.expires_hour", pflag.CommandLine.Lookup("jwt-expireshour"))
	viper.BindPFlag("redis.host", pflag.CommandLine.Lookup("redis-host"))
	viper.BindPFlag("redis.port", pflag.CommandLine.Lookup("redis-port"))
	viper.BindPFlag("redis.password", pflag.CommandLine.Lookup("redis-password"))
	viper.BindPFlag("redis.db", pflag.CommandLine.Lookup("redis-db"))
	viper.BindPFlag("upload.avatar_path", pflag.CommandLine.Lookup("upload-avatarpath"))
	viper.BindPFlag("upload.moment_path", pflag.CommandLine.Lookup("upload-momentpath"))
	viper.BindPFlag("upload.message_path", pflag.CommandLine.Lookup("upload-messagepath"))
	// 4. 默认值
	viper.SetDefault("system.reset", 0)
	viper.SetDefault("system.log_level", "info")
	viper.SetDefault("system.sms_valid_minutes", 5)
	viper.SetDefault("system.sms_cooldown_seconds", 60)
	viper.SetDefault("system.sms_hourly_limit", 10)
	viper.SetDefault("system.ip_register_hourly_limit", 10)
	viper.SetDefault("system.ip_login_minute_limit", 10)
	viper.SetDefault("system.login_failure_lock_minutes", 5)
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", 5432)
	viper.SetDefault("database.user", "postgres")
	viper.SetDefault("database.password", "")
	viper.SetDefault("database.dbname", "talkabc")
	viper.SetDefault("database.sslmode", "disable")
	viper.SetDefault("jwt.secret", "talkabc_secret_key")
	viper.SetDefault("jwt.expires_hour", 24)
	viper.SetDefault("redis.host", "localhost")
	viper.SetDefault("redis.port", 6379)
	viper.SetDefault("redis.password", "")
	viper.SetDefault("redis.db", 0)
	viper.SetDefault("upload.avatar_path", "./uploads/avatars")
	viper.SetDefault("upload.moment_path", "./uploads/moments")
	viper.SetDefault("upload.message_path", "./uploads/messages")
	viper.SetDefault("cors.origins", []string{"http://localhost:3000", "http://localhost:8080"})
	viper.SetDefault("cors.methods", []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"})
	viper.SetDefault("cors.headers", []string{"Origin", "Content-Type", "Authorization", "Accept", "X-Requested-With"})
	viper.SetDefault("cors.credentials", true)

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")

	if err := viper.ReadInConfig(); err != nil {
		log.Printf("Config file not found: %v", err)
	}

	if err := viper.Unmarshal(&AppConfig); err != nil {
		log.Fatalf("Failed to unmarshal: %v", err)
	}
}
