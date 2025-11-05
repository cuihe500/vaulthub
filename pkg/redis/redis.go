package redis

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/cuihe500/vaulthub/internal/config"
	"github.com/cuihe500/vaulthub/pkg/logger"
	"github.com/redis/go-redis/v9"
	"github.com/redis/go-redis/v9/maintnotifications"
)

// Client Redis客户端封装，支持单机、哨兵、集群三种模式
type Client struct {
	client redis.UniversalClient // 使用通用接口支持所有模式
	mode   string
}

// NewClient 根据配置创建Redis客户端，支持三种部署模式
func NewClient(cfg config.RedisConfig) (*Client, error) {
	mode := strings.ToLower(strings.TrimSpace(cfg.Mode))
	if mode == "" {
		mode = "standalone" // 默认单机模式
	}

	var rdb redis.UniversalClient
	var err error

	switch mode {
	case "standalone":
		rdb, err = newStandaloneClient(cfg)
	case "sentinel":
		rdb, err = newSentinelClient(cfg)
	case "cluster":
		rdb, err = newClusterClient(cfg)
	default:
		return nil, fmt.Errorf("不支持的Redis模式: %s，支持的模式: standalone, sentinel, cluster", cfg.Mode)
	}

	if err != nil {
		return nil, err
	}

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		logger.Error("Redis连接失败",
			logger.String("mode", mode),
			logger.Err(err))
		return nil, err
	}

	logger.Info("Redis连接成功", logger.String("mode", mode))

	return &Client{
		client: rdb,
		mode:   mode,
	}, nil
}

// newStandaloneClient 创建单机模式客户端
func newStandaloneClient(cfg config.RedisConfig) (redis.UniversalClient, error) {
	logger.Info("创建Redis单机客户端",
		logger.String("addr", cfg.Address()),
		logger.Int("db", cfg.DB))

	return redis.NewClient(&redis.Options{
		Addr:     cfg.Address(),
		Password: cfg.Password,
		DB:       cfg.DB,
		// https://github.com/redis/go-redis/issues/3536
		MaintNotificationsConfig: &maintnotifications.Config{
			Mode: maintnotifications.ModeDisabled,
		},
	}), nil
}

// newSentinelClient 创建哨兵模式客户端
func newSentinelClient(cfg config.RedisConfig) (redis.UniversalClient, error) {
	if cfg.MasterName == "" {
		return nil, fmt.Errorf("哨兵模式需要配置master_name")
	}
	if len(cfg.Sentinels) == 0 {
		return nil, fmt.Errorf("哨兵模式需要配置sentinels节点列表")
	}

	logger.Info("创建Redis哨兵客户端",
		logger.String("master_name", cfg.MasterName),
		logger.Any("sentinels", cfg.Sentinels),
		logger.Int("db", cfg.DB))

	return redis.NewFailoverClient(&redis.FailoverOptions{
		MasterName:       cfg.MasterName,
		SentinelAddrs:    cfg.Sentinels,
		SentinelPassword: cfg.SentinelPassword, // 哨兵节点密码
		Password:         cfg.Password,         // Redis主从节点密码
		DB:               cfg.DB,
		DisableIndentity: true, // 禁用客户端身份识别
	}), nil
}

// newClusterClient 创建集群模式客户端
func newClusterClient(cfg config.RedisConfig) (redis.UniversalClient, error) {
	if len(cfg.Addrs) == 0 {
		return nil, fmt.Errorf("集群模式需要配置addrs节点列表")
	}

	// Redis Cluster只支持db0，如果配置了其他DB则警告
	if cfg.DB != 0 {
		logger.Warn("Redis集群模式不支持DB参数，将被忽略",
			logger.Int("configured_db", cfg.DB),
			logger.String("actual_db", "0"))
	}

	logger.Info("创建Redis集群客户端",
		logger.Any("addrs", cfg.Addrs))

	return redis.NewClusterClient(&redis.ClusterOptions{
		Addrs:            cfg.Addrs,
		Password:         cfg.Password,
		DisableIndentity: true, // 禁用客户端身份识别
	}), nil
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
