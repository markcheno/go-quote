/*
Package quote is free quote downloader library and cli

Downloads daily/weekly/monthly/yearly historical price quotes from Yahoo
and daily/intraday data from Google,Tiingo, crypto from Gdax,Bittrex/Binance

Copyright 2018 Mark Chenoweth
Licensed under terms of MIT license

*/

package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/kjx98/go-quote"
)

var usage = `Usage:
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
  -outfile=<filename>  output filename
  -period=<period>     1m|3m|5m|15m|30m|1h|2h|4h|6h|8h|12h|d|3d|w|m [default=d]
  -source=<source>     yahoo|google|tiingo|tiingo-crypto|gdax|bittrex|binance|huobi|kraken [default=yahoo]
  -token=<tiingo_tok>  tingo api token [default=TIINGO_API_TOKEN]
  -format=<format>     (csv|json|hs|ami) [default=csv]
  -adjust=<bool>       adjust yahoo prices [default=true]
  -all=<bool>          all in one file (true|false) [default=false]
  -log=<dest>          filename|stdout|stderr|discard [default=stdout]
  -delay=<ms>          delay in milliseconds between quote requests

Note: not all periods work with all sources

Valid markets:
etfs:       etf
exchanges:  nasdaq,nyse,amex
market cap: megacap,largecap,midcap,smallcap,microcap,nanocap
sectors:    basicindustries,capitalgoods,consumerdurables,consumernondurable,
            consumerservices,energy,finance,healthcare,miscellaneous,
            utilities,technolog,transportation
crypto:     bittrex-btc,bittrex-eth,bittrex-usdt,
            binance-bnb,binance-btc,binance-eth,binance-usdt
            tiingo-btc,tiingo-eth,tiingo-usd,huobi-ht,huobi-btc
            huobi-eth,huobi-usdt,kraken
all:        allmarkets
`

const (
	version    = "0.2.1"
	dateFormat = "2006-01-02"
)

type quoteflags struct {
	years   int
	delay   int
	start   string
	end     string
	period  string
	source  string
	token   string
	infile  string
	outfile string
	format  string
	log     string
	all     bool
	adjust  bool
	version bool
}

func check(e error) {
	if e != nil {
		fmt.Printf("\nerror: %v\n\n", e)
		fmt.Println(usage)
		os.Exit(0)
		//panic(e)
	}
}

