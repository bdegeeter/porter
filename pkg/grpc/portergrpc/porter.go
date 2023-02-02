package portergrpc

import pGRPC "get.porter.sh/porter/gen/proto/go/porterapis/porter/v1alpha1"

type PorterServer struct {
	pGRPC.UnimplementedPorterServer
}

func NewPorterServer() (*PorterServer, error) {
	return &PorterServer{}, nil
}
