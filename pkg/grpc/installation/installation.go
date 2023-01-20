package installation

import (
	"context"
	"fmt"

	iGRPC "get.porter.sh/porter/gen/proto/go/porterapis/installation/v1alpha1"
	pGRPC "get.porter.sh/porter/gen/proto/go/porterapis/porter/v1alpha1"
	"get.porter.sh/porter/pkg/porter"
	"get.porter.sh/porter/pkg/tracing"
)

// server is used to implement helloworld.GreeterServer.
type PorterServer struct {
	Porter *porter.Porter
	pGRPC.UnimplementedPorterBundleServer
}

func NewPorterService(p *porter.Porter) (*PorterServer, error) {
	return &PorterServer{Porter: p}, nil
}

func (s *PorterServer) ListInstallations(ctx context.Context, in *iGRPC.ListInstallationsRequest) (*iGRPC.ListInstallationsResponse, error) {
	fmt.Println("IN LIST INSTALLATIONS")
	ctx, log := tracing.StartSpan(ctx)
	defer log.EndSpan()

	opts := porter.ListOptions{}
	fmt.Println("LISTING INSTALLATIONS")
	installations, err := s.Porter.ListInstallations(ctx, opts)
	if err != nil {
		return nil, err
	}
	insts := []*iGRPC.Installation{}
	for _, pInst := range installations {
		inst := iGRPC.Installation{
			Name:      pInst.Name,
			Namespace: pInst.Namespace,
			Bundle: &iGRPC.Bundle{
				Repository: pInst.Bundle.Repository,
				Version:    pInst.Bundle.Version,
			},
			//TODO: figure this out
			State:  iGRPC.InstallationState_INSTALLED,
			Status: iGRPC.InstallationStatus_SUCCEEDED,
		}
		insts = append(insts, &inst)

	}
	res := iGRPC.ListInstallationsResponse{
		Installation: insts,
	}
	return &res, nil
}
