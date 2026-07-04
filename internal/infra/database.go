package infra

import (
	"backend/pkg/logger"
	"fmt"
	"os"

	"github.com/jinzhu/gorm"
	_ "github.com/lib/pq"
)

// DatabaseConfig 数据库配置结构
// 用于存储连接 PostgreSQL 所需的所有参数
type DatabaseConfig struct {
	Host     string // 数据库主机地址
	Port     int    // 数据库端口
	User     string // 数据库用户名
	Password string // 数据库密码
	DBName   string // 数据库名称
	SSLMode  string // SSL模式（disable/require）
}

// ensureDatabaseExists 确保数据库存在
// 连接到默认的postgres数据库，检查目标数据库是否存在，不存在则创建
func ensureDatabaseExists(cfg DatabaseConfig) {
	defaultDSN := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=postgres sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.SSLMode)

	var err error
	var defaultDB *gorm.DB
	defaultDB, err = gorm.Open("postgres", defaultDSN)
	if err != nil {
		logger.Fatalf("Failed to connect to default database: %v", err)
	}
	defer defaultDB.Close()

	type existsResult struct {
		Exists bool
	}
	checkSQL := fmt.Sprintf("SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname='%s') as exists", cfg.DBName)
	var result existsResult
	defaultDB.Raw(checkSQL).Scan(&result)
	exists := result.Exists

	if !exists {
		createSQL := fmt.Sprintf("CREATE DATABASE \"%s\"", cfg.DBName)
		if err := defaultDB.Exec(createSQL).Error; err != nil {
			logger.Fatalf("Failed to create database: %v", err)
			os.Exit(1)
		}
		logger.Infof("Database '%s' created successfully", cfg.DBName)
	}
}

// NewDB 创建数据库连接
// 参数说明：
//   - cfg: 数据库配置
// 返回值：
//   - *gorm.DB: GORM数据库实例
//
// 业务流程：
//   1. 调用 ensureDatabaseExists 确保数据库存在
//   2. 使用配置参数构建DSN连接字符串
//   3. 调用 gorm.Open 连接数据库
//   4. 启用日志模式
//   5. 执行 Ping 验证连接
//   6. 返回数据库实例
func NewDB(cfg DatabaseConfig) *gorm.DB {
	ensureDatabaseExists(cfg)

	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode)

	db, err := gorm.Open("postgres", dsn)
	if err != nil {
		logger.Fatalf("Failed to connect database: %v", err)
	}

	db.LogMode(true)
	db.SetLogger(logger.NewGormLogger())

	// 数据库ping失败时应立即退出系统，确保应用不处于不一致状态
	if err := db.DB().Ping(); err != nil {
		logger.Fatalf("Failed to ping database: %v . System will exit now", err)
		os.Exit(1)
	}

	logger.Infof("Database connected successfully")
	return db
}