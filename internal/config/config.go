package config

import (
	"backend/pkg/logger"
	"fmt"
	"os"

	"github.com/go-redis/redis/v8"
	"github.com/jinzhu/gorm"
	"gopkg.in/yaml.v2"
)

// ==================== 配置结构体定义 ====================

type Config struct {
	System       SystemConfig       `yaml:"system" json:"system"`
	Logger       LoggerConfig       `yaml:"logger" json:"logger"`
	Security     SecurityConfig     `yaml:"security" json:"security"`
	SMSProvider  SMSProviderConfig  `yaml:"sms_provider" json:"sms_provider"`
	Server       ServerConfig       `yaml:"server" json:"server"`
	Database     DatabaseConfig     `yaml:"database" json:"database"`
	JWT          JWTConfig          `yaml:"jwt" json:"jwt"`
	Upload       UploadConfig       `yaml:"upload" json:"upload"`
	Redis        RedisConfig        `yaml:"redis" json:"redis"`
	CORS         CORSConfig         `yaml:"cors" json:"cors"`
}

type SystemConfig struct {
	Reset int `yaml:"reset" json:"reset"`
}

type LoggerConfig struct {
	Level       string `yaml:"level" json:"level"`
	Format      string `yaml:"format" json:"format"`
	Output      string `yaml:"output" json:"output"`
	FilePath    string `yaml:"file_path" json:"file_path"`
	MaxSize     int    `yaml:"max_size" json:"max_size"`
	MaxBackups  int    `yaml:"max_backups" json:"max_backups"`
	MaxAge      int    `yaml:"max_age" json:"max_age"`
	Compress    bool   `yaml:"compress" json:"compress"`
}

type SecurityConfig struct {
	SMSValidMinutes         int `yaml:"sms_valid_minutes" json:"sms_valid_minutes"`
	SMSCooldownSeconds      int `yaml:"sms_cooldown_seconds" json:"sms_cooldown_seconds"`
	SMSHourlyLimit          int `yaml:"sms_hourly_limit" json:"sms_hourly_limit"`
	IPRegisterHourlyLimit   int `yaml:"ip_register_hourly_limit" json:"ip_register_hourly_limit"`
	IPLoginMinuteLimit      int `yaml:"ip_login_minute_limit" json:"ip_login_minute_limit"`
	LoginFailureLockMinutes int `yaml:"login_failure_lock_minutes" json:"login_failure_lock_minutes"`
	RequireDailyCaptcha     int `yaml:"require_daily_captcha" json:"require_daily_captcha"`
}

type SMSProviderConfig struct {
	Default  string           `yaml:"default" json:"default"`
	Aliyun   AliyunSMSConfig  `yaml:"aliyun" json:"aliyun"`
	Huawei   HuaweiSMSConfig  `yaml:"huawei" json:"huawei"`
	Tencent  TencentSMSConfig `yaml:"tencent" json:"tencent"`
}

type AliyunSMSConfig struct {
	AccessKeyId     string `yaml:"access_key_id" json:"access_key_id"`
	AccessKeySecret string `yaml:"access_key_secret" json:"access_key_secret"`
	RegionID        string `yaml:"region_id" json:"region_id"`
	SignName        string `yaml:"sign_name" json:"sign_name"`
	TemplateCode    string `yaml:"template_code" json:"template_code"`
	SchemeName      string `yaml:"scheme_name" json:"scheme_name"`
	CountryCode     string `yaml:"country_code" json:"country_code"`
}

type HuaweiSMSConfig struct {
	AppKey    string `yaml:"app_key" json:"app_key"`
	AppSecret string `yaml:"app_secret" json:"app_secret"`
	SignName  string `yaml:"sign_name" json:"sign_name"`
	TemplateID string `yaml:"template_id" json:"template_id"`
}

type TencentSMSConfig struct {
	SecretID   string `yaml:"secret_id" json:"secret_id"`
	SecretKey  string `yaml:"secret_key" json:"secret_key"`
	RegionID   string `yaml:"region_id" json:"region_id"`
	SignName   string `yaml:"sign_name" json:"sign_name"`
	TemplateID string `yaml:"template_id" json:"template_id"`
}

type ServerConfig struct {
	Port int `yaml:"port" json:"port"`
}

type DatabaseConfig struct {
	Host     string `yaml:"host" json:"host"`
	Port     int    `yaml:"port" json:"port"`
	User     string `yaml:"user" json:"user"`
	Password string `yaml:"password" json:"password"`
	DBName   string `yaml:"dbname" json:"dbname"`
	SSLMode  string `yaml:"sslmode" json:"sslmode"`
}

