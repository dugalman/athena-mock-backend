package api

import (
	"github.com/gin-gonic/gin"
)

// setupRoutes registra todas las rutas de la aplicación.
func (s *Server) setupRoutes(router *gin.Engine) {
	// Rutas de Autenticación
	router.POST("/login", s.LoginHandler()) // Nota: ahora es s.LoginHandler
	router.POST("/logout", s.LogoutHandler())

	// Grupo de rutas para EGM
	egmRoutes := router.Group("/egms")
	{
		egmRoutes.POST("/:id/credit", s.addCreditToEGMHandler())
		egmRoutes.DELETE("/:id/credit", s.removeAllCreditFromEGMHandler())
		egmRoutes.POST("/:id/bind", s.bindEGMHandler())
		egmRoutes.POST("/:id/unbind", s.unbindEGMHandler())
	}

	// Grupo de rutas para Socio
	socioRoutes := router.Group("/socios")
	{
		socioRoutes.GET("/:id/balance", s.getBalanceHandler())
		socioRoutes.POST("/:id/balance/increment", s.incrementBalanceHandler())
		socioRoutes.POST("/:id/balance/decrement", s.decrementBalanceHandler())

		socioRoutes.GET("/:id/puntaje", s.getPuntajeHandler())
		socioRoutes.POST("/:id/puntaje", s.addPuntajeHandler())
	}
}
