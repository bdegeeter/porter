package installation

import (
	"context"

	iGRPC "get.porter.sh/porter/gen/proto/go/porterapis/installation/v1alpha1"
	pGRPC "get.porter.sh/porter/gen/proto/go/porterapis/porter/v1alpha1"
	pCtx "get.porter.sh/porter/pkg/grpc/context"
	"get.porter.sh/porter/pkg/porter"
	"get.porter.sh/porter/pkg/tracing"
)

// server is used to implement helloworld.GreeterServer.
type PorterServer struct {
	pGRPC.UnimplementedPorterBundleServer
}

func NewPorterService() (*PorterServer, error) {
	return &PorterServer{}, nil
}

func (s *PorterServer) ListInstallations(ctx context.Context, in *iGRPC.ListInstallationsRequest) (*iGRPC.ListInstallationsResponse, error) {
	ctx, log := tracing.StartSpan(ctx)
	defer log.EndSpan()
	p, err := pCtx.GetPorterConnectionFromContext(ctx)
	if err != nil {
		return nil, err
	}
	opts := porter.ListOptions{}
	installations, err := p.ListInstallations(ctx, opts)
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
			// State:  iGRPC.InstallationState_INSTALLED,
			// Status: iGRPC.InstallationStatus_SUCCEEDED,
		}
		insts = append(insts, &inst)

	}
	res := iGRPC.ListInstallationsResponse{
		Installation: insts,
	}
	return &res, nil
}
