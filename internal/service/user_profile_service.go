package service

import (
	"strings"

	"github.com/cuihe500/vaulthub/internal/database/models"
	apperrors "github.com/cuihe500/vaulthub/pkg/errors"
	"github.com/cuihe500/vaulthub/pkg/logger"
	"gorm.io/gorm"
)

// UserProfileService 用户档案服务
type UserProfileService struct {
	db *gorm.DB
}

// NewUserProfileService 创建用户档案服务实例
func NewUserProfileService(db *gorm.DB) *UserProfileService {
	return &UserProfileService{
		db: db,
	}
}

// CreateProfileRequest 创建用户档案请求
type CreateProfileRequest struct {
	Nickname string `json:"nickname" binding:"required,max=50"`
	Phone    string `json:"phone" binding:"omitempty,max=20"`
	Email    string `json:"email" binding:"required,email,max=100"`
}

// UpdateProfileRequest 更新用户档案请求
type UpdateProfileRequest struct {
	Nickname *string `json:"nickname" binding:"omitempty,max=50"`
	Phone    *string `json:"phone" binding:"omitempty,max=20"`
	Email    *string `json:"email" binding:"omitempty,email,max=100"`
}

// GetProfile 获取用户档案信息
func (s *UserProfileService) GetProfile(userID uint) (*models.SafeUserProfile, error) {
	var profile models.UserProfile
	if err := s.db.Where("user_id = ?", userID).First(&profile).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperrors.New(apperrors.CodeResourceNotFound, "用户档案不存在")
		}
		logger.Error("查询用户档案失败", logger.Uint("user_id", userID), logger.Err(err))
		return nil, apperrors.Wrap(apperrors.CodeDatabaseError, err)
	}

	return profile.ToSafeProfile(), nil
}

// CreateProfile 创建用户档案
func (s *UserProfileService) CreateProfile(userID uint, req *CreateProfileRequest) (*models.SafeUserProfile, error) {
	// 检查用户是否存在
	var user models.User
	if err := s.db.First(&user, userID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperrors.New(apperrors.CodeResourceNotFound, "用户不存在")
		}
		logger.Error("查询用户失败", logger.Uint("user_id", userID), logger.Err(err))
		return nil, apperrors.Wrap(apperrors.CodeDatabaseError, err)
	}

	// 检查是否已存在档案
	var existingProfile models.UserProfile
	if err := s.db.Where("user_id = ?", userID).First(&existingProfile).Error; err == nil {
		return nil, apperrors.New(apperrors.CodeResourceAlreadyExists, "用户档案已存在")
	}

	// 检查昵称是否已存在
	if err := s.checkNicknameExists(req.Nickname, 0); err != nil {
		return nil, err
	}

	// 检查邮箱是否已存在
	if err := s.checkEmailExists(req.Email, 0); err != nil {
		return nil, err
	}

	// 创建用户档案
	profile := &models.UserProfile{
		UserID:   userID,
		Nickname: strings.TrimSpace(req.Nickname),
		Phone:    strings.TrimSpace(req.Phone),
		Email:    strings.TrimSpace(req.Email),
	}

	if err := s.db.Create(profile).Error; err != nil {
		logger.Error("创建用户档案失败", logger.Uint("user_id", userID), logger.Err(err))
		return nil, apperrors.Wrap(apperrors.CodeDatabaseError, err)
	}

	logger.Info("用户档案已创建", logger.Uint("user_id", userID), logger.String("nickname", req.Nickname))

	return profile.ToSafeProfile(), nil
}

// UpdateProfile 更新用户档案
func (s *UserProfileService) UpdateProfile(userID uint, req *UpdateProfileRequest) (*models.SafeUserProfile, error) {
	var profile models.UserProfile
	if err := s.db.Where("user_id = ?", userID).First(&profile).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperrors.New(apperrors.CodeResourceNotFound, "用户档案不存在")
		}
		logger.Error("查询用户档案失败", logger.Uint("user_id", userID), logger.Err(err))
		return nil, apperrors.Wrap(apperrors.CodeDatabaseError, err)
	}

	// 更新字段
	updates := make(map[string]interface{})

	if req.Nickname != nil {
		nickname := strings.TrimSpace(*req.Nickname)
		if nickname == "" {
			return nil, apperrors.New(apperrors.CodeNicknameRequired, "")
		}
		if err := s.checkNicknameExists(nickname, profile.ID); err != nil {
			return nil, err
		}
		updates["nickname"] = nickname
	}

	if req.Phone != nil {
		phone := strings.TrimSpace(*req.Phone)
		updates["phone"] = phone
	}

	if req.Email != nil {
		email := strings.TrimSpace(*req.Email)
		if email == "" {
			return nil, apperrors.New(apperrors.CodeEmailRequired, "")
		}
		if err := s.checkEmailExists(email, profile.ID); err != nil {
			return nil, err
		}
		updates["email"] = email
	}

	if len(updates) == 0 {
		return profile.ToSafeProfile(), nil
	}

	// 更新数据库
	if err := s.db.Model(&profile).Updates(updates).Error; err != nil {
		logger.Error("更新用户档案失败", logger.Uint("user_id", userID), logger.Err(err))
		return nil, apperrors.Wrap(apperrors.CodeDatabaseError, err)
	}

	// 重新查询更新后的数据
	if err := s.db.Where("user_id = ?", userID).First(&profile).Error; err != nil {
		logger.Error("查询更新后的用户档案失败", logger.Uint("user_id", userID), logger.Err(err))
		return nil, apperrors.Wrap(apperrors.CodeDatabaseError, err)
	}

	logger.Info("用户档案已更新", logger.Uint("user_id", userID))

	return profile.ToSafeProfile(), nil
}

