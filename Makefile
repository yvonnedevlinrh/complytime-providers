BINARY_DIR ?= bin

.PHONY: build build-openscap-provider build-ampel-provider build-opa-provider test test-cross-repo vendor lint

build: build-openscap-provider build-ampel-provider build-opa-provider

build-openscap-provider:
	go build -o $(BINARY_DIR)/complyctl-provider-openscap ./cmd/openscap-provider

build-ampel-provider:
	go build -o $(BINARY_DIR)/complyctl-provider-ampel ./cmd/ampel-provider

build-opa-provider:
	go build -o $(BINARY_DIR)/complyctl-provider-opa ./cmd/opa-provider

test:
	go test ./...

test-cross-repo: ## run cross-repo integration test (requires COMPLYCTL_DIR and GITHUB_TOKEN)
ifndef COMPLYCTL_DIR
	$(error COMPLYCTL_DIR is not set. Set it to the root of a built complyctl checkout)
endif
	timeout 120 $(COMPLYCTL_DIR)/tests/cross-repo/cross_repo_integration_test.sh

vendor:
	go mod vendor

lint:
	golangci-lint run ./...
