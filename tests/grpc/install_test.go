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
	"github.com/cnabio/cnab-go/bundle"
	"github.com/cnabio/cnab-go/bundle/definition"
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
	writeOnly := true
	b := bundle.Bundle{
		Definitions: definition.Definitions{
			"foo": &definition.Schema{
				Type:      "string",
				WriteOnly: &writeOnly,
			},
			"bar": &definition.Schema{
				Type:      "string",
				WriteOnly: &writeOnly,
			},
		},
		Outputs: map[string]bundle.Output{
			"foo": {
				Definition: "foo",
				Path:       "/path/to/foo",
			},
			"bar": {
				Definition: "bar",
				Path:       "/path/to/bar",
			},
		},
	}
	extB := cnab.NewBundle(b)
	storeInst := grpcSvr.TestPorter.TestInstallations.CreateInstallation(storage.NewInstallation(namespace, name), grpcSvr.TestPorter.TestInstallations.SetMutableInstallationValues, func(i *storage.Installation) {
		i.Status.BundleVersion = "v0.1.0"
		i.Status.ResultStatus = cnab.StatusSucceeded
		i.Bundle.Repository = "test-bundle"
		i.Bundle.Version = "v0.1.0"
	})
	c := grpcSvr.TestPorter.TestInstallations.CreateRun(storeInst.NewRun(cnab.ActionInstall), func(sRun *storage.Run) {
		sRun.Bundle = b
		sRun.ParameterOverrides.Parameters = grpcSvr.TestPorter.SanitizeParameters(sRun.ParameterOverrides.Parameters, sRun.ID, extB)
	})
	sRes := grpcSvr.TestPorter.TestInstallations.CreateResult(c.NewResult(cnab.StatusSucceeded))
	grpcSvr.TestPorter.CreateOutput(sRes.NewOutput("foo", []byte("foo-output")), extB)
	grpcSvr.TestPorter.CreateOutput(sRes.NewOutput("bar", []byte("bar-output")), extB)
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

			//Call ListInstallationLatestOutputRequest
			req := &iGRPC.ListInstallationLatestOutputRequest{Name: test.instName, Namespace: &test.instNamespace}
			oresp, err := instClient.ListInstallationLatestOutputs(ctx, req)
			require.NoError(t, err)
			assert.Len(t, oresp.GetOutputs().GetOutput(), 2)

			oOpts := &porter.OutputListOptions{}
			oOpts.Name = test.instName
			oOpts.Namespace = test.instNamespace
			oOpts.Format = "json"
			dvs, err := grpcSvr.TestPorter.ListBundleOutputs(ctx, oOpts)
			require.NoError(t, err)

			//Validation
			validateOutputs(t, dvs, oresp)
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

func validateOutputs(t *testing.T, dvs porter.DisplayValues, actual *iGRPC.ListInstallationLatestOutputResponse) {
	//Get expected json
	bExpOuts, err := json.MarshalIndent(dvs, "", "  ")

	//Get actual json response
	pjm := protojson.MarshalOptions{EmitUnpopulated: true}
	bActOuts, err := pjm.Marshal(actual.GetOutputs())
	rActOuts := gjson.GetBytes(bActOuts, "output")
	var pJson bytes.Buffer
	//TODO: fix the layers of outputs in GRPC Response
	//json.Indent(&pJson, bActOuts, "", "  ")
	json.Indent(&pJson, []byte(rActOuts.String()), "", "  ")
	if true {
		fmt.Printf("GET OUTPUTS: %+v\n", actual.GetOutputs())
		fmt.Printf("GET OUTPUT: %+v\n", actual.GetOutputs().GetOutput())

		fmt.Printf("PORTER INSTALLATION Outputs:\n%s\n", string(bExpOuts))
		fmt.Printf("GRPC INSTALLATION Outputs:\n%s\n", string(pJson.Bytes()))
	}
	require.NoError(t, err)
	//TODO: fix the layers of outputs in GRPC Response
	assert.JSONEq(t, string(bExpOuts), rActOuts.String())
	//assert.JSONEq(t, string(bExpOuts), string(bActOuts))
}
