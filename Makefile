SHELL := /bin/bash
.SHELLFLAGS := -eu -o pipefail -c
include scripts/toolchain/versions.env
GATI_TOOLS_DIR ?= $(HOME)/.local/share/gati/tools
export PATH := $(GATI_TOOLS_DIR)/go-$(GATI_GO_VERSION)/bin:$(GATI_TOOLS_DIR)/node-$(GATI_NODE_VERSION)/bin:$(GATI_TOOLS_DIR)/compose-$(GATI_COMPOSE_VERSION):$(PATH)
CONFIG ?= config/gati.yaml
COMPOSE ?= docker-compose

.PHONY: bootstrap doctor deps test build config-check config-show config-check-simulation dev dev-native down compose-check verify-local
bootstrap:
	bash scripts/bootstrap.sh
doctor:
	bash scripts/doctor.sh
deps:
	go mod download
	npm --prefix web ci --ignore-scripts --no-audit --no-fund
test:
	go test -race ./...
	go test -tags simulation ./...
	go vet ./...
	npm --prefix web run check
build:
	mkdir -p bin
	CGO_ENABLED=0 go build -trimpath -buildvcs=false -o bin/gati ./cmd/gati
	npm --prefix web run build
config-check:
	go run ./cmd/gati -mode config-check -config "$(CONFIG)"
config-show:
	go run ./cmd/gati -mode config-show -config "$(CONFIG)"
config-check-simulation:
	go run -tags simulation ./cmd/gati -mode config-check -config config/simulation.yaml
dev:
	$(COMPOSE) -f compose.yaml up --build
dev-native:
	bash scripts/dev-native.sh
down:
	$(COMPOSE) -f compose.yaml down
compose-check:
	$(COMPOSE) -f compose.yaml config --quiet
verify-local: test build config-check config-check-simulation compose-check
	bash scripts/smoke.sh

.PHONY: check-containers
check-containers:
	COMPOSE="$(COMPOSE)" node scripts/check-containers.mjs

.PHONY: map-import map-check
map-import:
	go run ./cmd/map-import
map-check:
	bash scripts/check-map.sh
