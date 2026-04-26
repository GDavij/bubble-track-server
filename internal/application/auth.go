package application

import (
	"context"
	"fmt"
	"time"

	"github.com/bubbletrack/server/internal/domain"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthUseCase struct {
	userRepo   domain.UserRepository
	tokenRepo  domain.RefreshTokenRepository
	jwtSecret  []byte
	tokenTTL   time.Duration
	refreshTTL time.Duration
}

func NewAuthUseCase(
	userRepo domain.UserRepository,
	tokenRepo domain.RefreshTokenRepository,
	jwtSecret string,
) *AuthUseCase {
	return &AuthUseCase{
		userRepo:   userRepo,
		tokenRepo:  tokenRepo,
		jwtSecret:  []byte(jwtSecret),
		tokenTTL:   15 * time.Minute,
		refreshTTL: 7 * 24 * time.Hour,
	}
}

func (uc *AuthUseCase) Register(ctx context.Context, req *domain.RegisterRequest) (*domain.AuthResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	existing, _ := uc.userRepo.GetByEmail(ctx, req.Email)
	if existing != nil {
		return nil, &domain.ValidationError{Field: "email", Message: "email already registered"}
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	account := &domain.Account{
		Email:        req.Email,
		PasswordHash: string(hash),
		DisplayName:  req.DisplayName,
	}
	if err := uc.userRepo.Create(ctx, account); err != nil {
		return nil, fmt.Errorf("failed to create account: %w", err)
	}

	return uc.generateTokens(ctx, account)
}

func (uc *AuthUseCase) Login(ctx context.Context, req *domain.LoginRequest) (*domain.AuthResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	account, err := uc.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, &domain.ValidationError{Field: "credentials", Message: "invalid email or password"}
	}

	if err := bcrypt.CompareHashAndPassword([]byte(account.PasswordHash), []byte(req.Password)); err != nil {
		return nil, &domain.ValidationError{Field: "credentials", Message: "invalid email or password"}
	}

	return uc.generateTokens(ctx, account)
}

func (uc *AuthUseCase) RefreshToken(ctx context.Context, req *domain.RefreshRequest) (*domain.AuthResponse, error) {
	if req.RefreshToken == "" {
		return nil, &domain.ValidationError{Field: "refresh_token", Message: "refresh token is required"}
	}

	stored, err := uc.tokenRepo.GetByToken(ctx, req.RefreshToken)
	if err != nil {
		return nil, &domain.ValidationError{Field: "refresh_token", Message: "invalid refresh token"}
	}

	if stored.Revoked {
		return nil, &domain.ValidationError{Field: "refresh_token", Message: "token has been revoked"}
	}

	if time.Now().After(stored.ExpiresAt) {
		return nil, &domain.ValidationError{Field: "refresh_token", Message: "token has expired"}
	}

	_ = uc.tokenRepo.Revoke(ctx, stored.ID)

	account, err := uc.userRepo.GetByID(ctx, stored.UserID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	return uc.generateTokens(ctx, account)
}

func (uc *AuthUseCase) Logout(ctx context.Context, refreshToken string) error {
	if refreshToken == "" {
		return nil
	}
	stored, err := uc.tokenRepo.GetByToken(ctx, refreshToken)
	if err != nil {
		return nil
	}
	return uc.tokenRepo.Revoke(ctx, stored.ID)
}

func (uc *AuthUseCase) GetCurrentUser(ctx context.Context, userID string) (*domain.Account, error) {
	return uc.userRepo.GetByID(ctx, userID)
}

func (uc *AuthUseCase) ValidateAccessToken(tokenStr string) (string, error) {
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return uc.jwtSecret, nil
	})
	if err != nil {
		return "", err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return "", fmt.Errorf("invalid token")
	}

	userID, ok := claims["sub"].(string)
	if !ok {
		return "", fmt.Errorf("invalid token claims")
	}

	return userID, nil
}

func (uc *AuthUseCase) generateTokens(ctx context.Context, account *domain.Account) (*domain.AuthResponse, error) {
	now := time.Now().UTC()
	expiresAt := now.Add(uc.tokenTTL)

	claims := jwt.MapClaims{
		"sub":   account.ID,
		"email": account.Email,
		"name":  account.DisplayName,
		"iat":   now.Unix(),
		"exp":   expiresAt.Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	accessToken, err := token.SignedString(uc.jwtSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to sign token: %w", err)
	}

	refreshTokenStr := fmt.Sprintf("rt_%d_%s", now.UnixNano(), account.ID[:8])
	refreshToken := &domain.RefreshToken{
		UserID:    account.ID,
		Token:     refreshTokenStr,
		ExpiresAt: now.Add(uc.refreshTTL),
	}
	if err := uc.tokenRepo.Store(ctx, refreshToken); err != nil {
		return nil, fmt.Errorf("failed to store refresh token: %w", err)
	}

	return &domain.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshTokenStr,
		TokenType:    "Bearer",
		ExpiresIn:    int64(uc.tokenTTL.Seconds()),
		User:         account,
	}, nil
}
