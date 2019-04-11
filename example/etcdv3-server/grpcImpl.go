package main

import (
	"context"
	"net"

	log "github.com/tsingson/zaplogger"

	"github.com/tsingson/discovery-etcdv3/example/proto"
)

// server is used to implement helloworld.GreeterServer.
type server struct{}

// SayHello implements helloworld.GreeterServer
func (s *server) SayHello(ctx context.Context, in *proto.HelloRequest) (*proto.HelloReply, error) {
	log.Info(in.Name)
	return &proto.HelloReply{Message: "Hello " + in.Name + " from " + net.JoinHostPort(*host, *port)}, nil
}
