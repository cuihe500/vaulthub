package service

import (
	"github.com/cuihe500/vaulthub/internal/config"
	"github.com/cuihe500/vaulthub/internal/database/models"
	"github.com/cuihe500/vaulthub/pkg/errors"
	"github.com/cuihe500/vaulthub/pkg/logger"
	"gorm.io/gorm"
)

// SystemConfigService 系统配置服务
type SystemConfigService struct {
	db            *gorm.DB
	configManager *config.ConfigManager
}

// NewSystemConfigService 创建系统配置服务实例
func NewSystemConfigService(db *gorm.DB, configManager *config.ConfigManager) *SystemConfigService {
	return &SystemConfigService{
		db:            db,
		configManager: configManager,
	}
}

// ConfigItem 配置项
type ConfigItem struct {
	ConfigKey   string `json:"config_key"`
	ConfigValue string `json:"config_value"`
	Description string `json:"description"`
}

// ListConfigsResponse 配置列表响应
type ListConfigsResponse struct {
	Configs []ConfigItem `json:"configs"`
	Total   int64        `json:"total"`
}

// ListConfigs 获取所有配置列表
// 从数据库读取完整信息（包括description）
func (s *SystemConfigService) ListConfigs() (*ListConfigsResponse, error) {
	var configs []models.SystemConfig
	if err := s.db.Order("config_key").Find(&configs).Error; err != nil {
		logger.Error("查询配置列表失败", logger.Err(err))
		return nil, errors.Wrap(errors.CodeDatabaseError, err)
	}

	items := make([]ConfigItem, len(configs))
	for i, cfg := range configs {
		items[i] = ConfigItem{
			ConfigKey:   cfg.ConfigKey,
			ConfigValue: cfg.ConfigValue,
			Description: cfg.Description,
		}
	}

	return &ListConfigsResponse{
		Configs: items,
		Total:   int64(len(configs)),
	}, nil
}

// GetConfig 获取单个配置
func (s *SystemConfigService) GetConfig(key string) (*ConfigItem, error) {
	var cfg models.SystemConfig
	if err := s.db.Where("config_key = ?", key).First(&cfg).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			logger.Warn("配置项不存在", logger.String("key", key))
			return nil, errors.New(errors.CodeResourceNotFound, "配置项不存在")
		}
		logger.Error("查询配置失败", logger.String("key", key), logger.Err(err))
		return nil, errors.Wrap(errors.CodeDatabaseError, err)
	}

	return &ConfigItem{
		ConfigKey:   cfg.ConfigKey,
		ConfigValue: cfg.ConfigValue,
		Description: cfg.Description,
	}, nil
}

// UpdateConfigRequest 更新配置请求
type UpdateConfigRequest struct {
	ConfigValue string `json:"config_value" binding:"required"`
}

// UpdateConfig 更新配置
// 通过ConfigManager更新，自动触发热更新
func (s *SystemConfigService) UpdateConfig(key string, req *UpdateConfigRequest) error {
	// 检查配置是否存在
	var cfg models.SystemConfig
	if err := s.db.Where("config_key = ?", key).First(&cfg).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			logger.Warn("配置项不存在", logger.String("key", key))
			return errors.New(errors.CodeResourceNotFound, "配置项不存在")
		}
		logger.Error("查询配置失败", logger.String("key", key), logger.Err(err))
		return errors.Wrap(errors.CodeDatabaseError, err)
	}

	// 通过ConfigManager更新（会触发观察者）
	if err := s.configManager.Set(key, req.ConfigValue); err != nil {
		return err
	}

	logger.Info("配置更新成功",
		logger.String("key", key),
		logger.String("old_value", cfg.ConfigValue),
		logger.String("new_value", req.ConfigValue))

	return nil
}

// BatchUpdateConfigRequest 批量更新配置请求
type BatchUpdateConfigRequest struct {
	Configs []struct {
		ConfigKey   string `json:"config_key" binding:"required"`
		ConfigValue string `json:"config_value" binding:"required"`
	} `json:"configs" binding:"required,min=1"`
}

// BatchUpdateConfigs 批量更新配置
// 在事务中更新多个配置，全部成功或全部失败
func (s *SystemConfigService) BatchUpdateConfigs(req *BatchUpdateConfigRequest) error {
	// 先验证所有配置键都存在
	for _, cfg := range req.Configs {
		var existingCfg models.SystemConfig
		if err := s.db.Where("config_key = ?", cfg.ConfigKey).First(&existingCfg).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				logger.Warn("配置项不存在", logger.String("key", cfg.ConfigKey))
				return errors.New(errors.CodeResourceNotFound, "配置项不存在: "+cfg.ConfigKey)
			}
			logger.Error("查询配置失败", logger.String("key", cfg.ConfigKey), logger.Err(err))
			return errors.Wrap(errors.CodeDatabaseError, err)
		}
	}

	// 使用事务批量更新
	err := s.db.Transaction(func(tx *gorm.DB) error {
		for _, cfg := range req.Configs {
			// 通过ConfigManager更新（会触发观察者）
			if err := s.configManager.Set(cfg.ConfigKey, cfg.ConfigValue); err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		logger.Error("批量更新配置失败", logger.Err(err))
		return err
	}

	logger.Info("批量更新配置成功", logger.Int("count", len(req.Configs)))
	return nil
}

// ReloadConfigs 重新加载配置
// 从数据库重新加载所有配置到内存
func (s *SystemConfigService) ReloadConfigs() error {
	if err := s.configManager.Reload(); err != nil {
		logger.Error("重新加载配置失败", logger.Err(err))
		return err
	}

	logger.Info("配置已重新加载")
	return nil
}
