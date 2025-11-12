package config

import (
	"sync"

	"github.com/cuihe500/vaulthub/internal/database/models"
	"github.com/cuihe500/vaulthub/pkg/errors"
	"github.com/cuihe500/vaulthub/pkg/logger"
	"gorm.io/gorm"
)

// ConfigChangeCallback 配置变更回调函数
// key: 配置键, oldValue: 旧值, newValue: 新值
type ConfigChangeCallback func(key, oldValue, newValue string)

// ConfigManager 系统配置管理器
// 提供内存缓存、配置监控和热更新能力
type ConfigManager struct {
	db       *gorm.DB
	cache    map[string]string                 // 配置缓存
	watchers map[string][]ConfigChangeCallback // 配置观察者
	mu       sync.RWMutex                      // 读写锁，保护cache和watchers
}

// NewConfigManager 创建配置管理器
// 启动时从数据库加载所有配置到内存
func NewConfigManager(db *gorm.DB) (*ConfigManager, error) {
	manager := &ConfigManager{
		db:       db,
		cache:    make(map[string]string),
		watchers: make(map[string][]ConfigChangeCallback),
	}

	// 启动时加载所有配置
	if err := manager.loadAll(); err != nil {
		return nil, err
	}

	logger.Info("配置管理器初始化完成", logger.Int("config_count", len(manager.cache)))
	return manager, nil
}

// loadAll 从数据库加载所有配置到内存
func (m *ConfigManager) loadAll() error {
	var configs []models.SystemConfig
	if err := m.db.Find(&configs).Error; err != nil {
		logger.Error("加载系统配置失败", logger.Err(err))
		return errors.Wrap(errors.CodeDatabaseError, err)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	for _, config := range configs {
		m.cache[config.ConfigKey] = config.ConfigValue
	}

	logger.Debug("系统配置已加载到内存", logger.Int("count", len(configs)))
	return nil
}

// Get 获取配置值
// 先查内存缓存，未命中再查数据库并更新缓存
func (m *ConfigManager) Get(key string) (string, error) {
	// 先读缓存
	m.mu.RLock()
	value, exists := m.cache[key]
	m.mu.RUnlock()

	if exists {
		return value, nil
	}

	// 缓存未命中，查询数据库
	var config models.SystemConfig
	if err := m.db.Where("config_key = ?", key).First(&config).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return "", errors.New(errors.CodeResourceNotFound, "配置项不存在")
		}
		logger.Error("查询配置失败", logger.String("key", key), logger.Err(err))
		return "", errors.Wrap(errors.CodeDatabaseError, err)
	}

	// 更新缓存
	m.mu.Lock()
	m.cache[key] = config.ConfigValue
	m.mu.Unlock()

	return config.ConfigValue, nil
}

// GetWithDefault 获取配置值，不存在时返回默认值
// 不会返回错误，便于快速使用
func (m *ConfigManager) GetWithDefault(key, defaultValue string) string {
	value, err := m.Get(key)
	if err != nil {
		return defaultValue
	}
	return value
}

// Set 设置配置值
// 写数据库 + 更新缓存 + 触发所有观察者
func (m *ConfigManager) Set(key, value string) error {
	// 获取旧值用于通知观察者
	m.mu.RLock()
	oldValue := m.cache[key]
	m.mu.RUnlock()

	// 更新数据库
	result := m.db.Model(&models.SystemConfig{}).
		Where("config_key = ?", key).
		Update("config_value", value)

	if result.Error != nil {
		logger.Error("更新配置失败", logger.String("key", key), logger.Err(result.Error))
		return errors.Wrap(errors.CodeDatabaseError, result.Error)
	}

	if result.RowsAffected == 0 {
		logger.Warn("配置项不存在", logger.String("key", key))
		return errors.New(errors.CodeResourceNotFound, "配置项不存在")
	}

	// 更新缓存
	m.mu.Lock()
	m.cache[key] = value
	// 复制观察者列表，避免在锁内调用回调
	callbacks := make([]ConfigChangeCallback, len(m.watchers[key]))
	copy(callbacks, m.watchers[key])
	m.mu.Unlock()

	logger.Info("配置已更新",
		logger.String("key", key),
		logger.String("old_value", oldValue),
		logger.String("new_value", value))

	// 触发所有观察者（在锁外执行，避免死锁）
	for _, callback := range callbacks {
		callback(key, oldValue, value)
	}

	return nil
}

// Watch 注册配置变更观察者
// 当配置变更时，回调函数会被调用
func (m *ConfigManager) Watch(key string, callback ConfigChangeCallback) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.watchers[key] = append(m.watchers[key], callback)
	logger.Debug("配置观察者已注册", logger.String("key", key))
}

// Reload 重新加载配置
// 从数据库重新加载所有配置，用于手动刷新
func (m *ConfigManager) Reload() error {
	return m.loadAll()
}

// GetAll 获取所有配置
// 返回当前内存中的所有配置副本
func (m *ConfigManager) GetAll() map[string]string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// 创建副本避免外部修改
	result := make(map[string]string, len(m.cache))
	for k, v := range m.cache {
		result[k] = v
	}

	return result
}
