package helloworld

import (
	"context"
	hw "get.porter.sh/porter/gen/proto/go/helloworld/v1alpha"
	"log"
)

// server is used to implement helloworld.GreeterServer.
type HelloWorldServer struct {
	hw.UnimplementedGreeterServer
}

// SayHello implements helloworld.GreeterServer
func (s *HelloWorldServer) SayHello(ctx context.Context, in *hw.HelloRequest) (*hw.HelloReply, error) {
	log.Printf("Received: %v", in.GetName())
	return &hw.HelloReply{Message: "Hello " + in.GetName()}, nil
}
func (s *HelloWorldServer) SayHelloAgain(ctx context.Context, in *hw.HelloRequest) (*hw.HelloReply, error) {
	log.Printf("In SayHelloAgain, received: %v", in.GetName())
	return &hw.HelloReply{Message: "Hello again " + in.GetName()}, nil
}
