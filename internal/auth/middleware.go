package auth

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// AuthMiddleware extrae y valida el token JWT.
func AuthMiddleware(secretKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": 401, "message": "Falta la cabecera de autorización"})
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": 401, "message": "La cabecera de autorización debe tener el formato 'Bearer {token}'"})
			return
		}
		tokenString := parts[1]

		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(secretKey), nil
		})

		// Aquí manejamos específicamente los errores de token (inválido, expirado, etc.)
		if err != nil {
			if err == jwt.ErrTokenExpired {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": 401, "message": "El token ha expirado"})
			} else {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": 401, "message": "Token inválido"})
			}
			return
		}

		if !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": 401, "message": "Token inválido"})
			return
		}

		// Almacenamos tanto el ID del usuario (de los claims) como su "username" que usamos para las sesiones.
		// Asumimos que el "Subject" del token contiene el username/dni. Vamos a añadirlo.
		c.Set("userID", claims.UserID)
		c.Set("sessionID", claims.Subject) // Usaremos el 'Subject' para el mapa de sesiones

		c.Next()
	}
}
