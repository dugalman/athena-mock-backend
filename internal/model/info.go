package model

// InfoResponse define la estructura de la respuesta para el endpoint /info.
// Usamos `json:"..."` para controlar cómo se nombran los campos en la salida JSON.
type InfoResponse struct {
	Environment  string `json:"environment"`
	GoVersion    string `json:"goVersion"` // Cambiado de nodeVersion
	HostPlatform string `json:"hostPlatform"`
	Port         string `json:"port"`
	Version      string `json:"version"`
	Asistente    bool   `json:"asistente"`
}
