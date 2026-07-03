package config

import (
	"log"
	"strings"

	"github.com/go-redis/redis/v8"
	"github.com/jinzhu/gorm"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// ==================== 配置结构体定义 ====================

type Config struct {
	System       SystemConfig       `json:"system"`
	Security     SecurityConfig     `json:"security"`
	SMSProvider  SMSProviderConfig  `json:"sms_provider"`
	Server       ServerConfig       `json:"server"`
	Database     DatabaseConfig     `json:"database"`
	JWT          JWTConfig          `json:"jwt"`
	Upload       UploadConfig       `json:"upload"`
	Redis        RedisConfig        `json:"redis"`
	CORS         CORSConfig         `json:"cors"`
}

type SystemConfig struct {
	Reset    int    `json:"reset"`
	LogLevel string `json:"log_level"`
}

type SecurityConfig struct {
	SMSValidMinutes        int    `json:"sms_valid_minutes"`
	SMSCooldownSeconds     int    `json:"sms_cooldown_seconds"`
	SMSHourlyLimit         int    `json:"sms_hourly_limit"`
	IPRegisterHourlyLimit  int    `json:"ip_register_hourly_limit"`
	IPLoginMinuteLimit     int    `json:"ip_login_minute_limit"`
	LoginFailureLockMinutes int    `json:"login_failure_lock_minutes"`
}

type SMSProviderConfig struct {
	Default  string           `json:"default"`
	Aliyun   AliyunSMSConfig  `json:"aliyun"`
	Huawei   HuaweiSMSConfig  `json:"huawei"`
	Tencent  TencentSMSConfig `json:"tencent"`
}

type AliyunSMSConfig struct {
	AccessKeyID     string `json:"access_key_id"`
	AccessKeySecret string `json:"access_key_secret"`
	RegionID        string `json:"region_id"`
	SignName        string `json:"sign_name"`
	TemplateCode    string `json:"template_code"`
}

type HuaweiSMSConfig struct {
	AppKey    string `json:"app_key"`
	AppSecret string `json:"app_secret"`
	SignName  string `json:"sign_name"`
	TemplateID string `json:"template_id"`
}

type TencentSMSConfig struct {
	SecretID     string `json:"secret_id"`
	SecretKey    string `json:"secret_key"`
	RegionID     string `json:"region_id"`
	SignName     string `json:"sign_name"`
	TemplateID   string `json:"template_id"`
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

// ==================== 全局变量 ====================

var AppConfig Config
var DB *gorm.DB
var RDB *redis.Client

// ==================== 命令行参数解析工具函数 ====================

// parseArgv 统一解析命令行参数并绑定到viper
// 支持 int 和 string 类型，通过类型断言自动选择对应的 pflag 方法
func parseArgv(cfgKey string, defaultValue interface{}, usage string) {
	var argName = strings.ReplaceAll(cfgKey, ".", "-")
	switch v := defaultValue.(type) {
	case int:
		pflag.Int(argName, v, usage)
	case string:
		pflag.String(argName, v, usage)
	default:
		log.Printf("Unsupported type for config key: %s", cfgKey)
		return
	}
	pflag.Parse()
	viper.BindPFlag(cfgKey, pflag.CommandLine.Lookup(argName))
	viper.SetDefault(cfgKey, defaultValue)
}

// ==================== 配置初始化函数 ====================

func InitConfig() {
	// 系统配置
	parseArgv("system.reset", 0, "system reset flag")
	parseArgv("system.log_level", "info", "log level")
	// 安全配置
	parseArgv("security.sms_valid_minutes", 5, "sms valid minutes")
	parseArgv("security.sms_cooldown_seconds", 60, "sms cooldown seconds")
	parseArgv("security.sms_hourly_limit", 10, "sms hourly limit")
	parseArgv("security.ip_register_hourly_limit", 10, "ip register hourly limit")
	parseArgv("security.ip_login_minute_limit", 10, "ip login minute limit")
	parseArgv("security.login_failure_lock_minutes", 5, "login failure lock minutes")

	// 短信服务商配置
	parseArgv("sms_provider.default", "aliyun", "default sms provider (aliyun, huawei, tencent)")
	parseArgv("sms_provider.aliyun.access_key_id", "", "aliyun sms access key id")
	parseArgv("sms_provider.aliyun.access_key_secret", "", "aliyun sms access key secret")
	parseArgv("sms_provider.aliyun.region_id", "cn-hangzhou", "aliyun sms region id")
	parseArgv("sms_provider.aliyun.sign_name", "", "aliyun sms sign name")
	parseArgv("sms_provider.aliyun.template_code", "", "aliyun sms template code")

	// 服务器配置
	parseArgv("server.port", 8080, "server port")

	// 数据库配置
	parseArgv("database.host", "localhost", "database host")
	parseArgv("database.port", 5432, "database port")
	parseArgv("database.user", "postgres", "database user")
	parseArgv("database.password", "", "database password")
	parseArgv("database.dbname", "talkabc", "database name")
	parseArgv("database.sslmode", "disable", "database sslmode")

	// JWT配置
	parseArgv("jwt.secret", "talkabc_secret_key", "jwt secret")
	parseArgv("jwt.expires_hour", 24, "jwt expires hour")

	// Redis配置
	parseArgv("redis.host", "localhost", "redis host")
	parseArgv("redis.port", 6379, "redis port")
	parseArgv("redis.password", "", "redis password")
	parseArgv("redis.db", 0, "redis db")

	// 上传配置
	parseArgv("upload.avatar_path", "./uploads/avatars", "avatar path")
	parseArgv("upload.moment_path", "./uploads/moments", "moment path")
	parseArgv("upload.message_path", "./uploads/messages", "message path")

	// CORS配置（数组类型，单独设置默认值）
	viper.SetDefault("cors.origins", []string{"http://localhost:3000", "http://localhost:8080"})
	viper.SetDefault("cors.methods", []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"})
	viper.SetDefault("cors.headers", []string{"Origin", "Content-Type", "Authorization", "Accept", "X-Requested-With"})
	viper.SetDefault("cors.credentials", true)

	// 配置文件读取
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")

	if err := viper.ReadInConfig(); err != nil {
		log.Printf("Config file not found: %v", err)
	}

	// 配置反序列化到结构体
	if err := viper.Unmarshal(&AppConfig); err != nil {
		log.Fatalf("Failed to unmarshal: %v", err)
	}
}