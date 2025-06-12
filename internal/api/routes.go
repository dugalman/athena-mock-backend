package api

import (
	"athena.mock/backend/internal/config"
	"github.com/gin-gonic/gin"
)

func SetupRouter(cfg *config.Config) *gin.Engine {
	// gin.SetMode(gin.ReleaseMode) // Descomentar para producción
	r := gin.Default()

	// Definir rutas
	r.POST("/login", LoginHandler(cfg))
	// r.POST("/logout", LogoutHandler(cfg)) // Añadiríamos esto después

	return r
}
