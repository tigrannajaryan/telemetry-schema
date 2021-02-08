include ./Makefile.Common

RUN_CONFIG=local/config.yaml

CMD?=
GIT_SHA=$(shell git rev-parse --short HEAD)
BUILD_INFO_IMPORT_PATH=github.com/tigrannajaryan/telemetry-schema/internal/version
BUILD_X1=-X $(BUILD_INFO_IMPORT_PATH).GitHash=$(GIT_SHA)
ifdef VERSION
BUILD_X2=-X $(BUILD_INFO_IMPORT_PATH).Version=$(VERSION)
endif
BUILD_X3=-X go.opentelemetry.io/collector/internal/version.BuildType=$(BUILD_TYPE)
BUILD_INFO=-ldflags "${BUILD_X1} ${BUILD_X2} ${BUILD_X3}"
STATIC_CHECK=staticcheck
OTEL_VERSION=master

EXE_NAME=telschema

# Modules to run integration tests on.
# XXX: Find a way to automatically populate this. Too slow to run across all modules when there are just a few.
INTEGRATION_TEST_MODULES := \
	internal/common

.DEFAULT_GOAL := all

.PHONY: all
all: common build-exe

.PHONY: test-with-cover
unit-tests-with-cover:
	@echo Verifying that all packages have test files to count in coverage
	@internal/buildscripts/check-test-files.sh $(subst github.com/tigrannajaryan/telemetry-schema/,./,$(ALL_PKGS))
	@$(MAKE) for-all CMD="make do-unit-tests-with-cover"

.PHONY: integration-tests-with-cover
integration-tests-with-cover:
	@echo $(INTEGRATION_TEST_MODULES)
	@$(MAKE) for-all CMD="make do-integration-tests-with-cover" ALL_MODULES="$(INTEGRATION_TEST_MODULES)"

.PHONY: gotidy
gotidy:
	$(MAKE) for-all CMD="rm -fr go.sum"
	$(MAKE) for-all CMD="go mod tidy"

.PHONY: gofmt
gofmt:
	$(MAKE) for-all CMD="make fmt"

.PHONY: for-all
for-all:
	@echo "running $${CMD} in root"
	@$${CMD}
	@set -e; for dir in $(ALL_MODULES); do \
	  (cd "$${dir}" && \
	  	echo "running $${CMD} in $${dir}" && \
	 	$${CMD} ); \
	done

.PHONY: add-tag
add-tag:
	@[ "${TAG}" ] || ( echo ">> env var TAG is not set"; exit 1 )
	@echo "Adding tag ${TAG}"
	@git tag -a ${TAG} -s -m "Version ${TAG}"
	@set -e; for dir in $(ALL_MODULES); do \
	  (echo Adding tag "$${dir:2}/$${TAG}" && \
	 	git tag -a "$${dir:2}/$${TAG}" -s -m "Version ${dir:2}/${TAG}" ); \
	done

.PHONY: delete-tag
delete-tag:
	@[ "${TAG}" ] || ( echo ">> env var TAG is not set"; exit 1 )
	@echo "Deleting tag ${TAG}"
	@git tag -d ${TAG}
	@set -e; for dir in $(ALL_MODULES); do \
	  (echo Deleting tag "$${dir:2}/$${TAG}" && \
	 	git tag -d "$${dir:2}/$${TAG}" ); \
	done

GOMODULES = $(ALL_MODULES) $(PWD)
.PHONY: $(GOMODULES)
MODULEDIRS = $(GOMODULES:%=for-all-target-%)
for-all-target: $(MODULEDIRS)
$(MODULEDIRS):
	$(MAKE) -C $(@:for-all-target-%=%) $(TARGET)
.PHONY: for-all-target

.PHONY: install-tools
install-tools:
	go install github.com/client9/misspell/cmd/misspell
	go install github.com/golangci/golangci-lint/cmd/golangci-lint
	go install github.com/google/addlicense
	go install github.com/jstemmer/go-junit-report
	go install github.com/pavius/impi/cmd/impi
	go install github.com/tcnksm/ghr
	go install honnef.co/go/tools/cmd/staticcheck
	go install go.opentelemetry.io/collector/cmd/issuegenerator

.PHONY: run
run:
	./bin/$(EXE_NAME)_darwin_amd64 --config ./cmd/$(EXE_NAME)/config.yaml --log-level DEBUG

.PHONY: docker-component # Not intended to be used directly
docker-component: check-component
	GOOS=linux GOARCH=amd64 $(MAKE) $(COMPONENT)
	cp ./bin/$(COMPONENT)_linux_amd64 ./cmd/$(COMPONENT)/$(COMPONENT)
	docker build -t $(COMPONENT) ./cmd/$(COMPONENT)/
	rm ./cmd/$(COMPONENT)/$(COMPONENT)

.PHONY: check-component
check-component:
ifndef COMPONENT
	$(error COMPONENT variable was not defined)
endif

.PHONY: build-exe
build-exe:
	GO111MODULE=on CGO_ENABLED=0 go build -o ./bin/$(EXE_NAME)_$(GOOS)_$(GOARCH)$(EXTENSION) $(BUILD_INFO) ./cmd/$(EXE_NAME)

.PHONY: build-all-sys
build-all-sys: build-darwin_amd64 build-linux_amd64

.PHONY: build-darwin_amd64
build-darwin_amd64:
	GOOS=darwin  GOARCH=amd64 $(MAKE) build-exe

.PHONY: build-linux_amd64
build-linux_amd64:
	GOOS=linux   GOARCH=amd64 $(MAKE) build-exe

.PHONY: update-dep
update-dep:
	$(MAKE) for-all CMD="$(PWD)/internal/buildscripts/update-dep"
	$(MAKE) build-exe
	$(MAKE) gotidy

.PHONY: update-otel
update-otel:
	$(MAKE) update-dep MODULE=go.opentelemetry.io/collector VERSION=$(OTEL_VERSION)

