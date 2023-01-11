package main

import (
	"context"
	"fmt"
	"log"

	// Remote Packages with buf.build
	// petgrpc "buf.build/gen/go/gettys/petapis/grpc/go/pet/v1/petv1grpc"
	// petv1 "buf.build/gen/go/gettys/petapis/protocolbuffers/go/pet/v1"

	// Local generated files
	igrpc "get.porter.sh/porter/gen/proto/go/porterapis/installation/v1alpha1"
	"google.golang.org/grpc"
)

var grpcSvcPort = "8988"

func main() {
	log.Println("grpc client test")
	if err := run(); err != nil {
		log.Fatal(err)
	}
}
func run() error {
	connectTo := fmt.Sprintf("127.0.0.1:%s", grpcSvcPort)
	conn, err := grpc.Dial(connectTo, grpc.WithBlock(), grpc.WithInsecure())
	if err != nil {
		return fmt.Errorf("failed to connect to PetStoreService on %s: %w", connectTo, err)
	}
	log.Println("Connected to", connectTo)

	instClient := igrpc.NewInstallationsClient(conn)
	resp, err := instClient.ListInstallations(context.Background(), &igrpc.ListInstallationsRequest{})
	if err != nil {
		return fmt.Errorf("failed to ListInstallations: %w", err)
	}
	log.Println("Successfully Listed Installation")
	log.Println(resp)
	return nil
}
