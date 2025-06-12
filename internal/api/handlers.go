package api

import (
	"errors"
	"net/http"
	"strconv"
	"sync"

	"athena.mock/backend/internal/auth"
	"athena.mock/backend/internal/model"
	"athena.mock/backend/internal/service"
	"github.com/gin-gonic/gin"
)

// Usamos un mapa con un Mutex para manejar sesiones de forma segura para concurrencia.
var activeSessions = struct {
	sync.RWMutex
	m map[string]string
}{m: make(map[string]string)}

// --- AUTH HANDLERS ---

// Ahora son métodos del Server para que puedan acceder a la configuración (s.cfg)
func (s *Server) LoginHandler() gin.HandlerFunc {
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
		tokenString, err := auth.CreateToken(user, s.cfg.SecretKey)
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

// LogoutHandler elimina una sesión activa.
func (s *Server) LogoutHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		var body LogoutRequestBody
		// En una implementación real, el userID vendría del token JWT (c.GetString("userID"))
		// pero para replicar el test de Node.js, lo leemos del body.
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body, expecting userId"})
			return
		}

		userID := body.UserID

		// Eliminar la sesión
		activeSessions.Lock()
		delete(activeSessions.m, userID)
		activeSessions.Unlock()

		c.JSON(http.StatusOK, gin.H{
			"requestType": "logout",
			"error":       0,
			"message":     "Sesión finalizada. Usuario: " + userID,
		})
	}
}

// LoginRequestBody define la estructura del JSON que esperamos en el body.
type LoginRequestBody struct {
	Data struct {
		UserID   string `json:"userId"`
		Password string `json:"password"`
	} `json:"data"`
}

// LogoutRequestBody define el body para el logout
type LogoutRequestBody struct {
	UserID string `json:"userId"`
}

// ClearActiveSessions es una función helper para nuestras pruebas.
func ClearActiveSessions() {
	activeSessions.Lock()
	activeSessions.m = make(map[string]string)
	activeSessions.Unlock()
}

// --- EGM HANDLERS ---

type CreditRequest struct {
	Amount float64 `json:"amount"`
}

func (s *Server) addCreditToEGMHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "ID de EGM inválido"})
			return
		}

		var req CreditRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Cuerpo de la petición inválido"})
			return
		}

		if err := s.egmService.AddCredit(id, req.Amount); err != nil {
			// Manejar errores específicos del servicio
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Crédito añadido exitosamente"})
	}
}

func (s *Server) removeAllCreditFromEGMHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "ID de EGM inválido"})
			return
		}

		amount, err := s.egmService.RemoveAllCredit(id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Todo el crédito fue retirado", "amount_removed": amount})
	}
}

type BindRequest struct {
	UserID int `json:"userId"`
}

func (s *Server) bindEGMHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "ID de EGM inválido"})
			return
		}

		var req BindRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Se requiere userId en el cuerpo"})
			return
		}

		if err := s.egmService.BindEgmToUser(id, req.UserID); err != nil {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()}) // 409 Conflict es bueno para "ya ocupada"
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "EGM asignada al usuario"})
	}
}

func (s *Server) unbindEGMHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "ID de EGM inválido"})
			return
		}
		if err := s.egmService.UnbindEgmToUser(id); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "EGM liberada"})
	}
}

// --- SOCIO HANDLERS ---

func (s *Server) getBalanceHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "ID de Socio inválido"})
			return
		}
		balance, err := s.socioService.GetBalance(id)
		if err != nil {
			if errors.Is(err, service.ErrSocioNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"socio_id": id, "balance": balance})
	}
}

type BalanceRequest struct {
	Amount float64 `json:"amount"`
}

func (s *Server) incrementBalanceHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "ID de Socio inválido"})
			return
		}
		var req BalanceRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Se requiere 'amount' en el cuerpo"})
			return
		}

		if err := s.socioService.IncrementBalance(id, req.Amount); err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Balance incrementado"})
	}
}

func (s *Server) decrementBalanceHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "ID de Socio inválido"})
			return
		}
		var req BalanceRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Se requiere 'amount' en el cuerpo"})
			return
		}

		err = s.socioService.DecrementBalance(id, req.Amount)
		if err != nil {
			if errors.Is(err, service.ErrSocioNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
				return
			}
			if errors.Is(err, service.ErrInsufficientBalance) {
				c.JSON(http.StatusConflict, gin.H{"error": err.Error()}) // 409 Conflict es bueno para esto
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Balance decrementado"})
	}
}
