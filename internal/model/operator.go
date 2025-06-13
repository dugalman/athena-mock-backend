package model

// Operator representa a un usuario del sistema con rol de asistente.
type Operator struct {
	UserID     string   `json:"userId"`
	OperadorID int      `json:"operadorId"`
	Password   string   `json:"password"`
	Profiles   []string `json:"profiles"`
	Nickname   string   `json:"nickname"`
	RealName   string   `json:"realName"`
	DNI        int      `json:"dni"`
}
