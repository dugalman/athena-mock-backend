package model

// Socio representa a un miembro del club.
type Socio struct {
	ID       int     `json:"id"`
	RealName string  `json:"realName"`
	Name     string  `json:"name"`
	DNI      string  `json:"dni"`
	Password string  `json:"password"`
	Balance  float64 `json:"balance"`
	Puntaje  int     `json:"puntaje"`
}
