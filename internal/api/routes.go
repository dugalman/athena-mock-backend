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
	r.POST("/logout", LogoutHandler(cfg))

	// Ejemplo de una ruta protegida con el middleware
	// protected := r.Group("/api")
	// protected.Use(auth.AuthMiddleware(cfg.SecretKey))
	// {
	//    protected.GET("/me", func(c *gin.Context) {
	// 		userID, _ := c.Get("userID")
	// 		c.JSON(http.StatusOK, gin.H{"message": "Hello user " + userID.(string)})
	// 	})
	// }

	return r
}
