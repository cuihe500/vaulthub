package service

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/cuihe500/vaulthub/internal/database/models"
	"github.com/cuihe500/vaulthub/pkg/logger"
	"gorm.io/gorm"
)

// AuditService 审计服务
type AuditService struct {
	db         *gorm.DB
	auditChan  chan *models.AuditLog
	workerSize int
	wg         sync.WaitGroup
	ctx        context.Context
	cancel     context.CancelFunc
}

// NewAuditService 创建审计服务
func NewAuditService(db *gorm.DB, bufferSize, workerSize int) *AuditService {
	ctx, cancel := context.WithCancel(context.Background())
	return &AuditService{
		db:         db,
		auditChan:  make(chan *models.AuditLog, bufferSize),
		workerSize: workerSize,
		ctx:        ctx,
		cancel:     cancel,
	}
}

// Start 启动审计服务（启动worker goroutines）
func (s *AuditService) Start() {
	logger.Info("审计服务启动中",
		logger.Int("worker数量", s.workerSize),
		logger.Int("缓冲区大小", cap(s.auditChan)))

	for i := 0; i < s.workerSize; i++ {
		s.wg.Add(1)
		go s.worker(i)
	}

	logger.Info("审计服务启动成功")
}

// Stop 停止审计服务（优雅关闭）
func (s *AuditService) Stop() {
	logger.Info("审计服务关闭中")

	// 关闭channel，不再接收新的审计日志
	close(s.auditChan)

	// 取消context，通知所有worker退出
	s.cancel()

	// 等待所有worker处理完剩余的日志
	s.wg.Wait()

	logger.Info("审计服务已关闭")
}

// worker 处理审计日志的工作goroutine
func (s *AuditService) worker(id int) {
	defer s.wg.Done()

	logger.Debug("审计worker启动", logger.Int("worker_id", id))

	for {
		select {
		case <-s.ctx.Done():
			// context取消，退出worker
			logger.Debug("审计worker收到退出信号", logger.Int("worker_id", id))
			return

		case log, ok := <-s.auditChan:
			if !ok {
				// channel已关闭，退出worker
				logger.Debug("审计channel已关闭，worker退出", logger.Int("worker_id", id))
				return
			}

			// 写入数据库
			if err := s.writeLog(log); err != nil {
				logger.Error("写入审计日志失败",
					logger.Int("worker_id", id),
					logger.String("user_uuid", log.UserUUID),
					logger.String("action", string(log.ActionType)),
					logger.Err(err))
				// 写入失败不阻塞业务，只记录错误日志
			}
		}
	}
}

// writeLog 写入审计日志到数据库
func (s *AuditService) writeLog(log *models.AuditLog) error {
	// 设置默认值
	if log.CreatedAt.IsZero() {
		log.CreatedAt = time.Now().UTC()
	}

	// 如果Details是map或struct，确保能正确序列化为JSON
	if log.Details != nil {
		// GORM会自动处理JSON序列化，这里只是确保数据有效
		if _, err := json.Marshal(log.Details); err != nil {
			logger.Warn("审计日志Details序列化失败，将清空Details字段",
				logger.Err(err))
			log.Details = nil
		}
	}

	err := s.db.Create(log).Error
	if err != nil {
		logger.Error("审计日志数据库写入失败",
			logger.String("user_uuid", log.UserUUID),
			logger.String("action_type", string(log.ActionType)),
			logger.Err(err))
	} else {
		logger.Debug("审计日志数据库写入成功",
			logger.String("uuid", log.UUID),
			logger.String("user_uuid", log.UserUUID),
			logger.String("action_type", string(log.ActionType)))
	}

	return err
}

// LogAsync 异步记录审计日志（非阻塞）
func (s *AuditService) LogAsync(log *models.AuditLog) {
	select {
	case s.auditChan <- log:
		// 成功发送到channel
	default:
		// channel已满，记录警告但不阻塞业务
		logger.Warn("审计channel已满，丢弃审计日志",
			logger.String("user_uuid", log.UserUUID),
			logger.String("action", string(log.ActionType)),
			logger.String("resource_type", string(log.ResourceType)))
	}
}

