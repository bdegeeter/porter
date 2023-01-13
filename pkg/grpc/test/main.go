package main

import (
	"context"
	"os"

	"get.porter.sh/porter/pkg/grpc/installation"
	"get.porter.sh/porter/pkg/porter"
)

func main() {
	p := porter.New()
	if len(os.Args) > 1 {
		opts := porter.RunInternalPluginOpts{Key: "storage.porter.mongodb-docker"}
		p.RunInternalPlugins(context.Background(), opts)
	}
	s, err := installation.NewPorterService()
	if err != nil {
		panic(err)
	}
	s.Foo(p)
}
