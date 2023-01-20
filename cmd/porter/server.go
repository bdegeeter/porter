package main

import (
	"time"

	grpc "get.porter.sh/porter/pkg/grpc"
	"get.porter.sh/porter/pkg/porter"
	"get.porter.sh/porter/pkg/signals"
	"github.com/spf13/cobra"
)

func buildServerCommands(p *porter.Porter) *cobra.Command {
	cmd := &cobra.Command{
		Use:    "server",
		Short:  "Server commands",
		Long:   "Command for working with the gRPC server for porter",
		Hidden: true, // This is a hidden command and is currently only meant to be used by the porter operator
	}
	cmd.Annotations = map[string]string{
		"group": "server",
	}
	cmd.AddCommand(buildServerRunCommand(p))
	return cmd
}

func buildServerRunCommand(p *porter.Porter) *cobra.Command {
	opts := porter.ListOptions{}
	grpcOpts := grpc.Config{
		Port: 3333,
	}
	cmd := &cobra.Command{
		Use:   "run",
		Short: "Run the gRPC server",
		Long: `Run the gRPC server for porter.

This command starts the gRPC server for porter which is able to expose limited porter functionality via RPC.
Currently only data operations are supported, creation of resources such as installations, credential sets, or parameter sets is not supported.

A list of the supported RPCs can be found at <link?>
`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return opts.Validate()
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			srv, err := grpc.NewServer(cmd.Context(), &grpcOpts)
			if err != nil {
				return err
			}
			grpcServer, err := srv.ListenAndServe()
			stopCh := signals.SetupSignalHandler()
			serverShutdownTimeout := time.Duration(time.Second * 30)
			sd, _ := signals.NewShutdown(serverShutdownTimeout, cmd.Context())
			sd.Graceful(stopCh, grpcServer, cmd.Context())
			return err
		},
	}
	return cmd
}
