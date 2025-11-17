package models

import (
	"database/sql/driver"
	"encoding/json"
)

// Menu 菜单模型
type Menu struct {
	BaseModel
	UUID      string   `gorm:"type:char(36);uniqueIndex;not null" json:"uuid"`
	Path      string   `gorm:"type:varchar(128);not null" json:"path"`
	Name      string   `gorm:"type:varchar(64);not null" json:"name"`
	Icon      string   `gorm:"type:varchar(64);not null" json:"icon"`
	Title     string   `gorm:"type:varchar(64);not null" json:"title"`
	Roles     RoleList `gorm:"type:json;not null" json:"roles"`
	SortOrder int      `gorm:"type:int;not null;default:0" json:"sort_order"`
	IsVisible bool     `gorm:"type:boolean;not null;default:true" json:"is_visible"`
}

// TableName 指定表名
func (Menu) TableName() string {
	return "menus"
}

// RoleList 角色列表类型（用于JSON字段）
type RoleList []string

// Scan 实现sql.Scanner接口
func (r *RoleList) Scan(value interface{}) error {
	if value == nil {
		*r = []string{}
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}

	return json.Unmarshal(bytes, r)
}

// Value 实现driver.Valuer接口
func (r RoleList) Value() (driver.Value, error) {
	if r == nil {
		return json.Marshal([]string{})
	}
	return json.Marshal(r)
}

// MenuItem 菜单项（用于返回给前端）
type MenuItem struct {
	UUID      string   `json:"uuid"`
	Path      string   `json:"path"`
	Name      string   `json:"name"`
	Icon      string   `json:"icon"`
	Title     string   `json:"title"`
	Roles     []string `json:"roles"`
	SortOrder int      `json:"sort_order"`
}

// ToMenuItem 转换为菜单项
func (m *Menu) ToMenuItem() *MenuItem {
	return &MenuItem{
		UUID:      m.UUID,
		Path:      m.Path,
		Name:      m.Name,
		Icon:      m.Icon,
		Title:     m.Title,
		Roles:     m.Roles,
		SortOrder: m.SortOrder,
	}
}

// HasRole 检查菜单是否允许指定角色访问
func (m *Menu) HasRole(role string) bool {
	// 空数组表示所有角色都可访问
	if len(m.Roles) == 0 {
		return true
	}

	// 检查角色是否在允许列表中
	for _, r := range m.Roles {
		if r == role {
			return true
		}
	}

	return false
}
