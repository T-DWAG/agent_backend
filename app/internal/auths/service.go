package auths

import (
	"common/biz"
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"model"
	"net/smtp"
	"time"

	"github.com/google/uuid"
	"github.com/mszlu521/thunder/cache"
	"github.com/mszlu521/thunder/config"
	"github.com/mszlu521/thunder/errs"
	"github.com/mszlu521/thunder/logs"
	"github.com/mszlu521/thunder/tools/jwt"
	"github.com/mszlu521/thunder/tools/randoms"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// Service 是用户管理相关的服务层，实现了用户注册、发送邮箱验证、邮箱验证等功能。
// Service 只是服务的中间层，面向业务的逻辑；底层所有具体的数据存储操作，均通过 Repository（repo）的实例进行。
// repo 是通过 NewModel() 得到的，最终以数据库为后端，Service 是上层业务对外暴露的接口。
type Service struct {
	repo Repository // repo定义在 repository.go，实现了用户相关的数据操作，比如查找、保存、更新等
}

// NewService 创建 Service 实例
func NewService() *Service {
	return &Service{repo: NewModel()}
}

// register 注册新用户，依次完成账号查重、邮箱查重、密码加密、生成验证token、缓存token、创建用户、发送验证邮件
func (s *Service) register(req RegisterReq) (*RegisterResp, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// 1. 用户名查重
	u, err := s.repo.findByUserName(ctx, req.Username)
	if err != nil {
		logs.Errorf("find username err:%v", err)
		return nil, errs.DBError
	}
	if u != nil {
		return nil, biz.ErrUserNameExisted
	}

	// 2. 邮箱查重
	u, err = s.repo.findByEmail(ctx, req.Email)
	if err != nil {
		logs.Errorf("find email err:%v", err)
		return nil, errs.DBError
	}
	if u != nil {
		return nil, biz.ErrEmailExisted
	}

	// 3. 密码加密
	password, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, biz.ErrPasswordFormat
	}

	// 4. 生成邮箱验证token
	tokenBytes := make([]byte, 16)
	if _, err := rand.Read(tokenBytes); err != nil {
		return nil, errs.DBError
	}
	verifyToken := hex.EncodeToString(tokenBytes)
	userId := uuid.New()

	// 5. 验证token写入Redis缓存（24小时有效）
	redisCache := cache.NewRedisCache()
	tokenKey := fmt.Sprintf("verify_token:%s", verifyToken)
	logs.Infof("verify token key: %s, userId: %s", tokenKey, userId)
	if err := redisCache.Set(tokenKey, userId.String(), 24*60*60); err != nil {
		logs.Errorf("store verify token err:%v", err)
		return nil, errs.DBError
	}

	// 6. 构造新用户对象，初始状态为待验证
	u = &model.User{
		Id:            userId,
		Username:      req.Username,
		Password:      string(password),
		LastLoginTime: time.Now(),
		Status:        model.UserStatusPending,
		Avatar:        "default",
		CurrentPlan:   model.FreePlan,
		Email:         req.Email,
		EmailVerified: false,
	}

	// 7. 事务：保存用户（发邮件放在事务外，避免 SMTP 失败导致注册回滚）
	err = s.repo.transaction(func(tx *gorm.DB) error {
		return s.repo.saveUser(ctx, tx, u)
	})
	if err != nil {
		return nil, errs.DBError
	}
	if err := s.sendVerificationEmail(u.Email, u.Username, verifyToken); err != nil {
		logs.Errorf("send verification email err:%v", err)
	}

	return &RegisterResp{Message: "注册成功，请检查您的邮箱并点击验证链接完成注册"}, nil
}

// sendVerificationEmail 发送验证邮件，实际的邮件配置从配置中心获取
func (s *Service) sendVerificationEmail(email, username, token string) error {
	emailConfig := config.GetConfig().Email
	if emailConfig.Host == nil || emailConfig.Port == nil {
		logs.Warn("Email not configured, skipping verification email")
		return nil
	}

	subject := "请验证您的邮箱地址"
	verifyURL := fmt.Sprintf("%s/api/v1/auth/verify-email?token=%s", emailConfig.GetBaseURL(), token)
	body := fmt.Sprintf("尊敬的 %s，\n\n请点击链接验证邮箱：\n%s\n", username, verifyURL)

	auth := smtp.PlainAuth("", emailConfig.GetUsername(), emailConfig.GetPassword(), emailConfig.GetHost())
	msg := buildEmailMessage(emailConfig.GetFrom(), email, subject, body)
	addr := fmt.Sprintf("%s:%d", emailConfig.GetHost(), emailConfig.GetPort())
	return smtp.SendMail(addr, auth, emailConfig.GetFrom(), []string{email}, msg)
}

