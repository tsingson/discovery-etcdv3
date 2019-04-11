package registry

import (
	"fmt"
	"net"

	"github.com/hashicorp/consul/api"
	"github.com/sirupsen/logrus"
)

type consul struct {
	client *api.Client
	addr   string
}

// NewConsulRegistry returns a registryClient interface for given consul address
func NewConsulRegistry(c Config) (Registry, error) {
	addr := fmt.Sprintf("%s:%s", c.Host, c.Port)
	if addr == "" {
		addr = "consul:8500"
	}
	cfg := api.DefaultConfig()
	cfg.Address = addr
	cl, err := api.NewClient(cfg)
	if err != nil {
		logrus.Errorf("Can't connect to consul server at %s", addr)
		return nil, err
	}
	return consul{client: cl, addr: addr}, nil
}

func (r consul) Register(id string, name string, port int, tags ...string) error {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return fmt.Errorf("unable to determine local addr: %v", err)
	}
	defer conn.Close()
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	asr := &api.AgentServiceRegistration{
		ID:                name,
		Name:              name,
		Port:              port,
		EnableTagOverride: false,
		Tags:              tags,
		Address:           localAddr.IP.String(),
	}
	err = r.client.Agent().ServiceRegister(asr)
	if err != nil {
		logrus.Errorf("Failed to register service at '%s'. error: %v", r.addr, err)
	} else {
		logrus.Infof("Regsitered service '%s' at consul.", id)
	}
	return err
}

func (r consul) DeRegister(name string) error {
	err := r.client.Agent().ServiceDeregister(name)

	if err != nil {
		logrus.Errorf("Failed to deregister service by id: '%s'. Error: %v", name, err)
	} else {
		logrus.Infof("Deregistered service '%s' at consul.", name)
	}
	return err
}
