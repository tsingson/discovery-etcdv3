package discovery

import (
	"fmt"

	"github.com/grpc-ecosystem/grpc-opentracing/go/otgrpc"
	"github.com/hashicorp/consul/api"
	lb "github.com/tsingson/discovery-etcdv3/grpc-consul/discovery/lb"
	"google.golang.org/grpc"
)

// discovery provider
type discovery struct {
	*api.Client
	dialopts []grpc.DialOption
}

// NewConsulDiscovery returns discovery
func NewConsulDiscovery(cfg Config) (Discovery, error) {
	config := api.DefaultConfig()
	config.Address = fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)
	c, err := api.NewClient(config)
	if err != nil {
		return nil, err
	}

	opts := []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithBlock(),
	}

	if cfg.Tracer != nil {
		opts = append(opts, grpc.WithUnaryInterceptor(otgrpc.OpenTracingClientInterceptor(cfg.Tracer)))
		opts = append(opts, grpc.WithStreamInterceptor(otgrpc.OpenTracingStreamClientInterceptor(cfg.Tracer)))
	}

	return discovery{c, opts}, nil
}

// Dial grpc server
func (c discovery) Dial(name string, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	r, err := lb.NewResolver(c.Client, name, "")

	if err != nil {
		return nil, fmt.Errorf("Create balancer resolver for service %s failed. Error: %v", name, err)
	}
	c.dialopts = append(c.dialopts, grpc.WithBalancer(grpc.RoundRobin(r)))
	c.dialopts = append(c.dialopts, opts...)

	conn, err := grpc.Dial("", c.dialopts...)
	if err != nil {
		return nil, fmt.Errorf("Failed to dial %s: %v", name, err)
	}
	return conn, nil
}
