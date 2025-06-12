package model

// EGM representa una máquina de juego electrónica.
type EGM struct {
	ID         int     `json:"id"`
	IsOccupied bool    `json:"isOccupied"`
	OccupiedBy *int    `json:"occupiedBy"` // Usamos un puntero para poder tener 'null'
	Game       string  `json:"game"`
	Credits    float64 `json:"ca"` // 'ca' por 'cantidad de creditos'
}
