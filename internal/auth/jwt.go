package auth

import (
	"time"

	"athena.mock/backend/internal/model"
	"github.com/golang-jwt/jwt/v5"
)

// Claims define la estructura de nuestro token JWT.
type Claims struct {
	UserID   string   `json:"userId"`
	Nickname string   `json:"nickname"`
	Roles    []string `json:"roles"`
	Type     string   `json:"type"`
	ViewAPK  any      `json:"viewAPK"`
	jwt.RegisteredClaims
}

func CreateToken(user model.Authenticatable, secretKey string) (string, error) {

	expirationTime := time.Now().Add(1 * time.Hour) // Token válido por 1 hora

	claims := &Claims{
		UserID:   user.GetID(),
		Nickname: user.GetNickname(),
		Roles:    user.GetProfiles(),
		Type:     user.GetUserType(),
		ViewAPK:  user.GetViewAPK(),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			Subject:   user.GetUserID(), // <-- AÑADIMOS EL SUBJECT
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secretKey))

	return tokenString, err
}
