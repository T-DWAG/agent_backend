package auths

import (
	"context"
	"model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Repository 定义用户认证相关的数据操作接口，包含用户的增查改等操作。
type Repository interface {
	// transaction 启用事务执行传入的函数 f。
	// 入参: f func(tx *gorm.DB) error 事务操作函数，提供当前事务的 DB 对象
	// 出参: error 事务执行结果
	transaction(f func(tx *gorm.DB) error) error

	// findByUserName 通过用户名查找用户信息。
	// 入参: ctx 上下文信息, username 用户名
	// 出参: *model.User 用户对象指针, error 错误信息
	findByUserName(ctx context.Context, username string) (*model.User, error)

	// findByEmail 通过邮箱查找用户信息。
	// 入参: ctx 上下文信息, email 用户邮箱
	// 出参: *model.User 用户对象指针, error 错误信息
	findByEmail(ctx context.Context, email string) (*model.User, error)

	// findById 通过用户ID查找用户信息。
	// 入参: ctx 上下文信息, id 用户唯一ID
	// 出参: *model.User 用户对象指针, error 错误信息
	findById(ctx context.Context, id uuid.UUID) (*model.User, error)

	// findByUserNameOrEmail 通过用户名或邮箱查找用户信息（用户名优先）。
	// 入参: ctx 上下文信息, username 用户名或邮箱
	// 出参: *model.User 用户对象指针, error 错误信息
	findByUserNameOrEmail(ctx context.Context, username string) (*model.User, error)

	// saveUser 保存新用户信息（在提供的事务 tx 中操作）。
	// 入参: ctx 上下文信息, tx 事务对象, u 用户对象
	// 出参: error 错误信息
	saveUser(ctx context.Context, tx *gorm.DB, u *model.User) error

	// updateUser 更新已有用户信息（在提供的事务 tx 中操作）。
	// 入参: ctx 上下文信息, tx 事务对象, u 用户对象
	// 出参: error 错误信息
	updateUser(ctx context.Context, tx *gorm.DB, u *model.User) error
}
