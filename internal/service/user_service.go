package service

import (
	"github.com/cuihe500/vaulthub/internal/database/models"
	"github.com/cuihe500/vaulthub/pkg/errors"
	"github.com/cuihe500/vaulthub/pkg/logger"
	"gorm.io/gorm"
)

// UserService 用户服务
type UserService struct {
	db *gorm.DB
}

// NewUserService 创建用户服务实例
func NewUserService(db *gorm.DB) *UserService {
	return &UserService{
		db: db,
	}
}

// ListUsersRequest 用户列表请求
type ListUsersRequest struct {
	Page     int    `form:"page" binding:"omitempty,min=1"`
	PageSize int    `form:"page_size" binding:"omitempty,min=1,max=100"`
	Status   int    `form:"status" binding:"omitempty,min=1,max=3"`
	Role     string `form:"role" binding:"omitempty"`
}

// ListUsersResponse 用户列表响应
type ListUsersResponse struct {
	Users      []*models.SafeUser `json:"users"`
	Total      int64              `json:"total"`
	Page       int                `json:"page"`
	PageSize   int                `json:"page_size"`
	TotalPages int                `json:"total_pages"`
}

// ListUsers 获取用户列表
func (s *UserService) ListUsers(req *ListUsersRequest) (*ListUsersResponse, error) {
	// 设置默认值
	if req.Page == 0 {
		req.Page = 1
	}
	if req.PageSize == 0 {
		req.PageSize = 20
	}

	// 构建查询
	query := s.db.Model(&models.User{})

	// 应用过滤条件
	if req.Status > 0 {
		query = query.Where("status = ?", req.Status)
	}
	if req.Role != "" {
		query = query.Where("role = ?", req.Role)
	}

	// 获取总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		logger.Error("获取用户总数失败", logger.Err(err))
		return nil, errors.Wrap(errors.CodeDatabaseError, err)
	}

	// 分页查询
	var users []models.User
	offset := (req.Page - 1) * req.PageSize
	if err := query.Offset(offset).Limit(req.PageSize).Order("created_at DESC").Find(&users).Error; err != nil {
		logger.Error("查询用户列表失败", logger.Err(err))
		return nil, errors.Wrap(errors.CodeDatabaseError, err)
	}

	// 转换为安全用户信息
	safeUsers := make([]*models.SafeUser, len(users))
	for i, user := range users {
		safeUsers[i] = user.ToSafeUser()
	}

	// 计算总页数
	totalPages := int(total) / req.PageSize
	if int(total)%req.PageSize > 0 {
		totalPages++
	}

	return &ListUsersResponse{
		Users:      safeUsers,
		Total:      total,
		Page:       req.Page,
		PageSize:   req.PageSize,
		TotalPages: totalPages,
	}, nil
}

// GetUserByUUID 根据UUID获取用户信息
func (s *UserService) GetUserByUUID(userUUID string) (*models.SafeUser, error) {
	var user models.User
	if err := s.db.Where("uuid = ?", userUUID).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New(errors.CodeResourceNotFound, "user not found")
		}
		logger.Error("查询用户失败", logger.String("uuid", userUUID), logger.Err(err))
		return nil, errors.Wrap(errors.CodeDatabaseError, err)
	}

	return user.ToSafeUser(), nil
}

// UpdateUserStatusRequest 更新用户状态请求
type UpdateUserStatusRequest struct {
	Status models.UserStatus `json:"status" binding:"required,min=1,max=3"`
}

// UpdateUserStatus 更新用户状态
func (s *UserService) UpdateUserStatus(userUUID string, req *UpdateUserStatusRequest) (*models.SafeUser, error) {
	var user models.User
	if err := s.db.Where("uuid = ?", userUUID).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New(errors.CodeResourceNotFound, "user not found")
		}
		logger.Error("查询用户失败", logger.String("uuid", userUUID), logger.Err(err))
		return nil, errors.Wrap(errors.CodeDatabaseError, err)
	}

	// 更新状态
	user.Status = req.Status
	if err := s.db.Save(&user).Error; err != nil {
		logger.Error("更新用户状态失败", logger.String("uuid", userUUID), logger.Err(err))
		return nil, errors.Wrap(errors.CodeDatabaseError, err)
	}

	logger.Info("用户状态已更新", logger.String("uuid", userUUID), logger.Int("status", int(req.Status)))

	return user.ToSafeUser(), nil
}

// UpdateUserRoleRequest 更新用户角色请求
type UpdateUserRoleRequest struct {
	Role string `json:"role" binding:"required,oneof=admin user readonly"`
}

// UpdateUserRole 更新用户角色
func (s *UserService) UpdateUserRole(userUUID string, req *UpdateUserRoleRequest) (*models.SafeUser, error) {
	var user models.User
	if err := s.db.Where("uuid = ?", userUUID).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New(errors.CodeResourceNotFound, "user not found")
		}
		logger.Error("查询用户失败", logger.String("uuid", userUUID), logger.Err(err))
		return nil, errors.Wrap(errors.CodeDatabaseError, err)
	}

	// 更新角色
	user.Role = req.Role
	if err := s.db.Save(&user).Error; err != nil {
		logger.Error("更新用户角色失败", logger.String("uuid", userUUID), logger.Err(err))
		return nil, errors.Wrap(errors.CodeDatabaseError, err)
	}

	logger.Info("用户角色已更新", logger.String("uuid", userUUID), logger.String("role", req.Role))

	return user.ToSafeUser(), nil
}
