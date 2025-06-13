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

	"athena.mock/backend/internal/auth"
	"athena.mock/backend/internal/config"
	"athena.mock/backend/internal/model"
	"athena.mock/backend/internal/project"
	"athena.mock/backend/internal/service"
	"github.com/stretchr/testify/assert"

	"github.com/golang-jwt/jwt/v5" // Y el paquete jwt
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
	router := server.router

	// --- Test de Login ---
	loginBody := `{"data": {"userId": "12345678", "password": "pass123"}}`
	loginReq, _ := http.NewRequest("POST", "/login", bytes.NewBufferString(loginBody))
	loginReq.Header.Set("Content-Type", "application/json")

	wLogin := httptest.NewRecorder()
	router.ServeHTTP(wLogin, loginReq)

	assert.Equal(t, http.StatusOK, wLogin.Code, "El código de estado del login debería ser 200")

	var loginResponse map[string]interface{}
	err := json.Unmarshal(wLogin.Body.Bytes(), &loginResponse)
	assert.NoError(t, err)

	// --- ¡CAMBIO CLAVE! CAPTURAR EL TOKEN ---
	data, _ := loginResponse["data"].(map[string]interface{})
	token, tokenExists := data["token"].(string)
	assert.True(t, tokenExists, "La respuesta del login debe contener un token")

	// --- Test de Logout CON TOKEN ---
	// La prueba original enviaba un body, pero nuestro nuevo handler ya no lo necesita.
	// Enviaremos un body vacío `{}` ya que la petición es POST.
	logoutReq, _ := http.NewRequest("POST", "/logout", bytes.NewBufferString(`{}`))
	logoutReq.Header.Set("Content-Type", "application/json")
	// --- ¡CAMBIO CLAVE! AÑADIR LA CABECERA DE AUTORIZACIÓN ---
	logoutReq.Header.Set("Authorization", "Bearer "+token)

	wLogout := httptest.NewRecorder()
	router.ServeHTTP(wLogout, logoutReq)

	// Aserciones del Logout
	assert.Equal(t, http.StatusOK, wLogout.Code, "El código de estado del logout debería ser 200")
	var logoutResponse map[string]interface{}
	json.Unmarshal(wLogout.Body.Bytes(), &logoutResponse)
	assert.Equal(t, "logout", logoutResponse["requestType"], "El tipo de petición de logout debe ser 'logout'")
	assert.Contains(t, logoutResponse["message"], "Sesión finalizada", "El mensaje de logout debe indicar que la sesión finalizó")

	// --- Verificación Post-Logout ---
	// Esta parte ya era correcta.
	postLogoutLoginReq, _ := http.NewRequest("POST", "/login", bytes.NewBufferString(loginBody))
	postLogoutLoginReq.Header.Set("Content-Type", "application/json")
	wPostLogout := httptest.NewRecorder()
	router.ServeHTTP(wPostLogout, postLogoutLoginReq)

	assert.Equal(t, http.StatusOK, wPostLogout.Code, "Debería ser posible hacer login de nuevo después del logout")
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

