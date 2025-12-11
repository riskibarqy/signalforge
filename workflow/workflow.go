package workflow

import (
	"fmt"
	"math"
	"strings"

	"currency-report/internal/config"
	"currency-report/internal/prices"
)

type Report struct {
	Subject string
	Body    string
}

func MonthlyDCA(s config.Settings) Report {
	body := fmt.Sprintf(
		"Monthly DCA Plan (Pluang/Bibit)\n"+
			"- Gold: IDR %s\n"+
			"- Bitcoin: IDR %s\n"+
			"- XIIT: IDR %s\n",
		formatMoney(s.GoldDCA),
		formatMoney(s.BtcDCA),
		formatMoney(s.StockDCA),
	)
	return Report{
		Subject: "Monthly DCA Plan",
		Body:    body,
	}
}

func DailySignals(s config.Settings, gold, btc, xiit prices.Quote) Report {
	var lines []string
	lines = append(lines, "Daily Signals")
	lines = append(lines, fmt.Sprintf("Gold: %.2f %s (30d high %.2f)", gold.Price, gold.Currency, safeHigh(gold.High30)))
	lines = append(lines, fmt.Sprintf("BTC: %.2f %s (30d high %.2f)", btc.Price, btc.Currency, safeHigh(btc.High30)))
	lines = append(lines, fmt.Sprintf("XIIT: %.2f %s (30d high %.2f)", xiit.Price, currencyOrIDR(xiit.Currency), safeHigh(xiit.High30)))
	lines = append(lines, "")

	goldDrop := dropPct(gold.Price, gold.High30)
	btcDrop := dropPct(btc.Price, btc.High30)
	goldGain := gainPct(gold.Price, s.GoldAvgPrice)
	btcGain := gainPct(btc.Price, s.BtcAvgPrice)
	xiitGain := gainPct(xiit.Price, s.XiitAvgPrice)

	goldExtra := 0.0
	if goldDrop >= s.GoldExtraBuyDropPct {
		goldExtra = s.GoldExtraBuyAmount
	}
	goldSellPct := 0.0
	if s.GoldAvgPrice > 0 && goldGain >= s.GoldTakeProfitGainPct {
		goldSellPct = 10
	}

	btcExtra := 0.0
	if btcDrop >= s.BtcExtraBuyDropPct {
		btcExtra = s.BtcExtraBuyAmount
	}
	btcSellPct := 0.0
	if s.BtcAvgPrice > 0 && btcGain >= s.BtcTakeProfitGainPct {
		btcSellPct = 10
	}

	stockSignal := "Hold"
	if s.XiitAvgPrice > 0 {
		if xiitGain <= -s.StockBuyDropPct {
			stockSignal = "Optional Buy"
		} else if xiitGain >= s.StockTakeProfitPct {
			stockSignal = "Consider selling 10%"
		}
	} else {
		stockSignal = "Set XIIT average price to enable signals"
	}

	lines = append(lines, "Signals:")
	lines = append(lines, fmt.Sprintf("- Gold drop: %.2f%%, gain: %s, extra buy: IDR %s, sell: %.0f%%",
		goldDrop, gainLabel(goldGain, s.GoldAvgPrice), formatMoney(goldExtra), goldSellPct))
	lines = append(lines, fmt.Sprintf("- BTC drop: %.2f%%, gain: %s, extra buy: IDR %s, sell: %.0f%%",
		btcDrop, gainLabel(btcGain, s.BtcAvgPrice), formatMoney(btcExtra), btcSellPct))
	lines = append(lines, fmt.Sprintf("- XIIT gain vs avg: %s -> %s", gainLabel(xiitGain, s.XiitAvgPrice), stockSignal))

	return Report{
		Subject: "Daily Investment Signals",
		Body:    strings.Join(lines, "\n"),
	}
}

func MonthlyRebalance(s config.Settings) (Report, error) {
	total := s.GoldValueNow + s.BtcValueNow + s.StockValueNow
	if total <= 0 {
		return Report{}, fmt.Errorf("set GOLD_VALUE_NOW, BTC_VALUE_NOW, STOCK_VALUE_NOW to rebalance")
	}

	goldPct := s.GoldValueNow / total
	btcPct := s.BtcValueNow / total
	stockPct := s.StockValueNow / total

	goldDiff := goldPct - s.GoldTargetPct
	btcDiff := btcPct - s.BtcTargetPct
	stockDiff := stockPct - s.StockTargetPct

	var advice []string
	if goldPct > s.GoldTargetPct+0.05 {
		advice = append(advice, "Reduce gold or pause buys this month.")
	}
	if btcPct < s.BtcTargetPct-0.05 {
		advice = append(advice, "Increase BTC DCA slightly.")
	}
	if stockPct < s.StockTargetPct-0.05 {
		advice = append(advice, "Increase XIIT allocation next month.")
	}
	if len(advice) == 0 {
		advice = append(advice, "Portfolio within bands. Hold course.")
	}

	body := fmt.Sprintf(
		"Monthly Rebalance\n"+
			"Total: IDR %s\n"+
			"Gold: %s (diff %.2f%%)\n"+
			"BTC: %s (diff %.2f%%)\n"+
			"XIIT: %s (diff %.2f%%)\n"+
			"Recommendations:\n- %s\n",
		formatMoney(total),
		formatPct(goldPct), diffToPct(goldDiff),
		formatPct(btcPct), diffToPct(btcDiff),
		formatPct(stockPct), diffToPct(stockDiff),
		strings.Join(advice, "\n- "),
	)

	return Report{
		Subject: "Monthly Rebalance",
		Body:    body,
	}, nil
}

func dropPct(price, high float64) float64 {
	if high <= 0 {
		return 0
	}
	return (high - price) / high * 100
}

func gainPct(price, avg float64) float64 {
	if price <= 0 || avg <= 0 {
		return 0
	}
	return (price - avg) / avg * 100
}

func gainLabel(pct float64, avg float64) string {
	if avg <= 0 {
		return "avg price not set"
	}
	return fmt.Sprintf("%.2f%%", pct)
}

func currencyOrIDR(cur string) string {
	if cur == "" {
		return "IDR"
	}
	return cur
}

func safeHigh(high float64) float64 {
	if high == 0 {
		return math.NaN()
	}
	return high
}

func formatMoney(v float64) string {
	return commaSeparated(int64(v))
}

func commaSeparated(v int64) string {
	sign := ""
	if v < 0 {
		sign = "-"
		v = -v
	}
	s := fmt.Sprintf("%d", v)
	var out []byte
	for i, c := range s {
		if i != 0 && (len(s)-i)%3 == 0 {
			out = append(out, ',')
		}
		out = append(out, byte(c))
	}
	return sign + string(out)
}

func formatPct(v float64) string {
	return fmt.Sprintf("%.2f%%", v*100)
}

func diffToPct(v float64) float64 {
	return v * 100
}
