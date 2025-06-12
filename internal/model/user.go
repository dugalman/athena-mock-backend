package model

// User representa la estructura genérica de un usuario, ya sea Socio u Operador.
type User struct {
	ID       string   `json:"id"`
	UserID   string   `json:"userId"` // DNI para Socio, Username para Operador
	Password string   `json:"-"`      // El '-' evita que se serialice en JSON
	Profiles []string `json:"profiles"`
	Nickname string   `json:"nickname"`
	UserType string   `json:"-"`
	ViewAPK  any      `json:"-"` // Usamos 'any' para flexibilidad
}

// Mock de base de datos en memoria
var usersDB = map[string]User{
	"12345678": {
		ID:       "1",
		UserID:   "12345678",
		Password: "pass123",
		Profiles: []string{"socio"},
		Nickname: "Socio de Prueba",
		UserType: "partner",
		ViewAPK:  map[string]string{"menu": "socioMenu"},
	},
	"operador1": {
		ID:       "op-001",
		UserID:   "operador1",
		Password: "opass",
		Profiles: []string{"asistente"},
		Nickname: "Asistente de Prueba",
		UserType: "operator",
		ViewAPK:  map[string]string{"menu": "asistenteMenu"},
	},
}

// FindUserByID busca un usuario en nuestro mock de DB
func FindUserByID(userID string) (User, bool) {
	user, found := usersDB[userID]
	return user, found
}
