package conf

import (
	"github.com/BurntSushi/toml"
	"github.com/tsingson/discovery/lib/http"
	"github.com/tsingson/discovery/model"
)

// Config config.
type Config struct {
	Nodes      []string
	Zones      map[string][]string
	HTTPServer *ServerConfig
	HTTPClient *http.ClientConfig
	Env        *Env
	// Scheduler  []byte
	Schedulers map[string]*model.Scheduler
}

type DiscoveryConfig = Config

// Env is disocvery env.
type Env struct {
	Region    string
	Zone      string
	Host      string
	DeployEnv string
}

// ServerConfig Http Servers conf.
type ServerConfig struct {
	Addr string
}

var (
	confPath      string
	schedulerPath string
	region        string
	zone          string
	deployEnv     string
	hostname      string
	// Conf conf
	Conf = &Config{}
)

// LoadConfig load config from file toml
func LoadConfig(fh string) (*Config, error) {
	_, err := toml.DecodeFile(fh, &Conf)
	return Conf, err

}
