package auth

import (
	"testing"

	"athena.mock/backend/internal/model"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

func TestCreateToken(t *testing.T) {
	// 1. Arrange
	user := model.User{
		ID:       "user-123",
		Nickname: "Testy",
		Profiles: []string{"socio"},
		UserType: "partner",
		ViewAPK:  map[string]string{"menu": "testMenu"},
	}
	secret := "mi-clave-secreta-de-prueba"

	// 2. Act
	tokenString, err := CreateToken(user, secret)

	// 3. Assert
	assert.NoError(t, err, "La creación del token no debería fallar")
	assert.NotEmpty(t, tokenString, "El token string no debería estar vacío")

	// Opcional: Parsear el token para verificar su contenido
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})

	assert.NoError(t, err, "El token generado debería ser parseable")
	assert.True(t, token.Valid, "El token generado debería ser válido")
	assert.Equal(t, user.ID, claims.UserID, "El UserID en el claim debe coincidir")
	assert.Equal(t, user.Nickname, claims.Nickname, "El Nickname en el claim debe coincidir")
	assert.Equal(t, user.Profiles, claims.Roles, "Los Roles en el claim deben coincidir")
}
