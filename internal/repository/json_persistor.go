package repository

import (
	"encoding/json"
	"os"
	"sync"
)

// JSONPersistor maneja la lectura y escritura de datos en un archivo JSON.
// Es genérico y puede funcionar con cualquier tipo de slice (ej. []model.EGM, []model.Socio).
type JSONPersistor[T any] struct {
	filePath string
	mu       sync.RWMutex // Mutex para proteger el acceso al archivo
	data     []T
}

// NewJSONPersistor crea una nueva instancia del persistor y carga los datos iniciales.
func NewJSONPersistor[T any](filePath string, initialData []T) (*JSONPersistor[T], error) {
	p := &JSONPersistor[T]{
		filePath: filePath,
		data:     initialData,
	}

	// Intentar cargar datos existentes. Si no existe el archivo, lo crea con los datos iniciales.
	if err := p.load(); err != nil {
		if os.IsNotExist(err) {
			// El archivo no existe, lo creamos
			if err := p.save(); err != nil {
				return nil, err
			}
		} else {
			// Otro tipo de error al cargar
			return nil, err
		}
	}
	return p, nil
}

// load lee el archivo JSON y lo decodifica en p.data.
func (p *JSONPersistor[T]) load() error {
	p.mu.RLock()
	defer p.mu.RUnlock()

	file, err := os.ReadFile(p.filePath)
	if err != nil {
		return err
	}
	return json.Unmarshal(file, &p.data)
}

// save codifica p.data a JSON y lo escribe en el archivo.
func (p *JSONPersistor[T]) save() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	file, err := json.MarshalIndent(p.data, "", "  ") // MarshalIndent para un JSON legible
	if err != nil {
		return err
	}
	return os.WriteFile(p.filePath, file, 0644)
}

// GetData devuelve una copia de los datos para su manipulación.
// La lógica de negocio trabajará sobre esta copia y luego llamará a UpdateAll.
// @implement Persistor
func (p *JSONPersistor[T]) GetData() []T {
	p.mu.RLock()
	defer p.mu.RUnlock()
	// Devolvemos una copia para evitar que se modifique el slice original sin pasar por el save
	dataCopy := make([]T, len(p.data))
	copy(dataCopy, p.data)
	return dataCopy
}

// UpdateAll reemplaza todos los datos en memoria y los guarda en el archivo.
// @implement Persistor
func (p *JSONPersistor[T]) UpdateAll(newData []T) error {
	p.mu.Lock()
	p.data = newData
	p.mu.Unlock()
	return p.save()
}
