GRPCURL_VERSION=1.8.7
GRPCURL_CMD=grpcurl
GRPC_HOST=grpc.localtest.me
GRPC_PORT=8988


.PHONY: grpc-test
grpc-test: | $(GRPCURL_HOME)
	@echo "Testing gRPC service methods"
	@echo ""
	go run tests/grpc/client/main.go

.PHONY: grpc-list
grpc-list:
	$(GRPCURL_CMD) -plaintext -authority $(GRPC_HOST) $(GRPC_HOST):$(GRPC_PORT) list
	$(GRPCURL_CMD) -plaintext -authority $(GRPC_HOST) $(GRPC_HOST):$(GRPC_PORT) list installation.v1alpha1.Installations

.PHONY: grpc-serve
grpc-serve:
	@echo "Start GRPC server. Ctrl+C to exit..."
	@printf "#porter-service config.yaml\ngrpc-port: $(GRPC_PORT)\n" > $(PWD)/config.yaml
	./bin/porter-service --config-path=$(PWD)

.PHONY: grpc-regenerate
grpc-regenerate:
	@echo "Regenerating gRPC code from protobuf"
	@docker build --rm -t protoc:local -f protoc.Dockerfile .
	mkdir -p gen/proto/go/helloworld/v1alpha
	docker run -v $(PWD):/proto protoc:local protoc \
			--go_out=gen/proto/go/helloworld/v1alpha --go_opt=module=get.porter.sh/porter/gen/proto/go/helloworld/v1alpha \
    	--go-grpc_out=gen/proto/go/helloworld/v1alpha --go-grpc_opt=module=get.porter.sh/porter/gen/proto/go/helloworld/v1alpha \
    	proto/porterapis/helloword/v1alpha/helloworld.proto