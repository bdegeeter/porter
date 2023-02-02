package installation

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	iGRPC "get.porter.sh/porter/gen/proto/go/porterapis/installation/v1alpha1"
	"get.porter.sh/porter/pkg/cnab"
	pCtx "get.porter.sh/porter/pkg/grpc/context"
	"get.porter.sh/porter/pkg/porter"
	"get.porter.sh/porter/pkg/storage"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/encoding/protojson"
	"k8s.io/utils/pointer"
)

type instInfo struct {
	namespace string
	name      string
}

func TestListInstallationAllNamespacesReturnsListOfAllPorterInstallations(t *testing.T) {
	tests := []struct {
		name        string
		instInfo    []instInfo
		numExpInsts int
	}{
		{
			name:        "NoInstallationReturnsEmptyListOfInstallations",
			instInfo:    []instInfo{},
			numExpInsts: 0,
		},
		{
			name: "SingleInstallationSingleNamespace",
			instInfo: []instInfo{
				{namespace: "", name: "test"},
			},
			numExpInsts: 1,
		},
		{
			name: "SingleInstallationMultipleNamespace",
			instInfo: []instInfo{
				{namespace: "foo", name: "test"},
				{namespace: "bar", name: "test"},
			},
			numExpInsts: 2,
		},
		{
			name: "MultipleInstallationMultipleNamespace",
			instInfo: []instInfo{
				{namespace: "foo", name: "test1"},
				{namespace: "foo", name: "test2"},
				{namespace: "bar", name: "test3"},
				{namespace: "bar", name: "test4"},
			},
			numExpInsts: 4,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx, expInsts := setupTestPorterWithInstallations(t, test.instInfo, func(inst storage.Installation) bool {
				return true
			})
			instSvc := PorterServer{}
			req := &iGRPC.ListInstallationsRequest{AllNamespaces: pointer.Bool(true)}
			resp, err := instSvc.ListInstallations(ctx, req)
			installations := resp.GetInstallation()
			assert.Nil(t, err)
			assert.Len(t, installations, test.numExpInsts)
			verifyInstalltions(t, installations, expInsts)
		})
	}
}

func TestListInstallationReturnsOnlyRequestedInstallationIfNamespaceAndNameInListInstallationsRequest(t *testing.T) {
	reqInstInfo := instInfo{namespace: "test", name: "foo"}
	tests := []struct {
		name        string
		instInfo    []instInfo
		numExpInsts int
	}{
		{
			name:        "NoInstallationsReturnsEmptyListOfInstallationsWhenNamespaceAndNameSpecified",
			instInfo:    []instInfo{},
			numExpInsts: 0,
		},
		{
			name: "SingleInstallationSingleNamespaceReturnsEmptyListOfInstallationsWhenNamespaceAndNameInRequestDoNotMatch",
			instInfo: []instInfo{
				{namespace: "", name: "test"},
			},
			numExpInsts: 0,
		},
		{
			name: "SingleInstallationSingleNamespaceReturnsInstallationWhenNamespaceAndNameInRequestMatch",
			instInfo: []instInfo{
				{namespace: "test", name: "foo"},
			},
			numExpInsts: 1,
		},
		{
			name: "SingleInstallationMultipleNamespaceReturnsInstallationWhenNamespaceAndNameInRequestMatch",
			instInfo: []instInfo{
				{namespace: "test", name: "foo"},
				{namespace: "bar", name: "test"},
			},
			numExpInsts: 1,
		},
		{
			name: "MultipleInstallationMultipleNamespacesReturnsInstallationWhenNamespaceAndNameInRequestMatch",
			instInfo: []instInfo{
				{namespace: "foo", name: "test1"},
				{namespace: "foo", name: "test2"},
				{namespace: "test", name: "foo"},
				{namespace: "bar", name: "test4"},
			},
			numExpInsts: 1,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx, expInsts := setupTestPorterWithInstallations(t, test.instInfo, func(inst storage.Installation) bool {
				return inst.Namespace == reqInstInfo.namespace && inst.Name == reqInstInfo.name
			})
			instSvc := PorterServer{}
			req := &iGRPC.ListInstallationsRequest{Name: reqInstInfo.name, Namespace: &reqInstInfo.namespace}
			resp, err := instSvc.ListInstallations(ctx, req)
			installations := resp.GetInstallation()
			assert.Nil(t, err)
			assert.Len(t, installations, test.numExpInsts)
			verifyInstalltions(t, installations, expInsts)
		})
	}
}

