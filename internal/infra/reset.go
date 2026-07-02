package infra

import (
	"context"
	"fmt"
	"os"

	"github.com/go-redis/redis/v8"
	"github.com/jinzhu/gorm"
	_ "github.com/lib/pq"
)

// ResetDatabase 删除并重建数据库
// 参数说明：
//   - cfg: 数据库配置
//
// 业务流程：
//   1. 连接到默认的postgres数据库
//   2. 执行 DROP DATABASE IF EXISTS 删除目标数据库
//   3. 关闭数据库连接
func ResetDatabase(cfg DatabaseConfig) {
	defaultDSN := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=postgres sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.SSLMode)

	defaultDB, err := gorm.Open("postgres", defaultDSN)
	if err != nil {
		fmt.Printf("Failed to connect to default database for reset: %v\n", err)
		return
	}
	defer defaultDB.Close()

	dropSQL := fmt.Sprintf("DROP DATABASE IF EXISTS \"%s\"", cfg.DBName)
	if err := defaultDB.Exec(dropSQL).Error; err != nil {
		fmt.Printf("Failed to drop database: %v\n", err)
		return
	}
	fmt.Printf("Database '%s' dropped successfully\n", cfg.DBName)
}

// ResetRedis 清空Redis数据
// 参数说明：
//   - cfg: Redis配置
//
// 业务流程：
//   1. 创建Redis客户端
//   2. 执行 FLUSHDB 清空当前数据库
//   3. 关闭客户端
func ResetRedis(cfg RedisConfig) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	if err := rdb.FlushDB(context.Background()).Err(); err != nil {
		fmt.Printf("Failed to flush Redis: %v\n", err)
		return
	}
	fmt.Println("Redis data flushed successfully")
}

// ResetLogs 删除日志目录
// 删除 ./logs 目录及所有日志文件
func ResetLogs() {
	if err := os.RemoveAll("./logs"); err != nil {
		fmt.Printf("Failed to remove logs directory: %v\n", err)
		return
	}
	fmt.Println("Logs directory removed successfully")
}

// ResetUploads 删除上传文件目录
// 删除 ./uploads 目录及所有子目录和文件
func ResetUploads() {
	dirs := []string{"./uploads", "./uploads/avatars", "./uploads/moments", "./uploads/messages"}
	for _, dir := range dirs {
		if err := os.RemoveAll(dir); err != nil {
			fmt.Printf("Failed to remove directory %s: %v\n", dir, err)
		}
	}
	fmt.Println("Upload directories removed successfully")
}

// ResetAll 执行所有重置操作
// 参数说明：
//   - dbCfg: 数据库配置
//   - redisCfg: Redis配置
//
// 业务流程：
//   1. 调用 ResetLogs 删除日志
//   2. 调用 ResetUploads 删除上传文件
//   3. 调用 ResetRedis 清空Redis
//   4. 调用 ResetDatabase 删除数据库
//
// 注意：此操作会删除所有数据，谨慎使用
func ResetAll(dbCfg DatabaseConfig, redisCfg RedisConfig) {
	fmt.Println("=== Starting data reset ===")

	ResetLogs()
	ResetUploads()
	ResetRedis(redisCfg)
	ResetDatabase(dbCfg)

	fmt.Println("=== Data reset completed ===")
}