package resolver

import (
	"strings"

	"github.com/etcd-io/etcd/clientv3"
	"golang.org/x/xerrors"
	"google.golang.org/grpc/naming"
)

// Resolver is the implementaion of grpc.naming.Resolver
type Resolver struct {
	serviceName string // service name to resolve
}

// NewResolver return resolver with service name
func NewResolver(serviceName string) *Resolver {
	return &Resolver{serviceName: serviceName}
}

// Resolve to resolve the service from etcd, target is the dial address of etcd
// target example: "http://127.0.0.1:2379,http://127.0.0.1:12379,http://127.0.0.1:22379"
func (re *Resolver) Resolve(target string) (naming.Watcher, error) {
	if len(re.serviceName) == 0 {
		return nil, xerrors.New("grpclb: no service name provided")
	}

	// generate etcd client
	client, err := clientv3.New(clientv3.Config{
		Endpoints: strings.Split(target, ","),
	})
	if err != nil {
		return nil, xerrors.Errorf("grpclb: creat clientv3 client failed: %s", err.Error())
	}

	// Return watcher
	return &watcher{re: re, client: *client}, nil
}
