// Package auth 提供认证业务逻辑.
package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/ashwinyue/next-show/internal/model"
	"github.com/ashwinyue/next-show/internal/store"
)

// Biz 认证业务接口.
type Biz interface {
	// 用户管理
	Register(ctx context.Context, req *RegisterRequest) (*model.User, error)
	Login(ctx context.Context, req *LoginRequest) (*LoginResponse, error)
	Logout(ctx context.Context, token string) error
	GetCurrentUser(ctx context.Context, token string) (*model.User, error)
	UpdateProfile(ctx context.Context, userID string, req *UpdateProfileRequest) (*model.User, error)
	ChangePassword(ctx context.Context, userID string, req *ChangePasswordRequest) error

	// 用户列表（管理员）
	ListUsers(ctx context.Context, tenantID string) ([]*model.User, error)
	GetUser(ctx context.Context, id string) (*model.User, error)
	DeleteUser(ctx context.Context, id string) error

	// Token 验证
	ValidateToken(ctx context.Context, token string) (*Claims, error)
}

// RegisterRequest 注册请求.
type RegisterRequest struct {
	TenantID    string `json:"tenant_id"`
	Username    string `json:"username" binding:"required"`
	Email       string `json:"email" binding:"required,email"`
	Password    string `json:"password" binding:"required,min=6"`
	DisplayName string `json:"display_name"`
}

// LoginRequest 登录请求.
type LoginRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse 登录响应.
type LoginResponse struct {
	User        *model.User `json:"user"`
	AccessToken string      `json:"access_token"`
	TokenType   string      `json:"token_type"`
	ExpiresIn   int64       `json:"expires_in"` // 秒
}

// UpdateProfileRequest 更新资料请求.
type UpdateProfileRequest struct {
	DisplayName *string `json:"display_name"`
	Avatar      *string `json:"avatar"`
}

// ChangePasswordRequest 修改密码请求.
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

// Claims JWT Claims.
type Claims struct {
	UserID   string         `json:"user_id"`
	TenantID string         `json:"tenant_id"`
	Username string         `json:"username"`
	Email    string         `json:"email"`
	Role     model.UserRole `json:"role"`
	jwt.RegisteredClaims
}

// Config 认证配置.
type Config struct {
	JWTSecret     string
	TokenExpiry   time.Duration
	RefreshExpiry time.Duration
}

// DefaultConfig 默认配置.
func DefaultConfig() *Config {
	return &Config{
		JWTSecret:     "next-show-secret-key-change-in-production",
		TokenExpiry:   24 * time.Hour,
		RefreshExpiry: 7 * 24 * time.Hour,
	}
}

type bizImpl struct {
	store  store.Store
	config *Config
}

// NewBiz 创建认证业务实例.
func NewBiz(s store.Store, config *Config) Biz {
	if config == nil {
		config = DefaultConfig()
	}
	return &bizImpl{store: s, config: config}
}

func (b *bizImpl) Register(ctx context.Context, req *RegisterRequest) (*model.User, error) {
	// 检查用户名是否已存在
	if _, err := b.store.Users().GetByUsername(ctx, req.Username); err == nil {
		return nil, fmt.Errorf("username already exists")
	}

	// 检查邮箱是否已存在
	if _, err := b.store.Users().GetByEmail(ctx, req.Email); err == nil {
		return nil, fmt.Errorf("email already exists")
	}

	user := &model.User{
		ID:          uuid.New().String(),
		TenantID:    req.TenantID,
		Username:    req.Username,
		Email:       req.Email,
		DisplayName: req.DisplayName,
		Role:        model.UserRoleUser,
		Status:      model.UserStatusActive,
	}

	if err := user.SetPassword(req.Password); err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	if err := b.store.Users().Create(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (b *bizImpl) Login(ctx context.Context, req *LoginRequest) (*LoginResponse, error) {
	user, err := b.store.Users().GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("invalid email or password")
	}

	if user.Status != model.UserStatusActive {
		return nil, fmt.Errorf("user account is not active")
	}

	if !user.CheckPassword(req.Password) {
		return nil, fmt.Errorf("invalid email or password")
	}

	// 生成 JWT
	token, err := b.generateToken(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	// 更新最后登录时间
	_ = b.store.Users().UpdateLastLogin(ctx, user.ID)

	// 存储 token（用于撤销）
	userToken := &model.UserToken{
		ID:        uuid.New().String(),
		UserID:    user.ID,
		Token:     token,
		Type:      "access",
		ExpiresAt: time.Now().Add(b.config.TokenExpiry),
	}
	_ = b.store.Users().CreateToken(ctx, userToken)

	return &LoginResponse{
		User:        user,
		AccessToken: token,
		TokenType:   "Bearer",
		ExpiresIn:   int64(b.config.TokenExpiry.Seconds()),
	}, nil
}

func (b *bizImpl) Logout(ctx context.Context, token string) error {
	return b.store.Users().RevokeToken(ctx, token)
}

func (b *bizImpl) GetCurrentUser(ctx context.Context, token string) (*model.User, error) {
	claims, err := b.ValidateToken(ctx, token)
	if err != nil {
		return nil, err
	}
	return b.store.Users().Get(ctx, claims.UserID)
}

func (b *bizImpl) UpdateProfile(ctx context.Context, userID string, req *UpdateProfileRequest) (*model.User, error) {
	user, err := b.store.Users().Get(ctx, userID)
	if err != nil {
		return nil, err
	}

	if req.DisplayName != nil {
		user.DisplayName = *req.DisplayName
	}
	if req.Avatar != nil {
		user.Avatar = *req.Avatar
	}

	if err := b.store.Users().Update(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (b *bizImpl) ChangePassword(ctx context.Context, userID string, req *ChangePasswordRequest) error {
	user, err := b.store.Users().Get(ctx, userID)
	if err != nil {
		return err
	}

	if !user.CheckPassword(req.OldPassword) {
		return fmt.Errorf("old password is incorrect")
	}

	if err := user.SetPassword(req.NewPassword); err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	return b.store.Users().Update(ctx, user)
}

func (b *bizImpl) ListUsers(ctx context.Context, tenantID string) ([]*model.User, error) {
	return b.store.Users().List(ctx, tenantID)
}

func (b *bizImpl) GetUser(ctx context.Context, id string) (*model.User, error) {
	return b.store.Users().Get(ctx, id)
}

func (b *bizImpl) DeleteUser(ctx context.Context, id string) error {
	return b.store.Users().Delete(ctx, id)
}

func (b *bizImpl) ValidateToken(ctx context.Context, tokenStr string) (*Claims, error) {
	// 检查 token 是否被撤销
	if _, err := b.store.Users().GetToken(ctx, tokenStr); err != nil {
		return nil, fmt.Errorf("token is invalid or revoked")
	}

	// 解析 JWT
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(b.config.JWTSecret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}

	return claims, nil
}

func (b *bizImpl) generateToken(user *model.User) (string, error) {
	claims := &Claims{
		UserID:   user.ID,
		TenantID: user.TenantID,
		Username: user.Username,
		Email:    user.Email,
		Role:     user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(b.config.TokenExpiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "next-show",
			Subject:   user.ID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(b.config.JWTSecret))
}