// buildEmailMessage 构造 RFC5322 邮件正文；smtp.SendMail 的 from 仅用于 SMTP MAIL FROM，不会写入邮件头。
func buildEmailMessage(from, to, subject, body string) []byte {
	return []byte(
		"From: " + from + "\r\n" +
			"To: " + to + "\r\n" +
			"Subject: " + subject + "\r\n" +
			"MIME-Version: 1.0\r\n" +
			"Content-Type: text/plain; charset=UTF-8\r\n" +
			"\r\n" + body + "\r\n",
	)
}

// verifyEmail 检查邮箱验证token并激活用户
func (s *Service) verifyEmail(token string) (any, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	redisCache := cache.NewRedisCache()
	tokenKey := fmt.Sprintf("verify_token:%s", token)
	userIdStr, err := redisCache.Get(tokenKey)
	if err != nil || userIdStr == "" {
		return nil, biz.ErrInvalidToken
	}
	// 立即删除token
	redisCache.Set(tokenKey, "", 1)

	userId, err := uuid.Parse(userIdStr)
	if err != nil {
		return nil, biz.ErrInvalidToken
	}

	// 查询用户
	u, err := s.repo.findById(ctx, userId)
	if err != nil {
		return nil, errs.DBError
	}
	if u == nil {
		return nil, biz.ErrUserNotFound
	}
	// 已验证就直接返回
	if u.EmailVerified {
		return nil, nil
	}

	// 用户状态更新到正常
	u.EmailVerified = true
	u.Status = model.UserStatusNormal
	err = s.repo.transaction(func(tx *gorm.DB) error {
		return s.repo.updateUser(ctx, tx, u)
	})
	if err != nil {
		return nil, errs.DBError
	}
	return nil, nil
}

// 登录
// login 实现用户登录，包含如下流程：
// 1. 超时设置
// 2. repo实例查找有没有对应用户名或邮箱的账号
// 3. 如果报错或者找不到账号，返回用户不存在错误
// 4. 如果邮箱没认证，返回未认证错误
// 5. 都通过后，校验密码，校验通过发放token
func (s *Service) login(loginReq LoginReq) (*LoginResp, error) {
	// 1. 设置超时时间，避免请求卡死
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// 2. 根据用户名/邮箱查找用户
	u, err := s.repo.findByUserNameOrEmail(ctx, loginReq.Username)

	// 3. 查库出错或没找到用户，返回用户不存在
	if err != nil || u == nil {
		return nil, biz.ErrUserNotFound
	}

	// 4. 用户存在但未邮箱认证，直接返回未认证错误
	if !u.EmailVerified {
		return nil, biz.ErrEmailNotVerified
	}

	// 5. 校验密码，不通过则返回密码错误
	if err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(loginReq.Password)); err != nil {
		return nil, biz.ErrPasswordFormat
	}

	// 通过验证，返回登录token
	return s.token(u)
}

// token 服务，负责为已登录/认证用户发放 JWT 访问令牌和刷新令牌。
// 1. 入参 user 结构体，包含用户信息
// 2. 从 config 获取 jwt 访问令牌和刷新令牌的过期时间
// 3. 生成访问 token 和刷新 token
// 4. 返回登录响应 LoginResp，包含 token、过期时间、用户基础信息等
func (s *Service) token(u *model.User) (*LoginResp, error) {
	// 1. 获取配置中的 JWT 过期时间和刷新过期时间
	expire := config.GetConfig().Jwt.GetExpire()         // 正常 token 的过期时间
	refreshExpire := config.GetConfig().Jwt.GetRefresh() // 刷新 token 的过期时间

	// 2. 生成 JWT 访问 token
	token, err := jwt.GenToken(u.Id.String(), u.Username, expire)
	if err != nil {
		return nil, biz.ErrTokenGen
	}

	// 3. 生成 JWT 刷新 token
	refreshToken, err := jwt.GenToken(u.Id.String(), u.Username, refreshExpire)
	if err != nil {
		return nil, biz.ErrTokenGen
	}

	// 4. 封装并返回登录响应结构体，包括 token、刷新 token、过期时间和用户信息
	return &LoginResp{
		Expire:        time.Now().Add(expire).UnixMilli(),        // token 到期时间
		RefreshExpire: time.Now().Add(refreshExpire).UnixMilli(), // 刷新 token 到期时间
		Token:         token,                                     // 访问 token
		RefreshToken:  refreshToken,                              // 刷新 token
		UserInfo: model.UserDTO{ // 基本用户信息
			Id:       u.Id,
			Status:   u.Status,
			Username: u.Username,
			Role:     "admin", // 目前写死为 admin，实际可根据业务调整
		},
	}, nil
}

