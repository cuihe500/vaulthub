package service

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/cuihe500/vaulthub/internal/config"
	"github.com/cuihe500/vaulthub/internal/database/models"
	"github.com/cuihe500/vaulthub/pkg/crypto"
	"github.com/cuihe500/vaulthub/pkg/errors"
	"github.com/cuihe500/vaulthub/pkg/logger"
	"gorm.io/gorm"
)

// KeyRotationService 密钥轮换服务
type KeyRotationService struct {
	db                *gorm.DB
	encryptionService *EncryptionService
	configManager     *config.ConfigManager // 配置管理器
	migrationTasks    sync.Map              // 存储进行中的迁移任务，key: userUUID, value: *MigrationTask
	batchSize         int                   // 批次大小（缓存）
	batchSleepMS      int                   // 批次休眠时间（缓存）
	mu                sync.RWMutex          // 保护batchSize和batchSleepMS的并发访问
}

// NewKeyRotationService 创建密钥轮换服务实例
// 从ConfigManager读取配置并注册观察者监听配置变更
func NewKeyRotationService(db *gorm.DB, encryptionService *EncryptionService, configManager *config.ConfigManager) *KeyRotationService {
	s := &KeyRotationService{
		db:                db,
		encryptionService: encryptionService,
		configManager:     configManager,
		migrationTasks:    sync.Map{},
		batchSize:         100, // 默认值
		batchSleepMS:      100, // 默认值
	}

	// 从ConfigManager加载初始配置
	s.loadConfig()

	// 注册配置变更观察者
	configManager.Watch(models.ConfigKeyKeyRotationBatchSize, s.onBatchSizeChange)
	configManager.Watch(models.ConfigKeyKeyRotationBatchSleepMS, s.onBatchSleepChange)

	logger.Info("密钥轮换服务初始化完成",
		logger.Int("batch_size", s.batchSize),
		logger.Int("batch_sleep_ms", s.batchSleepMS))

	return s
}

// loadConfig 从ConfigManager加载配置
func (s *KeyRotationService) loadConfig() {
	// 加载批次大小
	if value := s.configManager.GetWithDefault(models.ConfigKeyKeyRotationBatchSize, models.ConfigValueKeyRotationBatchSizeDefault); value != "" {
		if size, err := strconv.Atoi(value); err == nil && size > 0 {
			s.mu.Lock()
			s.batchSize = size
			s.mu.Unlock()
		}
	}

	// 加载批次休眠时间
	if value := s.configManager.GetWithDefault(models.ConfigKeyKeyRotationBatchSleepMS, models.ConfigValueKeyRotationBatchSleepMSDefault); value != "" {
		if sleepMS, err := strconv.Atoi(value); err == nil && sleepMS >= 0 {
			s.mu.Lock()
			s.batchSleepMS = sleepMS
			s.mu.Unlock()
		}
	}
}

// onBatchSizeChange 批次大小配置变更回调
func (s *KeyRotationService) onBatchSizeChange(key, oldValue, newValue string) {
	size, err := strconv.Atoi(newValue)
	if err != nil || size <= 0 {
		logger.Warn("批次大小配置无效，保持原值",
			logger.String("new_value", newValue),
			logger.Int("current_value", s.batchSize))
		return
	}

	s.mu.Lock()
	oldSize := s.batchSize
	s.batchSize = size
	s.mu.Unlock()

	logger.Info("密钥轮换批次大小已更新",
		logger.Int("old_value", oldSize),
		logger.Int("new_value", size))
}

// onBatchSleepChange 批次休眠时间配置变更回调
func (s *KeyRotationService) onBatchSleepChange(key, oldValue, newValue string) {
	sleepMS, err := strconv.Atoi(newValue)
	if err != nil || sleepMS < 0 {
		logger.Warn("批次休眠时间配置无效，保持原值",
			logger.String("new_value", newValue),
			logger.Int("current_value", s.batchSleepMS))
		return
	}

	s.mu.Lock()
	oldSleepMS := s.batchSleepMS
	s.batchSleepMS = sleepMS
	s.mu.Unlock()

	logger.Info("密钥轮换批次休眠时间已更新",
		logger.Int("old_value", oldSleepMS),
		logger.Int("new_value", sleepMS))
}

// getKeyRotationConfig 获取密钥轮换配置
// 从内存缓存读取，无需访问数据库
func (s *KeyRotationService) getKeyRotationConfig() (batchSize int, batchSleepMS int) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.batchSize, s.batchSleepMS
}

