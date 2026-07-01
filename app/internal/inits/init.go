package inits

import (
	"app/internal/router"
	"model"

	"github.com/mszlu521/thunder/config"
	"github.com/mszlu521/thunder/database"
	"github.com/mszlu521/thunder/server"
	"github.com/mszlu521/thunder/tools/jwt"
)

// Init 初始化服务组件，包括数据库、JWT、路由等
func Init(s *server.Server, conf *config.Config) {
	// 初始化并连接 PostgreSQL 数据库
	database.InitPostgres(conf.DB.Postgres)

	// 初始化并连接 Redis
	database.InitRedis(conf.DB.Redis)

	// 使用 GORM 自动迁移，生成 users 表，无需手写 SQL
	database.GetPostgresDB().GormDB.AutoMigrate(&model.User{})

	// 初始化 JWT 相关配置
	jwt.Init(conf.Jwt.GetSecret())

	// 注册所有路由
	s.RegisterRouters(
		&router.Event{},
		&router.AuthRouter{},
		&router.SubscriptionRouter{},
	)
}
