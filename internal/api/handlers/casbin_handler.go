package handlers

import (
	"github.com/casbin/casbin/v2"
	"github.com/cuihe500/vaulthub/pkg/logger"
	"github.com/cuihe500/vaulthub/pkg/response"
	"github.com/gin-gonic/gin"
)

// CasbinHandler Casbin权限管理处理器
// 用于管理Casbin权限策略的热更新等操作
type CasbinHandler struct {
	enforcer *casbin.Enforcer
}

// NewCasbinHandler 创建Casbin处理器实例
func NewCasbinHandler(enforcer *casbin.Enforcer) *CasbinHandler {
	return &CasbinHandler{
		enforcer: enforcer,
	}
}

// ReloadPolicy 重新加载Casbin权限策略
// @Summary 重新加载权限策略
// @Description 从数据库重新加载Casbin权限策略到内存（管理员权限）。用于在运行时修改casbin_rule表后使策略生效，无需重启服务
// @Tags 系统管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response{data=map[string]interface{}}
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/configs/casbin/reload [post]
func (h *CasbinHandler) ReloadPolicy(c *gin.Context) {
	// 需要casbin:reload权限
	// 权限验证在路由层通过Casbin中间件完成

	// 重新加载策略
	if err := h.enforcer.LoadPolicy(); err != nil {
		logger.Error("重新加载Casbin策略失败", logger.Err(err))
		response.InternalError(c, "重新加载权限策略失败")
		return
	}

	logger.Info("Casbin权限策略重新加载成功")

	response.Success(c, gin.H{
		"message": "权限策略重新加载成功",
	})
}
