package portergrpc

import (
	"context"
	"encoding/json"
	"fmt"

	iGRPC "get.porter.sh/porter/gen/proto/go/porterapis/installation/v1alpha1"
	"get.porter.sh/porter/pkg/porter"
	"get.porter.sh/porter/pkg/tracing"
	"google.golang.org/protobuf/encoding/protojson"
)

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

func makeGRPCInstallation(inst porter.DisplayInstallation, gInst *iGRPC.Installation) error {
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

func makeGRPCInstallationOutputs(dv porter.DisplayValues, gInstOuts *iGRPC.InstallationOutputs) error {
	pjum := protojson.UnmarshalOptions{}

	gPorterValues := []*iGRPC.PorterValue{}
	for _, v := range dv {
		gInstOut := &iGRPC.PorterValue{}
		bInstOut, err := json.Marshal(v)
		if err != nil {
			return fmt.Errorf("PorterValue marshal error: %e", err)
		}
		err = pjum.Unmarshal(bInstOut, gInstOut)
		if err != nil {
			return fmt.Errorf("installation GRPC InstallationOutputs unmarshal error: %e", err)
		}
		gPorterValues = append(gPorterValues, gInstOut)
	}
	gInstOuts.Output = gPorterValues
	return nil
}

func (s *PorterServer) ListInstallations(ctx context.Context, req *iGRPC.ListInstallationsRequest) (*iGRPC.ListInstallationsResponse, error) {
	ctx, log := tracing.StartSpan(ctx)
	defer log.EndSpan()
	p, err := GetPorterConnectionFromContext(ctx)
	// Maybe try to setup a new porter connection instead of erring?
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
		gInst := &iGRPC.Installation{}
		err := makeGRPCInstallation(pInst, gInst)
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

func (s *PorterServer) ListInstallationLatestOutputs(ctx context.Context, req *iGRPC.ListInstallationLatestOutputRequest) (*iGRPC.ListInstallationLatestOutputResponse, error) {
	ctx, log := tracing.StartSpan(ctx)
	defer log.EndSpan()
	p, err := GetPorterConnectionFromContext(ctx)
	// Maybe try to setup a new porter connection instead of erring?
	if err != nil {
		return nil, err
	}

	opts := porter.OutputListOptions{}
	opts.Name = req.GetName()
	opts.Namespace = req.GetNamespace()
	opts.Format = "json"
	pdv, err := p.ListBundleOutputs(ctx, &opts)
	if err != nil {
		return nil, err
	}
	gInstOuts := &iGRPC.InstallationOutputs{}
	err = makeGRPCInstallationOutputs(pdv, gInstOuts)
	if err != nil {
		return nil, err
	}
	res := &iGRPC.ListInstallationLatestOutputResponse{
		Outputs: gInstOuts,
	}
	return res, nil
}
