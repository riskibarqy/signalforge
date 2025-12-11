# Simple helpers for local dev and Fly.io deploys

GO_RUN=GO111MODULE=on go run ./cmd/app
FLY_SECRET_VARS=SMTP_HOST SMTP_PORT SMTP_USER SMTP_PASS SMTP_FROM SMTP_TO OPENAI_API_KEY

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
	@if [ ! -f .env ]; then echo '.env not found; create it first'; exit 1; fi
	@echo 'Setting Fly secrets from .env'; \
	set -a; source .env; set +a; \
	fly secrets set $(foreach var,$(FLY_SECRET_VARS),$(var)=$${$(var)})