type JWTConfig struct {
	Secret      string `yaml:"secret" json:"secret"`
	ExpiresHour int    `yaml:"expires_hour" json:"expires_hour"`
}

type UploadConfig struct {
	AvatarPath  string `yaml:"avatar_path" json:"avatar_path"`
	MomentPath  string `yaml:"moment_path" json:"moment_path"`
	MessagePath string `yaml:"message_path" json:"message_path"`
}

type RedisConfig struct {
	Host     string `yaml:"host" json:"host"`
	Port     int    `yaml:"port" json:"port"`
	Password string `yaml:"password" json:"password"`
	DB       int    `yaml:"db" json:"db"`
}

type CORSConfig struct {
	Origins     []string `yaml:"origins" json:"origins"`
	Methods     []string `yaml:"methods" json:"methods"`
	Headers     []string `yaml:"headers" json:"headers"`
	Credentials bool     `yaml:"credentials" json:"credentials"`
}

// ==================== 全局变量 ====================

var AppConfig Config
var DB *gorm.DB
var RDB *redis.Client



func getDefaultConfig() *Config {
	return &Config{
		System: SystemConfig{Reset: 0},
		Logger: LoggerConfig{
			Level:       "info",
			Format:      "console",
			Output:      "console",
			FilePath:    "./logs/app.log",
			MaxSize:     100,
			MaxBackups:  30,
			MaxAge:      7,
			Compress:    true,
		},
		Security: SecurityConfig{
			SMSValidMinutes:         5,
			SMSCooldownSeconds:      60,
			SMSHourlyLimit:          10,
			IPRegisterHourlyLimit:   10,
			IPLoginMinuteLimit:      10,
			LoginFailureLockMinutes: 5,
			RequireDailyCaptcha:     1,
		},
		SMSProvider: SMSProviderConfig{
			Default: "aliyun",
			Aliyun: AliyunSMSConfig{
				RegionID: "cn-hangzhou",
			},
		},
		Server: ServerConfig{Port: 8080},
		Database: DatabaseConfig{
			Host:     "localhost",
			Port:     5432,
			User:     "postgres",
			Password: "",
			DBName:   "talkabc",
			SSLMode:  "disable",
		},
		JWT: JWTConfig{
			Secret:      "talkabc_secret_key",
			ExpiresHour: 24,
		},
		Redis: RedisConfig{
			Host:     "localhost",
			Port:     6379,
			Password: "",
			DB:       0,
		},
		Upload: UploadConfig{
			AvatarPath:  "./uploads/avatars",
			MomentPath:  "./uploads/moments",
			MessagePath: "./uploads/messages",
		},
		CORS: CORSConfig{
			Origins:     []string{"http://localhost:3000", "http://localhost:8080"},
			Methods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			Headers:     []string{"Origin", "Content-Type", "Authorization", "Accept", "X-Requested-With"},
			Credentials: true,
		},
	}
}
func loadConfig(filePath string, cfg *Config) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	if err := yaml.Unmarshal(content, cfg); err != nil {
		return err
	}

	return nil
}
// InitConfig 初始化全局配置
// 加载流程：
//   1. 先创建带有默认值的配置对象
//   2. 尝试从配置文件加载，配置文件中的值覆盖默认值
//   3. 配置文件优先级：./config.yaml
func InitConfigDefault() {
	InitConfig("./config.yaml")
}
func InitConfig(filePath string) {

	// 1. 先创建带有默认值的配置对象
	cfg := getDefaultConfig()

	// 2. 尝试从配置文件加载，覆盖默认值
	if err := loadConfig(filePath, cfg); err != nil {
		logger.Fatalf("Failed to load config from %s: %v . System will exit now", filePath, err)
		os.Exit(1)
	}

	// 3. 将加载的配置赋值给全局变量
	AppConfig = *cfg

	// 调试：输出加载的日志配置
	fmt.Printf("Logger config loaded - Level: '%s'\n",
		AppConfig.Logger.Level)

	// 4. 初始化日志
	logger.InitLogger(&AppConfig.Logger)

	// 5. 输出配置信息
	logger.Infof("[Config] System - Reset: %d", AppConfig.System.Reset)
	logger.Debugf("[Config] Debug level test - Logger.Level: %s", AppConfig.Logger.Level)
	str := fmt.Sprintf("%+v", AppConfig)
	logger.Infof("[Config] Full config: \n%s", str)
}

func InitConfigSafe(filePath string) {
	cfg := getDefaultConfig()

	if err := loadConfig(filePath, cfg); err != nil {
		logger.Fatalf("Failed to load config from %s: %v, System will exit now", filePath, err)
		os.Exit(1)
	}

	AppConfig = *cfg

	logger.InitLogger(&AppConfig.Logger)
}