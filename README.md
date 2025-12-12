# Investment Autopilot

Go utility that pulls prices for gold, BTC, and XIIT, applies your DCA/rebalance rules, optionally asks an OpenAI-compatible model for a short briefing, and emails the result. Module path: `github.com/riskibarqy/signalforge`.

## Data sources
- Gold: goldapi.io XAU/IDR (token required)
- BTC (IDR) & 30d high: CoinGecko market_chart
- XIIT ETF (IDR) & 30d high: Yahoo Finance chart (`XIIT.JK` by default, override with `XIIT_TICKER`)

## Configuration (env)
- Targets/DCA: `GOLD_TARGET_PCT` `BTC_TARGET_PCT` `STOCK_TARGET_PCT` `GOLD_DCA` `BTC_DCA` `STOCK_DCA`
- Signals: `GOLD_EXTRA_BUY_DROP_PCT` `GOLD_TAKE_PROFIT_GAIN_PCT` `BTC_EXTRA_BUY_DROP_PCT` `BTC_TAKE_PROFIT_GAIN_PCT` `STOCK_BUY_DROP_PCT` `STOCK_TAKE_PROFIT_PCT` `GOLD_EXTRA_BUY_AMOUNT` `BTC_EXTRA_BUY_AMOUNT`
- Prices/averages: `GOLD_AVG_PRICE` `BTC_AVG_PRICE` `XIIT_AVG_PRICE` (set to enable gain-based rules)
- Rebalance values: `GOLD_VALUE_NOW` `BTC_VALUE_NOW` `STOCK_VALUE_NOW` (or use CLI flags)
- Symbols: `XIIT_TICKER` (default `XIIT.JK`)
- APIs: `GOLD_API_TOKEN` (required), `OPENAI_API_KEY` (optional), `OPENAI_BASE_URL` (optional), `OPENAI_MODEL` (default `gpt-4o-mini`)
- SMTP: `SMTP_HOST` `SMTP_PORT` (default 587) `SMTP_USER` `SMTP_PASS` `SMTP_FROM` `SMTP_TO` (comma-separated)

## Run locally
```bash
# Daily signals with email (if SMTP configured)
go run ./cmd/app -mode daily

# Monthly DCA checklist without email
go run ./cmd/app -mode dca -no-email

# Rebalance (override values via flags)
go run ./cmd/app -mode rebalance -gold_value 12000000 -btc_value 6000000 -stock_value 10000000
```

## Fly.io deployment (outline)
1) Uses the included `Dockerfile` (Go build -> distroless) and `fly.toml` with worker processes:
```toml
[build]
  dockerfile = "Dockerfile"

[processes]
  app = "-mode daily" # ENTRYPOINT is /signalforge
```
2) Set secrets: `fly secrets set SMTP_HOST=... SMTP_PORT=587 SMTP_USER=... SMTP_PASS=... SMTP_FROM=... SMTP_TO=... GOLD_API_TOKEN=... OPENAI_API_KEY=...`
3) Deploy: `PACK_VOLUME_KEY=signalforge-cache fly deploy` (env var optional but speeds rebuilds).
4) Schedule with Fly cron/Machines, e.g.:
   - Daily 01:00 UTC (08:00 WIB): `fly m run -a <app> --config fly.toml --schedule "cron:0 1 * * *" -- -mode daily`
   - Monthly 01:00 UTC on the 1st: `fly m run -a <app> --config fly.toml --schedule "cron:0 1 1 * *" -- -mode rebalance`

## Notes
- AI is optional; if the key is missing the report still runs.
- If SMTP is not set, the app prints to stdout.
- metals.live does not expose a free 30d high, so gold drop logic uses the latest price as the high (drop = 0) until you supply a custom value.
