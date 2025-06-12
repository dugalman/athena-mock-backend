package service

import (
	"errors"
	"fmt"

	"athena.mock/backend/internal/model"
	"athena.mock/backend/internal/repository"
)

var (
	ErrSocioNotFound       = errors.New("socio no encontrado")
	ErrInsufficientBalance = errors.New("balance insuficiente")
)

type SocioService struct {
	repo *repository.JSONPersistor[model.Socio]
}

func NewSocioService() (*SocioService, error) {
	initialSocios := []model.Socio{
		{ID: 1, RealName: "Juan Perez", Name: "Juancho", DNI: "12345678", Password: "pass123", Balance: 1000.0},
		{ID: 2, RealName: "Maria Garcia", Name: "Mary", DNI: "87654321", Password: "mypass", Balance: 500.0},
	}
	repo, err := repository.NewJSONPersistor("db/socios.json", initialSocios)
	if err != nil {
		return nil, fmt.Errorf("fallo al inicializar el repositorio de Socios: %w", err)
	}
	return &SocioService{repo: repo}, nil
}

// findSocio busca un socio por ID y devuelve un puntero a él.
func (s *SocioService) findSocio(socios []model.Socio, socioID int) (*model.Socio, int) {
	for i := range socios {
		if socios[i].ID == socioID {
			return &socios[i], i
		}
	}
	return nil, -1
}

func (s *SocioService) IncrementBalance(socioID int, amount float64) error {
	socios := s.repo.GetData()
	socio, _ := s.findSocio(socios, socioID)
	if socio == nil {
		return ErrSocioNotFound
	}

	socio.Balance += amount
	return s.repo.UpdateAll(socios)
}

func (s *SocioService) DecrementBalance(socioID int, amount float64) error {
	socios := s.repo.GetData()
	socio, _ := s.findSocio(socios, socioID)
	if socio == nil {
		return ErrSocioNotFound
	}
	if socio.Balance < amount {
		return ErrInsufficientBalance
	}

	socio.Balance -= amount
	return s.repo.UpdateAll(socios)
}

func (s *SocioService) GetBalance(socioID int) (float64, error) {
	socios := s.repo.GetData()
	socio, _ := s.findSocio(socios, socioID)
	if socio == nil {
		return 0, ErrSocioNotFound
	}
	return socio.Balance, nil
}

func (s *SocioService) GetPuntaje(socioID int) (int, error) {
	socios := s.repo.GetData()
	socio, _ := s.findSocio(socios, socioID)
	if socio == nil {
		return 0, ErrSocioNotFound
	}
	return socio.Puntaje, nil
}

func (s *SocioService) AddPuntaje(socioID int, puntaje int) error {
	socios := s.repo.GetData()
	socio, _ := s.findSocio(socios, socioID)
	if socio == nil {
		return ErrSocioNotFound
	}
	socio.Puntaje += puntaje
	return s.repo.UpdateAll(socios)
}

func (s *SocioService) DelPuntaje(socioID int, puntaje int) error {
	socios := s.repo.GetData()
	socio, _ := s.findSocio(socios, socioID)
	if socio == nil {
		return ErrSocioNotFound
	}
	socio.Puntaje -= puntaje
	return s.repo.UpdateAll(socios)
}
