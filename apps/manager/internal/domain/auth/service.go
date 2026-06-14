package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"github.com/zalberix/cactus/apps/manager/config"
	db "github.com/zalberix/cactus/libs/storage/db"
)

var (
	ErrInvalidCredentials = errors.New("неверный email или пароль")
	ErrInvalidToken       = errors.New("неверный токен")
	ErrTokenExpired       = errors.New("токен истёк")
)

// Claims — содержимое JWT access токена.
type Claims struct {
	UserID int32  `json:"user_id"`
	Type   string `json:"type,omitempty"` // "access" или "refresh"
	jwt.RegisteredClaims
}

// Storage — интерфейс хранилища для auth сервиса.
type Storage interface {
	GetUserByEmail(ctx context.Context, email string) (db.User, error)
	GetUserByID(ctx context.Context, id int32) (db.User, error)
}

// Service реализует JWT аутентификацию и bcrypt хеширование паролей.
type Service struct {
	cfg        config.JWT
	store      Storage
	bcryptCost int
}

// NewService создаёт новый auth.Service.
func NewService(cfg config.JWT, store Storage) *Service {
	return NewServiceWithPasswordCost(cfg, store, bcrypt.DefaultCost)
}

func NewServiceWithPasswordCost(cfg config.JWT, store Storage, bcryptCost int) *Service {
	return &Service{cfg: cfg, store: store, bcryptCost: bcryptCost}
}

// IssueTokens выдаёт пару access + refresh JWT токенов для указанного userID.
func (s *Service) IssueTokens(userID int32) (*LoginResponse, error) {
	now := time.Now()

	// Access token
	accessClaims := Claims{
		UserID: userID,
		Type:   "access",
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   fmt.Sprintf("%d", userID),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.cfg.AccessTTL)),
		},
	}
	accessToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims).SignedString([]byte(s.cfg.Secret))
	if err != nil {
		return nil, fmt.Errorf("sign access token: %w", err)
	}

	// Refresh token — тоже JWT, но с type="refresh" и более длинным TTL
	refreshClaims := Claims{
		UserID: userID,
		Type:   "refresh",
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   fmt.Sprintf("%d", userID),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.cfg.RefreshTTL)),
		},
	}
	refreshToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).SignedString([]byte(s.cfg.Secret))
	if err != nil {
		return nil, fmt.Errorf("sign refresh token: %w", err)
	}

	return &LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(s.cfg.AccessTTL.Seconds()),
	}, nil
}

// ParseToken разбирает и валидирует JWT access токен, возвращает Claims.
func (s *Service) ParseToken(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %w", ErrInvalidToken)
		}
		return []byte(s.cfg.Secret), nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		return nil, fmt.Errorf("%w: %w", ErrInvalidToken, err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

// HashPassword хеширует пароль с bcrypt DefaultCost.
func (s *Service) HashPassword(password string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), s.bcryptCost)
	if err != nil {
		return "", fmt.Errorf("hash password: %w", err)
	}
	return string(hashed), nil
}

// CheckPassword сравнивает bcrypt хеш с открытым паролем.
func (s *Service) CheckPassword(hashed, plain string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashed), []byte(plain))
}

// Login аутентифицирует пользователя по email + пароль, возвращает токены.
func (s *Service) Login(ctx context.Context, req LoginRequest) (*LoginResponse, error) {
	user, err := s.store.GetUserByEmail(ctx, req.Email)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	if err := s.CheckPassword(user.Password, req.Password); err != nil {
		return nil, ErrInvalidCredentials
	}

	return s.IssueTokens(user.ID)
}

// Refresh принимает refresh JWT токен и выдаёт новую пару access + refresh токенов.
func (s *Service) Refresh(_ context.Context, refreshToken string) (*LoginResponse, error) {
	token, err := jwt.ParseWithClaims(refreshToken, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %w", ErrInvalidToken)
		}
		return []byte(s.cfg.Secret), nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		return nil, fmt.Errorf("%w: %w", ErrInvalidToken, err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	if claims.Type != "refresh" {
		return nil, fmt.Errorf("%w: не refresh токен", ErrInvalidToken)
	}

	return s.IssueTokens(claims.UserID)
}
