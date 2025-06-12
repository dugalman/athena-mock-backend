package main

import (
	"encoding/json"
	"log"
	"os"

	"athena.mock/backend/internal/model"
)

func main() {
	log.Println("Iniciando seeder...")

	// Asegurar que el directorio 'db' exista
	if err := os.MkdirAll("db", 0755); err != nil {
		log.Fatalf("No se pudo crear el directorio db: %v", err)
	}

	seedEGMs()
	seedSocios()

	log.Println("Seeding completado exitosamente.")
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
	saveToJSON("db/egms.json", egms)
	log.Println("-> db/egms.json sembrado con", len(egms), "registros.")
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

	saveToJSON("db/socios.json", socios)
	log.Println("-> db/socios.json sembrado con", len(socios), "registros.")
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
