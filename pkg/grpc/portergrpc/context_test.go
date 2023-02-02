package portergrpc

import (
	"context"
	"testing"

	"get.porter.sh/porter/pkg/porter"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
)

func TestGetPorterConnectionFromContextReturnsErrIfNoConnectionInContext(t *testing.T) {
	ctx := context.Background()
	p, err := GetPorterConnectionFromContext(ctx)
	assert.Nil(t, p)
	assert.EqualError(t, err, "Unable to find porter connection in context")
}

func TestGetPorterConnectionFromContextReturnsPorterConnection(t *testing.T) {
	p := porter.New()
	ctx := context.Background()
	ctx = context.WithValue(ctx, porterConnCtxKey, p)
	newP, err := GetPorterConnectionFromContext(ctx)
	assert.Nil(t, err)
	assert.Equal(t, p, newP)
}
func TestAddPorterConnectionToContextReturnsContextUpdatedWithPorterConnection(t *testing.T) {
	p := porter.New()
	ctx := context.Background()
	ctx = AddPorterConnectionToContext(p, ctx)
	newP, ok := ctx.Value(porterConnCtxKey).(*porter.Porter)
	assert.True(t, ok)
	assert.Equal(t, p, newP)
}

func TestNewConnectionInterceptorCallsNextHandlerInTheChainWithThePorterConnectionInTheContext(t *testing.T) {
	parentUnaryInfo := &grpc.UnaryServerInfo{FullMethod: "SomeService.StreamMethod"}
	input := "input"
	testHandler := func(ctx context.Context, req interface{}) (interface{}, error) {
		p, err := GetPorterConnectionFromContext(ctx)
		return p, err
	}
	ctx := context.Background()
	chain := grpc_middleware.ChainUnaryServer(NewConnectionInterceptor)
	newP, err := chain(ctx, input, parentUnaryInfo, testHandler)
	assert.Nil(t, err)
	assert.NotNil(t, newP)
	assert.IsType(t, &porter.Porter{}, newP)
}
