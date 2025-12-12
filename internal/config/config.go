package config

import (
	"os"
	"strconv"
	"strings"
)

type Settings struct {
	GoldTargetPct  float64
	BtcTargetPct   float64
	StockTargetPct float64

	GoldDCA  float64
	BtcDCA   float64
	StockDCA float64

	GoldExtraBuyDropPct   float64
	GoldTakeProfitGainPct float64
	BtcExtraBuyDropPct    float64
	BtcTakeProfitGainPct  float64
	StockBuyDropPct       float64
	StockTakeProfitPct    float64
	GoldExtraBuyAmount    float64
	BtcExtraBuyAmount     float64

	XiitTicker string

	GoldAvgPrice  float64
	BtcAvgPrice   float64
	XiitAvgPrice  float64
	GoldValueNow  float64
	BtcValueNow   float64
	StockValueNow float64

	GoldAPIToken string

	SMTPHost string
	SMTPPort int
	SMTPUser string
	SMTPPass string
	SMTPFrom string
	SMTPTo   []string

	OpenAIKey     string
	OpenAIBaseURL string
	OpenAIModel   string
}

func Load() Settings {
	s := Settings{
		GoldTargetPct:         0.40,
		BtcTargetPct:          0.20,
		StockTargetPct:        0.40,
		GoldDCA:               600000,
		BtcDCA:                300000,
		StockDCA:              600000,
		GoldExtraBuyDropPct:   5,
		GoldTakeProfitGainPct: 8,
		BtcExtraBuyDropPct:    10,
		BtcTakeProfitGainPct:  20,
		StockBuyDropPct:       10,
		StockTakeProfitPct:    20,
		GoldExtraBuyAmount:    100000,
		BtcExtraBuyAmount:     100000,
		XiitTicker:            "XIIT.JK",
		SMTPPort:              587,
		OpenAIBaseURL:         "https://api.openai.com/v1",
		OpenAIModel:           "gpt-4o-mini",
	}

	readFloat("GOLD_TARGET_PCT", &s.GoldTargetPct)
	readFloat("BTC_TARGET_PCT", &s.BtcTargetPct)
	readFloat("STOCK_TARGET_PCT", &s.StockTargetPct)

	readFloat("GOLD_DCA", &s.GoldDCA)
	readFloat("BTC_DCA", &s.BtcDCA)
	readFloat("STOCK_DCA", &s.StockDCA)

	readFloat("GOLD_EXTRA_BUY_DROP_PCT", &s.GoldExtraBuyDropPct)
	readFloat("GOLD_TAKE_PROFIT_GAIN_PCT", &s.GoldTakeProfitGainPct)
	readFloat("BTC_EXTRA_BUY_DROP_PCT", &s.BtcExtraBuyDropPct)
	readFloat("BTC_TAKE_PROFIT_GAIN_PCT", &s.BtcTakeProfitGainPct)
	readFloat("STOCK_BUY_DROP_PCT", &s.StockBuyDropPct)
	readFloat("STOCK_TAKE_PROFIT_PCT", &s.StockTakeProfitPct)

	readFloat("GOLD_EXTRA_BUY_AMOUNT", &s.GoldExtraBuyAmount)
	readFloat("BTC_EXTRA_BUY_AMOUNT", &s.BtcExtraBuyAmount)

	readString("XIIT_TICKER", &s.XiitTicker)

	readFloat("GOLD_AVG_PRICE", &s.GoldAvgPrice)
	readFloat("BTC_AVG_PRICE", &s.BtcAvgPrice)
	readFloat("XIIT_AVG_PRICE", &s.XiitAvgPrice)
	readFloat("GOLD_VALUE_NOW", &s.GoldValueNow)
	readFloat("BTC_VALUE_NOW", &s.BtcValueNow)
	readFloat("STOCK_VALUE_NOW", &s.StockValueNow)

	readString("SMTP_HOST", &s.SMTPHost)
	readInt("SMTP_PORT", &s.SMTPPort)
	readString("SMTP_USER", &s.SMTPUser)
	readString("SMTP_PASS", &s.SMTPPass)
	readString("SMTP_FROM", &s.SMTPFrom)
	if v := os.Getenv("SMTP_TO"); v != "" {
		s.SMTPTo = splitAndTrim(v)
	}

	readString("GOLD_API_TOKEN", &s.GoldAPIToken)

	readString("OPENAI_API_KEY", &s.OpenAIKey)
	readString("OPENAI_BASE_URL", &s.OpenAIBaseURL)
	readString("OPENAI_MODEL", &s.OpenAIModel)

	return s
}

func readFloat(name string, target *float64) {
	if v := os.Getenv(name); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			*target = f
		}
	}
}

func readInt(name string, target *int) {
	if v := os.Getenv(name); v != "" {
		if val, err := strconv.Atoi(v); err == nil {
			*target = val
		}
	}
}

func readString(name string, target *string) {
	if v := os.Getenv(name); v != "" {
		*target = v
	}
}

func splitAndTrim(v string) []string {
	parts := strings.Split(v, ",")
	var cleaned []string
	for _, p := range parts {
		t := strings.TrimSpace(p)
		if t != "" {
			cleaned = append(cleaned, t)
		}
	}
	return cleaned
}
