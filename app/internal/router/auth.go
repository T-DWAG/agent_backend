package router

import (
	"app/internal/auths"

	"github.com/gin-gonic/gin"
)

// AuthRouter 负责路由注册
type AuthRouter struct{}

// Register 方法初始化了一个用户管理的处理器，然后把它的方法注册成接口地址
func (u *AuthRouter) Register(engine *gin.Engine) {
	// 初始化用户管理 Handler（处理器）
	h := auths.NewHandler()
	// 分组路由，所有接口均以 /api/v1/auth 开头
	g := engine.Group("/api/v1/auth")
	{
		// 注册接口地址，将 Handler 的 Register 方法转成 /register 的 POST 接口
		g.POST("/register", h.Register)
		// 注册邮箱验证接口地址，将 Handler 的 VerifyEmail 方法转成 /verify-email 的 GET 接口
		g.GET("/verify-email", h.VerifyEmail)
		// 登录接口地址，将 Handler 的 Login 方法转成 /login 的 POST 接口
		g.POST("/login", h.Login)
		// 刷新令牌接口地址，将 Handler 的 RefreshToken 方法转成 /refresh-token 的 POST 接口
		g.POST("/refresh-token", h.RefreshToken)
		g.POST("/forgot-password", h.ForgotPassword)
		g.POST("/verify-code", h.VerifyCode)
		g.POST("/reset-password", h.ResetPassword)
	}
}
