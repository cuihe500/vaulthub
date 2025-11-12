package service

import (
	"time"

	"github.com/cuihe500/vaulthub/internal/database/models"
	"github.com/cuihe500/vaulthub/pkg/logger"
	"gorm.io/gorm"
)

// StatisticsService 统计服务
type StatisticsService struct {
	db *gorm.DB
}

// NewStatisticsService 创建统计服务
func NewStatisticsService(db *gorm.DB) *StatisticsService {
	return &StatisticsService{
		db: db,
	}
}

// AggregateDaily 聚合每日统计数据
// 基于audit_logs和encrypted_secrets表计算统计数据
func (s *StatisticsService) AggregateDaily() error {
	logger.Info("开始聚合每日统计数据")

	// 获取昨天的日期范围（UTC）
	now := time.Now().UTC()
	yesterday := now.AddDate(0, 0, -1)
	startOfDay := time.Date(yesterday.Year(), yesterday.Month(), yesterday.Day(), 0, 0, 0, 0, time.UTC)
	endOfDay := startOfDay.Add(24 * time.Hour)

	// 查询所有活跃用户（昨天有操作记录的用户）
	var activeUsers []string
	if err := s.db.Model(&models.AuditLog{}).
		Select("DISTINCT user_uuid").
		Where("created_at >= ? AND created_at < ?", startOfDay, endOfDay).
		Pluck("user_uuid", &activeUsers).Error; err != nil {
		logger.Error("查询活跃用户失败", logger.Err(err))
		return err
	}

	logger.Info("找到活跃用户",
		logger.Int("用户数", len(activeUsers)),
		logger.Time("统计日期", startOfDay))

	// 为每个用户聚合统计数据
	for _, userUUID := range activeUsers {
		if err := s.aggregateUserDaily(userUUID, startOfDay, endOfDay); err != nil {
			logger.Error("聚合用户每日统计失败",
				logger.String("user_uuid", userUUID),
				logger.Err(err))
			// 继续处理其他用户
			continue
		}
	}

	logger.Info("完成聚合每日统计数据")
	return nil
}

// aggregateUserDaily 聚合单个用户的每日统计
func (s *StatisticsService) aggregateUserDaily(userUUID string, startOfDay, endOfDay time.Time) error {
	// 统计密钥数量（当前快照）
	var secretStats struct {
		Total       int64
		APIKey      int64
		Password    int64
		Certificate int64
		SSHKey      int64
		DBCred      int64
		Token       int64
		Other       int64
	}

	// 总数
	if err := s.db.Model(&models.EncryptedSecret{}).
		Where("user_uuid = ?", userUUID).
		Count(&secretStats.Total).Error; err != nil {
		return err
	}

	// 按类型统计
	type TypeCount struct {
		SecretType models.SecretType
		Count      int64
	}
	var typeCounts []TypeCount
	if err := s.db.Model(&models.EncryptedSecret{}).
		Select("secret_type, COUNT(*) as count").
		Where("user_uuid = ?", userUUID).
		Group("secret_type").
		Scan(&typeCounts).Error; err != nil {
		return err
	}

	// 映射到结构体
	for _, tc := range typeCounts {
		switch tc.SecretType {
		case models.SecretTypeAPIKey:
			secretStats.APIKey = tc.Count
		case models.SecretTypePassword:
			secretStats.Password = tc.Count
		case models.SecretTypeCertificate:
			secretStats.Certificate = tc.Count
		case models.SecretTypeSSHKey:
			secretStats.SSHKey = tc.Count
		case models.SecretTypeDBCredential:
			secretStats.DBCred = tc.Count
		case models.SecretTypeToken:
			secretStats.Token = tc.Count
		case models.SecretTypeOther:
			secretStats.Other = tc.Count
		}
	}

	// 统计操作次数（从审计日志）
	var opStats struct {
		Create int64
		Update int64
		Delete int64
		Access int64
		Login  int64
		Failed int64
	}

	// 创建操作
	if err := s.db.Model(&models.AuditLog{}).
		Where("user_uuid = ? AND action_type = ? AND created_at >= ? AND created_at < ?",
			userUUID, models.ActionCreate, startOfDay, endOfDay).
		Count(&opStats.Create).Error; err != nil {
		return err
	}

	// 更新操作
	if err := s.db.Model(&models.AuditLog{}).
		Where("user_uuid = ? AND action_type = ? AND created_at >= ? AND created_at < ?",
			userUUID, models.ActionUpdate, startOfDay, endOfDay).
		Count(&opStats.Update).Error; err != nil {
		return err
	}

	// 删除操作
	if err := s.db.Model(&models.AuditLog{}).
		Where("user_uuid = ? AND action_type = ? AND created_at >= ? AND created_at < ?",
			userUUID, models.ActionDelete, startOfDay, endOfDay).
		Count(&opStats.Delete).Error; err != nil {
		return err
	}

	// 访问操作
	if err := s.db.Model(&models.AuditLog{}).
		Where("user_uuid = ? AND action_type = ? AND created_at >= ? AND created_at < ?",
			userUUID, models.ActionAccess, startOfDay, endOfDay).
		Count(&opStats.Access).Error; err != nil {
		return err
	}

	// 登录成功
	if err := s.db.Model(&models.AuditLog{}).
		Where("user_uuid = ? AND action_type = ? AND status = ? AND created_at >= ? AND created_at < ?",
			userUUID, models.ActionLogin, models.AuditSuccess, startOfDay, endOfDay).
		Count(&opStats.Login).Error; err != nil {
		return err
	}

	// 登录失败
	if err := s.db.Model(&models.AuditLog{}).
		Where("user_uuid = ? AND action_type = ? AND status = ? AND created_at >= ? AND created_at < ?",
			userUUID, models.ActionLogin, models.AuditFailed, startOfDay, endOfDay).
		Count(&opStats.Failed).Error; err != nil {
		return err
	}

	totalOps := opStats.Create + opStats.Update + opStats.Delete + opStats.Access

	// 创建或更新统计记录
	// 注意：数据库中PrivateKeyCount字段实际用于存储DBCredential+Token的总和
	combinedCount := int(secretStats.DBCred + secretStats.Token)
	stats := &models.UserStatistics{
		UserUUID:         userUUID,
		StatDate:         startOfDay,
		StatType:         models.StatDaily,
		TotalSecrets:     int(secretStats.Total),
		APIKeyCount:      int(secretStats.APIKey),
		PasswordCount:    int(secretStats.Password),
		CertificateCount: int(secretStats.Certificate),
		SSHKeyCount:      int(secretStats.SSHKey),
		PrivateKeyCount:  combinedCount, // 复用此字段存储DBCredential+Token
		OtherCount:       int(secretStats.Other),
		CreateCount:      int(opStats.Create),
		UpdateCount:      int(opStats.Update),
		DeleteCount:      int(opStats.Delete),
		AccessCount:      int(opStats.Access),
		TotalOperations:  int(totalOps),
		LoginCount:       int(opStats.Login),
		FailedLoginCount: int(opStats.Failed),
	}

	// 使用事务确保原子性
	return s.db.Transaction(func(tx *gorm.DB) error {
		// 检查是否已存在
		var existing models.UserStatistics
		err := tx.Where("user_uuid = ? AND stat_date = ? AND stat_type = ?",
			userUUID, startOfDay, models.StatDaily).
			First(&existing).Error

		if err == gorm.ErrRecordNotFound {
			// 不存在，创建新记录
			return tx.Create(stats).Error
		} else if err != nil {
			return err
		}

		// 已存在，更新记录
		stats.ID = existing.ID
		stats.UUID = existing.UUID
		return tx.Model(&existing).Updates(stats).Error
	})
}

