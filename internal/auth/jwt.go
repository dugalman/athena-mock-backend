package auth

import (
	"athena.mock/backend/internal/model"
	"github.com/golang-jwt/jwt/v5"
	"time"
)

// Claims define la estructura de nuestro token JWT.
type Claims struct {
	UserID   string `json:"userId"`
	Nickname string `json:"nickname"`
	Roles    []string `json:"roles"`
	Type     string `json:"type"`
	ViewAPK  any    `json:"viewAPK"`
	jwt.RegisteredClaims
}

func CreateToken(user model.User, secretKey string) (string, error) {
	expirationTime := time.Now().Add(1 * time.Hour) // Token válido por 1 hora

	claims := &Claims{
		UserID:   user.ID,
		Nickname: user.Nickname,
		Roles:    user.Profiles,
		Type:     user.UserType,
		ViewAPK:  user.ViewAPK,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secretKey))

	return tokenString, err
}
