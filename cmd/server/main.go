// @title TalkABC API
// @version 1.0
// @description TalkABC 聊天交友平台后端 API 文档
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization

package main

import (
	"backend/internal/config"
	"backend/internal/infra"
	"backend/internal/router"
	"backend/internal/sms"
	"backend/pkg/logger"
	"os"
	"strconv"
)

func main() {
	// 1. 加载配置文件
	config.InitConfig()

	defer logger.Sync()

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

	// 2. 重置应用系统（包括数据库、文件夹）
	if config.AppConfig.System.Reset == 1 {
		infra.ResetAll(dbCfg, redisCfg)
		createUploadDirs()
	}

	// 3. 连接数据库
	config.DB = infra.NewDB(dbCfg)
	infra.AutoMigrate(config.DB)
	config.RDB = infra.NewRedis(redisCfg)

	// 4. 连接短信网关
	if err := sms.InitSMSGateway(&config.AppConfig.SMSProvider); err != nil {
		logger.Fatalf("Failed to initialize SMS gateway: %v", err)
		os.Exit(1)
	}

	// 5. 初始化路由
	r := router.InitRouter()

	port := config.AppConfig.Server.Port

	logger.Infof("Server starting on port %d...", port)

	// 6. 启动服务器
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
			logger.Fatalf("InitLogger error: Failed to create directory %s: %v, system exit\n", dir, err)
			os.Exit(1)
		}
	}
}