func (s *Service) forgotPassword(req ForgotPasswordReq) error {
	u, err := s.repo.findByEmail(context.Background(), req.Email)
	if err != nil {
		logs.Errorf("find email err:%v", err)
		return errs.DBError
	}
	if u == nil {
		return nil
	}

	code, err := randoms.Gen6Code()
	if err != nil {
		logs.Errorf("gen code err:%v", err)
		return errs.DBError
	}

	redisCache := cache.NewRedisCache()
	codeKey := fmt.Sprintf("forgot_password_code:%s", req.Email)
	if err := redisCache.Set(codeKey, code, 5*60); err != nil {
		logs.Errorf("store forgot password code err:%v", err)
		return errs.DBError
	}

	if err := s.sendForgotPasswordEmail(u.Email, u.Username, code); err != nil {
		logs.Errorf("send forgot password email err:%v", err)
	}
	return nil
}

func (s *Service) sendForgotPasswordEmail(email, username, code string) error {
	emailConfig := config.GetConfig().Email
	if emailConfig.Host == nil || emailConfig.Port == nil {
		logs.Warn("Email not configured, skipping forgot password email")
		return nil
	}

	subject := "您的验证码"
	body := fmt.Sprintf("尊敬的 %s，\n\n您正在重置密码，验证码是：%s\n\n验证码5分钟内有效，如非本人操作请忽略。\n", username, code)
	auth := smtp.PlainAuth("", emailConfig.GetUsername(), emailConfig.GetPassword(), emailConfig.GetHost())
	msg := buildEmailMessage(emailConfig.GetFrom(), email, subject, body)
	addr := fmt.Sprintf("%s:%d", emailConfig.GetHost(), emailConfig.GetPort())
	return smtp.SendMail(addr, auth, emailConfig.GetFrom(), []string{email}, msg)
}

func (s *Service) verifyCode(req VerifyCodeReq) (string, error) {
	redisCache := cache.NewRedisCache()
	codeKey := fmt.Sprintf("forgot_password_code:%s", req.Email)
	storedCode, err := redisCache.Get(codeKey)
	if err != nil || storedCode == "" {
		return "", biz.ErrInvalidToken
	}
	if storedCode != req.Code {
		return "", biz.ErrInvalidToken
	}

	resetToken, err := s.generateResetToken(req.Email)
	if err != nil {
		return "", err
	}
	redisCache.Set(codeKey, "", 1)
	return resetToken, nil
}

func (s *Service) generateResetToken(email string) (string, error) {
	tokenBytes := make([]byte, 16)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", errs.DBError
	}
	resetToken := hex.EncodeToString(tokenBytes)

	redisCache := cache.NewRedisCache()
	tokenKey := fmt.Sprintf("reset_token:%s", resetToken)
	if err := redisCache.Set(tokenKey, email, 60*60); err != nil {
		logs.Errorf("store reset token err:%v", err)
		return "", errs.DBError
	}
	return resetToken, nil
}

func (s *Service) resetPassword(req ResetPasswordReq) error {
	email, err := s.validateResetToken(req.Token)
	if err != nil {
		return errs.NewError(400, "无效或已过期的验证码")
	}
	if email != req.Email {
		return errs.NewError(400, "邮箱不匹配")
	}

	user, err := s.repo.findByEmail(context.Background(), email)
	if err != nil || user == nil {
		return errs.NewError(500, "用户不存在")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return errs.NewError(500, "密码处理失败")
	}

	user.Password = string(hashedPassword)
	err = s.repo.transaction(func(tx *gorm.DB) error {
		return s.repo.updateUser(context.Background(), tx, user)
	})
	if err != nil {
		return errs.NewError(500, "更新密码失败")
	}
	return nil
}

func (s *Service) validateResetToken(token string) (string, error) {
	redisCache := cache.NewRedisCache()
	tokenKey := fmt.Sprintf("reset_token:%s", token)
	email, err := redisCache.Get(tokenKey)
	if err != nil || email == "" {
		return "", biz.ErrInvalidToken
	}
	redisCache.Set(tokenKey, "", 1)
	return email, nil
}
