package portergrpc

import (
	"context"
	"errors"

	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/porter"
	"get.porter.sh/porter/pkg/secrets"
	secretsplugin "get.porter.sh/porter/pkg/secrets/pluginstore"
	"get.porter.sh/porter/pkg/storage"
	storageplugin "get.porter.sh/porter/pkg/storage/pluginstore"
	"google.golang.org/grpc"
)

type ctxKey int

const porterConnCtxKey ctxKey = 0

func NewConnectionInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
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

func GetPorterConnectionFromContext(ctx context.Context) (*porter.Porter, error) {
	p, ok := ctx.Value(porterConnCtxKey).(*porter.Porter)
	if !ok {
		return nil, errors.New("Unable to find porter connection in context")
	}
	return p, nil
}

func AddPorterConnectionToContext(p *porter.Porter, ctx context.Context) context.Context {
	return context.WithValue(ctx, porterConnCtxKey, p)
}
