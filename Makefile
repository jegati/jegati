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
dev: secrets
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

.PHONY: secrets test-store
secrets:
	python3 scripts/init-secrets.py
test-store: secrets
	bash scripts/test-store.sh

.PHONY: browser-install test-browser
browser-install:
	cd web && npx playwright install chromium
test-browser:
	python3 scripts/lab.py --output "$(or $(OUTPUT),reports/local/browser-$(shell date -u +%Y%m%dT%H%M%S))" -- npm --prefix web exec -- playwright test --config web/playwright.config.ts

.PHONY: simulate
simulate:
	SCENARIO="$(or $(SCENARIO),tirana-evening)" SEED="$(SEED)" bash scripts/simulate.sh

.PHONY: privacy-probe
# Reproducing the counterexample is deliberately a failing privacy gate.
privacy-probe:
	GATI_PRIVACY_PROBE=1 SCENARIO=tirana-evening SEED=42 bash scripts/simulate.sh

.PHONY: simulate-population
simulate-population:
	SCENARIO="$(or $(SCENARIO),tirana-population)" SEED="$(SEED)" SIM_CONFIG="$(or $(SIM_CONFIG),config/gati.yaml)" OUTPUT="$(or $(OUTPUT),reports/local/tirana-population-$(shell date -u +%Y%m%dT%H%M%S)-seed-$(or $(SEED),scenario))" INTEGRATION_CHECKS=0 bash scripts/simulate.sh

.PHONY: check-population simulate-suite
check-population:
	node scripts/check-population.mjs "$(REPORT)"
simulate-suite:
	python3 scripts/run-simulation-suite.py --config "$(CONFIG)" --output "$(or $(OUTPUT),reports/local/tirana-suite-$(shell date -u +%Y%m%dT%H%M%S))"

.PHONY: push-keys dev-push security-check
push-keys: secrets
	go run ./cmd/gati-push-keys
dev-push: push-keys
	$(COMPOSE) -f compose.yaml -f compose.push.yaml up --build
security-check:
	go run golang.org/x/vuln/cmd/govulncheck@v1.8.0 ./...
	npm --prefix web audit --audit-level=low

.PHONY: test-fuzz test-monitor
test-fuzz:
	bash scripts/fuzz.sh
test-monitor:
	go test -race ./internal/monitor
	python3 -m unittest discover -s scripts -p 'test_monitor.py'

LOAD_SIZES ?= 1000,10000,100000
.PHONY: benchmark-scaling
benchmark-scaling:
	OUTPUT="$(or $(OUTPUT),reports/local/scaling-$(shell date -u +%Y%m%dT%H%M%S))" bash scripts/benchmark-scaling.sh

.PHONY: load
load:
	python3 scripts/load.py --output "$(or $(OUTPUT),reports/local/load-$(shell date -u +%Y%m%dT%H%M%S))" --sizes "$(or $(SIZES),$(LOAD_SIZES))" --seconds "$(or $(SECONDS),20)" --distribution "$(or $(DISTRIBUTION),uniform)" $(if $(BURST),--burst,) $(if $(MIXED),--mixed,)

.PHONY: test-full-journey
test-full-journey:
	python3 scripts/lab.py --output "$(or $(OUTPUT),reports/local/full-journey-$(shell date -u +%Y%m%dT%H%M%S))" -- npm --prefix web exec -- playwright test --config web/playwright.full.config.ts

.PHONY: test-recovery
test-recovery:
	python3 scripts/recovery.py --output "$(or $(OUTPUT),reports/local/recovery-$(shell date -u +%Y%m%dT%H%M%S))"

.PHONY: test-soak
test-soak: secrets
	mkdir -p "$(or $(OUTPUT),reports/local/soak)"
	GATI_SOAK_SECONDS="$(or $(SOAK_SECONDS),900)" GATI_SOAK_REPORT="$(abspath $(or $(OUTPUT),reports/local/soak))/soak.json" bash scripts/test-store.sh

.PHONY: test-browser-matrix
test-browser-matrix:
	GATI_BROWSER=firefox python3 scripts/lab.py --output "$(or $(OUTPUT),reports/local/matrix-$(shell date -u +%Y%m%dT%H%M%S))/firefox" -- npm --prefix web exec -- playwright test --config web/playwright.matrix.config.ts
	GATI_BROWSER=webkit python3 scripts/lab.py --output "$(or $(OUTPUT),reports/local/matrix-$(shell date -u +%Y%m%dT%H%M%S))/webkit" -- npm --prefix web exec -- playwright test --config web/playwright.matrix.config.ts

.PHONY: check-load reproduce-build
check-load:
	python3 scripts/check-load.py "$(REPORT)"
reproduce-build:
	python3 scripts/reproduce-build.py --output "$(or $(OUTPUT),reports/local/reproduction-$(shell date -u +%Y%m%dT%H%M%S))"

.PHONY: test-pressure
test-pressure:
	python3 scripts/pressure.py --output "$(or $(OUTPUT),reports/local/pressure-$(shell date -u +%Y%m%dT%H%M%S))"

.PHONY: production-build test-deployment
production-build:
	$(COMPOSE) -f compose.production.yaml config --quiet
	$(COMPOSE) -f compose.production.yaml build api web
test-deployment: production-build
	python3 scripts/deployment-lab.py --output "$(or $(OUTPUT),reports/local/deployment-$(shell date -u +%Y%m%dT%H%M%S))"
