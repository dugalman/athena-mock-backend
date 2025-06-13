package service

import (
	"errors"
	"fmt"
	"path/filepath"
	"sync" // Importamos sync

	"athena.mock/backend/internal/model"
	"athena.mock/backend/internal/project"
	"athena.mock/backend/internal/repository"
)

var (
	ErropertorNotFound      = errors.New("operador no encontrado")
	operatorServiceInstance *OperatorService
	operatorOnce            sync.Once
)

// ResetSocioServiceForTests reinicia el singleton. SOLO PARA USAR EN PRUEBAS.
func ResetOperatorServiceForTests() {
	operatorOnce = sync.Once{}
	operatorServiceInstance = nil
}

type OperatorService struct {
	// repo repository.Persistor[model.Operator]
	repo repository.OperatorPersistor // <-- ¡CAMBIO CLAVE!
}

func GetOperatorService() (*OperatorService, error) {
	var initErr error // Usaremos una variable fuera del 'Do'

	operatorOnce.Do(func() {

		defaultOperators := []model.Operator{
			{UserID: "asistenteUNO", OperadorID: 671, Password: "8d969eef6ecad3c29a3a629280e686cf0c3f5d5a86aff3ca12020c923adc6c92", Profiles: []string{"asistente"}, Nickname: "asistenteUNO", RealName: "Bruce Wayne", DNI: 30350516},
			{UserID: "asistenteDOS", OperadorID: 672, Password: "8d969eef6ecad3c29a3a629280e686cf0c3f5d5a86aff3ca12020c923adc6c92", Profiles: []string{"asistente"}, Nickname: "asistenteDOS", RealName: "Clark kent", DNI: 30350515},
		}

		// Asegúrate de que exista la carpeta 'db'
		// jsonRepo es un *JSONPersistor, que IMPLEMENTA la interfaz SocioPersistor.
		dbPath := filepath.Join(project.ProjectRoot, "db", "operadores.json") //<-- Ruta absoluta
		jsonRepo, repoErr := repository.NewJSONPersistor(dbPath, defaultOperators)
		if repoErr != nil {
			initErr = fmt.Errorf("fallo al inicializar el repositorio de Operadores: %w", repoErr)
			return
		}
		operatorServiceInstance = &OperatorService{repo: jsonRepo}
	})
	return operatorServiceInstance, initErr
}

// findSocio busca un socio por ID y devuelve un puntero a él.
func (s *SocioService) findOperator(operators []model.Operator, dni int) (*model.Operator, int) {
	for i := range operators {
		if operators[i].DNI == dni {
			return &operators[i], i
		}
	}
	return nil, -1
}