// GetUserStatistics 获取用户统计数据
func (s *StatisticsService) GetUserStatistics(req *GetStatisticsRequest) ([]*models.UserStatistics, error) {
	query := s.db.Model(&models.UserStatistics{})

	// 用户过滤
	if req.UserUUID != "" {
		query = query.Where("user_uuid = ?", req.UserUUID)
	}

	// 统计类型过滤
	if req.StatType != "" {
		query = query.Where("stat_type = ?", req.StatType)
	}

	// 时间范围过滤
	if !req.StartDate.IsZero() {
		query = query.Where("stat_date >= ?", req.StartDate)
	}
	if !req.EndDate.IsZero() {
		query = query.Where("stat_date <= ?", req.EndDate)
	}

	var stats []*models.UserStatistics
	err := query.Order("stat_date DESC").Find(&stats).Error
	return stats, err
}

// GetCurrentStatistics 获取用户当前统计（实时查询）
func (s *StatisticsService) GetCurrentStatistics(userUUID string) (*CurrentStatistics, error) {
	var result CurrentStatistics

	// 统计密钥总数
	if err := s.db.Model(&models.EncryptedSecret{}).
		Where("user_uuid = ?", userUUID).
		Count(&result.TotalSecrets).Error; err != nil {
		return nil, err
	}

	// 按类型统计
	type TypeCount struct {
		SecretType models.SecretType
		Count      int64
	}
	var typeCounts []TypeCount
	if err := s.db.Model(&models.EncryptedSecret{}).
		Select("secret_type, COUNT(*) as count").
		Where("user_uuid = ?", userUUID).
		Group("secret_type").
		Scan(&typeCounts).Error; err != nil {
		return nil, err
	}

	// 映射到结构体
	for _, tc := range typeCounts {
		switch tc.SecretType {
		case models.SecretTypeAPIKey:
			result.APIKeyCount = tc.Count
		case models.SecretTypePassword:
			result.PasswordCount = tc.Count
		case models.SecretTypeCertificate:
			result.CertificateCount = tc.Count
		case models.SecretTypeSSHKey:
			result.SSHKeyCount = tc.Count
		case models.SecretTypeDBCredential:
			result.PrivateKeyCount += tc.Count // 复用PrivateKeyCount字段
		case models.SecretTypeToken:
			result.PrivateKeyCount += tc.Count // 复用PrivateKeyCount字段
		case models.SecretTypeOther:
			result.OtherCount = tc.Count
		}
	}

	// 今日操作统计
	now := time.Now().UTC()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)

	// 今日总操作数
	if err := s.db.Model(&models.AuditLog{}).
		Where("user_uuid = ? AND created_at >= ?", userUUID, startOfDay).
		Count(&result.TodayOperations).Error; err != nil {
		return nil, err
	}

	return &result, nil
}

// GetStatisticsRequest 获取统计数据请求
type GetStatisticsRequest struct {
	UserUUID  string    `form:"user_uuid"`
	StatType  string    `form:"stat_type"`
	StartDate time.Time `form:"start_date" time_format:"2006-01-02"`
	EndDate   time.Time `form:"end_date" time_format:"2006-01-02"`
}

// CurrentStatistics 当前统计数据
type CurrentStatistics struct {
	TotalSecrets     int64 `json:"total_secrets"`
	APIKeyCount      int64 `json:"api_key_count"`
	PasswordCount    int64 `json:"password_count"`
	CertificateCount int64 `json:"certificate_count"`
	SSHKeyCount      int64 `json:"ssh_key_count"`
	PrivateKeyCount  int64 `json:"private_key_count"`
	OtherCount       int64 `json:"other_count"`
	TodayOperations  int64 `json:"today_operations"`
}
