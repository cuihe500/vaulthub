package app

import (
	"github.com/cuihe500/vaulthub/internal/service"
	"github.com/cuihe500/vaulthub/pkg/logger"
	"github.com/robfig/cron/v3"
)

// Scheduler 定时任务调度器
// 管理所有的定时任务，在应用启动时初始化
type Scheduler struct {
	cron               *cron.Cron
	keyRotationService *service.KeyRotationService
}

// NewScheduler 创建定时任务调度器实例
func NewScheduler(keyRotationService *service.KeyRotationService) *Scheduler {
	// 使用带秒级精度的cron
	c := cron.New(cron.WithSeconds())

	return &Scheduler{
		cron:               c,
		keyRotationService: keyRotationService,
	}
}

// Start 启动所有定时任务
func (s *Scheduler) Start() error {
	logger.Info("启动定时任务调度器")

	// 每天凌晨2点检查过期密钥
	// Cron表达式格式：秒 分 时 日 月 周
	// "0 0 2 * * *" = 每天2点0分0秒
	_, err := s.cron.AddFunc("0 0 2 * * *", func() {
		logger.Info("开始执行定时任务：检查过期密钥")
		if err := s.keyRotationService.CheckAndRotateExpiredKeys(); err != nil {
			logger.Error("检查过期密钥失败", logger.Err(err))
		} else {
			logger.Info("完成定时任务：检查过期密钥")
		}
	})

	if err != nil {
		logger.Error("添加密钥轮换定时任务失败", logger.Err(err))
		return err
	}

	// 启动cron调度器
	s.cron.Start()
	logger.Info("定时任务调度器已启动")

	return nil
}

// Stop 停止所有定时任务
func (s *Scheduler) Stop() {
	if s.cron != nil {
		logger.Info("停止定时任务调度器")
		s.cron.Stop()
		logger.Info("定时任务调度器已停止")
	}
}
