// Handler 是用户管理的控制器层（Controller），它包裹了 Service（服务层），用于处理来自 HTTP 的入参，将请求参数解析成结构体，然后调用底层的服务进行业务处理。
package auths

import (
	"github.com/gin-gonic/gin"
	"github.com/mszlu521/thunder/errs"
	"github.com/mszlu521/thunder/req"
	"github.com/mszlu521/thunder/res"
	"github.com/mszlu521/thunder/tools/jwt"
)

// Handler 结构体，持有用户相关的 Service 实例
type Handler struct {
	service *Service
}

// NewHandler 创建一个新的 Handler，并初始化其 service 字段
func NewHandler() *Handler {
	return &Handler{service: NewService()}
}

// Register 处理用户注册的 HTTP 接口
// 1. 从请求中解析出注册参数（JSON 格式，绑定到 RegisterReq 结构体）。
// 2. 调用 service 层进行注册的业务处理。
// 3. 将结果返回给前端（注册成功/失败）。
func (h *Handler) Register(c *gin.Context) {
	var reqData RegisterReq
	if err := req.JsonParam(c, &reqData); err != nil {
		return
	}
	resp, err := h.service.register(reqData)
	if err != nil {
		res.Error(c, err)
		return
	}
	res.Success(c, resp)
}

// VerifyEmail 处理邮箱验证的 HTTP 接口
// 1. 从 URL 查询参数中解析出邮箱验证参数（token）。
// 2. 调用 service 检查 token 并激活用户。
// 3. 验证通过后重定向到前端登录页。
func (h *Handler) VerifyEmail(c *gin.Context) {
	var reqData VerifyEmailReq
	if err := req.QueryParam(c, &reqData); err != nil {
		return
	}
	_, err := h.service.verifyEmail(reqData.Token)
	if err != nil {
		res.Error(c, err)
		return
	}
	c.Redirect(302, "http://localhost:5173/login")
}

// Login 处理用户登录的 HTTP 接口
// 1. 从请求中解析出登录参数（JSON 格式，绑定到 LoginReq 结构体）。
// 2. 调用 service 层进行登录的业务处理。
// 3. 将结果返回给前端（登录成功/失败）。
func (h *Handler) Login(c *gin.Context) {
	var reqData LoginReq
	if err := req.JsonParam(c, &reqData); err != nil {
		return
	}
	resp, err := h.service.login(reqData)
	if err != nil {
		res.Error(c, err)
		return
	}
	res.Success(c, resp)
}

// RefreshToken 处理刷新JWT令牌的 HTTP 接口，逻辑如下：
// 1. 从请求体中解析出刷新令牌参数（RefreshTokenReq）。
// 2. 校验刷新令牌的有效性（jwt.ParseToken），如无效或过期，返回错误。
// 3. 查找 claims.Username 对应的用户，若查无此人，返回用户不存在错误。
// 4. 使用 service.token 为该用户重新生成访问令牌与刷新令牌。
// 5. 返回新的令牌给前端。
func (h *Handler) RefreshToken(c *gin.Context) {
	// 1. 参数解析
	var reqData RefreshTokenReq
	if err := req.JsonParam(c, &reqData); err != nil {
		return
	}
	// 2. 校验刷新token
	claims, err := jwt.ParseToken(reqData.RefreshToken)
	if err != nil {
		res.Error(c, errs.NewError(401, "令牌已过期"))
		return
	}
	// 3. 查找用户
	user, err := h.service.repo.findByUserName(c.Request.Context(), claims.Username)
	if err != nil {
		res.Error(c, errs.NewError(500, "数据库异常"))
		return
	}
	if user == nil {
		res.Error(c, errs.NewError(401, "用户不存在"))
		return
	}
	// 4. 重新生成 token
	resp, err := h.service.token(user)
	if err != nil {
		res.Error(c, errs.NewError(500, "令牌生成失败"))
		return
	}
	// 5. 返回新token
	res.Success(c, resp)
}

func (h *Handler) ForgotPassword(c *gin.Context) {
	var reqData ForgotPasswordReq
	if err := req.JsonParam(c, &reqData); err != nil {
		return
	}
	if err := h.service.forgotPassword(reqData); err != nil {
		res.Error(c, err)
		return
	}
	res.Success(c, map[string]interface{}{"message": "验证码已发送到您的邮箱"})
}

func (h *Handler) VerifyCode(c *gin.Context) {
	var reqData VerifyCodeReq
	if err := req.JsonParam(c, &reqData); err != nil {
		return
	}
	token, err := h.service.verifyCode(reqData)
	if err != nil {
		res.Error(c, err)
		return
	}
	res.Success(c, map[string]interface{}{
		"message": "验证码验证成功",
		"token":   token,
	})
}

func (h *Handler) ResetPassword(c *gin.Context) {
	var reqData ResetPasswordReq
	if err := req.JsonParam(c, &reqData); err != nil {
		return
	}
	if err := h.service.resetPassword(reqData); err != nil {
		res.Error(c, err)
		return
	}
	res.Success(c, map[string]interface{}{"message": "密码重置成功"})
}
