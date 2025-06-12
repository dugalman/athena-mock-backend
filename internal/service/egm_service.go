package service

import (
	"errors"
	"fmt"

	"athena.mock/backend/internal/model"
	"athena.mock/backend/internal/repository"
)

var (
	ErrEGMNotFound = errors.New("EGM no encontrada")
)

type EGMService struct {
	repo *repository.JSONPersistor[model.EGM]
}

func NewEGMService() (*EGMService, error) {
	// Datos iniciales si el archivo no existe
	initialEGMs := []model.EGM{
		{ID: 1004, IsOccupied: false, OccupiedBy: nil, Game: "DIOSES DE AZAR", Credits: 0},
		{ID: 1005, IsOccupied: false, OccupiedBy: nil, Game: "FORTUNE COINS", Credits: 150.5},
	}
	// Asegúrate de que exista la carpeta 'db'
	repo, err := repository.NewJSONPersistor("db/egms.json", initialEGMs)
	if err != nil {
		return nil, fmt.Errorf("fallo al inicializar el repositorio de EGMs: %w", err)
	}
	return &EGMService{repo: repo}, nil
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
