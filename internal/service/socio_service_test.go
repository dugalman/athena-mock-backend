package service

import (
	"testing"

	"athena.mock/backend/internal/model"
	"athena.mock/backend/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Asegurarnos de que nuestro mock implementa la interfaz en tiempo de compilación.
// Es una buena práctica. Si no la implementa, la compilación fallará aquí.
var _ repository.SocioPersistor = (*repository.MockPersistor[model.Socio])(nil)

func TestSocioService_GetPuntaje(t *testing.T) {
	// 1. Arrange (Preparación)
	mockRepo := new(repository.MockPersistor[model.Socio])

	// ¡AHORA ESTO FUNCIONA!
	// Porque mockRepo (un *MockPersistor) implementa los métodos de la interfaz
	// SocioPersistor, que es lo que el campo 'repo' espera.
	//
	// Creamos un servicio de socio usando nuestro repositorio mockeado
	socioService := &SocioService{repo: mockRepo}

	// Datos de prueba
	testSocios := []model.Socio{
		{ID: 1, Puntaje: 100},
		{ID: 2, Puntaje: 200},
	}
	socioID := 2
	expectedPuntaje := 200

	// Configuramos el mock: "Cuando se llame a GetData, devuelve testSocios"
	mockRepo.On("GetData").Return(testSocios)

	// 2. Act (Actuación)
	puntaje, err := socioService.GetPuntaje(socioID)

	// 3. Assert (Verificación)
	assert.NoError(t, err, "No debería haber error al obtener el puntaje")
	assert.Equal(t, expectedPuntaje, puntaje, "El puntaje obtenido debe ser el esperado")

	// Verificamos que los métodos del mock fueron llamados como se esperaba
	mockRepo.AssertExpectations(t)
}

func TestSocioService_AddPuntaje(t *testing.T) {
	// 1. Arrange
	mockRepo := new(repository.MockPersistor[model.Socio])
	socioService := &SocioService{repo: mockRepo}

	socioID := 1
	initialPuntaje := 100
	addedPuntaje := 50

	// Datos de prueba
	testSocios := []model.Socio{
		{ID: socioID, Puntaje: initialPuntaje},
	}

	// Configuramos el mock para GetData
	mockRepo.On("GetData").Return(testSocios)

	// Configuramos el mock para UpdateAll.
	// Esperamos que se llame con un slice donde el puntaje del socio 1 sea 150.
	// El `mock.Anything` nos permite no ser tan estrictos con el slice exacto.
	// `Return(nil)` significa que la operación de guardado no devuelve error.
	mockRepo.On("UpdateAll", mock.MatchedBy(func(socios []model.Socio) bool {
		return len(socios) == 1 && socios[0].ID == socioID && socios[0].Puntaje == initialPuntaje+addedPuntaje
	})).Return(nil)

	// 2. Act
	err := socioService.AddPuntaje(socioID, addedPuntaje)

	// 3. Assert
	assert.NoError(t, err, "No debería haber error al añadir puntaje")
	mockRepo.AssertExpectations(t)
}

func TestSocioService_GetPuntaje_SocioNotFound(t *testing.T) {
	// 1. Arrange
	mockRepo := new(repository.MockPersistor[model.Socio])
	socioService := &SocioService{repo: mockRepo}

	// Devolvemos un slice vacío para simular que no se encuentra el socio
	mockRepo.On("GetData").Return([]model.Socio{})

	// 2. Act
	_, err := socioService.GetPuntaje(999) // ID que no existe

	// 3. Assert
	assert.Error(t, err, "Debería devolver un error si el socio no se encuentra")
	assert.Equal(t, ErrSocioNotFound, err, "El error debería ser ErrSocioNotFound")
	mockRepo.AssertExpectations(t)
}
