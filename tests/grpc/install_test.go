package grpc

import (
	"context"
	//"fmt"
	"testing"

	igrpc "get.porter.sh/porter/gen/proto/go/porterapis/installation/v1alpha1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

func TestInstall_installationMessage(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(bufDialer), grpc.WithInsecure())
	if err != nil {
		t.Fatalf("failed to dial bufnet: %v", err)
	}
	defer conn.Close()
	instClient := igrpc.NewInstallationsClient(conn)
	resp, err := instClient.ListInstallations(context.Background(), &igrpc.ListInstallationsRequest{})
	require.NoError(t, err)
	assert.Equal(t, resp.String(), `installation:{name:"test installation"  namespace:"foo"  bundle:{repository:"test.repo"  version:"v1.0.0"}}`)
}
