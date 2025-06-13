package service

import (
	"errors"
	"regexp"

	"athena.mock/backend/internal/model"
	"athena.mock/backend/internal/repository"
)

var (
	ErrUserNotFound = errors.New("usuario no encontrado")
)

// AuthService maneja la lógica de autenticación para diferentes tipos de usuarios.
type AuthService struct {
	socioRepo    repository.SocioPersistor
	operatorRepo repository.OperatorPersistor
}

// NewAuthService crea una nueva instancia de AuthService.
func NewAuthService(socioService *SocioService, operatorService *OperatorService) *AuthService {
	return &AuthService{
		socioRepo:    socioService.repo,
		operatorRepo: operatorService.repo,
	}
}

// esSocio determina si un userID es numérico (un DNI).
func (s *AuthService) esSocio(userID string) bool {
	isNumeric, _ := regexp.MatchString(`^[0-9]+$`, userID)
	return isNumeric
}

// AuthenticateUser busca y valida a un usuario, sea Socio u Operador.
func (s *AuthService) AuthenticateUser(userID, password string) (model.Authenticatable, error) {
	if s.esSocio(userID) {
		// Lógica para buscar un Socio
		socios := s.socioRepo.GetData()
		for _, socio := range socios {
			if socio.DNI == userID {
				// Caso 1 y 2: Contraseña inválida o vacía
				if socio.Password == password {
					return socio, nil
				}
				// Si encontramos el usuario pero la contraseña es incorrecta, salimos.
				return nil, ErrUserNotFound
			}
		}
	} else {
		// Lógica para buscar un Operador
		operators := s.operatorRepo.GetData()
		for _, op := range operators {
			if op.UserID == userID {
				if op.Password == password {
					return op, nil
				}
				return nil, ErrUserNotFound
			}
		}
	}

	// Caso 3: El bucle termina sin encontrar al usuario
	return nil, ErrUserNotFound
}
