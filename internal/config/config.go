package config

import "os"

type Config struct {
	Port        string
	SecretKey   string
	AppVersion  string // Versión de nuestra aplicación
	Environment string // development, production, etc.
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

	// Podemos pasar la versión a través de variables de entorno o flags de compilación
	appVersion := os.Getenv("APP_VERSION")
	if appVersion == "" {
		appVersion = "3.8.0" // Valor por defecto, como en tu package.json
	}

	environment := os.Getenv("GO_ENV")
	if environment == "" {
		environment = "development"
	}

	return &Config{
		Port:        port,
		SecretKey:   secret,
		AppVersion:  appVersion,
		Environment: environment,
	}
}