// Test 1: Verifica que un login de socio es exitoso y el JWT contiene los datos correctos.
func TestLoginSuccessAndJWTValidation(t *testing.T) {
	server := setupTestServer(t)
	router := server.router

	// 1. Arrange: Preparar el cuerpo de la petición
	loginBody := `{"data": {"userId": "12345678", "password": "pass123"}}`
	req, _ := http.NewRequest("POST", "/login", bytes.NewBufferString(loginBody))
	req.Header.Set("Content-Type", "application/json")

	// 2. Act: Ejecutar la petición
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 3. Assert: Verificar la respuesta y el token
	assert.Equal(t, http.StatusOK, w.Code, "El código de estado del login debería ser 200")

	var loginResponse map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &loginResponse)
	assert.NoError(t, err)

	data, _ := loginResponse["data"].(map[string]interface{})
	tokenString, ok := data["token"].(string)
	assert.True(t, ok, "La respuesta debe contener un token")

	// 4. Validar el contenido del JWT
	claims := &auth.Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		// Verificamos que el algoritmo de firma sea el esperado (HS256)
		_, ok := token.Method.(*jwt.SigningMethodHMAC)
		assert.True(t, ok, "El algoritmo de firma del token es inesperado")
		return []byte(server.cfg.SecretKey), nil
	})

	assert.NoError(t, err, "El token JWT debería ser válido y parseable")
	assert.True(t, token.Valid, "El campo 'valid' del token debe ser true")

	// Verificamos los claims (el "contenido" del token)
	assert.Equal(t, "1", claims.UserID, "El ID del usuario en el token es incorrecto") // El socio con DNI 12345678 tiene ID 1 en el seeder
	assert.Equal(t, "partner", claims.Type, "El tipo de usuario en el token debe ser 'partner'")
	assert.Contains(t, claims.Roles, "socio", "El rol 'socio' debe estar en el token")
}

// Test 2: Verifica que un segundo intento de login para el mismo usuario falla.
func TestLoginFailsIfAlreadyLoggedIn(t *testing.T) {
	server := setupTestServer(t)
	router := server.router

	// 1. Arrange: Preparar el cuerpo de la petición
	loginBody := `{"data": {"userId": "12345678", "password": "pass123"}}`
	req, _ := http.NewRequest("POST", "/login", bytes.NewBufferString(loginBody))
	req.Header.Set("Content-Type", "application/json")

	// 2. Act (Primer Login): Ejecutar la primera petición. Debería ser exitosa.
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req)

	// 3. Assert (Primer Login): Verificar que el primer login fue exitoso.
	assert.Equal(t, http.StatusOK, w1.Code, "El primer login debería ser exitoso")

	// 4. Act (Segundo Login): Ejecutar la misma petición de nuevo.
	req, _ = http.NewRequest("POST", "/login", bytes.NewBufferString(loginBody))
	req.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req)

	// 5. Assert (Segundo Login): Verificar que el segundo login falla con el mensaje correcto.
	assert.Equal(t, http.StatusUnauthorized, w2.Code, "El segundo login debería fallar con código 401")

	var errorResponse map[string]interface{}
	err := json.Unmarshal(w2.Body.Bytes(), &errorResponse)
	assert.NoError(t, err)

	expectedMessage := "Usuario ya logueado en otro dispositivo"
	assert.Equal(t, expectedMessage, errorResponse["message"], "El mensaje de error para doble login es incorrecto")
}

// Test 1: Verifica que el login falla con contraseña incorrecta.
func TestLoginFailsWithInvalidPassword(t *testing.T) {
	server := setupTestServer(t)
	router := server.router

	// 1. Arrange: Preparar el body con una contraseña que no corresponde.
	loginBody := `{"data": {"userId": "12345678", "password": "wrongpassword"}}`
	req, _ := http.NewRequest("POST", "/login", bytes.NewBufferString(loginBody))
	req.Header.Set("Content-Type", "application/json")

	// 2. Act
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 3. Assert
	assert.Equal(t, http.StatusUnauthorized, w.Code, "El login con contraseña incorrecta debe devolver 401")

	var errorResponse map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &errorResponse)
	assert.Equal(t, "Usuario o contraseña incorrecta", errorResponse["message"], "El mensaje de error es incorrecto")
}

// Test 2: Verifica que el login falla con contraseña vacía.
func TestLoginFailsWithEmptyPassword(t *testing.T) {
	server := setupTestServer(t)
	router := server.router

	// 1. Arrange: Preparar el body con una contraseña vacía.
	loginBody := `{"data": {"userId": "12345678", "password": ""}}`
	req, _ := http.NewRequest("POST", "/login", bytes.NewBufferString(loginBody))
	req.Header.Set("Content-Type", "application/json")

	// 2. Act
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 3. Assert
	assert.Equal(t, http.StatusUnauthorized, w.Code, "El login con contraseña vacía debe devolver 401")

	var errorResponse map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &errorResponse)
	assert.Equal(t, "Usuario o contraseña incorrecta", errorResponse["message"], "El mensaje de error es incorrecto")
}

