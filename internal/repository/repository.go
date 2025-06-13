package repository

import "athena.mock/backend/internal/model"

// Aquí definimos las interfaces. Estas son las abstracciones.

// Persistor es la interfaz que define cómo interactuamos con nuestra capa de persistencia.
// Cualquier tipo que implemente GetData y UpdateAll satisfará esta interfaz.
type Persistor[T any] interface {
	GetData() []T
	UpdateAll(data []T) error
}

// Interfaces específicas para nuestros modelos para mayor claridad.
type EGMPersistor interface {
	Persistor[model.EGM]
}

type SocioPersistor interface {
	Persistor[model.Socio]
}

type OperatorPersistor interface {
	Persistor[model.Operator]
}
