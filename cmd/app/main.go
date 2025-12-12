package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/riskibarqy/signalforge/internal/ai"
	"github.com/riskibarqy/signalforge/internal/config"
	"github.com/riskibarqy/signalforge/internal/email"
	"github.com/riskibarqy/signalforge/internal/prices"
	"github.com/riskibarqy/signalforge/internal/workflow"
)

func main() {
	ctx := context.Background()
	mode := flag.String("mode", "daily", "dca|daily|rebalance")
	skipEmail := flag.Bool("no-email", false, "do not send email, just print")
	goldValue := flag.Float64("gold_value", 0, "override gold portfolio value (IDR) for rebalance")
	btcValue := flag.Float64("btc_value", 0, "override btc portfolio value (IDR) for rebalance")
	stockValue := flag.Float64("stock_value", 0, "override stock portfolio value (IDR) for rebalance")
	flag.Parse()

	settings := config.Load()
	if *goldValue > 0 {
		settings.GoldValueNow = *goldValue
	}
	if *btcValue > 0 {
		settings.BtcValueNow = *btcValue
	}
	if *stockValue > 0 {
		settings.StockValueNow = *stockValue
	}

	httpClient := &http.Client{Timeout: 15 * time.Second}
	fetcher := prices.Fetcher{
		Client:       httpClient,
		GoldAPIToken: settings.GoldAPIToken,
	}

	var report workflow.Report
	var err error

	switch *mode {
	case "dca":
		report = workflow.MonthlyDCA(settings)
	case "daily":
		goldQuote, errGold := fetcher.FetchGold(ctx)
		btcQuote, errBTC := fetcher.FetchBTC(ctx)
		xiitQuote, errXiit := fetcher.FetchXiit(ctx, settings.XiitTicker)
		if errGold != nil || errBTC != nil || errXiit != nil {
			log.Printf("fetch errors: gold=%v btc=%v xiit=%v", errGold, errBTC, errXiit)
		}
		report = workflow.DailySignals(settings, goldQuote, btcQuote, xiitQuote)
	case "rebalance":
		report, err = workflow.MonthlyRebalance(settings)
	default:
		err = fmt.Errorf("unknown mode %q", *mode)
	}

	if err != nil {
		log.Fatalf("error: %v", err)
	}

	aiClient := ai.Client{
		BaseURL: settings.OpenAIBaseURL,
		APIKey:  settings.OpenAIKey,
		Model:   settings.OpenAIModel,
		Client:  httpClient,
	}
	if summary, err := aiClient.Summarize(ctx, report.Body); err == nil && summary != "" {
		report.Body += "\n\nAI Notes:\n" + summary
	} else if err != nil && settings.OpenAIKey != "" {
		log.Printf("ai summarize: %v", err)
	}

	fmt.Println(report.Subject)
	fmt.Println("")
	fmt.Println(report.Body)

	if *skipEmail {
		return
	}

	if settings.SMTPHost == "" || settings.SMTPFrom == "" || len(settings.SMTPTo) == 0 {
		log.Println("smtp config missing; skipping email")
		return
	}

	mailCfg := email.SMTPConfig{
		Host: settings.SMTPHost,
		Port: settings.SMTPPort,
		User: settings.SMTPUser,
		Pass: settings.SMTPPass,
		From: settings.SMTPFrom,
		To:   settings.SMTPTo,
	}
	if err := email.Send(ctx, mailCfg, report.Subject, report.Body); err != nil {
		log.Printf("email send failed: %v", err)
		os.Exit(1)
	}
}
