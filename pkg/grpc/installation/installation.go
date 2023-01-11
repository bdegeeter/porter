package installation

import (
	"context"
	"log"

	igrpc "get.porter.sh/porter/gen/proto/go/porterapis/installation/v1alpha1"
)

// server is used to implement helloworld.GreeterServer.
type InstallationServer struct {
	igrpc.UnimplementedInstallationsServer
}

// SayHello implements helloworld.GreeterServer
func (s *InstallationServer) ListInstallations(ctx context.Context, in *igrpc.ListInstallationsRequest) (*igrpc.ListInstallationsResponse, error) {
	log.Printf("IN LIST INSTALLATIONS")
	inst := igrpc.Installation{
		Name:      "test installation",
		Namespace: "foo",
		Bundle: &igrpc.Bundle{
			Repository: "test.repo",
			Version:    "v1.0.0",
		},
		State:  igrpc.InstallationState_INSTALLED,
		Status: igrpc.InstallationStatus_SUCCEEDED,
	}
	insts := []*igrpc.Installation{&inst}
	res := igrpc.ListInstallationsResponse{
		Installation: insts,
	}
	return &res, nil
}
