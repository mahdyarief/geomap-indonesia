package service

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/mahdyarief/geomap-indonesia/internal/config"
)

// ErrInvalidCredentials is returned when API keys do not match the config.
var ErrInvalidCredentials = errors.New("invalid credentials")

// TokenResponse is returned by POST /auth.
type TokenResponse struct {
	Token     string `json:"token"`
	TokenType string `json:"token_type"`
	ExpiresIn int    `json:"expires_in"`
}

// AuthService issues JWT tokens to authenticated clients.
type AuthService struct {
	cfg *config.Config
}

func NewAuthService(cfg *config.Config) *AuthService {
	return &AuthService{cfg: cfg}
}

// GenerateToken validates the client API keys and returns a signed JWT.
func (s *AuthService) GenerateToken(publicKey, privateKey string) (*TokenResponse, error) {
	if publicKey != s.cfg.APIPublicKey || privateKey != s.cfg.APIPrivateKey {
		return nil, ErrInvalidCredentials
	}

	now := time.Now()
	expiresAt := now.Add(time.Duration(s.cfg.JWTExpiresHours) * time.Hour)

	claims := jwt.MapClaims{
		"sub": publicKey,
		"iss": "geomap-indonesia",
		"iat": now.Unix(),
		"exp": expiresAt.Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(s.cfg.JWTSecret))
	if err != nil {
		return nil, err
	}

	return &TokenResponse{
		Token:     signed,
		TokenType: "Bearer",
		ExpiresIn: int(expiresAt.Sub(now).Seconds()),
	}, nil
}
