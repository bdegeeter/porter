package grpc

import (
	//"context"
	"context"
	"fmt"
	"net"
	"os"

	//"go.uber.org/zap/zapcore"

	//igrpc "get.porter.sh/porter/gen/proto/go/porterapis/installation/v1alpha1"
	pGRPC "get.porter.sh/porter/gen/proto/go/porterapis/porter/v1alpha1"
	"get.porter.sh/porter/pkg/grpc/installation"
	"get.porter.sh/porter/pkg/porter"
	"get.porter.sh/porter/pkg/tracing"

	//"go.opentelemetry.io/otel/attribute"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

type PorterGRPCService struct {
	Porter *porter.Porter
	config *Config
	ctx    context.Context
}

type Config struct {
	Port        int    `mapstructure:"grpc-port"`
	ServiceName string `mapstructure:"grpc-service-name"`
}

func NewServer(ctx context.Context, config *Config) (*PorterGRPCService, error) {
	log := tracing.LoggerFromContext(ctx)
	log.Debug("HELLO")
	p := porter.New()
	srv := &PorterGRPCService{
		Porter: p,
		config: config,
		ctx:    ctx,
	}

	return srv, nil
}

func (s *PorterGRPCService) ListenAndServe() (*grpc.Server, error) {
	ctx, log := tracing.StartSpan(s.ctx)
	fmt.Println("CONNECTING")
	err := s.Porter.Connect(ctx)
	if err != nil {
		return nil, err
	}
	fmt.Println("CONNECTED")
	defer s.Porter.Close()
	defer log.EndSpan()
	log.Infof("Starting gRPC on %v", s.config.Port)
	listener, err := net.Listen("tcp", fmt.Sprintf(":%v", s.config.Port))
	if err != nil {
		return nil, fmt.Errorf("failed to listen on %d: %s", s.config.Port, err)
	}

	srv := grpc.NewServer()
	healthServer := health.NewServer()
	reflection.Register(srv)
	grpc_health_v1.RegisterHealthServer(srv, healthServer)
	isrv, err := installation.NewPorterService(s.Porter)
	if err != nil {
		panic(err)
	}

	pGRPC.RegisterPorterBundleServer(srv, isrv)
	healthServer.SetServingStatus(s.config.ServiceName, grpc_health_v1.HealthCheckResponse_SERVING)

	go func() {
		if err := srv.Serve(listener); err != nil {
			log.Errorf("failed to serve: %s", err)
			os.Exit(1)
		}
	}()

	return srv, nil
}
