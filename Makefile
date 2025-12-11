# Simple helpers for local dev and Fly.io deploys

GO_RUN=GO111MODULE=on go run ./cmd/app

.PHONY: daily dca rebalance test fmt fly-init fly-deploy

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
