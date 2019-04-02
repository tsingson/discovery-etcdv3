package etcdv3

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/tsingson/zaplogger"

	etcd3 "github.com/coreos/etcd/clientv3"
)

const (
	_registerGap = 30 * time.Second
	_appid       = "infra.discovery"
)

type (
	Discovery struct {
		Endpoints  []string
		client     *etcd3.Client
		ctx        context.Context
		cancelFunc context.CancelFunc

		prefix     string
		serviceKey string
		interval   time.Duration
		ttl        int
	}
)

func New(prefix string, endPoints []string, interval time.Duration, ttl int) *Discovery {

	ctx, cancel := context.WithCancel(context.Background())
	return &Discovery{
		Endpoints:  endPoints,
		ctx:        ctx,
		cancelFunc: cancel,

		prefix: prefix,

		interval: interval,
		ttl:      ttl,
	}
}

var log = zaplogger.NewDevelopment()

// Prefix should start and end with no slash
const Prefix = "etcdv3_naming"

// Deregister  un register

// Register
func Register(name, host, port string, target string, interval time.Duration, ttl int) (cancelFunc context.CancelFunc, err error) {

	serviceValue := net.JoinHostPort(host, port)
	var n = New(Prefix, []string{target}, interval, ttl)
	return n.register(name, serviceValue)
}

func (n *Discovery) register(key, serviceValue string) (cancelFunc context.CancelFunc, err error) {

	serviceKey := fmt.Sprintf("/%s/%s/%s", n.prefix, key, serviceValue)

	ctx, cancel := context.WithCancel(n.ctx)

	// get endpoints for register dial address
	n.client, err = etcd3.New(etcd3.Config{
		Endpoints: n.Endpoints, // strings.Split(target, ","),
	})
	if err != nil {
		cancel()
		return // xerrors.Errorf("grpclb: create etcd3 client failed: %v", err)
	}

	ch := make(chan struct{}, 1)

	var resp *etcd3.LeaseGrantResponse

	resp, err = n.client.Grant(context.TODO(), int64(n.ttl))
	if err != nil {

		n.client.Close()
		cancel()
		<-ch
		return // xerrors.Errorf("grpclb: create etcd3 lease failed: %v", err)
	}

	if _, err = n.client.Put(context.TODO(), serviceKey, serviceValue, etcd3.WithLease(resp.ID)); err != nil {
		n.client.Close()
		cancel()
		<-ch
		return //  xerrors.Errorf("grpclb: set service '%s' with ttl to etcd3 failed: %s", key, err.Error())
	}

	if _, err = n.client.KeepAlive(context.TODO(), resp.ID); err != nil {
		n.client.Close()
		cancel()
		<-ch
		return // xerrors.Errorf("grpclb: refresh service '%s' with ttl to etcd3 failed: %s", key, err.Error())
	}

	// wait deregister then delete
	go func() {
		for {
			select {

			case <-ctx.Done():
				n.client.Delete(context.Background(), serviceKey)

				n.client.Close()

				ch <- struct{}{}
			}
		}

	}()

	return
}

// UnRegister delete registered service from etcd
func (n *Discovery) UnRegister() {
	n.client.Delete(context.Background(), n.serviceKey)

	n.client.Close()

	n.cancelFunc()
}

func registerTx(endpoints []string, dialTimeout, requestTimeout time.Duration) {
	cli, err := etcd3.New(etcd3.Config{
		Endpoints:   endpoints,
		DialTimeout: dialTimeout,
	})
	if err != nil {
		log.Fatal(err)
	}
	defer cli.Close()

	kvc := etcd3.NewKV(cli)

	_, err = kvc.Put(context.TODO(), "key", "xyz")
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	_, err = kvc.Txn(ctx).
		// txn value comparisons are lexical
		If(etcd3.Compare(etcd3.Value("key"), ">", "abc")).
		// the "Then" runs, since "xyz" > "abc"
		Then(etcd3.OpPut("key", "XYZ")).
		// the "Else" does not run
		Else(etcd3.OpPut("key", "ABC")).
		Commit()
	cancel()
	if err != nil {
		log.Fatal(err)
	}

	gresp, err := kvc.Get(context.TODO(), "key")
	cancel()
	if err != nil {
		log.Fatal(err)
	}
	for _, ev := range gresp.Kvs {
		fmt.Printf("%s : %s\n", ev.Key, ev.Value)
	}

}