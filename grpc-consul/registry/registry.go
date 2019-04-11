package registry

import (
	"fmt"
)

// Provider enum
type Provider int

const (
	// Consul provider
	Consul Provider = iota
)

// Config for Registry
type Config struct {
	Provider Provider
	Host     string
	Port     string
}

// Registry interface for extend
type Registry interface {
	Register(id string, name string, port int, tags ...string) error
	DeRegister(string) error
}

// NewRegistry returns Registry via provider, e.g. ConsulRegistry
func NewRegistry(cfg Config) (Registry, error) {
	switch cfg.Provider {
	case Consul:
		return NewConsulRegistry(cfg)
	default:
		return nil, fmt.Errorf("Unsupported registry provider: %v", cfg.Provider)
	}
}
