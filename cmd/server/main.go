package main

import (
	"log"
	"os"

	"athena.mock/backend/internal/api"
	"athena.mock/backend/internal/config"
	"athena.mock/backend/internal/service" // Importamos los servicios
)

func main() {
	// Asegurarse de que el directorio 'db' exista
	if err := os.MkdirAll("db", 0755); err != nil {
		log.Fatalf("No se pudo crear el directorio db: %v", err)
	}

	// Cargar configuración
	cfg := config.Load()

	// Inicializar servicios
	egmService, err := service.NewEGMService()
	if err != nil {
		log.Fatalf("Error al inicializar EGMService: %v", err)
	}
	socioService, err := service.NewSocioService()
	if err != nil {
		log.Fatalf("Error al inicializar SocioService: %v", err)
	}

	// Creamos la instancia del servidor con sus dependencias
	server := api.NewServer(cfg, egmService, socioService)

	log.Printf("Servidor Go escuchando en http://localhost:%s\n", cfg.Port)
	if err := server.Run(":" + cfg.Port); err != nil {
		log.Fatalf("Error al iniciar el servidor: %v", err)
	}
}
