package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"athena.mock/backend/internal/config"
	"athena.mock/backend/internal/service"
	"github.com/stretchr/testify/assert"
)

// setupTestServer es nuestra función helper clave para inicializar todo.
func setupTestServer(t *testing.T) *Server {

	// --- USAREMOS EL SEEDER PARA PREPARAR EL ENTORNO ---
	// Primero, creamos una función seeder que no termine el programa con log.Fatalf
	// Por ahora, para simplificar, llamaremos al `seeder` directamente desde el Makefile antes de testear.

	// Limpiar archivos de db y sesiones para un estado de prueba fresco y aislado.
	os.Remove("db/egms.json")
	os.Remove("db/socios.json")
	os.MkdirAll("db", 0755)
	ClearActiveSessions()

	cfg := config.Load()

	// Inicializamos los servicios reales para una prueba de integración completa.
	egmService, err := service.NewEGMService()
	assert.NoError(t, err) // Afirmamos que no hay error al crear el servicio
	socioService, err := service.NewSocioService()
	assert.NoError(t, err)

	// Creamos una instancia del servidor usando el nuevo patrón
	server := NewServer(cfg, egmService, socioService)
	return server
}

// TestAuthFlow prueba el ciclo completo de login y logout.
func TestAuthFlow(t *testing.T) {
	server := setupTestServer(t)
	router := server.router // Usamos el router del servidor que creamos

	// --- Test de Login ---
	loginBody := `{"data": {"userId": "12345678", "password": "pass123"}}`
	req, _ := http.NewRequest("POST", "/login", bytes.NewBufferString(loginBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Aserciones del Login
	assert.Equal(t, http.StatusOK, w.Code, "El código de estado del login debería ser 200")

	var loginResponse map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &loginResponse)
	assert.NoError(t, err, "La respuesta del login debería ser un JSON válido")

	data, _ := loginResponse["data"].(map[string]interface{})
	token, tokenExists := data["token"].(string)
	assert.True(t, tokenExists, "La respuesta del login debe contener un token")
	assert.NotEmpty(t, token, "El token no puede estar vacío")

	// --- Test de Logout ---
	logoutBody := `{"userId": "12345678"}`
	req, _ = http.NewRequest("POST", "/logout", bytes.NewBufferString(logoutBody))
	req.Header.Set("Content-Type", "application/json")

	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Aserciones del Logout
	assert.Equal(t, http.StatusOK, w.Code, "El código de estado del logout debería ser 200")
	var logoutResponse map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &logoutResponse)
	assert.Equal(t, "logout", logoutResponse["requestType"], "El tipo de petición de logout debe ser 'logout'")
	assert.Contains(t, logoutResponse["message"], "Sesión finalizada", "El mensaje de logout debe indicar que la sesión finalizó")

	// --- Verificación Post-Logout ---
	req, _ = http.NewRequest("POST", "/login", bytes.NewBufferString(loginBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code, "Debería ser posible hacer login de nuevo después del logout")
}

// TestEGMAndSocioFlow prueba el flujo de interacción entre socios y EGMs.
func TestEGMAndSocioFlow(t *testing.T) {
	server := setupTestServer(t)
	router := server.router

	// 1. Obtener balance inicial del socio 1 (ID 1, DNI 12345678)
	req, _ := http.NewRequest("GET", "/socios/1/balance", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// 2. Asignar EGM 1004 al socio 1
	bindBody := `{"userId": 1}`
	req, _ = http.NewRequest("POST", "/egms/1004/bind", bytes.NewBufferString(bindBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// 3. Intentar asignar la misma EGM de nuevo (debería fallar)
	req, _ = http.NewRequest("POST", "/egms/1004/bind", bytes.NewBufferString(bindBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusConflict, w.Code)

	// 4. Añadir 50 créditos a la EGM 1004
	creditBody := `{"amount": 50.0}`
	req, _ = http.NewRequest("POST", "/egms/1004/credit", bytes.NewBufferString(creditBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// 5. Retirar todos los créditos de la EGM
	req, _ = http.NewRequest("DELETE", "/egms/1004/credit", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	var cashoutResponse map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &cashoutResponse)
	assert.InDelta(t, 50.0, cashoutResponse["amount_removed"], 0.001, "La cantidad retirada debe ser 50")

	// 6. Liberar la EGM
	req, _ = http.NewRequest("POST", "/egms/1004/unbind", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

// ...
func TestPuntajeFlow(t *testing.T) {
	server := setupTestServer(t)
	router := server.router

	// 1. Obtener puntaje inicial del socio 2 (ID: 2)
	req, _ := http.NewRequest("GET", "/socios/2/puntaje", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var puntajeResponse map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &puntajeResponse)
	// El seeder inicializa al socio 2 con 6666 puntos
	assert.EqualValues(t, 6666, puntajeResponse["puntaje"])

	// 2. Añadir 100 puntos al socio 2
	addPuntajeBody := `{"puntaje": 100}`
	req, _ = http.NewRequest("POST", "/socios/2/puntaje", bytes.NewBufferString(addPuntajeBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// 3. Verificar el nuevo puntaje
	req, _ = http.NewRequest("GET", "/socios/2/puntaje", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	json.Unmarshal(w.Body.Bytes(), &puntajeResponse)
	// 6666 (inicial) + 100 (añadido) = 6766
	assert.EqualValues(t, 6766, puntajeResponse["puntaje"])
}
