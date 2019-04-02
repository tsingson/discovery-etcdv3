package main

import (
	"context"
	"flag"
	"fmt"
	"runtime"
	"strconv"
	"time"

	"google.golang.org/grpc"

	discovery "github.com/tsingson/discovery-etcdv3/etcdv3"
	pb "github.com/tsingson/discovery-etcdv3/etcdv3/helloworld"
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

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	conn, err := grpc.DialContext(ctx, etcdAddr, grpc.WithInsecure(), grpc.WithBalancer(balancer), grpc.WithBlock())
	cancel()
	if err != nil {
		panic(err)
	}
	ack := "ack"

	ticker := time.NewTicker(interval)
	for {
		select {
		case <-ticker.C:
			t := time.Now()
			client := pb.NewGreeterClient(conn)

			resp, err := client.SayHello(context.Background(), &pb.HelloRequest{Name: ack + strconv.Itoa(t.Second())})
			if err == nil {
				fmt.Printf("%v: Reply is %s\n", t, resp.Message)
			}
			// default:
			// 	t := time.Now()
			// 	fmt.Printf("%v: Reply is %s\n", t, " ------>")

		}
	}
}
