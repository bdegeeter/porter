package signals

import (
	"context"
	"time"

	"github.com/spf13/viper"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type Shutdown struct {
	logger                *zap.Logger
	tracerProvider        *sdktrace.TracerProvider
	serverShutdownTimeout time.Duration
}

func NewShutdown(serverShutdownTimeout time.Duration, logger *zap.Logger) (*Shutdown, error) {
	srv := &Shutdown{
		logger:                logger,
		serverShutdownTimeout: serverShutdownTimeout,
	}

	return srv, nil
}

func (s *Shutdown) Graceful(stopCh <-chan struct{}, grpcServer *grpc.Server) {
	ctx := context.Background()

	// wait for SIGTERM or SIGINT
	<-stopCh
	ctx, cancel := context.WithTimeout(ctx, s.serverShutdownTimeout)
	defer cancel()

	// wait for Kubernetes readiness probe to remove this instance from the load balancer
	// the readiness check interval must be lower than the timeout
	if viper.GetString("level") != "debug" {
		time.Sleep(3 * time.Second)
	}

	// stop OpenTelemetry tracer provider
	if s.tracerProvider != nil {
		if err := s.tracerProvider.Shutdown(ctx); err != nil {
			s.logger.Warn("stopping tracer provider", zap.Error(err))
		}
	}

	// determine if the GRPC was started
	if grpcServer != nil {
		s.logger.Info("Shutting down GRPC server", zap.Duration("timeout", s.serverShutdownTimeout))
		grpcServer.GracefulStop()
	}

}