// Test 3: Verifica que el login falla si el usuario no existe.
func TestLoginFailsWithNonExistentUser(t *testing.T) {
	server := setupTestServer(t)
	router := server.router

	// 1. Arrange: Preparar el body con un DNI que no está en nuestro seeder.
	loginBody := `{"data": {"userId": "99999999", "password": "anypassword"}}`
	req, _ := http.NewRequest("POST", "/login", bytes.NewBufferString(loginBody))
	req.Header.Set("Content-Type", "application/json")

	// 2. Act
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 3. Assert
	assert.Equal(t, http.StatusUnauthorized, w.Code, "El login con usuario inexistente debe devolver 401")

	var errorResponse map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &errorResponse)
	assert.Equal(t, "Usuario o contraseña incorrecta", errorResponse["message"], "El mensaje de error es incorrecto")
}

// func TestLogoutWithInvalidToken(t *testing.T) {
// 	server := setupTestServer(t)
// 	router := server.router

// 	// 1. Arrange: Crear una petición de logout con un token falso.
// 	invalidToken := "un.token.falso"
// 	req, _ := http.NewRequest("POST", "/logout", nil)
// 	req.Header.Set("Content-Type", "application/json")
// 	req.Header.Set("Authorization", "Bearer "+invalidToken)

// 	// 2. Act
// 	w := httptest.NewRecorder()
// 	router.ServeHTTP(w, req)

// 	// 3. Assert
// 	assert.Equal(t, http.StatusUnauthorized, w.Code, "Logout con token inválido debe devolver 401")

// 	var errorResponse map[string]interface{}
// 	json.Unmarshal(w.Body.Bytes(), &errorResponse)

// 	// Para replicar el mensaje exacto de la respuesta que pusiste.
// 	// La respuesta real del middleware es solo `{"error":401, "message":"Token inválido"}`
// 	// pero podemos verificar que el mensaje contenga la palabra clave.
// 	assert.Contains(t, errorResponse["message"], "Token inválido", "El mensaje de error es incorrecto")
// }

// func TestLogoutWithExpiredToken(t *testing.T) {
// 	server := setupTestServer(t)
// 	router := server.router

// 	// 1. Arrange: Crear un token que ya ha expirado.
// 	// Para esto, modificaremos temporalmente la función de creación de tokens
// 	// o crearemos uno manualmente. La forma manual es más limpia para un test.
// 	secretKey := []byte(server.cfg.SecretKey)

// 	// Creamos claims para un token que expiró hace una hora.
// 	expiredClaims := &auth.Claims{
// 		UserID:  "1",
// 		Subject: "12345678",
// 		RegisteredClaims: jwt.RegisteredClaims{
// 			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)),
// 		},
// 	}
// 	token := jwt.NewWithClaims(jwt.SigningMethodHS256, expiredClaims)
// 	expiredTokenString, err := token.SignedString(secretKey)
// 	assert.NoError(t, err)

// 	req, _ := http.NewRequest("POST", "/logout", nil)
// 	req.Header.Set("Authorization", "Bearer "+expiredTokenString)

// 	// 2. Act
// 	w := httptest.NewRecorder()
// 	router.ServeHTTP(w, req)

// 	// 3. Assert
// 	assert.Equal(t, http.StatusUnauthorized, w.Code, "Logout con token expirado debe devolver 401")

// 	var errorResponse map[string]interface{}
// 	json.Unmarshal(w.Body.Bytes(), &errorResponse)
// 	assert.Equal(t, "El token ha expirado", errorResponse["message"], "El mensaje para token expirado es incorrecto")
// }
