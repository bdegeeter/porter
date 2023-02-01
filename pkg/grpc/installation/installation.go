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
	tspb "google.golang.org/protobuf/types/known/timestamppb"
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
	//bInst, err := json.Marshal(inst)
	bInst, err := json.MarshalIndent(inst, "", "  ")
	if err != nil {
		return err
	}
	fmt.Printf("PORTER INSTALLATION:\n%s\n", string(bInst))
	pjum := protojson.UnmarshalOptions{}
	err = pjum.Unmarshal(bInst, gInst)
	if err != nil {
		return err
	}
	return nil
}

func makeInstResponse(inst porter.DisplayInstallation) *iGRPC.Installation {
	var uninstTime *tspb.Timestamp
	var instTime *tspb.Timestamp
	if inst.Status.Uninstalled != nil {
		uninstTime = tspb.New(*inst.Status.Uninstalled)
	}
	if inst.Status.Installed != nil {
		instTime = tspb.New(*inst.Status.Installed)
	}
	return &iGRPC.Installation{
		Id:            inst.ID,
		Name:          inst.Name,
		SchemaVersion: string(inst.SchemaVersion),
		Namespace:     inst.Namespace,
		Bundle: &iGRPC.Bundle{
			Repository: inst.Bundle.Repository,
			Version:    inst.Bundle.Version,
		},
		Status: &iGRPC.InstallationStatus{
			RunId:           inst.Status.RunID,
			Action:          inst.Status.Action,
			ResultId:        inst.Status.ResultID,
			ResultStatus:    inst.Status.ResultStatus,
			Created:         tspb.New(inst.Status.Created),
			Modified:        tspb.New(inst.Status.Modified),
			Installed:       instTime,
			Uninstalled:     uninstTime,
			BundleReference: inst.Status.BundleReference,
			BundleVersion:   inst.Status.BundleVersion,
			BundleDigest:    inst.Status.BundleDigest,
		},
		Calculated: &iGRPC.Calculated{
			ResolvedParameters:        makePorterValues(inst.ResolvedParameters),
			DisplayInstallationState:  inst.DisplayInstallationState,
			DisplayInstallationStatus: inst.DisplayInstallationStatus,
		},
	}

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
