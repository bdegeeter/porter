package context

import (
	"context"
	"errors"

	"get.porter.sh/porter/pkg/porter"
	"google.golang.org/grpc"
)

type ctxKey int

const porterConnCtxKey ctxKey = 0

func NewConnectionInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	p := porter.New()
	err := p.Connect(ctx)
	if err != nil {
		return nil, err
	}
	defer p.Close()
	ctx = context.WithValue(ctx, porterConnCtxKey, p)
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