func checkFlags(flags quoteflags) error {

	// validate source
	if flags.source != "yahoo" &&
		flags.source != "google" &&
		flags.source != "tiingo" &&
		flags.source != "tiingo-crypto" &&
		flags.source != "gdax" &&
		flags.source != "bittrex" &&
		flags.source != "huobi" &&
		flags.source != "kraken" &&
		flags.source != "binance" {
		return fmt.Errorf("invalid source, must be either 'yahoo', 'google', 'tiingo', 'gdax', 'bittrex', or 'binance', or 'huobi', 'kraken'")
	}

	// validate period
	if flags.source == "yahoo" &&
		(flags.period == "1m" || flags.period == "5m" || flags.period == "15m" || flags.period == "30m" || flags.period == "1h") {
		return fmt.Errorf("invalid source for yahoo, must be 'd'")
	}
	if flags.source == "tiingo" {
		// check period
		if flags.period != "d" {
			return fmt.Errorf("invalid source for tiingo, must be 'd'")
		}
		// check token
		if flags.token == "" {
			return fmt.Errorf("missing token for tiingo, must be passed or TIINGO_API_TOKEN must be set")
		}
	}

	if flags.source == "tiingo-crypto" &&
		!(flags.period == "1m" ||
			flags.period == "3m" ||
			flags.period == "5m" ||
			flags.period == "15m" ||
			flags.period == "30m" ||
			flags.period == "1h" ||
			flags.period == "2h" ||
			flags.period == "4h" ||
			flags.period == "6h" ||
			flags.period == "8h" ||
			flags.period == "12h" ||
			flags.period == "d") {
		return fmt.Errorf("invalid source for tiingo-crypto, must be '1m', '3m', '5m', '15m', '30m', '1h', '2h', '4h', '6h', '8h', '12h', '1d', '3d', '1w', or '1M'")
	}

	if flags.source == "tiingo-crypto" && flags.token == "" {
		return fmt.Errorf("missing token for tiingo-crypto, must be passed or TIINGO_API_TOKEN must be set")
	}

	if flags.source == "bittrex" && !(flags.period == "1m" || flags.period == "5m" || flags.period == "30m" || flags.period == "1h" || flags.period == "d") {
		return fmt.Errorf("invalid source for bittrex, must be '1m', '5m', '30m', '1h' or 'd'")
	}

	if flags.source == "binance" &&
		!(flags.period == "1m" ||
			flags.period == "3m" ||
			flags.period == "5m" ||
			flags.period == "15m" ||
			flags.period == "30m" ||
			flags.period == "1h" ||
			flags.period == "2h" ||
			flags.period == "4h" ||
			flags.period == "6h" ||
			flags.period == "8h" ||
			flags.period == "12h" ||
			flags.period == "d" ||
			flags.period == "3d" ||
			flags.period == "w" ||
			flags.period == "m") {
		return fmt.Errorf("invalid source for binance, must be '1m', '3m', '5m', '15m', '30m', '1h', '2h', '4h', '6h', '8h', '12h', '1d', '3d', '1w', or '1M'")
	}

	if flags.source == "huobi" &&
		!(flags.period == "1m" ||
			flags.period == "5m" ||
			flags.period == "15m" ||
			flags.period == "30m" ||
			flags.period == "1h" ||
			flags.period == "d" ||
			flags.period == "w" ||
			flags.period == "m") {
		return fmt.Errorf("invalid source for huobi, must be '1m', '5m', '15m', '30m', '1h', '1d', '1w', or '1M'")
	}

	if flags.source == "kraken" &&
		!(flags.period == "1m" ||
			flags.period == "5m" ||
			flags.period == "15m" ||
			flags.period == "30m" ||
			flags.period == "1h" ||
			flags.period == "4h" ||
			flags.period == "d" ||
			flags.period == "w") {
		return fmt.Errorf("invalid source for kraken, must be '1m', '5m', '15m', '30m', '1h', '4h', '1d', or '1w'")
	}

	//if flags.source == "google" && (flags.period == "w" || flags.period == "m") {
	//	return fmt.Errorf("invalid source for google, must be '1m' or '5m' or '15m' or '30m' or '1h' or 'd'")
	//}
	return nil
}