// CreateOrUpdateProfile 创建或更新用户档案
func (s *UserProfileService) CreateOrUpdateProfile(userID uint, req *CreateProfileRequest) (*models.SafeUserProfile, error) {
	// 尝试获取现有档案
	_, err := s.GetProfile(userID)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok && appErr.Type == "ResourceError" {
			// 档案不存在，创建新档案
			return s.CreateProfile(userID, req)
		}
		return nil, err
	}

	// 档案存在，转换为更新请求
	updateReq := &UpdateProfileRequest{
		Nickname: &req.Nickname,
		Phone:    &req.Phone,
		Email:    &req.Email,
	}

	return s.UpdateProfile(userID, updateReq)
}

// DeleteProfile 删除用户档案
func (s *UserProfileService) DeleteProfile(userID uint) error {
	var profile models.UserProfile
	if err := s.db.Where("user_id = ?", userID).First(&profile).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return apperrors.New(apperrors.CodeResourceNotFound, "用户档案不存在")
		}
		logger.Error("查询用户档案失败", logger.Uint("user_id", userID), logger.Err(err))
		return apperrors.Wrap(apperrors.CodeDatabaseError, err)
	}

	if err := s.db.Delete(&profile).Error; err != nil {
		logger.Error("删除用户档案失败", logger.Uint("user_id", userID), logger.Err(err))
		return apperrors.Wrap(apperrors.CodeDatabaseError, err)
	}

	logger.Info("用户档案已删除", logger.Uint("user_id", userID))

	return nil
}

// ListProfilesRequest 用户档案列表请求
type ListProfilesRequest struct {
	Page     int    `form:"page" binding:"omitempty,min=1"`
	PageSize int    `form:"page_size" binding:"omitempty,min=1,max=100"`
	Nickname string `form:"nickname" binding:"omitempty"`
	Email    string `form:"email" binding:"omitempty"`
}

// ListProfilesResponse 用户档案列表响应
type ListProfilesResponse struct {
	Profiles  []*models.SafeUserProfile `json:"profiles"`
	Total     int64                     `json:"total"`
	Page      int                       `json:"page"`
	PageSize  int                       `json:"page_size"`
	TotalPages int                      `json:"total_pages"`
}

// ListProfiles 获取用户档案列表（仅管理员）
func (s *UserProfileService) ListProfiles(req *ListProfilesRequest) (*ListProfilesResponse, error) {
	// 设置默认值
	if req.Page == 0 {
		req.Page = 1
	}
	if req.PageSize == 0 {
		req.PageSize = 20
	}

	// 构建查询
	query := s.db.Model(&models.UserProfile{}).
		Preload("User")

	// 应用过滤条件
	if req.Nickname != "" {
		query = query.Where("nickname LIKE ?", "%"+req.Nickname+"%")
	}
	if req.Email != "" {
		query = query.Where("email LIKE ?", "%"+req.Email+"%")
	}

	// 获取总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		logger.Error("获取用户档案总数失败", logger.Err(err))
		return nil, apperrors.Wrap(apperrors.CodeDatabaseError, err)
	}

	// 分页查询
	var profiles []models.UserProfile
	offset := (req.Page - 1) * req.PageSize
	if err := query.Offset(offset).Limit(req.PageSize).Order("created_at DESC").Find(&profiles).Error; err != nil {
		logger.Error("查询用户档案列表失败", logger.Err(err))
		return nil, apperrors.Wrap(apperrors.CodeDatabaseError, err)
	}

	// 转换为安全用户档案信息
	safeProfiles := make([]*models.SafeUserProfile, len(profiles))
	for i, profile := range profiles {
		safeProfiles[i] = profile.ToSafeProfile()
	}

	// 计算总页数
	totalPages := int(total) / req.PageSize
	if int(total)%req.PageSize > 0 {
		totalPages++
	}

	return &ListProfilesResponse{
		Profiles:   safeProfiles,
		Total:      total,
		Page:       req.Page,
		PageSize:   req.PageSize,
		TotalPages: totalPages,
	}, nil
}

// checkNicknameExists 检查昵称是否已存在
func (s *UserProfileService) checkNicknameExists(nickname string, excludeID uint) error {
	var count int64
	query := s.db.Model(&models.UserProfile{}).Where("nickname = ?", nickname)
	if excludeID > 0 {
		query = query.Where("id != ?", excludeID)
	}
	if err := query.Count(&count).Error; err != nil {
		return apperrors.Wrap(apperrors.CodeDatabaseError, err)
	}
	if count > 0 {
		return apperrors.New(apperrors.CodeNicknameExists, "")
	}
	return nil
}

// checkEmailExists 检查邮箱是否已存在
func (s *UserProfileService) checkEmailExists(email string, excludeID uint) error {
	var count int64
	query := s.db.Model(&models.UserProfile{}).Where("email = ?", email)
	if excludeID > 0 {
		query = query.Where("id != ?", excludeID)
	}
	if err := query.Count(&count).Error; err != nil {
		return apperrors.Wrap(apperrors.CodeDatabaseError, err)
	}
	if count > 0 {
		return apperrors.New(apperrors.CodeEmailExists, "")
	}
	return nil
}