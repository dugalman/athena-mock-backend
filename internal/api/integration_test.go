package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"athena.mock/backend/internal/config"
	"athena.mock/backend/internal/model"
	"athena.mock/backend/internal/project"
	"athena.mock/backend/internal/service"
	"github.com/stretchr/testify/assert"
)

// setupTestServer es nuestra función helper clave para inicializar todo.
func setupTestServer(t *testing.T) *Server {

	// 1. Resetear los singletons para forzar la reinicialización.
	service.ResetSocioServiceForTests()
	service.ResetEGMServiceForTests()

	// 2. Limpiar el sistema de archivos
	dbDir := filepath.Join(project.ProjectRoot, "db")
	os.RemoveAll(dbDir)
	os.MkdirAll(dbDir, 0755)

	// 3. Sembrar la base de datos con un estado conocido para ESTA prueba.
	// Esto es mejor que depender de `make seed`.
	seedForTests()

	// 4. Limpiar sesiones en memoria
	ClearActiveSessions()

	// 5. Crear las instancias de servicio (ahora se reinicializarán desde los archivos sembrados)
	cfg := config.Load()
	egmService, err := service.GetEGMService()
	assert.NoError(t, err)
	partnerService, err := service.GetSocioService()
	assert.NoError(t, err)
	operatorService, err := service.GetOperatorService()
	assert.NoError(t, err)

	// Inicializamos los servicios reales para una prueba de integración completa.
	// Creamos una instancia del servidor usando el nuevo patrón
	return NewServer(cfg, egmService, partnerService, operatorService)
}

// seedForTests es una versión del seeder que no usa `log.Fatalf` para poder
// usarla dentro de las pruebas.
func seedForTests() {
	// Esta es una versión simplificada. En un proyecto real, se reutilizaría
	// el código del seeder principal.

	// Seed EGMs
	egms := []model.EGM{{ID: 1004, Game: "DIOSES DE AZAR"} /* ... más egms ... */}
	egmsFile, _ := json.MarshalIndent(egms, "", "  ")
	os.WriteFile(filepath.Join(project.ProjectRoot, "db", "egms.json"), egmsFile, 0644)

	// Seed Socios
	socios := []model.Socio{
		{ID: 1, DNI: "12345678", Password: "pass123", RealName: "Test User", Balance: 1000, Puntaje: 100},
		{ID: 2, DNI: "20250514", RealName: "CVIP", Balance: 1000000, Puntaje: 6666, Password: "1234"},
	}
	sociosFile, _ := json.MarshalIndent(socios, "", "  ")
	os.WriteFile(filepath.Join(project.ProjectRoot, "db", "socios.json"), sociosFile, 0644)

	// Seed Operadores
	operadores := []model.Operator{
		{UserID: "asistenteUNO", OperadorID: 671, Password: "8d969eef6ecad3c29a3a629280e686cf0c3f5d5a86aff3ca12020c923adc6c92", Profiles: []string{"asistente"}, Nickname: "asistenteUNO", RealName: "Bruce Wayne", DNI: 30350516},
		{UserID: "asistenteDOS", OperadorID: 672, Password: "8d969eef6ecad3c29a3a629280e686cf0c3f5d5a86aff3ca12020c923adc6c92", Profiles: []string{"asistente"}, Nickname: "asistenteDOS", RealName: "Clark kent", DNI: 30350515},
	}
	operadoresFile, _ := json.MarshalIndent(operadores, "", "  ")
	os.WriteFile(filepath.Join(project.ProjectRoot, "db", "operadores.json"), operadoresFile, 0644)

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

func TestInfoEndpoint(t *testing.T) {
	server := setupTestServer(t)
	router := server.router

	// Hacemos la petición POST al endpoint /info
	req, _ := http.NewRequest("POST", "/info", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 1. Verificar el código de estado
	assert.Equal(t, http.StatusOK, w.Code, "El código de estado de /info debe ser 200")

	// 2. Verificar el cuerpo de la respuesta
	var response map[string]model.InfoResponse // La respuesta está envuelta en un objeto "data"
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err, "La respuesta de /info debe ser un JSON válido")

	// 3. Verificar los campos
	infoData, ok := response["data"]
	assert.True(t, ok, "La respuesta debe contener un campo 'data'")

	assert.NotEmpty(t, infoData.GoVersion, "goVersion no debe estar vacío")
	assert.Equal(t, "3.8.0", infoData.Version, "La versión de la app debe coincidir")
	assert.Equal(t, server.cfg.Port, infoData.Port, "El puerto debe coincidir con la configuración")
	assert.Equal(t, false, infoData.Asistente, "El campo asistente debe ser false")
	assert.Equal(t, runtime.GOOS+"/"+runtime.GOARCH, infoData.HostPlatform, "La plataforma del host debe ser correcta")
}

// ...
func TestOperatorLogin(t *testing.T) {
	server := setupTestServer(t)
	router := server.router

	// Contraseña para el seeder de operadores es el hash
	password := "8d969eef6ecad3c29a3a629280e686cf0c3f5d5a86aff3ca12020c923adc6c92"
	loginBody := fmt.Sprintf(`{"data": {"userId": "asistenteUNO", "password": "%s"}}`, password)

	req, _ := http.NewRequest("POST", "/login", bytes.NewBufferString(loginBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code, "El login del operador debería ser exitoso")

	var loginResponse map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &loginResponse)

	data, _ := loginResponse["data"].(map[string]interface{})
	profiles, _ := data["userProfiles"].([]interface{})

	assert.Equal(t, "asistente", profiles[0], "El perfil del usuario debe ser 'asistente'")
}
