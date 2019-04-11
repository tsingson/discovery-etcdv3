package main

import (
	"context"
	"flag"
	"runtime"
	"strconv"
	"time"

	log "github.com/tsingson/zaplogger"
	"google.golang.org/grpc"

	"github.com/tsingson/discovery-etcdv3/example/proto"
	"github.com/tsingson/discovery-etcdv3/grpc-etcdv3/resolver"
)

var (
	serv = flag.String("service", "goim.comet", "service name")
	reg  = flag.String("reg", "http://localhost:2379", "register etcd address")
)

func main() {
	runtime.MemProfileRate = 0
	runtime.GOMAXPROCS(128)
	noExit := make(chan struct{})

	flag.Parse()

	var ServerName string = *serv
	var EtcdServerAddr string = *reg

	go watch(ServerName, EtcdServerAddr, 5*time.Millisecond, 10*time.Second)
	<-noExit
}

func watch(serverName, etcdNamingServiceAddr string, interval, timeout time.Duration) {
	// timeout := 10*time.Second

	rl := resolver.NewResolver(serverName)
	bl := grpc.RoundRobin(rl)

	var conn *grpc.ClientConn
	var err error

	for {
		ctx, cancelFunc := context.WithTimeout(context.Background(), timeout)
		conn, err = grpc.DialContext(ctx, etcdNamingServiceAddr, grpc.WithInsecure(), grpc.WithBalancer(bl), grpc.WithBlock())
		cancelFunc()
		if err == nil {
			break
		}
	}

	var ack = "ack"
	var ticker = time.NewTicker(interval)
	for {
		select {
		case <-ticker.C:
			t := time.Now()
			client := proto.NewGreeterClient(conn)

			resp, err := client.SayHello(context.Background(), &proto.HelloRequest{Name: ack + strconv.Itoa(t.Second())})
			if err == nil {
				log.Info(resp.Message)
			}
			// default:
			// 	t := time.Now()
			// 	fmt.Printf("%v: Reply is %s\n", t, " ------>")

		}
	}
}
