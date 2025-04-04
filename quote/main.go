/*
Package quote is free quote downloader library and cli

Downloads historical price quotes from Tiingo and Coinbase

Copyright 2025 Mark Chenoweth
Licensed under terms of MIT license

*/

package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/markcheno/go-quote"
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
  -markets=<list>      list of valid markets to download (comma separated)
  -infile=<filename>   list of symbols to download
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
consumer_staples,industrials,basic_materials,energy,utilities,technology
coinbase,tiingo-usd,tiingo-btc,tiingo-eth
`

const (
	version    = "0.4"
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
	markets string
	infile  string
	outfile string
	format  string
	log     string
	all     bool
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
	if flags.source != "tiingo" &&
		flags.source != "tiingo-crypto" &&
		flags.source != "coinbase" {
		return fmt.Errorf("invalid source, must be either 'tiingo', 'tiingo-crypto', or 'coinbase'")
	}

	// validate period
	if flags.source == "tiingo" {
		// check period
		if flags.period != "d" {
			return fmt.Errorf("invalid period for tiingo, must be 'd'")
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

	return nil
}

func setOutput(flags quoteflags) error {
	var err error
	if flags.log == "stdout" {
		quote.Log.SetOutput(os.Stdout)
	} else if flags.log == "stderr" {
		quote.Log.SetOutput(os.Stderr)
	} else if flags.log == "discard" {
		quote.Log.SetOutput(io.Discard)
	} else {
		var f *os.File
		f, err = os.OpenFile(flags.log, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			quote.Log.Println(err)
		}
		defer f.Close()
		quote.Log.SetOutput(f)
	}
	return err
}

// func getSymbols(flags quoteflags, args []string) ([]string, error) {
// 	var err error
// 	var symbols []string
// 	if flags.infile != "" {
// 		symbols, err = quote.NewSymbolsFromFile(flags.infile)
// 		if err != nil {
// 			return symbols, err
// 		}
// 	} else {
// 		symbols = args
// 	}
// 	// make sure we found some symbols
// 	if len(symbols) == 0 {
// 		return symbols, fmt.Errorf("no symbols specified")
// 	}
// 	// validate outfileFlag
// 	if len(symbols) > 1 && flags.outfile != "" && !flags.all {
// 		return symbols, fmt.Errorf("outfile not valid with multiple symbols\nuse -all=true")
// 	}
// 	return symbols, nil
// }

func getSymbols(flags quoteflags, args []string) ([]string, error) {
	var err error
	var symbols []string

	if flags.infile != "" {
		// Check if infile contains wildcard characters
		if strings.Contains(flags.infile, "*") || strings.Contains(flags.infile, "?") {
			// Find all matching files
			matches, err := filepath.Glob(flags.infile)
			if err != nil {
				return nil, fmt.Errorf("error processing wildcard pattern: %v", err)
			}

			if len(matches) == 0 {
				return nil, fmt.Errorf("no files match pattern: %s", flags.infile)
			}

			// Read symbols from all matching files
			for _, file := range matches {
				fileSymbols, err := quote.NewSymbolsFromFile(file)
				if err != nil {
					return nil, fmt.Errorf("error reading symbols from %s: %v", file, err)
				}
				symbols = append(symbols, fileSymbols...)
			}
		} else {
			// Regular file handling
			symbols, err = quote.NewSymbolsFromFile(flags.infile)
			if err != nil {
				return symbols, err
			}
		}
	} else if flags.markets != "" {

		markets := strings.Split(flags.markets, ",")
		for _, cmd := range markets {
			if !quote.ValidMarket(cmd) {
				return symbols, fmt.Errorf("invalid market specified: " + cmd)
			}
			file := cmd + ".csv"
			switch cmd {
			case "etf":
				quote.NewEtfFile(file)

			default:
				quote.NewMarketFile(cmd, file)
			}
			fileSymbols, err := quote.NewSymbolsFromFile(file)
			if err != nil {
				return nil, fmt.Errorf("error reading symbols from %s: %v", file, err)
			}
			symbols = append(symbols, fileSymbols...)
		}
		return symbols, nil
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
		from = to.Add(-time.Duration(int(time.Hour) * 24 * 365 * flags.years))
	}
	return from, to
}

func outputAll(symbols []string, flags quoteflags) error {
	// output all in one file
	from, to := getTimes(flags)
	period := getPeriod(flags.period)
	quotes := quote.Quotes{}
	var err error
	if flags.source == "tiingo" {
		quotes, err = quote.NewQuotesFromTiingoSyms(symbols, from.Format(dateFormat), to.Format(dateFormat), flags.token)
	} else if flags.source == "tiingo-crypto" {
		quotes, err = quote.NewQuotesFromTiingoCryptoSyms(symbols, from.Format(dateFormat), to.Format(dateFormat), period, flags.token)
	} else if flags.source == "coinbase" {
		quotes, err = quote.NewQuotesFromCoinbaseSyms(symbols, from.Format(dateFormat), to.Format(dateFormat), period)
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
		if flags.source == "tiingo" {
			q, _ = quote.NewQuoteFromTiingo(sym, from.Format(dateFormat), to.Format(dateFormat), flags.token)
		} else if flags.source == "tiingo-crypto" {
			q, _ = quote.NewQuoteFromTiingoCrypto(sym, from.Format(dateFormat), to.Format(dateFormat), period, flags.token)
		} else if flags.source == "coinbase" {
			q, _ = quote.NewQuoteFromCoinbase(sym, from.Format(dateFormat), to.Format(dateFormat), period)
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

func handleCommand(symbols []string, flags quoteflags) bool {

	if flags.markets != "" {
		return false
	}

	// handle market special commands
	for _, cmd := range symbols {
		if !quote.ValidMarket(cmd) {
			return false
		}
		switch cmd {
		case "etf":
			quote.NewEtfFile(flags.outfile)
		default:
			quote.NewMarketFile(cmd, flags.outfile)
		}
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
	flag.StringVar(&flags.source, "source", "tiingo", "tiingo|tiingo-crypto|coinbase")
	flag.StringVar(&flags.token, "token", os.Getenv("TIINGO_API_TOKEN"), "tiingo api token")
	flag.StringVar(&flags.infile, "infile", "", "input filename")
	flag.StringVar(&flags.outfile, "outfile", "", "output filename")
	flag.StringVar(&flags.markets, "markets", "", "list of valid markets (comma separated)")
	flag.StringVar(&flags.format, "format", "csv", "csv|json")
	flag.StringVar(&flags.log, "log", "stdout", "<filename>|stdout")
	flag.BoolVar(&flags.all, "all", false, "all output in one file")
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

	// check for and handle special commands
	if handleCommand(symbols, flags) {
		os.Exit(0)
	}

	//fmt.Println("Downloading quotes for", len(symbols), "symbols")

	// main output
	if flags.all {
		outputAll(symbols, flags)
	} else {
		outputIndividual(symbols, flags)
	}
}
