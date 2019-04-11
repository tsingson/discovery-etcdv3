package dao

import (
	"github.com/etcd-io/etcd/clientv3"
)

func New(Endpoints string) (*clientv3.Client, error) {
	return clientv3.New(clientv3.Config{
		Endpoints: Endpoints, // strings.Split(endPoints, ","),
	})

}