// MigrationTask 数据迁移任务状态
type MigrationTask struct {
	UserUUID        string     `json:"user_uuid"`
	OldVersion      int        `json:"old_version"`
	NewVersion      int        `json:"new_version"`
	TotalSecrets    int64      `json:"total_secrets"`
	MigratedSecrets int64      `json:"migrated_secrets"`
	FailedSecrets   int64      `json:"failed_secrets"`
	Status          string     `json:"status"` // running, completed, failed
	StartedAt       time.Time  `json:"started_at"`
	CompletedAt     *time.Time `json:"completed_at,omitempty"`
	Error           string     `json:"error,omitempty"`
	ctx             context.Context
	cancel          context.CancelFunc
	mu              sync.RWMutex
}

// RotateDEKRequest 密钥轮换请求
type RotateDEKRequest struct {
	UserUUID    string `json:"-"`                               // 由handler从上下文设置
	SecurityPIN string `json:"security_pin" binding:"required"` // 安全密码，用于验证和解密DEK
}

// RotateDEKResponse 密钥轮换响应
type RotateDEKResponse struct {
	UserEncryptionKey *models.SafeUserEncryptionKey `json:"user_encryption_key"`
	Message           string                        `json:"message"`
}

// RotateDEK 手动触发密钥轮换
// 限制：每30天最多轮换一次（可配置）
func (s *KeyRotationService) RotateDEK(req *RotateDEKRequest) (*RotateDEKResponse, error) {
	// 1. 获取用户密钥配置
	var userKey models.UserEncryptionKey
	if err := s.db.Where("user_uuid = ?", req.UserUUID).First(&userKey).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			logger.Warn("用户加密密钥不存在", logger.String("user_uuid", req.UserUUID))
			return nil, errors.New(errors.CodeResourceNotFound, "用户加密密钥不存在")
		}
		logger.Error("查询用户加密密钥失败", logger.Err(err))
		return nil, errors.Wrap(errors.CodeDatabaseError, err)
	}

	// 2. 检查是否有正在进行的轮换
	if userKey.RotationStatus == string(models.RotationStatusInProgress) {
		logger.Warn("密钥轮换已在进行中", logger.String("user_uuid", req.UserUUID))
		return nil, errors.New(errors.CodeResourceConflict, "密钥轮换已在进行中，请等待完成")
	}

	// 3. 检查轮换频率限制（每30天最多一次）
	if userKey.LastRotationAt != nil {
		timeSinceLastRotation := time.Since(*userKey.LastRotationAt)
		if timeSinceLastRotation < 30*24*time.Hour {
			remainingDays := int((30*24*time.Hour - timeSinceLastRotation).Hours() / 24)
			logger.Warn("密钥轮换过于频繁",
				logger.String("user_uuid", req.UserUUID),
				logger.Int("remaining_days", remainingDays))
			return nil, errors.New(errors.CodeTooManyRequests,
				fmt.Sprintf("密钥轮换过于频繁，请在%d天后再试", remainingDays))
		}
	}

	// 4. 验证安全密码（快速失败）
	if userKey.SecurityPINHash != "" {
		if !crypto.VerifyPassword(req.SecurityPIN, userKey.SecurityPINHash) {
			logger.Warn("安全密码验证失败", logger.String("user_uuid", req.UserUUID))
			return nil, errors.New(errors.CodeInvalidCredentials, "安全密码错误")
		}
	}

	// 5. 从安全密码派生KEK并解密旧DEK
	kek, err := crypto.DeriveKEK(req.SecurityPIN, userKey.KEKSalt)
	if err != nil {
		logger.Error("派生KEK失败", logger.Err(err))
		return nil, errors.WithMessage(errors.CodeKeyDerivationError, "密钥派生失败", err)
	}
	defer crypto.ClearBytes(kek)

	oldDEK, err := s.encryptionService.decryptDEK(userKey.EncryptedDEK, kek)
	if err != nil {
		logger.Warn("解密DEK失败，安全密码可能错误", logger.String("user_uuid", req.UserUUID), logger.Err(err))
		return nil, errors.New(errors.CodeInvalidCredentials, "安全密码错误")
	}
	defer crypto.ClearBytes(oldDEK)

	// 5. 生成新的DEK
	newDEK, err := crypto.GenerateRandomBytes(crypto.AESKeySize)
	if err != nil {
		logger.Error("生成新DEK失败", logger.Err(err))
		return nil, err
	}
	defer crypto.ClearBytes(newDEK)

	newVersion := userKey.DEKVersion + 1

	// 6. 用当前KEK加密新DEK
	encryptedNewDEK, nonce, authTag, err := crypto.EncryptAESGCM(newDEK, kek)
	if err != nil {
		logger.Error("加密新DEK失败", logger.Err(err))
		return nil, err
	}

	newEncryptedDEKBlob := make([]byte, 0, len(encryptedNewDEK)+len(nonce)+len(authTag))
	newEncryptedDEKBlob = append(newEncryptedDEKBlob, encryptedNewDEK...)
	newEncryptedDEKBlob = append(newEncryptedDEKBlob, nonce...)
	newEncryptedDEKBlob = append(newEncryptedDEKBlob, authTag...)

	// 7. 同时用恢复密钥加密新DEK（更新备份）
	// 从恢复密钥哈希重新生成恢复密钥（注意：这里需要用户提供恢复密钥，或者保持不变）
	// 为简化流程，这里保持EncryptedDEKRecovery不变，在实际应用中可能需要更新
	// TODO: 考虑是否需要在轮换时同时更新恢复密钥加密的DEK

	now := time.Now()

	// 8. 在事务中更新数据库
	err = s.db.Transaction(func(tx *gorm.DB) error {
		// 保存旧DEK用于数据迁移
		updates := map[string]interface{}{
			"encrypted_dek":       newEncryptedDEKBlob,
			"encrypted_dek_old":   userKey.EncryptedDEK, // 保存旧DEK
			"dek_version":         newVersion,
			"rotation_status":     string(models.RotationStatusInProgress),
			"rotation_started_at": now,
			"last_rotation_at":    now,
		}

		if err := tx.Model(&models.UserEncryptionKey{}).
			Where("user_uuid = ?", req.UserUUID).
			Updates(updates).Error; err != nil {
			logger.Error("更新用户密钥失败", logger.Err(err))
			return errors.Wrap(errors.CodeDatabaseError, err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	logger.Info("密钥轮换启动成功",
		logger.String("user_uuid", req.UserUUID),
		logger.Int("old_version", userKey.DEKVersion),
		logger.Int("new_version", newVersion))

	// 9. 启动后台数据迁移任务
	// 注意：需要重新读取oldDEK和newDEK，因为上面的defer会清零
	oldDEKForMigration, _ := s.encryptionService.decryptDEK(userKey.EncryptedDEK, kek)
	newDEKForMigration, _ := s.encryptionService.decryptDEK(newEncryptedDEKBlob, kek)

	go s.migrateSecretsToNewDEK(req.UserUUID, userKey.DEKVersion, newVersion, oldDEKForMigration, newDEKForMigration)

	// 10. 重新查询更新后的用户密钥
	var updatedKey models.UserEncryptionKey
	if err := s.db.Where("user_uuid = ?", req.UserUUID).First(&updatedKey).Error; err != nil {
		logger.Error("查询更新后的用户密钥失败", logger.Err(err))
		return nil, errors.Wrap(errors.CodeDatabaseError, err)
	}

	return &RotateDEKResponse{
		UserEncryptionKey: updatedKey.ToSafe(),
		Message:           "密钥轮换已启动，数据迁移正在后台进行",
	}, nil
}

// migrateSecretsToNewDEK 后台数据迁移任务
// 该函数在goroutine中运行，渐进式地将所有旧数据重新加密
func (s *KeyRotationService) migrateSecretsToNewDEK(userUUID string, oldVersion, newVersion int, oldDEK, newDEK []byte) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	task := &MigrationTask{
		UserUUID:   userUUID,
		OldVersion: oldVersion,
		NewVersion: newVersion,
		Status:     "running",
		StartedAt:  time.Now(),
		ctx:        ctx,
		cancel:     cancel,
	}

	// 存储任务状态
	s.migrationTasks.Store(userUUID, task)
	defer s.migrationTasks.Delete(userUUID)

	// 确保密钥使用完毕后清零
	defer crypto.ClearBytes(oldDEK)
	defer crypto.ClearBytes(newDEK)

	logger.Info("开始数据迁移任务",
		logger.String("user_uuid", userUUID),
		logger.Int("old_version", oldVersion),
		logger.Int("new_version", newVersion))

	// 获取需要迁移的秘密总数
	var total int64
	if err := s.db.Model(&models.EncryptedSecret{}).
		Where("user_uuid = ? AND dek_version = ?", userUUID, oldVersion).
		Count(&total).Error; err != nil {
		logger.Error("获取待迁移秘密总数失败", logger.Err(err))
		s.markMigrationFailed(task, err)
		return
	}

	task.mu.Lock()
	task.TotalSecrets = total
	task.mu.Unlock()

	logger.Info("待迁移秘密总数", logger.String("user_uuid", userUUID), logger.Int64("total", total))

	// 如果没有需要迁移的数据，直接标记完成
	if total == 0 {
		s.markMigrationCompleted(task, userUUID)
		return
	}

	// 从数据库读取密钥轮换配置
	batchSize, batchSleepMS := s.getKeyRotationConfig()
	batchSleep := time.Duration(batchSleepMS) * time.Millisecond

	logger.Info("使用密钥轮换配置",
		logger.String("user_uuid", userUUID),
		logger.Int("batch_size", batchSize),
		logger.Int("batch_sleep_ms", batchSleepMS))

	// 分批处理，避免一次性锁定大量数据
	offset := 0

	for {
		// 检查是否被取消
		select {
		case <-ctx.Done():
			logger.Warn("数据迁移任务被取消", logger.String("user_uuid", userUUID))
			return
		default:
		}

		var secrets []models.EncryptedSecret
		err := s.db.Where("user_uuid = ? AND dek_version = ?", userUUID, oldVersion).
			Limit(batchSize).
			Offset(offset).
			Find(&secrets).Error

		if err != nil {
			logger.Error("查询待迁移秘密失败", logger.Err(err))
			s.markMigrationFailed(task, err)
			return
		}

		// 如果没有更多数据，结束迁移
		if len(secrets) == 0 {
			break
		}

		// 处理这一批数据
		for _, secret := range secrets {
			// 用旧DEK解密
			plainData, err := crypto.DecryptAESGCM(secret.EncryptedData, oldDEK, secret.Nonce, secret.AuthTag)
			if err != nil {
				logger.Error("用旧DEK解密失败",
					logger.Err(err),
					logger.String("secret_uuid", secret.SecretUUID))
				task.mu.Lock()
				task.FailedSecrets++
				task.mu.Unlock()
				continue // 继续处理下一个
			}

			// 用新DEK重新加密
			newEncryptedData, newNonce, newAuthTag, err := crypto.EncryptAESGCM(plainData, newDEK)
			if err != nil {
				logger.Error("用新DEK加密失败",
					logger.Err(err),
					logger.String("secret_uuid", secret.SecretUUID))
				crypto.ClearBytes(plainData)
				task.mu.Lock()
				task.FailedSecrets++
				task.mu.Unlock()
				continue
			}

			// 清理明文数据
			crypto.ClearBytes(plainData)

			// 更新数据库
			if err := s.db.Model(&models.EncryptedSecret{}).
				Where("id = ?", secret.ID).
				Updates(map[string]interface{}{
					"encrypted_data": newEncryptedData,
					"nonce":          newNonce,
					"auth_tag":       newAuthTag,
					"dek_version":    newVersion,
				}).Error; err != nil {
				logger.Error("更新秘密失败",
					logger.Err(err),
					logger.Uint("secret_id", secret.ID))
				task.mu.Lock()
				task.FailedSecrets++
				task.mu.Unlock()
				continue
			}

			// 更新进度
			task.mu.Lock()
			task.MigratedSecrets++
			task.mu.Unlock()
		}

		offset += batchSize

		// 避免CPU占用过高，每批次间隔休息（使用配置的休眠时间）
		time.Sleep(batchSleep)

		// 记录进度
		task.mu.RLock()
		progress := float64(task.MigratedSecrets) / float64(task.TotalSecrets) * 100
		task.mu.RUnlock()
		logger.Info("数据迁移进度",
			logger.String("user_uuid", userUUID),
			logger.Float64("progress", progress),
			logger.Int64("migrated", task.MigratedSecrets),
			logger.Int64("total", task.TotalSecrets))
	}

	// 迁移完成
	s.markMigrationCompleted(task, userUUID)
}

// markMigrationCompleted 标记迁移完成
func (s *KeyRotationService) markMigrationCompleted(task *MigrationTask, userUUID string) {
	now := time.Now()
	task.mu.Lock()
	task.Status = "completed"
	task.CompletedAt = &now
	task.mu.Unlock()

	// 更新数据库，清理旧DEK
	if err := s.db.Model(&models.UserEncryptionKey{}).
		Where("user_uuid = ?", userUUID).
		Updates(map[string]interface{}{
			"encrypted_dek_old": nil,
			"rotation_status":   string(models.RotationStatusCompleted),
		}).Error; err != nil {
		logger.Error("更新轮换状态为已完成失败", logger.Err(err))
	}

	logger.Info("数据迁移任务完成",
		logger.String("user_uuid", userUUID),
		logger.Int64("migrated", task.MigratedSecrets),
		logger.Int64("failed", task.FailedSecrets),
		logger.Int64("total", task.TotalSecrets))
}

// markMigrationFailed 标记迁移失败
func (s *KeyRotationService) markMigrationFailed(task *MigrationTask, err error) {
	now := time.Now()
	task.mu.Lock()
	task.Status = "failed"
	task.CompletedAt = &now
	task.Error = err.Error()
	task.mu.Unlock()

	logger.Error("数据迁移任务失败",
		logger.String("user_uuid", task.UserUUID),
		logger.Err(err))
}

// GetRotationStatus 获取密钥轮换状态
func (s *KeyRotationService) GetRotationStatus(userUUID string) (*MigrationTask, error) {
	// 首先检查内存中的任务状态（运行中的任务）
	if value, ok := s.migrationTasks.Load(userUUID); ok {
		task := value.(*MigrationTask)
		task.mu.RLock()
		defer task.mu.RUnlock()

		// 创建副本返回，避免并发问题
		taskCopy := &MigrationTask{
			UserUUID:        task.UserUUID,
			OldVersion:      task.OldVersion,
			NewVersion:      task.NewVersion,
			TotalSecrets:    task.TotalSecrets,
			MigratedSecrets: task.MigratedSecrets,
			FailedSecrets:   task.FailedSecrets,
			Status:          task.Status,
			StartedAt:       task.StartedAt,
			CompletedAt:     task.CompletedAt,
			Error:           task.Error,
		}
		return taskCopy, nil
	}

	// 如果内存中没有，查询数据库获取历史状态
	var userKey models.UserEncryptionKey
	if err := s.db.Where("user_uuid = ?", userUUID).First(&userKey).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New(errors.CodeResourceNotFound, "用户加密密钥不存在")
		}
		logger.Error("查询用户加密密钥失败", logger.Err(err))
		return nil, errors.Wrap(errors.CodeDatabaseError, err)
	}

	// 构造历史状态响应
	task := &MigrationTask{
		UserUUID:    userUUID,
		NewVersion:  userKey.DEKVersion,
		Status:      userKey.RotationStatus,
		StartedAt:   *userKey.RotationStartedAt,
		CompletedAt: userKey.LastRotationAt,
	}

	return task, nil
}

