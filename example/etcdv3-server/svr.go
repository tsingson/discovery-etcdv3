package main

import (
	"context"
	"flag"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	log "github.com/tsingson/zaplogger"
	"google.golang.org/grpc"

	"github.com/tsingson/discovery-etcdv3/example/proto"
	"github.com/tsingson/discovery-etcdv3/naming"
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

	cancel, err := naming.Register(*serv, *host, *port, *reg, time.Second*10, 15)
	if err != nil {

		os.Exit(-1)
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

	// log.Printf("starting hello service at %grpcServ", *port)

	grpcServ := grpc.NewServer()

	proto.RegisterGreeterServer(grpcServ, &server{})
	err = grpcServ.Serve(lis)
	if err != nil {
		// TODO: handle errors
	}
}

// server is used to implement helloworld.GreeterServer.
type server struct{}

// SayHello implements helloworld.GreeterServer
func (s *server) SayHello(ctx context.Context, in *proto.HelloRequest) (*proto.HelloReply, error) {
	log.Info(in.Name)
	return &proto.HelloReply{Message: "Hello " + in.Name + " from " + net.JoinHostPort(*host, *port)}, nil
}
