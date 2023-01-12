package installation

import (
	"context"
	"log"

	iGRPC "get.porter.sh/porter/gen/proto/go/porterapis/installation/v1alpha1"
	pGRPC "get.porter.sh/porter/gen/proto/go/porterapis/porter/v1alpha1"
)

// server is used to implement helloworld.GreeterServer.
type PorterServer struct {
	pGRPC.UnimplementedPorterBundleServer
}

// SayHello implements helloworld.GreeterServer
func (s *PorterServer) ListInstallations(ctx context.Context, in *iGRPC.ListInstallationsRequest) (*iGRPC.ListInstallationsResponse, error) {
	log.Printf("IN LIST INSTALLATIONS")
	inst := iGRPC.Installation{
		Name:      "test installation",
		Namespace: "foo",
		Bundle: &iGRPC.Bundle{
			Repository: "test.repo",
			Version:    "v1.0.0",
		},
		State:  iGRPC.InstallationState_INSTALLED,
		Status: iGRPC.InstallationStatus_SUCCEEDED,
	}
	insts := []*iGRPC.Installation{&inst}
	res := iGRPC.ListInstallationsResponse{
		Installation: insts,
	}
	return &res, nil
}
