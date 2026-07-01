package model

import (
	"time"

	"github.com/google/uuid"
)

// StatusEnum 用户状态的枚举类型
type StatusEnum int

const (
	UserStatusNormal  StatusEnum = 1 // 正常用户
	UserStatusDisable StatusEnum = 2 // 禁用用户
	UserStatusPending StatusEnum = 3 // 待邮箱验证
)

// User 用户表的模型定义
type User struct {
	// Id 用户唯一标识，使用 uuid，主键，默认自动生成
	Id uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	// Username 用户名，唯一且不能为空
	Username string `json:"username" gorm:"uniqueIndex;not null"`
	// Password 密码，存储加密后的密码字符串
	Password string `json:"password"`
	// Avatar 用户头像，存储头像URL
	Avatar string `json:"avatar"`
	// Status 用户状态，使用 StatusEnum 枚举，默认待邮箱验证
	Status StatusEnum `json:"status" gorm:"type:smallint;default:3"`
	// LastLoginTime 上次登录时间
	LastLoginTime time.Time `json:"lastLoginTime"`
	// CurrentPlan 当前用户订阅的套餐，引用订阅类型，默认 free
	CurrentPlan SubscriptionPlan `json:"currentPlan" gorm:"type:varchar(20);default:'free'"`
	// Email 用户邮箱，唯一且不能为空
	Email string `json:"email" gorm:"type:varchar(100);uniqueIndex;not null"`
	// EmailVerified 邮箱是否已验证
	EmailVerified bool `json:"emailVerified" gorm:"type:boolean;default:false"`
}

// TableName 指定 User 结构体对应的数据库表名为 "users"
func (User) TableName() string {
	return "users"
}

// UserDTO 用户数据传输对象，用于对外展示用户信息（屏蔽敏感信息）
type UserDTO struct {
	// Id 用户唯一标识
	Id uuid.UUID `json:"id"`
	// Username 用户名
	Username string `json:"username"`
	// Email 邮箱
	Email string `json:"email"`
	// Avatar 头像
	Avatar string `json:"avatar"`
	// Status 用户状态
	Status StatusEnum `json:"status"`
	// Role 用户角色，可选
	Role string `json:"role,omitempty"`
}
