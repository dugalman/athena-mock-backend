package config

import "os"

type Config struct {
	Port      string
	SecretKey string
}

func Load() *Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Valor por defecto
	}

	secret := os.Getenv("SECRET_KEY")
	if secret == "" {
		secret = "tu_clave_secreta" // Valor por defecto (no para producción)
	}

	return &Config{
		Port:      port,
		SecretKey: secret,
	}
}