func TestListInstallationOnlyReturnsInstallationsInNamespaceSpecifiedInListInstallationRequest(t *testing.T) {
	reqInstInfo := instInfo{namespace: "test"}
	tests := []struct {
		name        string
		instInfo    []instInfo
		numExpInsts int
	}{
		{
			name:        "NoInstallationsReturnsEmptyListOfInstallationsWhenNamespaceSpecified",
			instInfo:    []instInfo{},
			numExpInsts: 0,
		},
		{
			name: "SingleInstallationSingleNamespaceReturnsEmptyListOfInstallationsWhenNamespaceInRequestDoesNotMatch",
			instInfo: []instInfo{
				{namespace: "", name: "test"},
			},
			numExpInsts: 0,
		},
		{
			name: "SingleInstallationSingleNamespaceReturnsInstallationWhenNamespaceInRequestMatches",
			instInfo: []instInfo{
				{namespace: "test", name: "foo"},
			},
			numExpInsts: 1,
		},
		{
			name: "SingleInstallationMultipleNamespaceReturnsInstallationWhenNamespaceInRequestMatches",
			instInfo: []instInfo{
				{namespace: "test", name: "foo"},
				{namespace: "bar", name: "test"},
			},
			numExpInsts: 1,
		},
		{
			name: "MultipleInstallationMultipleNamespacesReturnsInstallationWhenNamespaceInRequestMatches",
			instInfo: []instInfo{
				{namespace: "foo", name: "test1"},
				{namespace: "test", name: "test2"},
				{namespace: "test", name: "foo"},
				{namespace: "bar", name: "test4"},
			},
			numExpInsts: 2,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx, expInsts := setupTestPorterWithInstallations(t, test.instInfo, func(inst storage.Installation) bool {
				return reqInstInfo.namespace == inst.Namespace
			})
			instSvc := PorterServer{}
			req := &iGRPC.ListInstallationsRequest{Namespace: &reqInstInfo.namespace}
			resp, err := instSvc.ListInstallations(ctx, req)
			installations := resp.GetInstallation()
			assert.Nil(t, err)
			assert.Len(t, installations, test.numExpInsts)
			verifyInstalltions(t, installations, expInsts)
		})
	}
}

func TestListInstallationsReturnsErrorIfUnableToGetPorterConnectionFromRequestContext(t *testing.T) {
	instSvc := PorterServer{}
	req := &iGRPC.ListInstallationsRequest{}
	ctx := context.TODO()
	resp, err := instSvc.ListInstallations(ctx, req)
	assert.Error(t, err)
	assert.Nil(t, resp)

}

func setupTestPorterWithInstallations(t *testing.T, installations []instInfo, match func(storage.Installation) bool) (context.Context, map[string]porter.DisplayInstallation) {
	p := porter.NewTestPorter(t)
	expInsts := map[string]porter.DisplayInstallation{}
	for _, inst := range installations {
		installation := storage.NewInstallation(inst.namespace, inst.name)
		storeInst := p.TestInstallations.CreateInstallation(installation, p.TestInstallations.SetMutableInstallationValues, func(i *storage.Installation) {
			i.Status.BundleVersion = "v0.1.0"
			i.Status.ResultStatus = cnab.StatusSucceeded
			i.Bundle.Repository = "test-bundle"
			i.Bundle.Version = "v0.1.0"
		})
		if match(storeInst) {
			instName := fmt.Sprintf("%s-%s", installation.Namespace, installation.Name)
			expInsts[instName] = porter.NewDisplayInstallation(storeInst)
		}
	}
	ctx := pCtx.AddPorterConnectionToContext(p.Porter, context.TODO())
	return ctx, expInsts
}

func verifyInstalltions(t *testing.T, installations []*iGRPC.Installation, expInsts map[string]porter.DisplayInstallation) {
	for _, inst := range installations {
		instName := fmt.Sprintf("%s-%s", inst.Namespace, inst.Name)
		i, ok := expInsts[instName]
		assert.True(t, ok)
		bExpInst, err := json.Marshal(i)
		assert.NoError(t, err)
		pjm := protojson.MarshalOptions{EmitUnpopulated: true}
		bActInst, err := pjm.Marshal(inst)
		assert.NoError(t, err)
		assert.JSONEq(t, string(bExpInst), string(bActInst))
	}
}
