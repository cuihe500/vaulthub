package redis

import (
	"context"
	"time"

	"github.com/cuihe500/vaulthub/internal/config"
	"github.com/cuihe500/vaulthub/pkg/logger"
	"github.com/redis/go-redis/v9"
)

// Client Redis客户端封装
type Client struct {
	client *redis.Client
}

// NewClient 创建Redis客户端
func NewClient(cfg config.RedisConfig) (*Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Address(),
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		logger.Error("Redis连接失败",
			logger.String("addr", cfg.Address()),
			logger.Err(err))
		return nil, err
	}

	logger.Info("Redis连接成功",
		logger.String("addr", cfg.Address()),
		logger.Int("db", cfg.DB))

	return &Client{client: rdb}, nil
}

// Close 关闭Redis连接
func (c *Client) Close() error {
	return c.client.Close()
}

// Set 设置键值对,带过期时间
func (c *Client) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return c.client.Set(ctx, key, value, expiration).Err()
}

// Get 获取键对应的值
func (c *Client) Get(ctx context.Context, key string) (string, error) {
	return c.client.Get(ctx, key).Result()
}

// Del 删除键
func (c *Client) Del(ctx context.Context, keys ...string) error {
	return c.client.Del(ctx, keys...).Err()
}

// Exists 检查键是否存在
func (c *Client) Exists(ctx context.Context, keys ...string) (int64, error) {
	return c.client.Exists(ctx, keys...).Result()
}

// Expire 设置键的过期时间
func (c *Client) Expire(ctx context.Context, key string, expiration time.Duration) error {
	return c.client.Expire(ctx, key, expiration).Err()
}

// TTL 获取键的剩余过期时间
func (c *Client) TTL(ctx context.Context, key string) (time.Duration, error) {
	return c.client.TTL(ctx, key).Result()
}
