package main

import (
	"context"
	"flag"
	"runtime"
	"strconv"
	"time"

	log "github.com/tsingson/zaplogger"
	"google.golang.org/grpc"

	discovery "github.com/tsingson/discovery-etcdv3/etcdv3"
	"github.com/tsingson/discovery-etcdv3/example/helloworld"
)

var (
	serv = flag.String("service", "goim.comet", "service name")
	reg  = flag.String("reg", "http://localhost:2379", "register etcd address")
)

func main() {
	runtime.MemProfileRate = 0
	runtime.GOMAXPROCS(128)
	signal := make(chan struct{})

	flag.Parse()

	go watch(*serv, *reg, 5*time.Millisecond, 10*time.Second)
	<-signal
}

func watch(discoveryKey, etcdAddr string, interval, timeout time.Duration) {
	// timeout := 10*time.Second
	resolver := discovery.NewResolver(discoveryKey)
	balancer := grpc.RoundRobin(resolver)

	var conn *grpc.ClientConn
	var err error

	for {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		conn, err = grpc.DialContext(ctx, etcdAddr, grpc.WithInsecure(), grpc.WithBalancer(balancer), grpc.WithBlock())
		cancel()
		if err == nil {
			break
		}
	}

	ack := "ack"

	ticker := time.NewTicker(interval)
	for {
		select {
		case <-ticker.C:
			t := time.Now()
			client := helloworld.NewGreeterClient(conn)

			resp, err := client.SayHello(context.Background(), &helloworld.HelloRequest{Name: ack + strconv.Itoa(t.Second())})
			if err == nil {
				log.Info(resp.Message)
			}
			// default:
			// 	t := time.Now()
			// 	fmt.Printf("%v: Reply is %s\n", t, " ------>")

		}
	}
}
