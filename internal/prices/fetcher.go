package prices

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
)

type Quote struct {
	Symbol      string
	Price       float64
	High30      float64
	Currency    string
	Source      string
	RetrievedAt time.Time
}

type Fetcher struct {
	Client *http.Client
}

func (f Fetcher) FetchGold(ctx context.Context) (Quote, error) {
	client := f.httpClient()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.metals.live/v1/spot/gold", nil)
	if err != nil {
		return Quote{}, err
	}
	req.Header.Set("User-Agent", "currency-report/0.1")

	resp, err := client.Do(req)
	if err != nil {
		return Quote{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return Quote{}, fmt.Errorf("gold api status %d", resp.StatusCode)
	}

	var payload []interface{}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return Quote{}, err
	}

	price, err := extractNumber(payload)
	if err != nil {
		return Quote{}, fmt.Errorf("gold price unavailable: %w", err)
	}

	return Quote{
		Symbol:      "XAU",
		Price:       price,
		High30:      price,
		Currency:    "USD",
		Source:      "metals.live",
		RetrievedAt: time.Now(),
	}, nil
}

func (f Fetcher) FetchBTC(ctx context.Context) (Quote, error) {
	client := f.httpClient()
	url := "https://api.coingecko.com/api/v3/coins/bitcoin/market_chart?vs_currency=idr&days=30&interval=daily"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return Quote{}, err
	}
	req.Header.Set("User-Agent", "currency-report/0.1")

	resp, err := client.Do(req)
	if err != nil {
		return Quote{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return Quote{}, fmt.Errorf("btc api status %d", resp.StatusCode)
	}

	var payload struct {
		Prices [][]float64 `json:"prices"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return Quote{}, err
	}
	if len(payload.Prices) == 0 {
		return Quote{}, errors.New("btc price list empty")
	}

	var price float64
	var high float64
	for _, entry := range payload.Prices {
		if len(entry) < 2 {
			continue
		}
		p := entry[1]
		price = p
		if p > high {
			high = p
		}
	}

	return Quote{
		Symbol:      "BTC",
		Price:       price,
		High30:      high,
		Currency:    "IDR",
		Source:      "coingecko",
		RetrievedAt: time.Now(),
	}, nil
}

func (f Fetcher) FetchXiit(ctx context.Context, ticker string) (Quote, error) {
	client := f.httpClient()
	url := fmt.Sprintf("https://query1.finance.yahoo.com/v8/finance/chart/%s?range=1mo&interval=1d", ticker)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return Quote{}, err
	}
	req.Header.Set("User-Agent", "currency-report/0.1")

	resp, err := client.Do(req)
	if err != nil {
		return Quote{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return Quote{}, fmt.Errorf("xiit api status %d", resp.StatusCode)
	}

	var payload yahooChartResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return Quote{}, err
	}
	if len(payload.Chart.Result) == 0 {
		return Quote{}, errors.New("xiit result empty")
	}

	meta := payload.Chart.Result[0].Meta
	var price float64
	if meta.RegularMarketPrice != nil {
		price = *meta.RegularMarketPrice
	}
	quote := payload.Chart.Result[0].Indicators.Quote
	var high float64
	if len(quote) > 0 {
		for _, h := range quote[0].High {
			if h > high {
				high = h
			}
		}
		if len(quote[0].Close) > 0 {
			price = quote[0].Close[len(quote[0].Close)-1]
		}
	}

	return Quote{
		Symbol:      ticker,
		Price:       price,
		High30:      high,
		Currency:    meta.Currency,
		Source:      "yahoo-finance",
		RetrievedAt: time.Now(),
	}, nil
}

func (f Fetcher) httpClient() *http.Client {
	if f.Client != nil {
		return f.Client
	}
	return &http.Client{Timeout: 12 * time.Second}
}

func extractNumber(payload []interface{}) (float64, error) {
	for _, entry := range payload {
		switch v := entry.(type) {
		case float64:
			if v > 0 {
				return v, nil
			}
		case []interface{}:
			for _, inner := range v {
				if f, ok := inner.(float64); ok && f > 0 {
					return f, nil
				}
			}
		case map[string]interface{}:
			for _, inner := range v {
				if f, ok := inner.(float64); ok && f > 0 {
					return f, nil
				}
			}
		}
	}
	return 0, errors.New("no numeric value found")
}

type yahooChartResponse struct {
	Chart struct {
		Result []struct {
			Meta struct {
				Currency           string   `json:"currency"`
				Symbol             string   `json:"symbol"`
				RegularMarketPrice *float64 `json:"regularMarketPrice"`
			} `json:"meta"`
			Indicators struct {
				Quote []struct {
					Close []float64 `json:"close"`
					High  []float64 `json:"high"`
				} `json:"quote"`
			} `json:"indicators"`
		} `json:"result"`
		Error interface{} `json:"error"`
	} `json:"chart"`
}
