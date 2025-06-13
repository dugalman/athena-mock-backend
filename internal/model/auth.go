package model

import "strconv"

// Authenticatable define el contrato para cualquier entidad que pueda autenticarse.
type Authenticatable interface {
	GetID() string
	GetUserID() string
	GetPassword() string
	GetProfiles() []string
	GetNickname() string
	GetUserType() string
	GetViewAPK() any
}

// Implementación de la interfaz para Socio
func (s Socio) GetID() string         { return strconv.Itoa(s.ID) } // strconv.Itoa convierte int a string
func (s Socio) GetUserID() string     { return s.DNI }
func (s Socio) GetPassword() string   { return s.Password }
func (s Socio) GetProfiles() []string { return []string{"socio"} }
func (s Socio) GetNickname() string   { return s.Name }
func (s Socio) GetUserType() string   { return "partner" }
func (s Socio) GetViewAPK() any       { return map[string]string{"menu": "socioMenu"} } // Simulado

// Implementación de la interfaz para Operator
func (o Operator) GetID() string         { return strconv.Itoa(o.OperadorID) }
func (o Operator) GetUserID() string     { return o.UserID }
func (o Operator) GetPassword() string   { return o.Password }
func (o Operator) GetProfiles() []string { return o.Profiles }
func (o Operator) GetNickname() string   { return o.Nickname }
func (o Operator) GetUserType() string   { return "operator" }
func (o Operator) GetViewAPK() any       { return map[string]string{"menu": "asistenteMenu"} } // Simulado
