package installation

import (
	"context"
	"encoding/json"
	"fmt"

	iGRPC "get.porter.sh/porter/gen/proto/go/porterapis/installation/v1alpha1"
	pGRPC "get.porter.sh/porter/gen/proto/go/porterapis/porter/v1alpha1"
	pCtx "get.porter.sh/porter/pkg/grpc/context"
	"get.porter.sh/porter/pkg/porter"
	"get.porter.sh/porter/pkg/tracing"
	"google.golang.org/protobuf/encoding/protojson"
	//anypb "google.golang.org/protobuf/types/known/anypb"
)

// server is used to implement helloworld.GreeterServer.
type PorterServer struct {
	pGRPC.UnimplementedPorterBundleServer
}

func NewPorterService() (*PorterServer, error) {
	return &PorterServer{}, nil
}

func makeInstOptsLabels(labels map[string]string) []string {
	var retLabels []string
	for k, v := range labels {
		retLabels = append(retLabels, fmt.Sprintf("%s=%s", k, v))
	}
	return retLabels
}

func displayValueToProtoPorterValue(value porter.DisplayValue) (*iGRPC.PorterValue, error) {
	b, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	pv := &iGRPC.PorterValue{}
	err = protojson.Unmarshal(b, pv)
	if err != nil {
		return nil, err
	}
	return pv, nil
}

func makePorterValues(values porter.DisplayValues) []*iGRPC.PorterValue {
	var retPVs []*iGRPC.PorterValue
	for _, dv := range values {
		//TODO: handle error
		pv, _ := displayValueToProtoPorterValue(dv)
		retPVs = append(retPVs, pv)
	}
	return retPVs
}

func jsonMakeInstResponse(inst porter.DisplayInstallation, gInst *iGRPC.Installation) error {
	bInst, err := json.Marshal(inst)
	if err != nil {
		return err
	}
	pjum := protojson.UnmarshalOptions{}
	err = pjum.Unmarshal(bInst, gInst)
	if err != nil {
		return err
	}
	return nil
}

func (s *PorterServer) ListInstallations(ctx context.Context, req *iGRPC.ListInstallationsRequest) (*iGRPC.ListInstallationsResponse, error) {
	ctx, log := tracing.StartSpan(ctx)
	defer log.EndSpan()
	p, err := pCtx.GetPorterConnectionFromContext(ctx)
	if err != nil {
		return nil, err
	}
	opts := porter.ListOptions{
		Name:          req.GetName(),
		Namespace:     req.GetNamespace(),
		Labels:        makeInstOptsLabels(req.GetLabels()),
		AllNamespaces: req.GetAllNamespaces(),
		Skip:          req.GetSkip(),
		Limit:         req.GetLimit(),
	}
	installations, err := p.ListInstallations(ctx, opts)
	if err != nil {
		return nil, err
	}
	insts := []*iGRPC.Installation{}
	for _, pInst := range installations {
		//inst := makeInstResponse(pInst)
		gInst := &iGRPC.Installation{}
		err := jsonMakeInstResponse(pInst, gInst)
		if err != nil {
			return nil, err
		}
		insts = append(insts, gInst)
	}
	res := iGRPC.ListInstallationsResponse{
		Installation: insts,
	}
	return &res, nil
}
