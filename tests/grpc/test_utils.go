package grpc

import (
	"context"
	"fmt"
	"net"
	"testing"

	pGRPC "get.porter.sh/porter/gen/proto/go/porterapis/porter/v1alpha1"
	pCtx "get.porter.sh/porter/pkg/grpc/context"
	"get.porter.sh/porter/pkg/grpc/installation"
	"get.porter.sh/porter/pkg/porter"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"

	//"go.opentelemetry.io/otel/attribute"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/test/bufconn"
)

type TestPorterGRPCServer struct {
	TestPorter *porter.TestPorter
	t          *testing.T
}

const bufSize = 1024 * 1024

var lis *bufconn.Listener

func NewTestGRPCServer(t *testing.T) (*TestPorterGRPCServer, error) {
	srv := &TestPorterGRPCServer{
		TestPorter: porter.NewTestPorter(t),
	}
	return srv, nil
}

func (s *TestPorterGRPCServer) newTestInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	ctx = pCtx.AddPorterConnectionToContext(s.TestPorter.Porter, ctx)
	h, err := handler(ctx, req)
	return h, err
}

func (s *TestPorterGRPCServer) ListenAndServe() *grpc.Server {
	lis = bufconn.Listen(bufSize)

	srv := grpc.NewServer(
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
			s.newTestInterceptor,
		)))

	healthServer := health.NewServer()
	reflection.Register(srv)
	grpc_health_v1.RegisterHealthServer(srv, healthServer)
	pSvc, err := installation.NewPorterService()
	if err != nil {
		panic(err)
	}
	//pSvc.Porter.Config.SetPorterPath("porter")
	pGRPC.RegisterPorterBundleServer(srv, pSvc)
	healthServer.SetServingStatus("test-health", grpc_health_v1.HealthCheckResponse_SERVING)

	go func() {
		if err := srv.Serve(lis); err != nil {
			fmt.Println("failed to serve")
			//s.log.Fatal("failed to serve", zap.Error(err))
		}
	}()
	return srv
}

func bufDialer(context.Context, string) (net.Conn, error) {
	return lis.Dial()
}
