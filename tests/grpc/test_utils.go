package grpc

import (
	"context"
	"fmt"
	"net"

	//"go.uber.org/zap"
	//"go.uber.org/zap/zapcore"

	pGRPC "get.porter.sh/porter/gen/proto/go/porterapis/porter/v1alpha1"
	"get.porter.sh/porter/pkg/grpc/installation"
	"get.porter.sh/porter/pkg/porter"

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

func (s *TestPorterGRPCService) ListenAndServe() *grpc.Server {
	lis = bufconn.Listen(bufSize)

	srv := grpc.NewServer()
	healthServer := health.NewServer()
	reflection.Register(srv)
	grpc_health_v1.RegisterHealthServer(srv, healthServer)
	pSvc, err := installation.NewPorterService()
	if err != nil {
		panic(err)
	}
	pGRPC.RegisterPorterBundleServer(srv, pSvc)
	healthServer.SetServingStatus("test-health", grpc_health_v1.HealthCheckResponse_SERVING)
	// Setup the storage plugin
	// testP := porter.New()
	// var cfg interface{}
	// plug, err := mongodb_docker.NewPlugin(testP.Context, cfg)
	// if err != nil {
	// 	panic(err)
	// }
	// go func() {
	// 	plugins.Serve(testP.Context, storageplugins.PluginInterface, plug, storageplugins.PluginProtocolVersion)
	// }()
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
