package repository

import (
	"github.com/stretchr/testify/mock"
)

// MockPersistor es nuestro mock para JSONPersistor.
type MockPersistor[T any] struct {
	mock.Mock
}

func (m *MockPersistor[T]) GetData() []T {
	args := m.Called()
	return args.Get(0).([]T)
}

func (m *MockPersistor[T]) UpdateAll(data []T) error {
	args := m.Called(data)
	return args.Error(0)
}
