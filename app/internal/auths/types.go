package auths

import "model"

// 本文件定义了用户认证和管理相关的请求与返回结构体，主要用于用户注册、登录、邮箱验证、密码重置等接口的请求(A)和响应参数(B)。

// RegisterReq 用户注册的请求参数
type RegisterReq struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
}

// RegisterResp 用户注册的响应数据
type RegisterResp struct {
	Message string `json:"message"`
}

// VerifyEmailReq 邮箱验证请求参数（如点击邮箱链接验证）
type VerifyEmailReq struct {
	Token string `form:"token" binding:"required"`
}

// LoginReq 用户登录的请求参数
type LoginReq struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResp 用户登录的响应数据
type LoginResp struct {
	Expire        int64         `json:"expire"`        // 访问 token 过期时间
	RefreshExpire int64         `json:"refreshExpire"` // refresh token 过期时间
	Token         string        `json:"token"`         // 访问 token
	RefreshToken  string        `json:"refreshToken"`  // 刷新 token
	UserInfo      model.UserDTO `json:"userInfo"`      // 用户简要信息
	Message       string        `json:"message,omitempty"`
}

// RefreshTokenReq 刷新 Token 的请求参数
type RefreshTokenReq struct {
	RefreshToken string `json:"refreshToken" binding:"required"`
}

// ForgotPasswordReq 忘记密码时发送邮件请求参数
type ForgotPasswordReq struct {
	Email string `json:"email" binding:"required,email"`
}

// VerifyCodeReq 验证邮箱验证码请求参数
type VerifyCodeReq struct {
	Email string `json:"email" binding:"required,email"`
	Code  string `json:"code" binding:"required,len=6"`
}

// ResetPasswordReq 重设密码请求参数
type ResetPasswordReq struct {
	Email       string `json:"email" binding:"required,email"`
	Token       string `json:"token" binding:"required"`
	NewPassword string `json:"newPassword" binding:"required,min=6"`
}
