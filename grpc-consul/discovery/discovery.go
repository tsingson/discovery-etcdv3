package discovery

import (
	"fmt"

	"github.com/opentracing/opentracing-go"
	"google.golang.org/grpc"
)

// Provider enum
type Provider int

const (
	// Consul provider
	Consul Provider = iota
)

// Config for discovery
type Config struct {
	Provider Provider
	Host     string
	Port     string
	Tracer   opentracing.Tracer
}

// Discovery service
type Discovery interface {
	Dial(name string, opts ...grpc.DialOption) (*grpc.ClientConn, error)
}

// NewDiscovery returns Discovery
func NewDiscovery(cfg Config) (Discovery, error) {
	switch cfg.Provider {
	case Consul:
		return NewConsulDiscovery(cfg)
	default:
		return nil, fmt.Errorf("Unsupported provider %v", cfg.Provider)
	}
}
