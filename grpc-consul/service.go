package gf

import (
	"fmt"
	"net/http"

	"github.com/tsingson/discovery-etcdv3/grpc-consul/registry"

	"github.com/opentracing/opentracing-go"
	"github.com/tsingson/uuid"
	"google.golang.org/grpc"
)

// Addr for host port
type Addr struct {
	Port int
	Host string
}

// Addr returns host port
func (a *Addr) String() string {
	return fmt.Sprintf("%s:%d", a.Host, a.Port)
}

// Service define.
type Service struct {
	ID   string
	Name string

	ServiceAdrr    Addr
	PrometheusAdrr Addr
	// Interceptors
	UnaryInterceptors  []grpc.UnaryServerInterceptor
	StreamInterceptors []grpc.StreamServerInterceptor
	// The RPC server implementation
	GRPCServer func(*grpc.Server)
	// Tracing
	Tracer opentracing.Tracer
	//Registry
	Registry registry.Registry

	// for shutdown
	grpcServer       *grpc.Server
	prometheusServer *http.Server
}

func generateID(n string) string {
	uid := uuid.NewV4()
	return n + "-" + uid.String()
}

func defaultOptions(n string) *Service {
	return &Service{
		ID:             generateID(n),
		Name:           n,
		GRPCServer:     func(s *grpc.Server) {},
		ServiceAdrr:    Addr{Host: "0.0.0.0", Port: 9100},
		PrometheusAdrr: Addr{Host: "0.0.0.0", Port: 9000},
	}
}

// NewService creates a service
func NewService(name string) *Service {
	return defaultOptions(name)
}

// UseServiceAdrr to set ServiceAdrr
func (service *Service) UseServiceAdrr(addr Addr) {
	service.ServiceAdrr = addr
}

// UsePrometheusAdrr to set PrometheusAdrr
func (service *Service) UsePrometheusAdrr(addr Addr) {
	service.PrometheusAdrr = addr
}

// GRPCImplementation for GRPCServer
func (service *Service) GRPCImplementation(r func(*grpc.Server)) {
	service.GRPCServer = r
}

// UseTracer to set tracer
func (service *Service) UseTracer(t opentracing.Tracer) {
	service.Tracer = t
}

// UseRegistry to set registry
func (service *Service) UseRegistry(r registry.Registry) {
	service.Registry = r
}

// AddUnaryInterceptor adds a unary interceptor to the RPC server
func (service *Service) AddUnaryInterceptor(interceptor grpc.UnaryServerInterceptor) {
	service.UnaryInterceptors = append(service.UnaryInterceptors, interceptor)
}

// AddStreamInterceptor adds a stream interceptor to the RPC server
func (service *Service) AddStreamInterceptor(interceptor grpc.StreamServerInterceptor) {
	service.StreamInterceptors = append(service.StreamInterceptors, interceptor)
}
