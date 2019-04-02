package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"google.golang.org/grpc"

	discovery "github.com/tsingson/discovery-etcdv3/etcdv3"
	pb "github.com/tsingson/discovery-etcdv3/etcdv3/helloworld"
)

var (
	serv = flag.String("service", "goim.comet", "service name")
	host = flag.String("host", "localhost", "listening host")
	port = flag.String("port", "50001", "listening port")
	reg  = flag.String("reg", "http://10.0.0.11:2379", "register etcd address")
)

func main() {
	flag.Parse()

	lis, err := net.Listen("tcp", net.JoinHostPort(*host, *port))
	if err != nil {
		panic(err)
	}

	cancel, err := discovery.Register(*serv, *host, *port, *reg, time.Second*10, 15)
	if err != nil {
		panic(err)
	}

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGTERM, syscall.SIGINT, syscall.SIGKILL, syscall.SIGHUP, syscall.SIGQUIT)

	go func() {
		s := <-signalCh
		log.Printf("receive signal '%v'", s)
		if cancel != nil {
			cancel()
		}
		os.Exit(1)
	}()

	log.Printf("starting hello service at %grpcServ", *port)

	grpcServ := grpc.NewServer()

	pb.RegisterGreeterServer(grpcServ, &server{})
	err = grpcServ.Serve(lis)
	if err != nil {
		// TODO: handle errors
	}
}

// server is used to implement helloworld.GreeterServer.
type server struct{}

// SayHello implements helloworld.GreeterServer
func (s *server) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	fmt.Printf("%v: Receive is %s\n", time.Now(), in.Name)
	return &pb.HelloReply{Message: "Hello " + in.Name + " from " + net.JoinHostPort(*host, *port)}, nil
}
