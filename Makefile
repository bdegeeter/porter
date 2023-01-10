GRPCURL_VERSION=1.8.7
GRPCURL_HOME=$(PWD)/.grpcurl
GRPCURL_CMD=$(PWD)/.grpcurl/grpcurl
GRPC_HOST=grpc.localtest.me
GRPC_PORT=8988

OS=$(shell uname -s)
ifeq ($(OS),Darwin)
	MAGE_OS=macOS
	SYS_OS=darwin
	GRPCURL_OS=osx
else ifeq ($(OS),Linux)
	MAGE_OS=Linux
	SYS_OS=linux
	GRPCURL_OS=linux
else
	MAGE_OS=unknown
endif

$(GRPCURL_HOME):
	@mkdir -p $(GRPCURL_HOME)
	curl -Lo grpcurl.tgz "https://github.com/fullstorydev/grpcurl/releases/download/v1.8.7/grpcurl_$(GRPCURL_VERSION)_$(GRPCURL_OS)_x86_64.tar.gz"
	tar xzvf grpcurl.tgz -C $(GRPCURL_HOME)
	rm -f grpcurl.tgz
	chmod +x $(GRPCURL_CMD)

.PHONY: build-grpc
build-grpc:
	@docker build --rm -t vulcan-grpc:local -f Dockerfile .

.PHONY: load-grpc-image
load-grpc-image:
	@$(KIND) load docker-image vulcan-grpc:local

.PHONY: grpc-generate-cert
grpc-generate-cert:
	@openssl req -new -nodes -keyout grpc-porter-tls.key -out grpc-porter.csr -config tests/grpc/openssl.cnf -subj "/C=CN/ST=Wa/L=Seattle/O=Porter-Dev/OU=ContainerService/CN=$(GRPC_HOST)"
	@openssl x509 -req -days 3650 -in grpc-porter.csr -signkey grpc-porter-tls.key -out grpc-porter-tls.crt -extensions v3_req -extfile tests/grpc/openssl.cnf

.PHONY: grpc-test
grpc-test: | $(GRPCURL_HOME)
	@echo "Testing gRPC service methods"
	@echo ""
	@echo ""
	@echo "Plain Text Testing SayHello"
	$(GRPCURL_CMD) -plaintext -d '{"name": "porter"}' -authority $(GRPC_HOST) $(GRPC_HOST):$(GRPC_PORT) helloworld.Greeter/SayHello
	@echo "Plain Text Testing SayHelloAgain"
	$(GRPCURL_CMD) -plaintext -d '{"name": "porter"}' -authority $(GRPC_HOST) $(GRPC_HOST):$(GRPC_PORT) helloworld.Greeter/SayHelloAgain

.PHONY: grpc-list
grpc-list:
	$(GRPCURL_CMD) -plaintext -authority $(GRPC_HOST) $(GRPC_HOST):$(GRPC_PORT) list
	$(GRPCURL_CMD) -plaintext -authority $(GRPC_HOST) $(GRPC_HOST):$(GRPC_PORT) list helloworld.Greeter

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