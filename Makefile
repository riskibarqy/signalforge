# Simple helpers for local dev and Fly.io deploys

SHELL := /bin/sh
.ONESHELL:

GO_RUN=GO111MODULE=on go run ./cmd/app
FLY_APP ?= signalforge-late-fire-4638
ENV_FILE ?= $(abspath .env)
FLY_SECRET_VARS=SMTP_HOST SMTP_PORT SMTP_USER SMTP_PASS SMTP_FROM SMTP_TO GOLD_API_TOKEN OPENAI_API_KEY

.PHONY: daily dca rebalance test fmt fly-init fly-deploy fly-secrets

daily:
	$(GO_RUN) -mode daily

dca:
	$(GO_RUN) -mode dca

rebalance:
	$(GO_RUN) -mode rebalance

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
