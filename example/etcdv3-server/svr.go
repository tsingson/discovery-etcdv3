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
	discovery "github.com/tsingson/discovery-etcdv3/grpc-etcdv3/naming"
)

var (
	serv = flag.String("service", "goim.comet", "service name")
	host = flag.String("host", "localhost", "listening host")
	port = flag.String("port", "50001", "listening port")
	reg  = flag.String("reg", "http://10.0.0.11:2379", "register etcd address")
)

func main() {
	flag.Parse()
	noExit := make(chan struct{})
	var err error
	var grpcServerName string = *serv
	var grpcServerHost string = *host
	var grpcServerPort string = *port
	var etcdServerAddr string = *reg
	var cancelFunc context.CancelFunc



	// start server
	{
		var lis net.Listener
		var err error
		lis, err = net.Listen("tcp", net.JoinHostPort(*host, *port))
		if err != nil {
			panic(err)
		}

		// log.Printf("starting hello service at %grpcServ", *port)
		grpcServer := grpc.NewServer()
		proto.RegisterGreeterServer(grpcServer, &server{})
		err = grpcServer.Serve(lis)
		if err != nil {
			// TODO: handle errors
		}
	}
	{ // register GRPC to etcd
		cancelFunc, err = discovery.Register(grpcServerName, grpcServerHost, grpcServerPort, etcdServerAddr, time.Second*10, 15)
		if err != nil {
			os.Exit(-1)
		}
	}
	// control system singal and exit
	{
		signalCh := make(chan os.Signal, 1)
		signal.Notify(signalCh, syscall.SIGTERM, syscall.SIGINT, syscall.SIGKILL, syscall.SIGHUP, syscall.SIGQUIT)

		go func() {
			s := <-signalCh
			log.Printf("receive signal '%v'", s)
			if cancelFunc != nil {
				cancelFunc()
			}
			os.Exit(1)
		}()
	}
	<-noExit

}