// QueryLogs 查询审计日志（支持分页和过滤）
func (s *AuditService) QueryLogs(req *QueryAuditLogsRequest) ([]*AuditLogDTO, int64, error) {
	query := s.db.Model(&models.AuditLog{})

	// 用户过滤
	if req.UserUUID != "" {
		query = query.Where("user_uuid = ?", req.UserUUID)
	}

	// 操作类型过滤
	if req.ActionType != "" {
		query = query.Where("action_type = ?", req.ActionType)
	}

	// 资源类型过滤
	if req.ResourceType != "" {
		query = query.Where("resource_type = ?", req.ResourceType)
	}

	// 状态过滤
	if req.Status != "" {
		query = query.Where("status = ?", req.Status)
	}

	// 时间范围过滤
	if !req.StartTime.IsZero() {
		query = query.Where("created_at >= ?", req.StartTime)
	}
	if !req.EndTime.IsZero() {
		query = query.Where("created_at <= ?", req.EndTime)
	}

	// 统计总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	var logs []*models.AuditLog
	query = query.Order("created_at DESC")

	// 如果未传分页参数，则全量导出（添加安全上限10000）
	if req.Page <= 0 && req.PageSize <= 0 {
		query = query.Limit(10000)
	} else {
		// 设置默认值
		if req.Page <= 0 {
			req.Page = 1
		}
		if req.PageSize <= 0 {
			req.PageSize = 20
		}
		offset := (req.Page - 1) * req.PageSize
		query = query.Limit(req.PageSize).Offset(offset)
	}

	err := query.Find(&logs).Error
	if err != nil {
		return nil, 0, err
	}

	// 转换为DTO
	dtos := make([]*AuditLogDTO, 0, len(logs))
	for _, log := range logs {
		dtos = append(dtos, toAuditLogDTO(log))
	}

	return dtos, total, nil
}

// toAuditLogDTO 将AuditLog模型转换为DTO
func toAuditLogDTO(log *models.AuditLog) *AuditLogDTO {
	dto := &AuditLogDTO{
		UUID:         log.UUID,
		UserUUID:     log.UserUUID,
		Username:     log.Username,
		ActionType:   string(log.ActionType),
		ResourceType: string(log.ResourceType),
		Status:       string(log.Status),
		CreatedAt:    log.CreatedAt,
	}

	if log.ResourceUUID != nil {
		dto.ResourceUUID = *log.ResourceUUID
	}
	if log.ResourceName != nil {
		dto.ResourceName = *log.ResourceName
	}
	if log.ErrorCode != nil {
		dto.ErrorCode = *log.ErrorCode
	}
	if log.ErrorMessage != nil {
		dto.ErrorMessage = *log.ErrorMessage
	}
	if log.IPAddress != nil {
		dto.IPAddress = *log.IPAddress
	}
	if log.UserAgent != nil {
		dto.UserAgent = *log.UserAgent
	}
	if log.RequestID != nil {
		dto.RequestID = *log.RequestID
	}

	dto.Details = log.Details

	return dto
}

// QueryAuditLogsRequest 查询审计日志请求
type QueryAuditLogsRequest struct {
	UserUUID     string    `form:"user_uuid"`
	ActionType   string    `form:"action_type"`
	ResourceType string    `form:"resource_type"`
	Status       string    `form:"status"`
	StartTime    time.Time `form:"start_time" time_format:"2006-01-02T15:04:05Z07:00"`
	EndTime      time.Time `form:"end_time" time_format:"2006-01-02T15:04:05Z07:00"`
	Page         int       `form:"page" binding:"omitempty,min=1"`
	PageSize     int       `form:"page_size" binding:"omitempty,min=1,max=10000"`
}

