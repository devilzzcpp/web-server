package jwt

import (
	"fmt"
	"strings"
	"time"
	"web-server/internal/config"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID int `json:"user_id"`
	jwt.RegisteredClaims
}

func GenerateToken(userID int, cfg *config.Config) (string, error) {
	secret := cfg.JwtSecret
	exp := cfg.JwtExpires

	claims := Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * time.Duration(exp))),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))

}

func ParseToken(cfg *config.Config, headers map[string]string) (*Claims, error) {
	auth := headers["Authorization"]
	if !strings.HasPrefix(auth, "Bearer ") {
		return nil, fmt.Errorf("invalid authorization header")
	}
	tokenStr := strings.TrimPrefix(auth, "Bearer ")

	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return []byte(cfg.JwtSecret), nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}
	return claims, nil
}