// CheckAndRotateExpiredKeys 检查并自动轮换过期的密钥
// 该函数由定时任务调用，每天运行一次
func (s *KeyRotationService) CheckAndRotateExpiredKeys() error {
	// 查找180天未轮换的密钥
	var userKeys []models.UserEncryptionKey
	expiredDate := time.Now().AddDate(0, 0, -180) // 180天前

	err := s.db.Where("(last_rotation_at IS NULL AND created_at < ?) OR (last_rotation_at < ?)",
		expiredDate, expiredDate).
		Where("rotation_status != ?", string(models.RotationStatusInProgress)).
		Find(&userKeys).Error

	if err != nil {
		logger.Error("查询过期密钥失败", logger.Err(err))
		return errors.Wrap(errors.CodeDatabaseError, err)
	}

	if len(userKeys) == 0 {
		logger.Info("没有需要自动轮换的密钥")
		return nil
	}

	logger.Info("发现需要自动轮换的密钥", logger.Int("count", len(userKeys)))

	// TODO: 对于自动轮换，需要实现无密码轮换机制（使用主密钥或其他安全机制）
	// 当前实现需要用户密码，因此自动轮换需要额外的安全机制
	// 这里仅记录日志，实际轮换需要用户主动触发

	for _, userKey := range userKeys {
		lastRotation := "never"
		if userKey.LastRotationAt != nil {
			lastRotation = userKey.LastRotationAt.Format(time.RFC3339)
		}
		logger.Warn("用户密钥需要轮换",
			logger.String("user_uuid", userKey.UserUUID),
			logger.String("last_rotation", lastRotation))
		// TODO: 发送通知提醒用户轮换密钥
	}

	return nil
}
