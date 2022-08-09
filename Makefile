# -----------------------------------------------------------------------------
# global

.DEFAULT_GOAL := help

comma := ,
empty :=
space := $(empty) $(empty)

# -----------------------------------------------------------------------------
# go

GO_PATH ?= $(shell go env GOPATH)
GO_OS ?= $(shell go env GOOS)
GO_ARCH ?= $(shell go env GOARCH)

PKG := $(subst $(GO_PATH)/src/,,$(CURDIR))
ORG_PKG := $(subst /$(notdir $(PKG)),,$(PKG))
CGO_ENABLED ?= 0
GO_BUILDTAGS=osusergo netgo static
GO_LDFLAGS=-s -w
ifeq (${GO_OS},linux)
GO_LDFLAGS+=-d
endif
GO_LDFLAGS+=-buildid= "-extldflags=-static"
GO_FLAGS ?= -tags='$(subst $(space),$(comma),${GO_BUILDTAGS})' -ldflags='${GO_LDFLAGS}'

GO_PKGS := $(shell go list ./...)
GO_TEST_PKGS ?= $(shell go list -f='{{if or .TestGoFiles .XTestGoFiles}}{{.ImportPath}}{{end}}' ./...)

GO_TEST ?= ${TOOLS_BIN}/gotestsum --
GO_TEST_FLAGS ?= -race -count=1 -shuffle=on
GO_TEST_FUNC ?= .
GO_COVERAGE_OUT ?= coverage.out
GO_BENCH_FLAGS ?= -benchmem
GO_BENCH_FUNC ?= .
GO_LINT_FLAGS ?=

TOOLS_DIR := ${CURDIR}/hack/tools
TOOLS_BIN := ${TOOLS_DIR}/bin
TOOLS = $(shell cd ${TOOLS_DIR}; go list -f '{{ join .Imports " " }}' -tags=tools)

JOBS = $(shell getconf _NPROCESSORS_CONF)

# -----------------------------------------------------------------------------
# defines

define target
@printf "+ $(patsubst ,$@,$(1))\\n" >&2
endef

# -----------------------------------------------------------------------------
# target

##@ build

.PHONY: ${GOBIN}/$(notdir ${PKG})
${GOBIN}/$(notdir ${PKG}):
	$(call target,build)
	@mkdir -p ./bin
	CGO_ENABLED=0 go build -v ${GO_FLAGS} -o ./bin/$(notdir ${PKG}) ${PKG}

build: ${GOBIN}/$(notdir ${PKG})  ## Build binary.

##@ proto

.PHONY: protoc
protoc:
	PATH=${CURDIR}/bin:$$PATH protoc -I . -I ${HOME}/src/github.com/googleapis/googleapis --go_out=testdata --go-grpc_out=testdata --proxy_out=testdata ${HOME}/src/github.com/googleapis/googleapis/google/spanner/v1/*.proto

##@ test, bench, coverage

.PHONY: test
test: CGO_ENABLED=1
test: GO_FLAGS=-tags='$(subst ${space},${comma},${GO_BUILDTAGS})'
test: ${TOOLS_BIN}/gotestsum  ## Runs package test including race condition.
	$(call target)
	@CGO_ENABLED=${CGO_ENABLED} GOTESTSUM_FORMAT=standard-verbose ${GO_TEST} ${GO_TEST_FLAGS} -run=${GO_TEST_FUNC} $(strip ${GO_FLAGS}) ${GO_TEST_PKGS}

.PHONY: coverage
coverage: CGO_ENABLED=1
coverage: GO_FLAGS=-tags='$(subst ${space},${comma},${GO_BUILDTAGS})'
coverage: ${TOOLS_BIN}/gotestsum  ## Takes packages test coverage.
	$(call target)
	@CGO_ENABLED=${CGO_ENABLED} ${GO_TEST} ${GO_TEST_FLAGS} -covermode=atomic -coverpkg=${PKG}/... -coverprofile=${GO_COVERAGE_OUT} $(strip ${GO_FLAGS}) ${GO_PKGS}

##@ fmt, lint

.PHONY: fmt
fmt: ${TOOLS_BIN}/goimportz ${TOOLS_BIN}/gofumpt  ## Run goimportz and gofumpt.
	$(call target)
	${TOOLS_BIN}/goimportz -local=${PKG},${ORG_PKG} -w $(shell find . -iname "*.go" -not -iname "*pb..go" -not -path "./vendor/**")
	${TOOLS_BIN}/gofumpt -s -extra -w $(shell find . -iname "*.go" -not -iname "*pb..go" -not -path "./vendor/**")

.PHONY: lint
lint: lint/golangci-lint  ## Run all linters.

.PHONY: lint/golangci-lint
lint/golangci-lint: ${TOOLS_BIN}/golangci-lint .golangci.yaml  ## Run golangci-lint.
	$(call target)
	${TOOLS_BIN}/golangci-lint -j ${JOBS} run $(strip ${GO_LINT_FLAGS}) ./...

##@ tools

.PHONY: tools
tools: ${TOOLS_BIN}/''  ## Install tools

${TOOLS_BIN}/%: ${TOOLS_DIR}/go.sum ${TOOLS_DIR}/go.mod
	@cd ${TOOLS_DIR}; \
	  for t in ${TOOLS}; do \
			if [ ! $$(basename $$t) = 'cmd' ] && [ -z '$*' ] || [ $$(basename $$t) = '$*' ]; then \
				echo "Install $$t"; \
				GOBIN=${TOOLS_BIN} CGO_ENABLED=0 go install -v -mod=readonly ${GO_FLAGS} "$${t}"; \
			fi \
	  done

##@ clean

.PHONY: clean
clean:  ## Cleanups binaries and extra files in the package.
	$(call target)
	@rm -rf *.out *.test *.prof trace.txt ${TOOLS_BIN}

##@ help

.PHONY: help
help:  ## Show this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[33m<target>\033[0m\n"} /^[a-zA-Z_0-9\/%_-]+:.*?##/ { printf "  \033[1;32m%-20s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ miscellaneous

.PHONY: todo
todo:  ## Print the all of (TODO|BUG|XXX|FIXME|NOTE) in packages.
	@grep -E '(TODO|BUG|XXX|FIXME)(\(.+\):|:)' $(shell find . -type f -name '*.go' -and -not -iwholename '*vendor*')

.PHONY: nolint
nolint:  ## Print the all of //nolint:... pragma in packages.
	@rg -t go -C 3 -e '//nolint.+' --follow --hidden --glob='!vendor' --glob='!internal'

.PHONY: env/%
env/%: ## Print the value of MAKEFILE_VARIABLE. Use `make env/GO_FLAGS` or etc.
	@echo $($*)
