package grpc

import (
	"context"
	"fmt"
	"net"

	//"go.uber.org/zap"
	//"go.uber.org/zap/zapcore"

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

type TestPorterGRPCService struct {
	Porter *porter.Porter
}

const bufSize = 1024 * 1024

var lis *bufconn.Listener

func NewServer() (*TestPorterGRPCService, error) {
	srv := &TestPorterGRPCService{}
	return srv, nil
}

func init() {
	grpcSvc, _ := NewServer()
	grpcSvc.ListenAndServe()
}

func newTestInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	p, err := pCtx.GetPorterConnectionFromContext(ctx)
	if err != nil {
		return nil, err
	}
	p.Config.SetPorterPath("porter")
	ctx = pCtx.AddPorterConnectionToContext(p, ctx)
	h, err := handler(ctx, req)
	return h, err
}

func (s *TestPorterGRPCService) ListenAndServe() *grpc.Server {
	lis = bufconn.Listen(bufSize)

	srv := grpc.NewServer(
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
			pCtx.NewConnectionInterceptor,
			newTestInterceptor,
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
