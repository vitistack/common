# Makefile for the project
# inspired by kubebuilder.io

# Setting SHELL to bash allows bash commands to be executed by recipes.
# Options are set to exit when a recipe line exits non-zero or a piped command fails.
SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec

# Basic colors
BLACK=\033[0;30m
RED=\033[0;31m
GREEN=\033[0;32m
YELLOW=\033[0;33m
BLUE=\033[0;34m
PURPLE=\033[0;35m
CYAN=\033[0;36m
WHITE=\033[0;37m

# Text formatting
BOLD=\033[1m
UNDERLINE=\033[4m
RESET=\033[0m

## Location to install dependencies to
LOCALBIN ?= $(shell pwd)/bin
$(LOCALBIN):
	mkdir -p $(LOCALBIN)

CONTROLLER_GEN = $(LOCALBIN)/controller-gen
GOLANGCI_LINT = $(LOCALBIN)/golangci-lint
GOSEC ?= $(LOCALBIN)/gosec
GOVULNCHECK ?= $(LOCALBIN)/govulncheck

# Use the Go toolchain version declared in go.mod when building tools
GO_VERSION := $(shell awk '/^go /{print $$2}' go.mod)
GO_TOOLCHAIN := go$(GO_VERSION)
GOSEC_VERSION ?= latest
GOVULNCHECK_VERSION ?= latest
GOLANGCI_LINT_VERSION ?= latest
CONTROLLER_TOOLS_VERSION ?= latest

##@ Help
.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Generate code and manifests

.PHONY: manifests
manifests: generate ## Generate CRD manifests
	
.PHONY: generate
generate: gen-deepcopy gen-manifests   ## Generate code and manifests.

.PHONY: gen-manifests
gen-manifests: controller-gen ## Generate manifests
	$(CONTROLLER_GEN) rbac:roleName=manager-role crd webhook paths="./..." output:crd:artifacts:config=crds
	@hack/sanitize-crds.sh crds
	@echo "Cleaning up charts/vitistack-crds/templates..."
	@rm -f charts/vitistack-crds/templates/*.yaml
	@echo "Copying CRDs to charts/vitistack-crds/templates..."
	@cp crds/*.yaml charts/vitistack-crds/templates/
	@echo "CRDs copied successfully!"

.PHONY: sanitize-crds
sanitize-crds: ## Sanitize generated CRDs to remove unsupported int32/int64 formats.
	hack/sanitize-crds.sh crds

.PHONY: verify-crds
verify-crds: ## Verify CRDs are sanitized (no int32/int64 format lines present).
	@set -e; \
	if grep -R --include='*.yaml' -E '^[[:space:]]*format:[[:space:]]*"?int(32|64)"?[[:space:]]*$$' crds >/dev/null; then \
	  echo "CRDs contain int32/int64 format lines. Run 'make sanitize-crds' or 'make manifests'."; \
	  exit 1; \
	else \
	  echo "CRDs are sanitized."; \
	fi

.PHONY: gen-deepcopy
gen-deepcopy: controller-gen ## Generate code
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."


##@ Build
.PHONY: build
build: ## Build the manager binary.
	go build ./...

##@ Code sanity

.PHONY: fmt
fmt: ## Run go fmt against code.
	go fmt ./...

.PHONY: vet
vet: ## Run go vet against code.
	go vet ./...

.PHONY: lint
lint: golangci-lint ## Run go vet against code.
	$(GOLANGCI_LINT) run --timeout 5m ./... --config .golangci.yml

##@ Tests
.PHONY: test
test: ## Run unit tests.
	go test -v ./... -coverprofile coverage.out
	go tool cover -html=coverage.out -o coverage.html

.PHONY: bench
bench: ## Run benchmarks (override with BENCH=<regex>, PKG=<package pattern>, COUNT=<n>)
	@bench_regex=$${BENCH:-.}; \
	pkg_pattern=$${PKG:-./...}; \
	count=$${COUNT:-1}; \
	echo "Running benchmarks: regex=$${bench_regex} packages=$${pkg_pattern} count=$${count}"; \
	go test -run=^$$ -bench=$${bench_regex} -benchmem -count=$${count} $${pkg_pattern}

.PHONY: bench-profile
bench-profile: ## Run benchmarks with CPU & memory profiles (outputs bench.cpu, bench.mem)
	@bench_regex=$${BENCH:-.}; \
	pkg_pattern=$${PKG:-./pkg/loggers/vlog}; \
	echo "Profiling benchmarks: regex=$${bench_regex} packages=$${pkg_pattern}"; \
	go test -run=^$$ -bench=$${bench_regex} -cpuprofile bench.cpu -memprofile bench.mem -benchmem $${pkg_pattern}

deps: ## Download and verify dependencies
	@echo "Downloading dependencies..."
	@go mod download
	@go mod verify
	@go mod tidy
	@echo "Dependencies updated!"

update-deps: ## Update dependencies
	@echo "Updating dependencies..."
	@go get -u ./...
	@go mod tidy
	@echo "Dependencies updated!"

##@ Helm
.PHONY: helm-lint
helm-lint: ## Lint Helm charts
	helm lint charts/vitistack-crds

##@ kubernetes
.PHONY: install-crds
install-crds: manifests ## Install CRDs into a Kubernetes cluster.
	kubectl apply -f crds

.PHONY: uninstall-crds
uninstall-crds: ## Uninstall CRDs from a Kubernetes cluster.
	kubectl delete -f crds

##@ Tools

.PHONY: controller-gen
controller-gen: $(CONTROLLER_GEN) ## Download controller-gen locally if necessary.
$(CONTROLLER_GEN): $(LOCALBIN)
	$(call go-install-tool,$(CONTROLLER_GEN),sigs.k8s.io/controller-tools/cmd/controller-gen,$(CONTROLLER_TOOLS_VERSION))

.PHONY: golangci-lint
golangci-lint: $(LOCALBIN) ## Download golangci-lint locally if necessary.
	$(call go-install-tool,$(GOLANGCI_LINT),github.com/golangci/golangci-lint/v2/cmd/golangci-lint,$(GOLANGCI_LINT_VERSION))

.PHONY: install-security-scanner
install-security-scanner: $(GOSEC) ## Install gosec security scanner locally (static analysis for security issues)
$(GOSEC): $(LOCALBIN)
	@set -e; echo "Attempting to install gosec $(GOSEC_VERSION)"; \
	if ! GOBIN=$(LOCALBIN) go install github.com/securego/gosec/v2/cmd/gosec@$(GOSEC_VERSION) 2>/dev/null; then \
		echo "Primary install failed, attempting install from @main (compatibility fallback)"; \
		if ! GOBIN=$(LOCALBIN) go install github.com/securego/gosec/v2/cmd/gosec@main; then \
			echo "gosec installation failed for versions $(GOSEC_VERSION) and @main"; \
			exit 1; \
		fi; \
	fi; \
	echo "gosec installed at $(GOSEC)"; \
	chmod +x $(GOSEC)

.PHONY: install-govulncheck
install-govulncheck: $(GOVULNCHECK) ## Install govulncheck locally (vulnerability scanner for Go)
$(GOVULNCHECK): $(LOCALBIN)
	@set -e; echo "Attempting to install govulncheck $(GOVULNCHECK_VERSION)"; \
	if ! GOBIN=$(LOCALBIN) go install golang.org/x/vuln/cmd/govulncheck@$(GOVULNCHECK_VERSION) 2>/dev/null; then \
		echo "Primary install failed, attempting install from @latest (compatibility fallback)"; \
		if ! GOBIN=$(LOCALBIN) go install golang.org/x/vuln/cmd/govulncheck@latest; then \
			echo "govulncheck installation failed for versions $(GOVULNCHECK_VERSION) and @latest"; \
			exit 1; \
		fi; \
	fi; \
	echo "govulncheck installed at $(GOVULNCHECK)"; \
	chmod +x $(GOVULNCHECK)

##@ Security
.PHONY: gosec
gosec: install-security-scanner ## Run gosec security scan (fails on findings)
	$(GOSEC) ./...

.PHONY: govulncheck
govulncheck: install-govulncheck ## Run govulncheck vulnerability scan (fails on findings)
	$(GOVULNCHECK) ./...
# go-install-tool will 'go install' any package with custom target and name of binary, if it doesn't exist
# $1 - target path with name of binary
# $2 - package url which can be installed
# $3 - specific version of package
define go-install-tool
@[ -f "$(1)-$(3)" ] || { \
set -e; \
package=$(2)@$(3) ;\
echo "Downloading $${package}" ;\
rm -f $(1) || true ;\
GOTOOLCHAIN=$(GO_TOOLCHAIN) GOBIN=$(LOCALBIN) go install $${package} ;\
mv $(1) $(1)-$(3) ;\
} ;\
ln -sf $(1)-$(3) $(1)
endef
