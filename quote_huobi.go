/*
Package quote is free quote downloader library and cli

Downloads daily/weekly/monthly historical price quotes from Yahoo
and daily/intraday data from Google/Tiingo/Bittrex/Binance/Huobi

Copyright 2017 Mark Chenoweth
Licensed under terms of MIT license (see LICENSE)
*/
package quote

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"
)

func getHuobiMarket(market, rawdata string) ([]string, error) {

	type Symbol struct {
		BaseAsset       string `json:"base-currency"`
		QuoteAsset      string `json:"quote-currency"`
		PricePrecision  int    `json:"price-precision"`
		AmountPrecision int    `json:"amount-precision"`
		Partition       string `json:"symbol-partition"`
		Symbol          string `json:"symbol"`
	}

	type Markets struct {
		Status  string   `json:"status"`
		Symbols []Symbol `json:"data"`
	}

	var markets Markets
	err := json.Unmarshal([]byte(rawdata), &markets)
	if err != nil {
		fmt.Println(err)
	}

	var symbols []string
	for _, mkt := range markets.Symbols {
		if strings.HasSuffix(market, "ht") && mkt.QuoteAsset == "ht" {
			symbols = append(symbols, mkt.Symbol)
		} else if strings.HasSuffix(market, "btc") && mkt.QuoteAsset == "btc" {
			symbols = append(symbols, mkt.Symbol)
		} else if strings.HasSuffix(market, "eth") && mkt.QuoteAsset == "eth" {
			symbols = append(symbols, mkt.Symbol)
		} else if strings.HasSuffix(market, "usdt") && mkt.QuoteAsset == "usdt" {
			symbols = append(symbols, mkt.Symbol)
		}
	}

	return symbols, err
}

// NewQuoteFromHuobi - Huobi historical prices for a symbol
func NewQuoteFromHuobi(symbol string, period Period) (Quote, error) {

	var interval string

	switch period {
	case Min1:
		interval = "1min"
	case Min5:
		interval = "5min"
	case Min15:
		interval = "15min"
	case Min30:
		interval = "30min"
	case Min60:
		interval = "60min"
	case Daily:
		interval = "1day"
	case Weekly:
		interval = "1week"
	case Monthly:
		interval = "1mon"
	default:
		interval = "1day"
	}

	var quote Quote
	quote.Symbol = symbol

	maxBars := 1990

	url := fmt.Sprintf(
		"https://api.huobi.br.com/market/history/kline?symbol=%s&period=%s&size=%d",
		//"https://api.huobipro.com/market/history/kline?symbol=%s&period=%s&size=%d",
		symbol,
		interval,
		maxBars)
	//log.Println(url)
	client := &http.Client{Timeout: ClientTimeout}
	req, _ := http.NewRequest("GET", url, nil)
	resp, err := client.Do(req)

	if err != nil {
		Log.Printf("huobi error: %v\n", err)
		return NewQuote("", 0), err
	}
	defer resp.Body.Close()

	contents, _ := ioutil.ReadAll(resp.Body)

	type OHLC struct {
		Id     int64
		Open   float64
		Close  float64
		Low    float64
		High   float64
		Amount float64
		Vol    float64
		Count  int64
	}
	type Result struct {
		Status string `json:"status"`
		Ch     string `json:"ch"`
		Ts     int64  `json:"ts"`
		OHLC   []OHLC `json:"data"`
	}

	var result Result

	err = json.Unmarshal(contents, &result)
	if err != nil {
		Log.Printf("huobi error: %v\n", err)
	}

	bars := result.OHLC
	numrows := len(bars)
	q := NewQuote(symbol, numrows)
	//Log.Printf("numrows=%d, bars=%v\n", numrows, bars)

	/*
		0       OpenTime                 int64
		1 			Open                     float64
		2 			Close                    float64
		3		 	Low                      float64
		4 			High                     float64
		5 			Amount         float64
		6 			Vol                   float64
		7 			Count                int64
	*/

	for bar := 0; bar < numrows; bar++ {
		q.Date[numrows-1-bar] = time.Unix(bars[bar].Id, 0)
		q.Open[numrows-1-bar] = bars[bar].Open
		q.High[numrows-1-bar] = bars[bar].High
		q.Low[numrows-1-bar] = bars[bar].Low
		q.Close[numrows-1-bar] = bars[bar].Close
		q.Volume[numrows-1-bar] = bars[bar].Vol
	}
	quote.Date = append(quote.Date, q.Date...)
	quote.Open = append(quote.Open, q.Open...)
	quote.High = append(quote.High, q.High...)
	quote.Low = append(quote.Low, q.Low...)
	quote.Close = append(quote.Close, q.Close...)
	quote.Volume = append(quote.Volume, q.Volume...)

	return quote, nil
}

// NewQuotesFromHuobi - create a list of prices from symbols in file
func NewQuotesFromHuobi(filename string, period Period) (Quotes, error) {
	quotes := Quotes{}
	inFile, err := os.Open(filename)
	if err != nil {
		return quotes, err
	}
	defer inFile.Close()
	scanner := bufio.NewScanner(inFile)
	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {
		sym := scanner.Text()
		quote, err := NewQuoteFromHuobi(sym, period)
		if err == nil {
			quotes = append(quotes, quote)
		} else {
			Log.Println("error downloading " + sym)
		}
		time.Sleep(Delay * time.Millisecond)
	}
	return quotes, nil
}

// NewQuotesFromHuobiSyms - create a list of prices from symbols in string array
func NewQuotesFromHuobiSyms(symbols []string, period Period) (Quotes, error) {

	quotes := Quotes{}
	for _, symbol := range symbols {
		quote, err := NewQuoteFromHuobi(symbol, period)
		if err == nil {
			quotes = append(quotes, quote)
		} else {
			Log.Println("error downloading " + symbol)
		}
		time.Sleep(Delay * time.Millisecond)
	}
	return quotes, nil
}
