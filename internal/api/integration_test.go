package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"athena.mock/backend/internal/config"
	"github.com/stretchr/testify/assert" // Una librería de aserciones muy popular
)

// Para instalar testify: go get github.com/stretchr/testify
func TestLoginLogoutIntegration(t *testing.T) {
	// 1. Setup
	// Limpiamos las sesiones antes de cada test para asegurar un estado limpio.
	ClearActiveSessions()

	cfg := config.Load()
	router := SetupRouter(cfg)

	// --- 2. Test de Login ---
	loginBody := `{"data": {"userId": "12345678", "password": "pass123"}}`
	req, _ := http.NewRequest("POST", "/login", bytes.NewBufferString(loginBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder() // Un "grabador" de respuestas HTTP
	router.ServeHTTP(w, req)    // Ejecutamos la petición contra nuestro router

	// Aserciones del Login
	assert.Equal(t, http.StatusOK, w.Code)

	var loginResponse map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &loginResponse)
	assert.NoError(t, err)

	// Extraemos el token para la siguiente petición
	data, _ := loginResponse["data"].(map[string]interface{})
	token, tokenExists := data["token"].(string)
	assert.True(t, tokenExists)
	assert.NotEmpty(t, token)

	// --- 3. Test de Logout ---
	logoutBody := `{"userId": "12345678"}`
	req, _ = http.NewRequest("POST", "/logout", bytes.NewBufferString(logoutBody))
	req.Header.Set("Content-Type", "application/json")
	// En una ruta protegida, añadiríamos el token aquí:
	// req.Header.Set("Authorization", "Bearer "+token)

	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Aserciones del Logout
	assert.Equal(t, http.StatusOK, w.Code)
	var logoutResponse map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &logoutResponse)
	assert.Equal(t, "logout", logoutResponse["requestType"])
	assert.Contains(t, logoutResponse["message"], "Sesión finalizada")

	// --- 4. Verificación Post-Logout ---
	// Intentamos hacer login de nuevo, debería funcionar porque la sesión se borró.
	req, _ = http.NewRequest("POST", "/login", bytes.NewBufferString(loginBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}
