package api

import (
	"athena.mock/backend/internal/auth"
	"athena.mock/backend/internal/config"
	"athena.mock/backend/internal/model"
	"github.com/gin-gonic/gin"
	"net/http"
	"sync"
)

// Usamos un mapa con un Mutex para manejar sesiones de forma segura para concurrencia.
var activeSessions = struct {
	sync.RWMutex
	m map[string]string
}{m: make(map[string]string)}

// LoginRequestBody define la estructura del JSON que esperamos en el body.
type LoginRequestBody struct {
	Data struct {
		UserID   string `json:"userId"`
		Password string `json:"password"`
	} `json:"data"`
}

func LoginHandler(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		var body LoginRequestBody
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		userID := body.Data.UserID
		password := body.Data.Password

		// Verificar si ya hay sesión activa
		activeSessions.RLock()
		_, loggedIn := activeSessions.m[userID]
		activeSessions.RUnlock()
		if loggedIn {
			c.JSON(http.StatusUnauthorized, gin.H{"error": 401, "message": "Usuario ya logueado en otro dispositivo"})
			return
		}

		// Buscar usuario y validar contraseña
		user, found := model.FindUserByID(userID)
		if !found || user.Password != password {
			c.JSON(http.StatusUnauthorized, gin.H{"error": 401, "message": "Usuario o contraseña incorrecta"})
			return
		}

		// Generar el token JWT
		tokenString, err := auth.CreateToken(user, cfg.SecretKey)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not create token"})
			return
		}

		// Almacenar sesión activa
		activeSessions.Lock()
		activeSessions.m[userID] = tokenString
		activeSessions.Unlock()

		// Respuesta exitosa (similar a responseSocio)
		c.JSON(http.StatusOK, gin.H{
			"requestType": "login",
			"error":       0,
			"message":     "Usuario: " + user.UserID + " logueado",
			"data": gin.H{
				"token":        tokenString,
				"userId":       user.UserID,
				"userProfiles": user.Profiles,
				// ... otros campos de la respuesta
			},
		})
	}
}
