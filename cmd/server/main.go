package main

import (
	"log"

	"athena.mock/backend/internal/api"
	"athena.mock/backend/internal/config"
)

func main() {
	// Cargar configuración desde variables de entorno o valores por defecto
	cfg := config.Load()

	// Configurar el enrutador
	router := api.SetupRouter(cfg)

	// Iniciar el servidor
	log.Printf("Servidor Go escuchando en http://localhost:%s\n", cfg.Port)
	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatalf("Error al iniciar el servidor: %v", err)
	}
}
