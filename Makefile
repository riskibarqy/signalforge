# Simple helpers for local dev and Fly.io deploys

SHELL := /bin/sh
.ONESHELL:

BIN=./bin/signalforge
GO_BUILD_ENV=GO111MODULE=on
FLY_APP ?= signalforge-late-fire-4638
ENV_FILE ?= $(abspath .env)
FLY_SECRET_VARS=SMTP_HOST SMTP_PORT SMTP_USER SMTP_PASS SMTP_FROM SMTP_TO GOLD_API_TOKEN OPENAI_API_KEY

.PHONY: build daily dca rebalance test fmt fly-init fly-deploy fly-secrets

build: $(BIN)

$(BIN):
	@mkdir -p $(dir $(BIN))
	$(GO_BUILD_ENV) go build -o $(BIN) ./cmd/app

daily: $(BIN)
	@if [ ! -f $(ENV_FILE) ]; then echo "$(ENV_FILE) not found; create it or set ENV_FILE"; exit 1; fi
	set -a; . $(ENV_FILE); set +a; $(BIN) -mode daily

dca: $(BIN)
	@if [ ! -f $(ENV_FILE) ]; then echo "$(ENV_FILE) not found; create it or set ENV_FILE"; exit 1; fi
	set -a; . $(ENV_FILE); set +a; $(BIN) -mode dca

rebalance: $(BIN)
	@if [ ! -f $(ENV_FILE) ]; then echo "$(ENV_FILE) not found; create it or set ENV_FILE"; exit 1; fi
	set -a; . $(ENV_FILE); set +a; $(BIN) -mode rebalance

test:
	go test ./...

fmt:
	gofmt -w cmd internal

fly-init:
	@if [ ! -f fly.toml ]; then echo 'Creating minimal fly.toml'; \
	echo '[build]\n  builder = "paketobuildpacks/builder-jammy-base"\n\n[processes]\n  app = "signalforge -mode daily"\n' > fly.toml; \
	else echo 'fly.toml already exists'; fi

fly-deploy:
	fly deploy

fly-secrets:
	@if [ ! -f $(ENV_FILE) ]; then echo '$(ENV_FILE) not found; create it first'; exit 1; fi
	@echo "Using ENV_FILE=$(ENV_FILE)"
	@echo "Setting Fly secrets from $(ENV_FILE)"; \
	set -a; . $(ENV_FILE); set +a; \
	fly secrets set -a $(FLY_APP) $(foreach var,$(FLY_SECRET_VARS),$(var)="$${$(var)}")
