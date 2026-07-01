package auths

import (
	"context"
	"model"

	"github.com/google/uuid"
	"github.com/mszlu521/thunder/database"
	"github.com/mszlu521/thunder/gorms"
	"gorm.io/gorm"
)

// Model 这是repo里面接口的实例，实现了Repository接口，负责用户认证相关的数据操作
type Model struct {
	db *gorm.DB
}

// NewModel 返回 Model 的实例，初始化数据库连接
func NewModel() *Model {
	return &Model{db: database.GetPostgresDB().GormDB}
}

// findByUserName 根据用户名查找用户信息
func (m *Model) findByUserName(ctx context.Context, username string) (*model.User, error) {
	var user model.User
	err := m.db.WithContext(ctx).Where("username = ?", username).First(&user).Error
	if gorms.IsRecordNotFoundError(err) {
		return nil, nil
	}
	return &user, err
}

// findByEmail 根据邮箱查找用户信息
func (m *Model) findByEmail(ctx context.Context, email string) (*model.User, error) {
	var user model.User
	err := m.db.WithContext(ctx).Where("email = ?", email).First(&user).Error
	if gorms.IsRecordNotFoundError(err) {
		return nil, nil
	}
	return &user, err
}

// findById 根据用户ID查找用户信息
func (m *Model) findById(ctx context.Context, id uuid.UUID) (*model.User, error) {
	var user model.User
	err := m.db.WithContext(ctx).Where("id = ?", id).First(&user).Error
	if gorms.IsRecordNotFoundError(err) {
		return nil, nil
	}
	return &user, err
}

// findByUserNameOrEmail 根据用户名或邮箱查找用户信息（用户名或邮箱匹配任一均可）
func (m *Model) findByUserNameOrEmail(ctx context.Context, username string) (*model.User, error) {
	var user model.User
	err := m.db.WithContext(ctx).Where("username = ? OR email = ?", username, username).First(&user).Error
	if gorms.IsRecordNotFoundError(err) {
		return nil, nil
	}
	return &user, err
}

// saveUser 保存新用户信息（支持事务）
func (m *Model) saveUser(ctx context.Context, tx *gorm.DB, u *model.User) error {
	if tx == nil {
		tx = m.db
	}
	return tx.WithContext(ctx).Create(u).Error
}

// updateUser 更新已有用户信息（支持事务）
func (m *Model) updateUser(ctx context.Context, tx *gorm.DB, u *model.User) error {
	if tx == nil {
		tx = m.db
	}
	return tx.WithContext(ctx).Save(u).Error
}

// transaction 启用事务执行传入的函数 f
func (m *Model) transaction(f func(tx *gorm.DB) error) error {
	return m.db.Transaction(f)
}
