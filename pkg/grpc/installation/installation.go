package installation

import (
	"context"

	igrpc "get.porter.sh/porter/gen/proto/go/porterapis/installation/v1alpha1"
)

// server is used to implement helloworld.GreeterServer.
type InstallationServer struct {
	igrpc.UnimplementedInstallationsServer
}

// SayHello implements helloworld.GreeterServer
func (s *InstallationServer) ListInstallations(ctx context.Context, in *igrpc.ListInstallationsRequest) (*igrpc.ListInstallationsResponse, error) {
	return &igrpc.ListInstallationsResponse{}, nil
}
