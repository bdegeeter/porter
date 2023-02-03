package grpc

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"testing"

	iGRPC "get.porter.sh/porter/gen/proto/go/porterapis/installation/v1alpha1"
	pGRPC "get.porter.sh/porter/gen/proto/go/porterapis/porter/v1alpha1"
	"get.porter.sh/porter/pkg/cnab"
	"get.porter.sh/porter/pkg/porter"
	"get.porter.sh/porter/pkg/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/encoding/protojson"
)

func grpcInstallationExpectedJSON(inst storage.Installation) ([]byte, error) {
	expInst := porter.NewDisplayInstallation(inst)
	bExpInst, err := json.MarshalIndent(expInst, "", "  ")
	if err != nil {
		return nil, err
	}
	// if no credentialSets or parameterSets add as empty list to
	// match GRPCInstallation expectations
	empty := make([]string, 0)
	emptySets := []string{"credentialSets", "parameterSets"}
	for _, es := range emptySets {
		res := gjson.GetBytes(bExpInst, es)
		if !res.Exists() {
			bExpInst, err = sjson.SetBytes(bExpInst, es, empty)
			if err != nil {
				return nil, err
			}
		}
	}
	return bExpInst, nil
}

// TODO: add opts structure for different installation options
func newTestInstallation(t *testing.T, namespace, name string, grpcSvr *TestPorterGRPCServer) storage.Installation {
	i1 := storage.NewInstallation(namespace, name)
	storeInst := grpcSvr.TestPorter.TestInstallations.CreateInstallation(i1, grpcSvr.TestPorter.TestInstallations.SetMutableInstallationValues, func(i *storage.Installation) {
		i.Status.BundleVersion = "v0.1.0"
		i.Status.ResultStatus = cnab.StatusSucceeded
		i.Bundle.Repository = "test-bundle"
		i.Bundle.Version = "v0.1.0"
	})
	return storeInst
}

func TestInstall_installationMessage(t *testing.T) {
	tests := []struct {
		testName      string
		instName      string
		instNamespace string ""
	}{
		{
			testName: "basic installation",
			instName: "test",
		},
		{
			testName: "another installation",
			instName: "another-test",
		},
	}
	//t.Parallel()
	for _, test := range tests {
		t.Run(fmt.Sprintf("%s", test.testName), func(t *testing.T) {
			//Server setup
			grpcSvr, err := NewTestGRPCServer(t)
			require.NoError(t, err)
			server := grpcSvr.ListenAndServe()
			defer server.Stop()

			//Client setup
			ctx := context.TODO()
			client, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(bufDialer), grpc.WithTransportCredentials(insecure.NewCredentials()))
			require.NoError(t, err)
			defer client.Close()
			instClient := pGRPC.NewPorterClient(client)

			inst := newTestInstallation(t, test.instNamespace, test.instName, grpcSvr)

			//Call ListInstallations
			resp, err := instClient.ListInstallations(ctx, &iGRPC.ListInstallationsRequest{})
			require.NoError(t, err)
			assert.Len(t, resp.Installation, 1)

			// Validation
			validateInstallations(t, inst, resp.GetInstallation()[0])
		})
	}
}

func validateInstallations(t *testing.T, expected storage.Installation, actual *iGRPC.Installation) {
	assert.Equal(t, actual.Name, expected.Name)
	bExpInst, err := grpcInstallationExpectedJSON(expected)
	require.NoError(t, err)
	pjm := protojson.MarshalOptions{EmitUnpopulated: true}
	bActInst, err := pjm.Marshal(actual)
	var pJson bytes.Buffer
	json.Indent(&pJson, bActInst, "", "  ")
	if false {
		fmt.Printf("PORTER INSTALLATION:\n%s\n", string(bExpInst))
		fmt.Printf("GRPC INSTALLATION:\n%s\n", string(pJson.Bytes()))
	}
	require.NoError(t, err)
	assert.JSONEq(t, string(bExpInst), string(bActInst))
}
