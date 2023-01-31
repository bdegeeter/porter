package grpc

import (
	"context"
	//"fmt"
	"testing"

	iGRPC "get.porter.sh/porter/gen/proto/go/porterapis/installation/v1alpha1"
	pGRPC "get.porter.sh/porter/gen/proto/go/porterapis/porter/v1alpha1"
	"get.porter.sh/porter/pkg/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

func TestInstall_installationMessage(t *testing.T) {
	t.Parallel()
	grpcSvr, _ := NewTestGRPCServer(t)
	i1 := storage.NewInstallation("", "test")
	grpcSvr.TestPorter.TestInstallations.CreateInstallation(i1)
	ctx := grpcSvr.TestPorter.SetupIntegrationTest()
	grpcSvr.TestPorter.AddTestBundleDir("../integration/testdata/bundles/bundle-with-custom-action", true)
	grpcSvr.ListenAndServe()

	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(bufDialer), grpc.WithInsecure())
	if err != nil {
		t.Fatalf("failed to dial bufnet: %v", err)
	}
	defer conn.Close()
	instClient := pGRPC.NewPorterBundleClient(conn)
	resp, err := instClient.ListInstallations(context.Background(), &iGRPC.ListInstallationsRequest{})
	require.NoError(t, err)
	//assert.Equal(t, `installation:{name:"test installation" namespace:"foo" bundle:{repository:"test.repo" version:"v1.0.0"}}`, resp.String())
	assert.Len(t, resp.Installation, 1)
	assert.Equal(t, resp.Installation[0].Name, "test")
}
