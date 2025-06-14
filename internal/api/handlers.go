package api

import (
	"errors"
	"net/http"
	"runtime"
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

		// ... (parseo del body sin cambios) ...
		userID := body.Data.UserID
		password := body.Data.Password

		// --- VERIFICACIÓN DE SESIÓN ACTIVA ---
		activeSessions.RLock()
		_, loggedIn := activeSessions.m[userID]
		activeSessions.RUnlock()
		if loggedIn {
			c.JSON(http.StatusUnauthorized, gin.H{"error": 401, "message": "Usuario ya logueado en otro dispositivo"})
			return
		}

		// Usamos el nuevo servicio de autenticación
		user, err := s.authService.AuthenticateUser(userID, password)
		if err != nil {
			// Los 3 casos que queremos probar terminan aquí.
			c.JSON(http.StatusUnauthorized, gin.H{"error": 401, "message": "Usuario o contraseña incorrecta"})
			return
		}

		// Generar el token JWT usando la interfaz
		tokenString, err := auth.CreateToken(user, s.cfg.SecretKey)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not create token"})
			return
		}

		// --- ALMACENAMIENTO DE SESIÓN ACTIVA ---
		activeSessions.Lock()
		activeSessions.m[userID] = tokenString
		activeSessions.Unlock()

		// Respuesta exitosa (similar a responseSocio)
		c.JSON(http.StatusOK, gin.H{
			"requestType": "login",
			"error":       0,
			"message":     "Usuario: " + user.GetUserID() + " logueado",
			"data": gin.H{
				"token":        tokenString,
				"userId":       user.GetUserID(),
				"userProfiles": user.GetProfiles(),
				// ... otros campos de la respuesta
			},
		})
	}
}

// LogoutHandler elimina una sesión activa.
func (s *Server) LogoutHandler() gin.HandlerFunc {
	return func(c *gin.Context) {

		// El userID ahora viene del middleware, que ya validó el token.
		// Esto es mucho más seguro que leerlo del body.
		sessionID, exists := c.Get("sessionID")

		if !exists {
			// Este caso no debería ocurrir si el middleware está bien configurado.
			c.JSON(http.StatusInternalServerError, gin.H{"error": "No se pudo obtener la ID de sesión del token"})
			return
		}

		userIDStr := sessionID.(string)

		// Eliminar la sesión
		activeSessions.Lock()
		delete(activeSessions.m, userIDStr)
		activeSessions.Unlock()

		c.JSON(http.StatusOK, gin.H{
			"requestType": "logout",
			"error":       0,
			"message":     "Sesión finalizada. Usuario: " + userIDStr,
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

func (s *Server) getPuntajeHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "ID de Socio inválido"})
			return
		}
		puntaje, err := s.socioService.GetPuntaje(id)
		if err != nil {
			if errors.Is(err, service.ErrSocioNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"socio_id": id, "puntaje": puntaje})
	}
}

type PuntajeRequest struct {
	Puntaje int `json:"puntaje"`
}

func (s *Server) addPuntajeHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "ID de Socio inválido"})
			return
		}
		var req PuntajeRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Se requiere 'puntaje' en el cuerpo"})
			return
		}

		if err := s.socioService.AddPuntaje(id, req.Puntaje); err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Puntaje añadido exitosamente"})
	}
}

// --- INFO HANDLER ---

func (s *Server) InfoHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Construimos la respuesta usando datos de la configuración y del sistema
		response := model.InfoResponse{
			Environment:  s.cfg.Environment,
			GoVersion:    runtime.Version(),                   // Obtiene la versión de Go (ej: go1.21.0)
			HostPlatform: runtime.GOOS + "/" + runtime.GOARCH, // ej: linux/amd64
			Port:         s.cfg.Port,
			Version:      s.cfg.AppVersion,
			Asistente:    false, // Valor estático como en el ejemplo
		}

		// En Go, es más idiomático devolver el objeto directamente.
		// Gin se encargará de envolverlo en un campo "data" si es necesario,
		// pero la práctica común es devolver el objeto tal cual.
		// Para replicar exactamente la salida, lo envolvemos.
		c.JSON(http.StatusOK, gin.H{
			"data": response,
		})
	}
}
