package main

import (
	//"context"
	"fmt"
	"log"
	"net"

	"go.uber.org/zap"
	//"go.uber.org/zap/zapcore"

	igrpc "get.porter.sh/porter/gen/proto/go/porterapis/installation/v1alpha1"
	"get.porter.sh/porter/pkg/grpc/installation"
	"get.porter.sh/porter/pkg/porter"

	//"go.opentelemetry.io/otel/attribute"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

type PorterGRPCService struct {
	Porter *porter.Porter
	config *Config
	log    *zap.Logger
}

type Config struct {
	Port        int    `mapstructure:"grpc-port"`
	ServiceName string `mapstructure:"grpc-service-name"`
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	logger, err := zap.NewDevelopmentConfig().Build()
	if err != nil {
		return fmt.Errorf("Unable to build logger")
	}
	cfg := Config{
		Port:        8080,
		ServiceName: "local",
	}
	srv, err := NewServer(&cfg, logger)
	if err != nil {
		return fmt.Errorf("Unable to start server")
	}
	srv.ListenAndServe()
	return nil
}

func NewServer(config *Config, logger *zap.Logger) (*PorterGRPCService, error) {
	srv := &PorterGRPCService{
		config: config,
		log:    logger,
	}

	return srv, nil
}

func (s *PorterGRPCService) ListenAndServe() *grpc.Server {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%v", s.config.Port))
	if err != nil {
		fmt.Println("failed to listen")
		s.log.Fatal("failed to listen", zap.Int("port", s.config.Port))
	}
	s.log.Info(fmt.Sprintf("Serving on 0.0.0.0:%v", s.config.Port))
	srv := grpc.NewServer()
	healthServer := health.NewServer()
	reflection.Register(srv)
	grpc_health_v1.RegisterHealthServer(srv, healthServer)
	isrv := &installation.InstallationServer{}
	igrpc.RegisterInstallationsServer(srv, isrv)
	healthServer.SetServingStatus(s.config.ServiceName, grpc_health_v1.HealthCheckResponse_SERVING)
	if err := srv.Serve(listener); err != nil {
		fmt.Println("failed to serve")
		s.log.Fatal("failed to serve", zap.Error(err))
	}
	return srv
}