func setOutput(flags quoteflags) error {
	var err error
	if flags.log == "stdout" {
		quote.Log.SetOutput(os.Stdout)
	} else if flags.log == "stderr" {
		quote.Log.SetOutput(os.Stderr)
	} else if flags.log == "discard" {
		quote.Log.SetOutput(ioutil.Discard)
	} else {
		var f *os.File
		f, err = os.OpenFile(flags.log, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		defer f.Close()
		quote.Log.SetOutput(f)
	}
	return err
}

func getSymbols(flags quoteflags, args []string) ([]string, error) {

	var err error
	var symbols []string

	if flags.infile != "" {
		symbols, err = quote.NewSymbolsFromFile(flags.infile)
		if err != nil {
			return symbols, err
		}
	} else {
		symbols = args
	}

	// make sure we found some symbols
	if len(symbols) == 0 {
		return symbols, fmt.Errorf("no symbols specified")
	}

	// validate outfileFlag
	if len(symbols) > 1 && flags.outfile != "" && !flags.all {
		return symbols, fmt.Errorf("outfile not valid with multiple symbols\nuse -all=true")
	}

	return symbols, nil
}

func getPeriod(periodFlag string) quote.Period {
	period := quote.Daily
	switch periodFlag {
	case "1m":
		period = quote.Min1
	case "3m":
		period = quote.Min3
	case "5m":
		period = quote.Min5
	case "15m":
		period = quote.Min15
	case "30m":
		period = quote.Min30
	case "1h":
		period = quote.Min60
	case "2h":
		period = quote.Hour2
	case "4h":
		period = quote.Hour4
	case "6h":
		period = quote.Hour6
	case "8h":
		period = quote.Hour8
	case "12h":
		period = quote.Hour12
	case "d":
		period = quote.Daily
	case "1d":
		period = quote.Daily
	case "3d":
		period = quote.Day3
	case "w":
		period = quote.Weekly
	case "1w":
		period = quote.Weekly
	case "m":
		period = quote.Monthly
	case "1M":
		period = quote.Monthly
	}
	return period
}

func getTimes(flags quoteflags) (time.Time, time.Time) {
	// determine start/end times
	to := quote.ParseDateString(flags.end)
	var from time.Time
	if flags.start != "" {
		from = quote.ParseDateString(flags.start)
	} else { // use years
		yearSec := int64(time.Hour) * 24 * 365
		from = to.Add(-time.Duration(yearSec * int64(flags.years)))
	}
	return from, to
}

func outputAll(symbols []string, flags quoteflags) error {
	// output all in one file
	from, to := getTimes(flags)
	period := getPeriod(flags.period)
	quotes := quote.Quotes{}
	var err error
	if flags.source == "yahoo" {
		quotes, err = quote.NewQuotesFromYahooSyms(symbols, from.Format(dateFormat), to.Format(dateFormat), period, flags.adjust)
	} else if flags.source == "google" {
		quotes, err = quote.NewQuotesFromGoogleSyms(symbols, from.Format(dateFormat), to.Format(dateFormat), period)
	} else if flags.source == "tiingo" {
		quotes, err = quote.NewQuotesFromTiingoSyms(symbols, from.Format(dateFormat), to.Format(dateFormat), flags.token)
	} else if flags.source == "tiingo-crypto" {
		quotes, err = quote.NewQuotesFromTiingoCryptoSyms(symbols, from.Format(dateFormat), to.Format(dateFormat), period, flags.token)
	} else if flags.source == "gdax" {
		quotes, err = quote.NewQuotesFromGdaxSyms(symbols, from.Format(dateFormat), to.Format(dateFormat), period)
	} else if flags.source == "bittrex" {
		quotes, err = quote.NewQuotesFromBittrexSyms(symbols, period)
	} else if flags.source == "binance" {
		quotes, err = quote.NewQuotesFromBinanceSyms(symbols, from.Format(dateFormat), to.Format(dateFormat), period)
	} else if flags.source == "huobi" {
		quotes, err = quote.NewQuotesFromHuobiSyms(symbols, period)
	} else if flags.source == "kraken" {
		quotes, err = quote.NewQuotesFromKrakenSyms(symbols, period)
	}
	if err != nil {
		return err
	}

	if flags.format == "csv" {
		err = quotes.WriteCSV(flags.outfile)
	} else if flags.format == "json" {
		err = quotes.WriteJSON(flags.outfile, false)
	} else if flags.format == "hs" {
		err = quotes.WriteHighstock(flags.outfile)
	} else if flags.format == "ami" {
		err = quotes.WriteAmibroker(flags.outfile)
	}
	return err
}

func outputIndividual(symbols []string, flags quoteflags) error {
	// output individual symbol files

	from, to := getTimes(flags)
	period := getPeriod(flags.period)

	for _, sym := range symbols {
		var q quote.Quote
		if flags.source == "yahoo" {
			q, _ = quote.NewQuoteFromYahoo(sym, from.Format(dateFormat), to.Format(dateFormat), period, flags.adjust)
		} else if flags.source == "google" {
			q, _ = quote.NewQuoteFromGoogle(sym, from.Format(dateFormat), to.Format(dateFormat), period)
		} else if flags.source == "tiingo" {
			q, _ = quote.NewQuoteFromTiingo(sym, from.Format(dateFormat), to.Format(dateFormat), flags.token)
		} else if flags.source == "tiingo-crypto" {
			q, _ = quote.NewQuoteFromTiingoCrypto(sym, from.Format(dateFormat), to.Format(dateFormat), period, flags.token)
		} else if flags.source == "gdax" {
			q, _ = quote.NewQuoteFromGdax(sym, from.Format(dateFormat), to.Format(dateFormat), period)
		} else if flags.source == "bittrex" {
			q, _ = quote.NewQuoteFromBittrex(sym, period)
		} else if flags.source == "binance" {
			q, _ = quote.NewQuoteFromBinance(sym, from.Format(dateFormat), to.Format(dateFormat), period)
		} else if flags.source == "huobi" {
			q, _ = quote.NewQuoteFromHuobi(sym, period)
		} else if flags.source == "kraken" {
			q, _ = quote.NewQuoteFromKraken(sym, period)
		}
		var err error
		if flags.format == "csv" {
			err = q.WriteCSV(flags.outfile)
		} else if flags.format == "json" {
			err = q.WriteJSON(flags.outfile, false)
		} else if flags.format == "hs" {
			err = q.WriteHighstock(flags.outfile)
		} else if flags.format == "ami" {
			err = q.WriteAmibroker(flags.outfile)
		}
		if err != nil {
			fmt.Printf("Error writing file: %v\n", err)
		}
		time.Sleep(quote.Delay * time.Millisecond)
	}
	return nil
}

func handleCommand(cmd string, flags quoteflags) bool {
	// handle market special commands
	if !quote.ValidMarket(cmd) {
		return false
	}
	switch cmd {
	case "etf":
		quote.NewEtfFile(flags.outfile)
	default:
		quote.NewMarketFile(cmd, flags.outfile)
	}
	return true
}

func main() {

	var err error
	var symbols []string
	var flags quoteflags

	flag.IntVar(&flags.years, "years", 5, "number of years to download")
	flag.IntVar(&flags.delay, "delay", 100, "milliseconds to delay between requests")
	flag.StringVar(&flags.start, "start", "", "start date (yyyy[-mm[-dd]])")
	flag.StringVar(&flags.end, "end", "", "end date (yyyy[-mm[-dd]])")
	flag.StringVar(&flags.period, "period", "d", "1m|5m|15m|30m|1h|d")
	flag.StringVar(&flags.source, "source", "yahoo", "yahoo|google|tiingo|gdax|bittrex|binance|huobi|kraken")
	flag.StringVar(&flags.token, "token", os.Getenv("TIINGO_API_TOKEN"), "tiingo api token")
	flag.StringVar(&flags.infile, "infile", "", "input filename")
	flag.StringVar(&flags.outfile, "outfile", "", "output filename")
	flag.StringVar(&flags.format, "format", "csv", "csv|json")
	flag.StringVar(&flags.log, "log", "stdout", "<filename>|stdout")
	flag.BoolVar(&flags.all, "all", false, "all output in one file")
	flag.BoolVar(&flags.adjust, "adjust", true, "adjust Yahoo prices")
	flag.BoolVar(&flags.version, "v", false, "show version")
	flag.BoolVar(&flags.version, "version", false, "show version")
	flag.Parse()

	if flags.version {
		fmt.Println(version)
		os.Exit(0)
	}

	quote.Delay = time.Duration(flags.delay)

	err = setOutput(flags)
	check(err)

	err = checkFlags(flags)
	check(err)

	symbols, err = getSymbols(flags, flag.Args())
	check(err)

	// check for and handled special commands
	if handleCommand(symbols[0], flags) {
		os.Exit(0)
	}

	// main output
	if flags.all {
		err = outputAll(symbols, flags)
	} else {
		err = outputIndividual(symbols, flags)
	}
}
