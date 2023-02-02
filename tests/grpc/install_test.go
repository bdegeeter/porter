package grpc

import (
	"bytes"
	"encoding/json"
	"testing"

	iGRPC "get.porter.sh/porter/gen/proto/go/porterapis/installation/v1alpha1"
	pGRPC "get.porter.sh/porter/gen/proto/go/porterapis/porter/v1alpha1"
	"get.porter.sh/porter/pkg/cnab"
	"get.porter.sh/porter/pkg/porter"
	"get.porter.sh/porter/pkg/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/encoding/protojson"
)

func TestInstall_installationMessage(t *testing.T) {
	t.Parallel()
	grpcSvr, _ := NewTestGRPCServer(t)
	i1 := storage.NewInstallation("", "test")
	storeInst := grpcSvr.TestPorter.TestInstallations.CreateInstallation(i1, grpcSvr.TestPorter.TestInstallations.SetMutableInstallationValues, func(i *storage.Installation) {
		i.Status.BundleVersion = "v0.1.0"
		i.Status.ResultStatus = cnab.StatusSucceeded
		i.Bundle.Repository = "test-bundle"
		i.Bundle.Version = "v0.1.0"
	})
	expInst := porter.NewDisplayInstallation(storeInst)

	ctx := grpcSvr.TestPorter.SetupIntegrationTest()
	//ctx := context.Background()
	grpcSvr.TestPorter.AddTestBundleDir("../integration/testdata/bundles/bundle-with-custom-action", true)
	grpcSvr.ListenAndServe()

	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(bufDialer), grpc.WithInsecure())
	if err != nil {
		t.Fatalf("failed to dial bufnet: %v", err)
	}
	defer conn.Close()
	instClient := pGRPC.NewPorterClient(conn)
	resp, err := instClient.ListInstallations(ctx, &iGRPC.ListInstallationsRequest{})
	require.NoError(t, err)
	assert.Len(t, resp.Installation, 1)
	assert.Equal(t, resp.Installation[0].Name, expInst.Name)

	bExpInst, err := json.MarshalIndent(expInst, "", "  ")
	require.NoError(t, err)
	//fmt.Printf("PORTER INSTALLATION:\n%s\n", string(bExpInst))
	pjm := protojson.MarshalOptions{EmitUnpopulated: true}
	bActInst, err := pjm.Marshal(resp.GetInstallation()[0])
	var pJson bytes.Buffer
	json.Indent(&pJson, bActInst, "", "  ")
	//fmt.Printf("GRPC INSTALLATION:\n%s\n", string(pJson.Bytes()))
	require.NoError(t, err)
	assert.JSONEq(t, string(bExpInst), string(bActInst))
}