// ExportStatistics 导出各类型密钥统计总量
// 该方法用于全局统计，返回所有用户或指定用户的密钥类型分布情况
func (s *AuditService) ExportStatistics(userUUID string) (*SecretStatisticsExport, error) {
	query := s.db.Model(&models.EncryptedSecret{})

	// 如果指定用户UUID，则只统计该用户的数据
	if userUUID != "" {
		query = query.Where("user_uuid = ?", userUUID)
	}

	// 统计总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	// 按类型分组统计
	type TypeCount struct {
		SecretType models.SecretType
		Count      int64
	}
	var typeCounts []TypeCount
	if err := query.Select("secret_type, COUNT(*) as count").
		Group("secret_type").
		Scan(&typeCounts).Error; err != nil {
		return nil, err
	}

	// 构建返回结果
	result := &SecretStatisticsExport{
		TotalSecrets: total,
		ByType:       make(map[string]int64),
	}

	// 映射到结果
	for _, tc := range typeCounts {
		result.ByType[string(tc.SecretType)] = tc.Count
	}

	// 确保所有类型都有值（即使为0）
	allTypes := []models.SecretType{
		models.SecretTypeAPIKey,
		models.SecretTypeDBCredential,
		models.SecretTypeCertificate,
		models.SecretTypeSSHKey,
		models.SecretTypeToken,
		models.SecretTypePassword,
		models.SecretTypeOther,
	}
	for _, t := range allTypes {
		if _, exists := result.ByType[string(t)]; !exists {
			result.ByType[string(t)] = 0
		}
	}

	return result, nil
}

// AuditLogDTO 审计日志数据传输对象
type AuditLogDTO struct {
	UUID         string      `json:"uuid"`
	UserUUID     string      `json:"user_uuid"`
	Username     string      `json:"username"`
	ActionType   string      `json:"action_type"`
	ResourceType string      `json:"resource_type"`
	ResourceUUID string      `json:"resource_uuid,omitempty"`
	ResourceName string      `json:"resource_name,omitempty"`
	Status       string      `json:"status"`
	ErrorCode    int         `json:"error_code,omitempty"`
	ErrorMessage string      `json:"error_message,omitempty"`
	IPAddress    string      `json:"ip_address,omitempty"`
	UserAgent    string      `json:"user_agent,omitempty"`
	RequestID    string      `json:"request_id,omitempty"`
	Details      interface{} `json:"details,omitempty"`
	CreatedAt    time.Time   `json:"created_at"`
}

// SecretStatisticsExport 密钥统计导出数据
type SecretStatisticsExport struct {
	TotalSecrets int64            `json:"total_secrets"` // 密钥总数
	ByType       map[string]int64 `json:"by_type"`       // 按类型统计
}

// ExportOperationStatistics 导出操作统计（按时间范围）
// 该方法用于统计指定时间范围内的操作分布情况
func (s *AuditService) ExportOperationStatistics(userUUID string, startTime, endTime time.Time) (*OperationStatisticsExport, error) {
	query := s.db.Model(&models.AuditLog{})

	// 记录查询参数用于调试
	logger.Debug("导出操作统计",
		logger.String("user_uuid", userUUID),
		logger.Time("start_time", startTime),
		logger.Time("end_time", endTime))

	// 如果指定用户UUID，则只统计该用户的数据
	if userUUID != "" {
		query = query.Where("user_uuid = ?", userUUID)
	}

	// 时间范围过滤
	if !startTime.IsZero() {
		query = query.Where("created_at >= ?", startTime)
	}
	if !endTime.IsZero() {
		query = query.Where("created_at < ?", endTime)
	}

	// 统计总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	logger.Debug("操作统计查询结果", logger.Int64("total", total))

	// 按操作类型分组统计
	type ActionCount struct {
		ActionType models.ActionType
		Count      int64
	}
	var actionCounts []ActionCount
	if err := query.Select("action_type, COUNT(*) as count").
		Group("action_type").
		Scan(&actionCounts).Error; err != nil {
		return nil, err
	}

	// 构建返回结果
	result := &OperationStatisticsExport{
		TotalOperations: total,
		ByAction:        make(map[string]int64),
	}

	// 映射到结果
	for _, ac := range actionCounts {
		result.ByAction[string(ac.ActionType)] = ac.Count
	}

	// 确保所有操作类型都有值（即使为0）
	allActions := []models.ActionType{
		models.ActionCreate,
		models.ActionUpdate,
		models.ActionDelete,
		models.ActionAccess,
		models.ActionLogin,
		models.ActionLogout,
	}
	for _, a := range allActions {
		if _, exists := result.ByAction[string(a)]; !exists {
			result.ByAction[string(a)] = 0
		}
	}

	return result, nil
}

// OperationStatisticsExport 操作统计导出数据
type OperationStatisticsExport struct {
	TotalOperations int64            `json:"total_operations"` // 操作总数
	ByAction        map[string]int64 `json:"by_action"`        // 按操作类型统计
}
