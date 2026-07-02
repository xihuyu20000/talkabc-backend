package main

import (
	"backend/internal/config"
	"backend/internal/infra"
	"backend/internal/router"
	"backend/pkg/logger"
	"os"
	"strconv"
)

func main() {
	config.InitConfig()

	if config.AppConfig.System.Reset == 1 {
		dbCfg := infra.DatabaseConfig{
			Host:     config.AppConfig.Database.Host,
			Port:     config.AppConfig.Database.Port,
			User:     config.AppConfig.Database.User,
			Password: config.AppConfig.Database.Password,
			DBName:   config.AppConfig.Database.DBName,
			SSLMode:  config.AppConfig.Database.SSLMode,
		}
		redisCfg := infra.RedisConfig{
			Host:     config.AppConfig.Redis.Host,
			Port:     config.AppConfig.Redis.Port,
			Password: config.AppConfig.Redis.Password,
			DB:       config.AppConfig.Redis.DB,
		}
		infra.ResetAll(dbCfg, redisCfg)
	}

	logger.InitLogger(config.AppConfig.System.LogLevel)

	dbCfg := infra.DatabaseConfig{
		Host:     config.AppConfig.Database.Host,
		Port:     config.AppConfig.Database.Port,
		User:     config.AppConfig.Database.User,
		Password: config.AppConfig.Database.Password,
		DBName:   config.AppConfig.Database.DBName,
		SSLMode:  config.AppConfig.Database.SSLMode,
	}
	config.DB = infra.NewDB(dbCfg)
	infra.AutoMigrate(config.DB)

	redisCfg := infra.RedisConfig{
		Host:     config.AppConfig.Redis.Host,
		Port:     config.AppConfig.Redis.Port,
		Password: config.AppConfig.Redis.Password,
		DB:       config.AppConfig.Redis.DB,
	}
	config.RDB = infra.NewRedis(redisCfg)

	createUploadDirs()

	r := router.InitRouter()

	port := config.AppConfig.Server.Port

	logger.Info("Server starting on port %d...", port)

	r.Run(":" + strconv.Itoa(port))
}

// createUploadDirs 创建上传文件所需的目录
// 目录说明：
//   - ./uploads: 根上传目录
//   - ./uploads/avatars: 用户头像存储目录
//   - ./uploads/moments: 动态图片/视频存储目录
//   - ./uploads/messages: 聊天消息文件存储目录
func createUploadDirs() {
	// 定义所有需要创建的目录
	dirs := []string{"./uploads", "./uploads/avatars", "./uploads/moments", "./uploads/messages"}

	// 遍历创建每个目录
	for _, dir := range dirs {
		// MkdirAll会递归创建目录
		// 0755是目录权限：所有者可读写执行，其他人可读执行
		if err := os.MkdirAll(dir, 0755); err != nil {
			// 如果创建失败，打印错误日志但继续运行
			logger.Error("Failed to create directory %s: %v", dir, err)
		}
	}
}
