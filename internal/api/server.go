package api

import (
	"athena.mock/backend/internal/config"
	"athena.mock/backend/internal/service"
	"github.com/gin-gonic/gin"
)

// Server contiene las dependencias de la API, como el router y los servicios.
type Server struct {
	router       *gin.Engine
	cfg          *config.Config
	egmService   *service.EGMService
	socioService *service.SocioService
	authService  *service.AuthService // <-- NUEVO

}

// NewServer crea una nueva instancia del servidor y configura las rutas.
func NewServer(cfg *config.Config, egmService *service.EGMService, socioService *service.SocioService, operatorService *service.OperatorService) *Server {

	authService := service.NewAuthService(socioService, operatorService) // <-- NUEVO

	server := &Server{
		cfg:          cfg,
		egmService:   egmService,
		socioService: socioService,
		authService:  authService, // <-- NUEVO
	}

	router := gin.Default()
	// Aquí llamamos a una función para registrar las rutas
	server.setupRoutes(router)

	server.router = router
	return server
}

// Run inicia el servidor HTTP.
func (s *Server) Run(addr string) error {
	return s.router.Run(addr)
}
