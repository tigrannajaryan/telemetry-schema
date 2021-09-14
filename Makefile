include ./Makefile.Common

RUN_CONFIG=local/config.yaml

CMD?=
GIT_SHA=$(shell git rev-parse --short HEAD)
STATIC_CHECK=staticcheck

EXE_NAME=telschema

.DEFAULT_GOAL := all

.PHONY: all
all: build-all test

.PHONY: build-all
build-all: build-scheck

.PHONY: build-scheck
build-scheck:
	go build -o bin/scheck ./cmd/scheck

.PHONY: install-tools
install-tools:
	go install honnef.co/go/tools/cmd/staticcheck

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
