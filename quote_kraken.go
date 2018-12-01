/*
Package quote is free quote downloader library and cli

Downloads daily/weekly/monthly historical price quotes from Yahoo
and daily/intraday data from Google/Tiingo/Bittrex/Binance/Huobi/Kraken

Copyright 2018 Jesse Kuang
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
	"strconv"
	"time"
)

func getKrakenMarket(market, rawdata string) ([]string, error) {

	type Symbol struct {
		BaseAsset       string `json:"base"`
		QuoteAsset      string `json:"quote"`
		PricePrecision  int    `json:"pair_decimals"`
		AmountPrecision int    `json:"lot_decimals"`
	}

	type Markets struct {
		Symbols map[string]Symbol `json:"result"`
	}

	var markets Markets
	err := json.Unmarshal([]byte(rawdata), &markets)
	if err != nil {
		Log.Println(err)
		return nil, err
	}

	var symbols []string
	for symbol, mkt := range markets.Symbols {
		fmt.Printf("Pair: %s, PriceDigits: %d, AmoutDigits: %d\n",
			symbol, mkt.PricePrecision, mkt.AmountPrecision)
		symbols = append(symbols, symbol)
	}

	return symbols, err
}

// NewQuoteFromKraken - Kraken historical prices for a symbol
func NewQuoteFromKraken(symbol string, period Period) (Quote, error) {

	var interval string

	switch period {
	case Min1:
		interval = "1"
	case Min5:
		interval = "5"
	case Min15:
		interval = "15"
	case Min30:
		interval = "30"
	case Min60:
		interval = "60"
	case Hour4:
		interval = "240"
	case Daily:
		interval = "1440"
	case Weekly:
		interval = "10080"
	default:
		interval = "1440"
	}

	var quote Quote
	quote.Symbol = symbol

	// kraken id used for continue download, no use "since" param

	url := fmt.Sprintf(
		"https://api.kraken.com/0/public/OHLC?pair=%s&interval=%s",
		symbol, interval)
	//Log.Println(url)
	client := &http.Client{Timeout: ClientTimeout}
	req, _ := http.NewRequest("GET", url, nil)
	resp, err := client.Do(req)

	if err != nil {
		Log.Printf("Kraken OHLC error: %v\n", err)
		return NewQuote("", 0), err
	}
	defer resp.Body.Close()

	contents, _ := ioutil.ReadAll(resp.Body)

	type OHLC [8]interface{}
	type Result struct {
		Error  []interface{} `json:"error"`
		Result interface{}   `json:"result"`
		LastId int64         `json:"last"`
	}

	var result Result

	err = json.Unmarshal(contents, &result)
	if err != nil {
		Log.Printf("kraken error: %v\n", err)
		return NewQuote("", 0), err
	}

	Bars := result.Result.(map[string]interface{})
	var iBars []interface{}
	if iibars, ok := Bars[symbol]; !ok {
		Log.Println("kraken not found symbol")
		return NewQuote("", 0), nil
	} else {
		iBars = iibars.([]interface{})
	}
	numrows := len(iBars)
	q := NewQuote(symbol, numrows)
	//Log.Printf("numrows=%d, bars=%v\n", numrows, iBars)

	/*
		0		OpenTime		int64
		1 		Open			float64
		2 		High			float64
		3		Low				float64
		4 		Close			float64
		5 		vwap			float64
		6 		Vol				float64
		7 		Count			int64
	*/

	for bar := 0; bar < numrows; bar++ {
		bars := iBars[bar].([]interface{})
		q.Date[bar] = time.Unix(int64(bars[0].(float64)), 0).UTC()
		q.Open[bar], _ = strconv.ParseFloat(bars[1].(string), 64)
		q.High[bar], _ = strconv.ParseFloat(bars[2].(string), 64)
		q.Low[bar], _ = strconv.ParseFloat(bars[3].(string), 64)
		q.Close[bar], _ = strconv.ParseFloat(bars[4].(string), 64)
		q.Volume[bar], _ = strconv.ParseFloat(bars[6].(string), 64)
	}
	quote.Date = append(quote.Date, q.Date...)
	quote.Open = append(quote.Open, q.Open...)
	quote.High = append(quote.High, q.High...)
	quote.Low = append(quote.Low, q.Low...)
	quote.Close = append(quote.Close, q.Close...)
	quote.Volume = append(quote.Volume, q.Volume...)

	return quote, nil
}

// NewQuotesFromKraken - create a list of prices from symbols in file
func NewQuotesFromKraken(filename string, period Period) (Quotes, error) {
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
		quote, err := NewQuoteFromKraken(sym, period)
		if err == nil {
			quotes = append(quotes, quote)
		} else {
			Log.Println("error downloading " + sym)
		}
		time.Sleep(Delay * time.Millisecond)
	}
	return quotes, nil
}

// NewQuotesFromKrakenSyms - create a list of prices from symbols in string array
func NewQuotesFromKrakenSyms(symbols []string, period Period) (Quotes, error) {

	quotes := Quotes{}
	for _, symbol := range symbols {
		quote, err := NewQuoteFromKraken(symbol, period)
		if err == nil {
			quotes = append(quotes, quote)
		} else {
			Log.Println("error downloading " + symbol)
		}
		time.Sleep(Delay * time.Millisecond)
	}
	return quotes, nil
}
