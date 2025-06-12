package api

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"athena.mock/backend/internal/config"
	"athena.mock/backend/internal/service"
	"github.com/stretchr/testify/assert"
)

// setupTestServer es una función helper para inicializar todo lo necesario para un test.
func setupTestServer(t *testing.T) *Server {
	// Limpiar archivos de db para un estado fresco
	os.Remove("db/egms.json")
	os.Remove("db/socios.json")
	os.MkdirAll("db", 0755)

	cfg := config.Load()
	egmService, err := service.NewEGMService()
	assert.NoError(t, err)
	socioService, err := service.NewSocioService()
	assert.NoError(t, err)

	return NewServer(cfg, egmService, socioService)
}

func TestEGMAndSocioFlow(t *testing.T) {
	server := setupTestServer(t)

	// 1. Obtener balance inicial del socio 1
	req, _ := http.NewRequest("GET", "/socios/1/balance", nil)
	w := httptest.NewRecorder()
	server.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	// Podríamos verificar el valor exacto del balance aquí

	// 2. Asignar EGM 1004 al socio 1
	bindBody := `{"userId": 1}`
	req, _ = http.NewRequest("POST", "/egms/1004/bind", bytes.NewBufferString(bindBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	server.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// 3. Intentar asignar la misma EGM de nuevo (debería fallar)
	req, _ = http.NewRequest("POST", "/egms/1004/bind", bytes.NewBufferString(bindBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	server.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)

	// 4. Añadir 50 créditos a la EGM 1004
	creditBody := `{"amount": 50.0}`
	req, _ = http.NewRequest("POST", "/egms/1004/credit", bytes.NewBufferString(creditBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	server.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// 5. Retirar todos los créditos de la EGM
	req, _ = http.NewRequest("DELETE", "/egms/1004/credit", nil)
	w = httptest.NewRecorder()
	server.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	// Podríamos verificar que el amount_removed es 50

	// 6. Liberar la EGM
	req, _ = http.NewRequest("POST", "/egms/1004/unbind", nil)
	w = httptest.NewRecorder()
	server.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}
