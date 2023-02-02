package portergrpc

import (
	"context"

	pGRPC "get.porter.sh/porter/gen/proto/go/porterapis/porter/v1alpha1"
	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/porter"
	"get.porter.sh/porter/pkg/secrets"
	secretsplugin "get.porter.sh/porter/pkg/secrets/pluginstore"
	"get.porter.sh/porter/pkg/storage"
	storageplugin "get.porter.sh/porter/pkg/storage/pluginstore"
	"google.golang.org/grpc"
)

type PorterServer struct {
	pGRPC.UnimplementedPorterServer
}

func NewPorterServer() (*PorterServer, error) {
	return &PorterServer{}, nil
}

func (s *PorterServer) NewConnectionInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	// Load some sane default?
	pCfg := config.New()
	storage := storage.NewPluginAdapter(storageplugin.NewStore(pCfg))
	secretStorage := secrets.NewPluginAdapter(secretsplugin.NewStore(pCfg))

	p := porter.NewFor(pCfg, storage, secretStorage)
	err := p.Connect(ctx)
	if err != nil {
		return nil, err
	}
	defer p.Close()

	ctx = AddPorterConnectionToContext(p, ctx)
	h, err := handler(ctx, req)
	return h, err
}
