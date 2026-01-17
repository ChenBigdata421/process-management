package di

import (
	"sync"

	"go.uber.org/dig"
)

var (
	instance *dig.Container
	once     sync.Once
)

// GetContainer returns the singleton instance of the dig container
func GetContainer() *dig.Container {
	once.Do(func() {
		instance = dig.New()
	})
	return instance
}

// Provide is a helper function to add dependencies to the container
func Provide(constructor interface{}, opts ...dig.ProvideOption) error {
	return GetContainer().Provide(constructor, opts...)
}

// Invoke is a helper function to resolve and use dependencies from the container
func Invoke(function interface{}, opts ...dig.InvokeOption) error {
	return GetContainer().Invoke(function, opts...)
}
