package main

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"

	"athena.mock/backend/internal/model"
	"athena.mock/backend/internal/project"
)

func main() {
	log.Println("Iniciando seeder...")

	// Asegurar que el directorio 'db' exista
	if err := os.MkdirAll("db", 0755); err != nil {
		log.Fatalf("No se pudo crear el directorio db: %v", err)
	}

	seedEGMs()
	seedSocios()
	seedOperadores()

	log.Println("Seeding completado exitosamente.")
}

func seedOperadores() {
	operadores := []model.Operator{
		{UserID: "asistenteDOS", OperadorID: 672, Password: "8d969eef6ecad3c29a3a629280e686cf0c3f5d5a86aff3ca12020c923adc6c92", Profiles: []string{"asistente"}, Nickname: "asistenteDOS", RealName: "Clark kent", DNI: 30350515},
		{UserID: "asistenteTRES", OperadorID: 673, Password: "8d969eef6ecad3c29a3a629280e686cf0c3f5d5a86aff3ca12020c923adc6c92", Profiles: []string{"asistente"}, Nickname: "asistenteTRES", RealName: "Lana Lan", DNI: 30350516},
		{UserID: "asistenteUNO", OperadorID: 671, Password: "8d969eef6ecad3c29a3a629280e686cf0c3f5d5a86aff3ca12020c923adc6c92", Profiles: []string{"asistente"}, Nickname: "asistenteUNO", RealName: "Bruce Wayne", DNI: 30350516},
	}

	dbPath := filepath.Join(project.ProjectRoot, "db", "operadores.json")
	saveToJSON(dbPath, operadores)
	log.Println("->", dbPath, "sembrado con", len(operadores), "registros.")
}

func seedEGMs() {
	egms := []model.EGM{
		{ID: 1000, Game: "CLEOPATRA"},
		{ID: 1001, Game: "ZAPITO RULES"},
		{ID: 1002, Game: "ENDLESS FORTUNA"},
		{ID: 1003, Game: "TESOROS DEL INFRAMUNDO"},
		{ID: 1004, Game: "DIOSES DE AZAR"},
		{ID: 1005, Game: "LUCKY VALKYRIA"},
		{ID: 1006, Game: "ORO ETERNO"},
		{ID: 1007, Game: "SEVEN PYRAMIDS"},
		{ID: 1008, Game: "CAZADORES DE JACKPOT"},
		{ID: 1009, Game: "RUEDA DEL DESTINO"},
		{ID: 1010, Game: "MAGIA DEL DRAGÓN"},
	}

	// Guardar en archivo JSON
	dbPath := filepath.Join(project.ProjectRoot, "db", "egms.json")
	saveToJSON(dbPath, egms)
	log.Println("->", dbPath, "sembrado con", len(egms), "registros.")
}

func seedSocios() {
	socios := []model.Socio{
		{ID: 1, DNI: "20250513", RealName: "CASH", Balance: 1000000, Puntaje: 777},
		{ID: 2, DNI: "20250514", RealName: "CVIP", Balance: 1000000, Puntaje: 6666},
		{ID: 3, DNI: "48572913", RealName: "Juan Salvo", Balance: 1000000, Puntaje: 0},
		{ID: 4, DNI: "23650198", RealName: "Inodoro Pereyra", Balance: 1000000, Puntaje: 0},
		{ID: 5, DNI: "39158276", RealName: "Jorge Lavandina Pérez", Balance: 1000000, Puntaje: 0},
		{ID: 6, DNI: "41736084", RealName: "Gabriel David León", Balance: 1000000, Puntaje: 0},
		{ID: 7, DNI: "30498127", RealName: "Esteban Espinosa", Balance: 1000000, Puntaje: 0},
	}

	// Asignar contraseñas por defecto si es necesario
	for i := range socios {
		if socios[i].Password == "" {
			socios[i].Password = "123456"
		}
	}

	dbPath := filepath.Join(project.ProjectRoot, "db", "socios.json") // <-- Ruta absoluta
	saveToJSON(dbPath, socios)
	log.Println("->", dbPath, "sembrado con", len(socios), "registros.")
}

// saveToJSON es una función helper para escribir los datos.
func saveToJSON(filePath string, data interface{}) {
	file, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		log.Fatalf("Error al serializar datos para %s: %v", filePath, err)
	}
	if err := os.WriteFile(filePath, file, 0644); err != nil {
		log.Fatalf("Error al escribir el archivo %s: %v", filePath, err)
	}
}
