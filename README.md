# go-quote

[![GoDoc](http://godoc.org/github.com/markcheno/go-quote?status.svg)](http://godoc.org/github.com/markcheno/go-quote)

A free quote downloader library and cli

Downloads daily historical price quotes from Tiingo and daily/intraday data from various api's. Written in pure Go. No external dependencies. Now downloads crypto coin historical data from various exchanges.

- Update: 04/01/2025 - added markets flag, wildcard input files, multiple inputs

- Update: 03/02/2025 - Removed obsolete Yahoo support

- Update: 02/15/2024 - Major update: updated to Go 1.22, removed bittrex/binance support, fixed nasdaq/tiingo markets

- Update: 11/15/2021 - Removed obsolete markets, converted to go modules

- Update: 7/18/2021 - Removed obsolete Google support

- Update: 6/26/2019 - updated GDAX to Coinbase, added coinbase market

- Update: 4/26/2018 - Added preliminary [tiingo](https://api.tiingo.com/) CRYPTO support. Use -source=tiingo-crypto -token=<your_tingo_token> You can also set env variable TIINGO_API_TOKEN. To get symbol lists, use market: tiingo-btc, tiingo-eth or tiingo-usd

- Update: 12/21/2017 - Added Amibroker format option (creates csv file with separate date and time). Use -format=ami

- Update: 12/20/2017 - Added [Binance](https://www.binance.com/trade.html) exchange support. Use -source=binance

- Update: 12/18/2017 - Added [Bittrex](https://bittrex.com/home/markets) exchange support. Use -source=bittrex

- Update: 10/21/2017 - Added Coinbase [GDAX](https://www.gdax.com/trade/BTC-USD) exchange support. Use -source=gdax All times are in UTC. Automatically rate limited.

- Update: 7/19/2017 - Added preliminary [tiingo](https://api.tiingo.com/) support. Use -source=tiingo -token=<your_tingo_token> You can also set env variable TIINGO_API_TOKEN

- Update: 5/24/2017 - Now works with the new Yahoo download format. Beware - Yahoo data quality is now questionable and the free Yahoo quotes are likely to permanently go away in the near future. Use with caution!

Still very much in alpha mode. Expect bugs and API changes. Comments/suggestions/pull requests welcome!

Copyright 2024 Mark Chenoweth

Install CLI utility (quote) with:

```bash
go install github.com/markcheno/go-quote/quote@latest
```

```
Usage:
  quote -h | -help
  quote -v | -version
  quote <market> [-output=<outputFile>]
  quote [-years=<years>|(-start=<datestr> [-end=<datestr>])] [options] [-infile=<filename>|<symbol> ...]

Options:
  -h -help             show help
  -v -version          show version
  -years=<years>       number of years to download [default=5]
  -start=<datestr>     yyyy[-[mm-[dd]]]
  -end=<datestr>       yyyy[-[mm-[dd]]] [default=today]
  -infile=<filename>   list of symbols to download
  -markets=<list>      list of markets to download (comma separated)
  -outfile=<filename>  output filename
  -period=<period>     1m|3m|5m|15m|30m|1h|2h|4h|6h|8h|12h|d|3d|w|m [default=d]
  -source=<source>     tiingo|tiingo-crypto|coinbase [default=tiingo]
  -token=<tiingo_tok>  tingo api token [default=TIINGO_API_TOKEN]
  -format=<format>     (csv|json|hs|ami) [default=csv]
  -all=<bool>          all in one file (true|false) [default=false]
  -log=<dest>          filename|stdout|stderr|discard [default=stdout]
  -delay=<ms>          delay in milliseconds between quote requests

Note: not all periods work with all sources

Valid markets:
etf,nasdaq,nasdaq100,amex,nyse,megacap,largecap,midcap,smallcap,microcap,nanocap,
telecommunications,health_care,finance,real_estate,consumer_discretionary,
consumer_staples,industrials,basic_materials,energy,utilities
coinbase,tiingo-usd,tiingo-btc,tiingo-eth
```

## CLI Examples

```bash
# display usage
quote -help

# downloads 10 years of smallcap, midcap, largecap and megacap stocks to stocks.csv
quote -markets=smallcap,midcap,largecap,megacap -all=true -years=10 -outfile=stocks.csv

# downloads 10 years of spy,qqq and djia to indexes.csv
quote -years=10 -outfile=indexes.csv -all=true spy qqq djia

# downloads 5 years of Tiingo SPY history to spy.csv (TIINGO_API_TOKEN must be set)
quote spy

# downloads 1 year of bitcoin history to BTC-USD.csv
quote -years=1 -source=coinbase BTC-USD


# downloads full etf symbol list to etf.txt, also works for nasdaq,nasdaq100,nyse,amex
quote etf

# download fresh etf list and 5 years of etf data all in one file
quote -markets=etf -all=true -outfile=etf.csv
```

## Install library

Install the package with:

```bash
go get github.com/markcheno/go-quote@latest
```

## Library example

```go
package main

import (
	"fmt"
	"github.com/markcheno/go-quote"
	"github.com/markcheno/go-talib"
)

func main() {
	spy, _ := quote.NewQuoteFromTiingo("spy", "2016-01-01", "2016-04-01", quote.Daily, true)
	fmt.Print(spy.CSV())
	rsi2 := talib.Rsi(spy.Close, 2)
	fmt.Println(rsi2)
}
```

## License

MIT License - see LICENSE for more details
