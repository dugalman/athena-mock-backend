package service

import (
	"errors"
	"fmt"
	"path/filepath"
	"sync"

	"athena.mock/backend/internal/model"
	"athena.mock/backend/internal/project"
	"athena.mock/backend/internal/repository"
)

var (
	ErrEGMNotFound     = errors.New("EGM no encontrada")
	egmServiceInstance *EGMService
	egmOnce            sync.Once
)

type EGMService struct {
	repo *repository.JSONPersistor[model.EGM]
}

// GetEGMService initializes and returns a singleton instance of EGMService.
// It ensures that the EGMService is only created once using sync.Once.
// If the underlying repository file does not exist, it initializes it with default EGM data.
// Returns the EGMService instance and any error encountered during initialization.
func GetEGMService() (*EGMService, error) {
	var err error
	egmOnce.Do(func() {
		// Datos iniciales si el archivo no existe
		initialEGMs := []model.EGM{
			{ID: 1004, IsOccupied: false, OccupiedBy: nil, Game: "DIOSES DE AZAR", Credits: 0},
			{ID: 1005, IsOccupied: false, OccupiedBy: nil, Game: "FORTUNE COINS", Credits: 150.5},
		}
		// Asegúrate de que exista la carpeta 'db'
		dbPath := filepath.Join(project.ProjectRoot, "db", "egms.json") // <-- Ruta absoluta
		repo, repoErr := repository.NewJSONPersistor(dbPath, initialEGMs)
		if repoErr != nil {
			err = fmt.Errorf("fallo al inicializar repositorio EGM: %w", repoErr)
			return
		}
		egmServiceInstance = &EGMService{repo: repo}
	})
	return egmServiceInstance, err
}

// findEGM busca una EGM por ID y devuelve un puntero a ella para modificarla.
func (s *EGMService) findEGM(egms []model.EGM, egmID int) (*model.EGM, int) {
	for i := range egms {
		if egms[i].ID == egmID {
			return &egms[i], i
		}
	}
	return nil, -1
}

func (s *EGMService) AddCredit(egmID int, amount float64) error {
	egms := s.repo.GetData()
	egm, _ := s.findEGM(egms, egmID)
	if egm == nil {
		return ErrEGMNotFound
	}

	egm.Credits += amount
	return s.repo.UpdateAll(egms)
}

func (s *EGMService) RemoveAllCredit(egmID int) (float64, error) {
	egms := s.repo.GetData()
	egm, _ := s.findEGM(egms, egmID)
	if egm == nil {
		return 0, ErrEGMNotFound
	}

	removedAmount := egm.Credits
	egm.Credits = 0
	err := s.repo.UpdateAll(egms)
	return removedAmount, err
}

func (s *EGMService) RemoveCredit(egmID int, amount float64) error {
	egms := s.repo.GetData()
	egm, _ := s.findEGM(egms, egmID)
	if egm == nil {
		return ErrEGMNotFound
	}
	if egm.Credits < amount {
		return errors.New("créditos insuficientes en la EGM")
	}

	egm.Credits -= amount
	return s.repo.UpdateAll(egms)
}

func (s *EGMService) BindEgmToUser(egmID int, userID int) error {
	egms := s.repo.GetData()
	egm, _ := s.findEGM(egms, egmID)
	if egm == nil {
		return ErrEGMNotFound
	}
	if egm.IsOccupied {
		return fmt.Errorf("EGM %d ya está ocupada por el usuario %d", egmID, *egm.OccupiedBy)
	}

	egm.IsOccupied = true
	egm.OccupiedBy = &userID
	return s.repo.UpdateAll(egms)
}

func (s *EGMService) UnbindEgmToUser(egmID int) error {
	egms := s.repo.GetData()
	egm, _ := s.findEGM(egms, egmID)
	if egm == nil {
		return ErrEGMNotFound
	}

	egm.IsOccupied = false
	egm.OccupiedBy = nil
	return s.repo.UpdateAll(egms)
}
