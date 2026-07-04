package test

import (
	"backend/internal/config"
	"testing"
)

func TestInitConfig_LoadsCorrectly(t *testing.T) {
	config.InitConfig("../../config.yaml")

	t.Run("System config", func(t *testing.T) {
		if config.AppConfig.System.Reset != 1 {
			t.Errorf("System.Reset = %d, want 1", config.AppConfig.System.Reset)
		}
	})

	t.Run("Logger config", func(t *testing.T) {
		if config.AppConfig.Logger.Level != "debug" {
			t.Errorf("Logger.Level = %q, want %q", config.AppConfig.Logger.Level, "debug")
		}
		if config.AppConfig.Logger.Format != "console" {
			t.Errorf("Logger.Format = %q, want %q", config.AppConfig.Logger.Format, "console")
		}
		if config.AppConfig.Logger.Output != "both" {
			t.Errorf("Logger.Output = %q, want %q", config.AppConfig.Logger.Output, "both")
		}
		if config.AppConfig.Logger.FilePath != "./logs/app.log" {
			t.Errorf("Logger.FilePath = %q, want %q", config.AppConfig.Logger.FilePath, "./logs/app.log")
		}
		if config.AppConfig.Logger.MaxSize != 100 {
			t.Errorf("Logger.MaxSize = %d, want 100", config.AppConfig.Logger.MaxSize)
		}
		if config.AppConfig.Logger.MaxBackups != 30 {
			t.Errorf("Logger.MaxBackups = %d, want 30", config.AppConfig.Logger.MaxBackups)
		}
		if config.AppConfig.Logger.MaxAge != 7 {
			t.Errorf("Logger.MaxAge = %d, want 7", config.AppConfig.Logger.MaxAge)
		}
		if !config.AppConfig.Logger.Compress {
			t.Error("Logger.Compress = false, want true")
		}
	})

	t.Run("Security config", func(t *testing.T) {
		if config.AppConfig.Security.SMSValidMinutes != 5 {
			t.Errorf("Security.SMSValidMinutes = %d, want 5", config.AppConfig.Security.SMSValidMinutes)
		}
		if config.AppConfig.Security.SMSCooldownSeconds != 60 {
			t.Errorf("Security.SMSCooldownSeconds = %d, want 60", config.AppConfig.Security.SMSCooldownSeconds)
		}
		if config.AppConfig.Security.SMSHourlyLimit != 10 {
			t.Errorf("Security.SMSHourlyLimit = %d, want 10", config.AppConfig.Security.SMSHourlyLimit)
		}
		if config.AppConfig.Security.IPRegisterHourlyLimit != 10 {
			t.Errorf("Security.IPRegisterHourlyLimit = %d, want 10", config.AppConfig.Security.IPRegisterHourlyLimit)
		}
		if config.AppConfig.Security.IPLoginMinuteLimit != 10 {
			t.Errorf("Security.IPLoginMinuteLimit = %d, want 10", config.AppConfig.Security.IPLoginMinuteLimit)
		}
		if config.AppConfig.Security.LoginFailureLockMinutes != 5 {
			t.Errorf("Security.LoginFailureLockMinutes = %d, want 5", config.AppConfig.Security.LoginFailureLockMinutes)
		}
		if config.AppConfig.Security.RequireDailyCaptcha != 1 {
			t.Errorf("Security.RequireDailyCaptcha = %d, want 1", config.AppConfig.Security.RequireDailyCaptcha)
		}
	})

	t.Run("SMS Provider config", func(t *testing.T) {
		if config.AppConfig.SMSProvider.Default != "aliyun" {
			t.Errorf("SMSProvider.Default = %q, want %q", config.AppConfig.SMSProvider.Default, "aliyun")
		}
		if config.AppConfig.SMSProvider.Aliyun.RegionID != "cn-hangzhou" {
			t.Errorf("SMSProvider.Aliyun.RegionID = %q, want %q", config.AppConfig.SMSProvider.Aliyun.RegionID, "cn-hangzhou")
		}
		if config.AppConfig.SMSProvider.Aliyun.AccessKeyID != "LTAI5t5mcECvLgmdRVxh3Z84" {
			t.Errorf("SMSProvider.Aliyun.AccessKeyID = %q, want %q", config.AppConfig.SMSProvider.Aliyun.AccessKeyID, "LTAI5t5mcECvLgmdRVxh3Z84")
		}
		if config.AppConfig.SMSProvider.Aliyun.AccessKeySecret != "JB6g95jjzjb92AHOmzEbcsJ0PRTyo" {
			t.Errorf("SMSProvider.Aliyun.AccessKeySecret = %q, want %q", config.AppConfig.SMSProvider.Aliyun.AccessKeySecret, "JB6g95jjzjb92AHOmzEbcsJ0PRTyo")
		}
	})

	t.Run("Server config", func(t *testing.T) {
		if config.AppConfig.Server.Port != 8080 {
			t.Errorf("Server.Port = %d, want 8080", config.AppConfig.Server.Port)
		}
	})

	t.Run("Database config", func(t *testing.T) {
		if config.AppConfig.Database.Host != "localhost" {
			t.Errorf("Database.Host = %q, want %q", config.AppConfig.Database.Host, "localhost")
		}
		if config.AppConfig.Database.Port != 5432 {
			t.Errorf("Database.Port = %d, want 5432", config.AppConfig.Database.Port)
		}
		if config.AppConfig.Database.User != "postgres" {
			t.Errorf("Database.User = %q, want %q", config.AppConfig.Database.User, "postgres")
		}
		if config.AppConfig.Database.Password != "admin" {
			t.Errorf("Database.Password = %q, want %q", config.AppConfig.Database.Password, "admin")
		}
		if config.AppConfig.Database.DBName != "talkabc" {
			t.Errorf("Database.DBName = %q, want %q", config.AppConfig.Database.DBName, "talkabc")
		}
		if config.AppConfig.Database.SSLMode != "disable" {
			t.Errorf("Database.SSLMode = %q, want %q", config.AppConfig.Database.SSLMode, "disable")
		}
	})

	t.Run("JWT config", func(t *testing.T) {
		if config.AppConfig.JWT.Secret != "talkabc_jwt_secret_key" {
			t.Errorf("JWT.Secret = %q, want %q", config.AppConfig.JWT.Secret, "talkabc_jwt_secret_key")
		}
		if config.AppConfig.JWT.ExpiresHour != 24 {
			t.Errorf("JWT.ExpiresHour = %d, want 24", config.AppConfig.JWT.ExpiresHour)
		}
	})

	t.Run("Upload config", func(t *testing.T) {
		if config.AppConfig.Upload.AvatarPath != "./uploads/avatars" {
			t.Errorf("Upload.AvatarPath = %q, want %q", config.AppConfig.Upload.AvatarPath, "./uploads/avatars")
		}
		if config.AppConfig.Upload.MomentPath != "./uploads/moments" {
			t.Errorf("Upload.MomentPath = %q, want %q", config.AppConfig.Upload.MomentPath, "./uploads/moments")
		}
		if config.AppConfig.Upload.MessagePath != "./uploads/messages" {
			t.Errorf("Upload.MessagePath = %q, want %q", config.AppConfig.Upload.MessagePath, "./uploads/messages")
		}
	})

	t.Run("Redis config", func(t *testing.T) {
		if config.AppConfig.Redis.Host != "localhost" {
			t.Errorf("Redis.Host = %q, want %q", config.AppConfig.Redis.Host, "localhost")
		}
		if config.AppConfig.Redis.Port != 6379 {
			t.Errorf("Redis.Port = %d, want 6379", config.AppConfig.Redis.Port)
		}
		if config.AppConfig.Redis.Password != "" {
			t.Errorf("Redis.Password = %q, want %q", config.AppConfig.Redis.Password, "")
		}
		if config.AppConfig.Redis.DB != 0 {
			t.Errorf("Redis.DB = %d, want 0", config.AppConfig.Redis.DB)
		}
	})

	t.Run("CORS config", func(t *testing.T) {
		if len(config.AppConfig.CORS.Origins) != 1 || config.AppConfig.CORS.Origins[0] != "*" {
			t.Errorf("CORS.Origins = %v, want [\"*\"]", config.AppConfig.CORS.Origins)
		}
		if !config.AppConfig.CORS.Credentials {
			t.Error("CORS.Credentials = false, want true")
		}
	})
}
