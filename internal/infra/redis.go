package infra

import (
	"backend/pkg/logger"
	"context"
	"fmt"

	"github.com/go-redis/redis/v8"
)

// RedisConfig Redis配置结构
// 用于存储连接 Redis 所需的所有参数
type RedisConfig struct {
	Host     string // Redis主机地址
	Port     int    // Redis端口
	Password string // Redis密码（无密码时为空）
	DB       int    // Redis数据库编号
}

// NewRedis 创建Redis客户端
// 参数说明：
//   - cfg: Redis配置
// 返回值：
//   - *redis.Client: Redis客户端实例
//
// 业务流程：
//   1. 构建Redis地址（host:port）
//   2. 创建redis.Client实例
//   3. 执行Ping验证连接
//   4. 返回Redis客户端实例
func NewRedis(cfg RedisConfig) *redis.Client {
	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)

	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	_, err := rdb.Ping(context.Background()).Result()
	if err != nil {
		logger.Fatalf("Failed to connect to Redis: %v", err)
	}

	logger.Infof("Redis connection successful")
	return rdb
}