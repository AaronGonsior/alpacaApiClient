package alpacaApiClient

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

var (
	APIKeyID     string
	APISecretKey string
)

func init() {
	handleApiKeyInit()
}

type Config struct {
	APIKeyID     string `json:"api_key_id"`
	APISecretKey string `json:"api_secret_key"`
}

func loadConfig() (*Config, error) {
	file, err := os.Open("alpacaConfig.json")
	if err != nil {
		return nil, fmt.Errorf("error opening config file: %v", err)
	}
	defer file.Close()

	var config Config
	if err := json.NewDecoder(file).Decode(&config); err != nil {
		return nil, fmt.Errorf("error decoding config file: %v", err)
	}

	return &config, nil
}

func main() {
	config, err := loadConfig()
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		return
	}

	//uuid := "799238118"

	//TSLA quote
	url := "https://data.alpaca.markets/v2/stocks/quotes/latest?symbols=TSLA"

	req, _ := http.NewRequest("GET", url, nil)

	req.Header.Add("accept", "application/json")
	req.Header.Add("APCA-API-KEY-ID", config.APIKeyID)
	req.Header.Add("APCA-API-SECRET-KEY", config.APISecretKey)

	res, _ := http.DefaultClient.Do(req)

	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)

	fmt.Println(string(body))
	/*

		//TSLA options chain
		url = "https://data.alpaca.markets/v1beta1/options/snapshots/TSLA?feed=indicative&limit=10&type=call&page_token="

		req, _ = http.NewRequest("GET", url, nil)

		req.Header.Add("accept", "application/json")
		req.Header.Add("APCA-API-KEY-ID", "PKAKUI1RYWHYUSLU49LW")
		req.Header.Add("APCA-API-SECRET-KEY", "SCbRWgh5XAad3QTyS7fIksVK1e9H4x2ggjScE9c6")

		res, _ = http.DefaultClient.Do(req)

		defer res.Body.Close()
		body, _ = io.ReadAll(res.Body)

		fmt.Println(string(body))

		//TSLA call buy
		url = "https://paper-api.alpaca.markets/v2/orders"

		payload := strings.NewReader("{\"type\":\"market\",\"time_in_force\":\"day\",\"symbol\":\"TSLA250523C00335000\",\"qty\":\"1\",\"side\":\"buy\"}")

		req, _ = http.NewRequest("POST", url, payload)

		req.Header.Add("accept", "application/json")
		req.Header.Add("content-type", "application/json")

		res, _ = http.DefaultClient.Do(req)

		defer res.Body.Close()
		body, _ = io.ReadAll(res.Body)

		fmt.Println(string(body))
	*/

	//Test functions

	//url = "https://data.alpaca.markets/v1beta1/options/snapshots/TSLA?feed=indicative&limit=1000&type=call&page_token="
	optreq := OptionURLReq{
		Ticker:        "TSLA",
		Contract_type: "call",
		StrikeRange:   []int{0, 10000},
		DateRange:     []string{"2025-05-23", "2027-01-23"},
	}

	options, _, err := GetOptions(optreq, -1)
	check(err)
	fmt.Println(len(options))

	fmt.Println(options[0])
	fmt.Println(options[0].LatestQuote.AskPrice)

}

func check(err error) {
	if err != nil {
		fmt.Println(err)
	}
}

func WriteJson(path string, content string) {
	// Open a file for writing
	file, err := os.Create(path)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer file.Close()
	// Encode the string as JSON and write it to the file
	if err := json.NewEncoder(file).Encode(content); err != nil {
		fmt.Println(err)
		return
	}
}

func LoadJson(path string) string {
	// Open the file for reading
	file, err := os.Open(path)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return ""
	}
	defer file.Close()

	// Read the entire file content
	content, err := io.ReadAll(file)
	if err != nil {
		fmt.Println("Error reading file:", err)
		return ""
	}

	return string(content)
}

func JsonToOptions(path string) []Option {
	content := LoadJson(path)
	if content == "" {
		return nil
	}

	//fmt.Println("content: ", content)
	//os.Exit(123)

	// Parse the JSON into a map
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(content), &data); err != nil {
		fmt.Println("Error parsing JSON:", err)
		return nil
	}

	// Get the options array
	optionsData, ok := data["options"].([]interface{})
	if !ok {
		fmt.Println("Error: options array not found in JSON")
		return nil
	}

	var options []Option
	for _, opt := range optionsData {
		optMap, ok := opt.(map[string]interface{})
		if !ok {
			continue
		}

		// Create new Option with basic data and initialize all market data structures
		newOption := Option{
			ID:                getString(optMap["id"]),
			Symbol:            getString(optMap["symbol"]),
			Name:              getString(optMap["name"]),
			Status:            getString(optMap["status"]),
			Tradable:          getBool(optMap["tradable"]),
			ExpirationDate:    getString(optMap["expiration_date"]),
			RootSymbol:        getString(optMap["root_symbol"]),
			UnderlyingSymbol:  getString(optMap["underlying_symbol"]),
			UnderlyingAssetID: getString(optMap["underlying_asset_id"]),
			Type:              getString(optMap["type"]),
			Style:             getString(optMap["style"]),
			StrikePrice:       getFloat64(optMap["strike_price"]),
			Multiplier:        getInt(optMap["multiplier"]),
			Size:              getInt(optMap["size"]),
			OpenInterest:      getInt(optMap["open_interest"]),
			OpenInterestDate:  getString(optMap["open_interest_date"]),
			ClosePrice:        getFloat64(optMap["close_price"]),
			ClosePriceDate:    getString(optMap["close_price_date"]),
			PPIND:             getBool(optMap["ppind"]),

			// Initialize all market data structures with empty values
			DailyBar: &Bar{
				Close:          0,
				High:           0,
				Low:            0,
				NumberOfTrades: 0,
				Open:           0,
				Timestamp:      time.Time{},
				Volume:         0,
				VWAP:           0,
			},
			PrevDailyBar: &Bar{
				Close:          0,
				High:           0,
				Low:            0,
				NumberOfTrades: 0,
				Open:           0,
				Timestamp:      time.Time{},
				Volume:         0,
				VWAP:           0,
			},
			MinuteBar: &Bar{
				Close:          0,
				High:           0,
				Low:            0,
				NumberOfTrades: 0,
				Open:           0,
				Timestamp:      time.Time{},
				Volume:         0,
				VWAP:           0,
			},
			Greeks: &Greeks{
				Delta: 0,
				Gamma: 0,
				Rho:   0,
				Theta: 0,
				Vega:  0,
			},
			ImpliedVol: 0,
			LatestQuote: &Quote{
				AskPrice:    0,
				AskSize:     0,
				AskExchange: "",
				BidPrice:    0,
				BidSize:     0,
				BidExchange: "",
				Condition:   "",
				Timestamp:   time.Time{},
			},
			LatestTrade: &Trade{
				Condition: "",
				Price:     0,
				Size:      0,
				Timestamp: time.Time{},
				Exchange:  "",
			},
		}

		// Parse DailyBar if it exists
		if dailyBar, ok := optMap["dailyBar"].(map[string]interface{}); ok {
			timestamp, _ := time.Parse(time.RFC3339, getString(dailyBar["t"]))
			newOption.DailyBar.Close = getFloat64(dailyBar["c"])
			newOption.DailyBar.High = getFloat64(dailyBar["h"])
			newOption.DailyBar.Low = getFloat64(dailyBar["l"])
			newOption.DailyBar.NumberOfTrades = getInt(dailyBar["n"])
			newOption.DailyBar.Open = getFloat64(dailyBar["o"])
			newOption.DailyBar.Timestamp = timestamp
			newOption.DailyBar.Volume = getInt(dailyBar["v"])
			newOption.DailyBar.VWAP = getFloat64(dailyBar["vw"])
		}

		// Parse PrevDailyBar if it exists
		if prevDailyBar, ok := optMap["prevDailyBar"].(map[string]interface{}); ok {
			timestamp, _ := time.Parse(time.RFC3339, getString(prevDailyBar["t"]))
			newOption.PrevDailyBar.Close = getFloat64(prevDailyBar["c"])
			newOption.PrevDailyBar.High = getFloat64(prevDailyBar["h"])
			newOption.PrevDailyBar.Low = getFloat64(prevDailyBar["l"])
			newOption.PrevDailyBar.NumberOfTrades = getInt(prevDailyBar["n"])
			newOption.PrevDailyBar.Open = getFloat64(prevDailyBar["o"])
			newOption.PrevDailyBar.Timestamp = timestamp
			newOption.PrevDailyBar.Volume = getInt(prevDailyBar["v"])
			newOption.PrevDailyBar.VWAP = getFloat64(prevDailyBar["vw"])
		}

		// Parse MinuteBar if it exists
		if minuteBar, ok := optMap["minuteBar"].(map[string]interface{}); ok {
			timestamp, _ := time.Parse(time.RFC3339, getString(minuteBar["t"]))
			newOption.MinuteBar.Close = getFloat64(minuteBar["c"])
			newOption.MinuteBar.High = getFloat64(minuteBar["h"])
			newOption.MinuteBar.Low = getFloat64(minuteBar["l"])
			newOption.MinuteBar.NumberOfTrades = getInt(minuteBar["n"])
			newOption.MinuteBar.Open = getFloat64(minuteBar["o"])
			newOption.MinuteBar.Timestamp = timestamp
			newOption.MinuteBar.Volume = getInt(minuteBar["v"])
			newOption.MinuteBar.VWAP = getFloat64(minuteBar["vw"])
		}

		// Parse Greeks if they exist
		if greeks, ok := optMap["greeks"].(map[string]interface{}); ok {
			newOption.Greeks.Delta = getFloat64(greeks["delta"])
			newOption.Greeks.Gamma = getFloat64(greeks["gamma"])
			newOption.Greeks.Rho = getFloat64(greeks["rho"])
			newOption.Greeks.Theta = getFloat64(greeks["theta"])
			newOption.Greeks.Vega = getFloat64(greeks["vega"])
		}

		// Parse ImpliedVolatility if it exists
		newOption.ImpliedVol = getFloat64(optMap["impliedVolatility"])

		// Parse LatestQuote if it exists
		if quote, ok := optMap["latestQuote"].(map[string]interface{}); ok {
			timestamp, _ := time.Parse(time.RFC3339Nano, getString(quote["t"]))
			newOption.LatestQuote.AskPrice = getFloat64(quote["ap"])
			newOption.LatestQuote.AskSize = getInt(quote["as"])
			newOption.LatestQuote.AskExchange = getString(quote["ax"])
			newOption.LatestQuote.BidPrice = getFloat64(quote["bp"])
			newOption.LatestQuote.BidSize = getInt(quote["bs"])
			newOption.LatestQuote.BidExchange = getString(quote["bx"])
			newOption.LatestQuote.Condition = getString(quote["c"])
			newOption.LatestQuote.Timestamp = timestamp
		}

		// Parse LatestTrade if it exists
		if trade, ok := optMap["latestTrade"].(map[string]interface{}); ok {
			timestamp, _ := time.Parse(time.RFC3339Nano, getString(trade["t"]))
			newOption.LatestTrade.Condition = getString(trade["c"])
			newOption.LatestTrade.Price = getFloat64(trade["p"])
			newOption.LatestTrade.Size = getInt(trade["s"])
			newOption.LatestTrade.Timestamp = timestamp
			newOption.LatestTrade.Exchange = getString(trade["x"])
		}

		options = append(options, newOption)
	}

	return options
}

// Helper functions to safely extract values from interface{}
func getFloat64(v interface{}) float64 {
	if f, ok := v.(float64); ok {
		return f
	}
	return 0
}

func getInt(v interface{}) int {
	if f, ok := v.(float64); ok {
		return int(f)
	}
	return 0
}

func getString(v interface{}) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

type OptionURLReq struct {
	Ticker        string
	Contract_type string
	ApiKey        string
	StrikeRange   []int
	DateRange     []string
}

// Bar represents price/volume data for a time period
type Bar struct {
	Close          float64   `json:"c"`
	High           float64   `json:"h"`
	Low            float64   `json:"l"`
	NumberOfTrades int       `json:"n"`
	Open           float64   `json:"o"`
	Timestamp      time.Time `json:"t"`
	Volume         int       `json:"v"`
	VWAP           float64   `json:"vw"`
}

// Greeks represents the option Greeks
type Greeks struct {
	Delta float64 `json:"delta"`
	Gamma float64 `json:"gamma"`
	Rho   float64 `json:"rho"`
	Theta float64 `json:"theta"`
	Vega  float64 `json:"vega"`
}

// Quote represents the latest quote data
type Quote struct {
	AskPrice    float64   `json:"ap"`
	AskSize     int       `json:"as"`
	AskExchange string    `json:"ax"`
	BidPrice    float64   `json:"bp"`
	BidSize     int       `json:"bs"`
	BidExchange string    `json:"bx"`
	Condition   string    `json:"c"`
	Timestamp   time.Time `json:"t"`
}

// Trade represents the latest trade data
type Trade struct {
	Condition string    `json:"c"`
	Price     float64   `json:"p"`
	Size      int       `json:"s"`
	Timestamp time.Time `json:"t"`
	Exchange  string    `json:"x"`
}

type Option struct {
	// Basic contract information
	ID                string  `json:"id"`
	Symbol            string  `json:"symbol"`
	Name              string  `json:"name"`
	Status            string  `json:"status"`
	Tradable          bool    `json:"tradable"`
	ExpirationDate    string  `json:"expiration_date"`
	RootSymbol        string  `json:"root_symbol"`
	UnderlyingSymbol  string  `json:"underlying_symbol"`
	UnderlyingAssetID string  `json:"underlying_asset_id"`
	Type              string  `json:"type"`
	Style             string  `json:"style"`
	StrikePrice       float64 `json:"strike_price"`
	Multiplier        int     `json:"multiplier"`
	Size              int     `json:"size"`
	OpenInterest      int     `json:"open_interest"`
	OpenInterestDate  string  `json:"open_interest_date"`
	ClosePrice        float64 `json:"close_price"`
	ClosePriceDate    string  `json:"close_price_date"`
	PPIND             bool    `json:"ppind"`

	// Market data
	DailyBar     *Bar    `json:"dailyBar"`
	PrevDailyBar *Bar    `json:"prevDailyBar"`
	MinuteBar    *Bar    `json:"minuteBar"`
	Greeks       *Greeks `json:"greeks"`
	ImpliedVol   float64 `json:"impliedVolatility"`
	LatestQuote  *Quote  `json:"latestQuote"`
	LatestTrade  *Trade  `json:"latestTrade"`
}

func (o Option) Print() string {
	readStr := fmt.Sprint(o)
	readStr = strings.Replace(readStr, "} {", "\n", -1)
	readStr = strings.Replace(readStr, "}]", "", -1)
	readStr = strings.Replace(readStr, "[{", "", -1)
	return readStr
}

func GetOptions(optreq OptionURLReq, nMax int) ([]Option, string, error) {
	print := true
	var options []Option
	var log string

	// Validate date format
	for _, date := range optreq.DateRange {
		_, err := time.Parse("2006-01-02", date)
		if err != nil {
			return nil, log, fmt.Errorf("invalid date format in DateRange: %s. Expected format: YYYY-MM-DD", date)
		}
	}

	if print {
		fmt.Println("Pulling options for option request:")
		fmt.Printf("ticker=%v\nContract_type=%v\nStrikeRange=%v\nDateRange=%v\n",
			optreq.Ticker,
			optreq.Contract_type,
			optreq.StrikeRange,
			optreq.DateRange)
	}

	if nMax == -1 {
		nMax = 10000
	}

	// Initial URL with all parameters
	url := fmt.Sprintf("https://paper-api.alpaca.markets/v2/options/contracts?underlying_symbols=%s&show_deliverables=false&expiration_date_gte=%s&expiration_date_lte=%s&type=%s&strike_price_gte=%v&strike_price_lte=%v&page_token=%s&limit=1000",
		optreq.Ticker,
		optreq.DateRange[0],
		optreq.DateRange[1],
		optreq.Contract_type,
		optreq.StrikeRange[0],
		optreq.StrikeRange[1],
		"")

	// Keep track of processed options to avoid duplicates
	processedIDs := make(map[string]bool)
	requestCounter := 0
	optionCounter := 0

	// Set timeout for the entire operation
	timeout := time.After(5 * time.Minute)
	tick := time.Tick(1 * time.Second)

	// Continue fetching until no more pages or max options reached
	for {
		select {
		case <-timeout:
			return options, log, fmt.Errorf("operation timed out after 5 minutes. Fetched %d options", len(options))
		case <-tick:
			// Make API request
			_, bodyStr, err := APIRequest(url, 1)
			if err != nil {
				return nil, log, err
			}

			// Parse the response
			var data map[string]interface{}
			if err := json.Unmarshal([]byte(bodyStr), &data); err != nil {
				return nil, log, err
			}

			// Get the option_contracts array
			contracts, ok := data["option_contracts"].([]interface{})
			if !ok {
				return nil, log, fmt.Errorf("invalid response format: option_contracts not found or not an array")
			}

			// Process each option in the current page
			for _, contract := range contracts {
				contractMap, ok := contract.(map[string]interface{})
				if !ok {
					continue
				}

				id := getString(contractMap["id"])
				// Skip if we've already processed this option
				if processedIDs[id] {
					continue
				}
				processedIDs[id] = true

				// Create new Option
				newOption := Option{
					ID:                id,
					Symbol:            getString(contractMap["symbol"]),
					Name:              getString(contractMap["name"]),
					Status:            getString(contractMap["status"]),
					Tradable:          getBool(contractMap["tradable"]),
					ExpirationDate:    getString(contractMap["expiration_date"]),
					RootSymbol:        getString(contractMap["root_symbol"]),
					UnderlyingSymbol:  getString(contractMap["underlying_symbol"]),
					UnderlyingAssetID: getString(contractMap["underlying_asset_id"]),
					Type:              getString(contractMap["type"]),
					Style:             getString(contractMap["style"]),
					StrikePrice:       parseFloat64(getString(contractMap["strike_price"])),
					Multiplier:        parseInt(getString(contractMap["multiplier"])),
					Size:              parseInt(getString(contractMap["size"])),
					OpenInterest:      parseInt(getString(contractMap["open_interest"])),
					OpenInterestDate:  getString(contractMap["open_interest_date"]),
					ClosePrice:        parseFloat64(getString(contractMap["close_price"])),
					ClosePriceDate:    getString(contractMap["close_price_date"]),
					PPIND:             getBool(contractMap["ppind"]),

					// Initialize all pointer fields with empty but non-nil structs
					DailyBar: &Bar{
						Close:          0,
						High:           0,
						Low:            0,
						NumberOfTrades: 0,
						Open:           0,
						Timestamp:      time.Time{},
						Volume:         0,
						VWAP:           0,
					},
					PrevDailyBar: &Bar{
						Close:          0,
						High:           0,
						Low:            0,
						NumberOfTrades: 0,
						Open:           0,
						Timestamp:      time.Time{},
						Volume:         0,
						VWAP:           0,
					},
					MinuteBar: &Bar{
						Close:          0,
						High:           0,
						Low:            0,
						NumberOfTrades: 0,
						Open:           0,
						Timestamp:      time.Time{},
						Volume:         0,
						VWAP:           0,
					},
					Greeks: &Greeks{
						Delta: 0,
						Gamma: 0,
						Rho:   0,
						Theta: 0,
						Vega:  0,
					},
					ImpliedVol: 0,
					LatestQuote: &Quote{
						AskPrice:    0,
						AskSize:     0,
						AskExchange: "",
						BidPrice:    0,
						BidSize:     0,
						BidExchange: "",
						Condition:   "",
						Timestamp:   time.Time{},
					},
					LatestTrade: &Trade{
						Condition: "",
						Price:     0,
						Size:      0,
						Timestamp: time.Time{},
						Exchange:  "",
					},
				}

				options = append(options, newOption)
				optionCounter++

				if print {
					fmt.Printf("\r %v API requests successfully made - %v available options found", requestCounter+1, optionCounter)
				}

				// Check if we've reached the maximum number of options
				if nMax > 0 && len(options) >= nMax {
					if print {
						fmt.Println("\nReached maximum number of options")
					}
					goto MARKET_DATA
				}
			}

			requestCounter++

			// Check for next page token
			nextToken, ok := data["next_page_token"].(string)
			if !ok || nextToken == "" {
				if print {
					fmt.Println("\nCompleted fetching basic option data")
				}
				goto MARKET_DATA
			}

			// Update URL for next page
			url = fmt.Sprintf("https://paper-api.alpaca.markets/v2/options/contracts?underlying_symbols=%s&show_deliverables=false&expiration_date_gte=%s&expiration_date_lte=%s&type=%s&strike_price_gte=%v&strike_price_lte=%v&page_token=%s&limit=1000",
				optreq.Ticker,
				optreq.DateRange[0],
				optreq.DateRange[1],
				optreq.Contract_type,
				optreq.StrikeRange[0],
				optreq.StrikeRange[1],
				nextToken)
		}
	}

MARKET_DATA:
	// Now get the market data for these options
	if print {
		fmt.Println("Fetching market data for options...")
	}

	// Create a map for quick option lookup by symbol
	optionMap := make(map[string]*Option)
	for i := range options {
		optionMap[options[i].Symbol] = &options[i]
		if print && i < 5 {
			//fmt.Printf("Sample option symbol from first API: %s\n", options[i].Symbol)
		}
	}

	nextToken := ""
	// Initial market data URL
	marketDataURL := fmt.Sprintf("https://data.alpaca.markets/v1beta1/options/snapshots/%s?feed=indicative&limit=1000&page_token=%s&strike_price_gte=%v&strike_price_lte=%v&expiration_date_gte=%s&expiration_date_lte=%s&type=%s",
		optreq.Ticker,
		nextToken,
		optreq.StrikeRange[0],
		optreq.StrikeRange[1],
		optreq.DateRange[0],
		optreq.DateRange[1],
		optreq.Contract_type,
	)

	marketRequestCounter := 0
	marketDataProcessed := 0
	symbolsNotFound := make(map[string]bool)

	// Continue fetching market data until no more pages
	for {
		_, bodyStr, err := APIRequest(marketDataURL, 1)
		if err != nil {
			return options, log + "\nError fetching market data: " + err.Error(), nil
		}

		// Parse the market data response
		var marketData map[string]interface{}
		if err := json.Unmarshal([]byte(bodyStr), &marketData); err != nil {
			return options, log + "\nError parsing market data: " + err.Error(), nil
		}

		// Update options with market data
		snapshots, ok := marketData["snapshots"].(map[string]interface{})
		if !ok {
			return options, log + "\nNo snapshots found in market data", nil
		}

		/*
			if marketRequestCounter == 0 && print {
				//fmt.Println("Sample symbols from second API:")
				count := 0
				for symbol := range snapshots {
					if count < 5 {
						//fmt.Printf("Sample market data symbol: %s\n", symbol)
						count++
					} else {
						break
					}
				}
			}
		*/

		for symbol, data := range snapshots {
			//fmt.Println("\n symbol: ", symbol)

			option, exists := optionMap[symbol]
			//fmt.Println("\n option found with same symbol: ", option, "(symbol: ", option.Symbol, ")")

			if !exists {
				if !symbolsNotFound[symbol] {
					symbolsNotFound[symbol] = true
					if print {
						fmt.Printf("Warning: No matching option found for symbol: %s\n", symbol)
					}
				}
				continue
			}

			snapshot, ok := data.(map[string]interface{})
			if !ok {
				continue
			}

			// Parse DailyBar
			if dailyBar, ok := snapshot["dailyBar"].(map[string]interface{}); ok {
				timestamp, _ := time.Parse(time.RFC3339, getString(dailyBar["t"]))
				option.DailyBar.Close = getFloat64(dailyBar["c"])
				option.DailyBar.High = getFloat64(dailyBar["h"])
				option.DailyBar.Low = getFloat64(dailyBar["l"])
				option.DailyBar.NumberOfTrades = getInt(dailyBar["n"])
				option.DailyBar.Open = getFloat64(dailyBar["o"])
				option.DailyBar.Timestamp = timestamp
				option.DailyBar.Volume = getInt(dailyBar["v"])
				option.DailyBar.VWAP = getFloat64(dailyBar["vw"])
			}

			// Parse PrevDailyBar
			if prevDailyBar, ok := snapshot["prevDailyBar"].(map[string]interface{}); ok {
				timestamp, _ := time.Parse(time.RFC3339, getString(prevDailyBar["t"]))
				option.PrevDailyBar.Close = getFloat64(prevDailyBar["c"])
				option.PrevDailyBar.High = getFloat64(prevDailyBar["h"])
				option.PrevDailyBar.Low = getFloat64(prevDailyBar["l"])
				option.PrevDailyBar.NumberOfTrades = getInt(prevDailyBar["n"])
				option.PrevDailyBar.Open = getFloat64(prevDailyBar["o"])
				option.PrevDailyBar.Timestamp = timestamp
				option.PrevDailyBar.Volume = getInt(prevDailyBar["v"])
				option.PrevDailyBar.VWAP = getFloat64(prevDailyBar["vw"])
			}

			// Parse MinuteBar
			if minuteBar, ok := snapshot["minuteBar"].(map[string]interface{}); ok {
				timestamp, _ := time.Parse(time.RFC3339, getString(minuteBar["t"]))
				option.MinuteBar.Close = getFloat64(minuteBar["c"])
				option.MinuteBar.High = getFloat64(minuteBar["h"])
				option.MinuteBar.Low = getFloat64(minuteBar["l"])
				option.MinuteBar.NumberOfTrades = getInt(minuteBar["n"])
				option.MinuteBar.Open = getFloat64(minuteBar["o"])
				option.MinuteBar.Timestamp = timestamp
				option.MinuteBar.Volume = getInt(minuteBar["v"])
				option.MinuteBar.VWAP = getFloat64(minuteBar["vw"])
			}

			// Parse Greeks
			if greeks, ok := snapshot["greeks"].(map[string]interface{}); ok {
				option.Greeks.Delta = getFloat64(greeks["delta"])
				option.Greeks.Gamma = getFloat64(greeks["gamma"])
				option.Greeks.Rho = getFloat64(greeks["rho"])
				option.Greeks.Theta = getFloat64(greeks["theta"])
				option.Greeks.Vega = getFloat64(greeks["vega"])
			}

			// Parse ImpliedVolatility
			option.ImpliedVol = getFloat64(snapshot["impliedVolatility"])

			// Parse LatestQuote
			if quote, ok := snapshot["latestQuote"].(map[string]interface{}); ok {
				timestamp, _ := time.Parse(time.RFC3339Nano, getString(quote["t"]))
				option.LatestQuote.AskPrice = getFloat64(quote["ap"])
				option.LatestQuote.AskSize = getInt(quote["as"])
				option.LatestQuote.AskExchange = getString(quote["ax"])
				option.LatestQuote.BidPrice = getFloat64(quote["bp"])
				option.LatestQuote.BidSize = getInt(quote["bs"])
				option.LatestQuote.BidExchange = getString(quote["bx"])
				option.LatestQuote.Condition = getString(quote["c"])
				option.LatestQuote.Timestamp = timestamp
			}

			// Parse LatestTrade
			if trade, ok := snapshot["latestTrade"].(map[string]interface{}); ok {
				timestamp, _ := time.Parse(time.RFC3339Nano, getString(trade["t"]))
				option.LatestTrade.Condition = getString(trade["c"])
				option.LatestTrade.Price = getFloat64(trade["p"])
				option.LatestTrade.Size = getInt(trade["s"])
				option.LatestTrade.Timestamp = timestamp
				option.LatestTrade.Exchange = getString(trade["x"])
			}

			marketDataProcessed++
			if print {
				fmt.Printf("\rProcessed market data for %s", symbol)
			}
		}

		marketRequestCounter++
		if print {
			fmt.Printf("\n%v market data API requests made - %v options updated\n", marketRequestCounter, marketDataProcessed)
			fmt.Printf("Total symbols not found: %d\n", len(symbolsNotFound))
		}

		// Check for next page token
		nextToken, ok = marketData["next_page_token"].(string)
		if !ok || nextToken == "" {
			if print {
				fmt.Println("\nMarket data fetching completed")
			}
			break
		}

		// Update URL with next page token
		marketDataURL = fmt.Sprintf("https://data.alpaca.markets/v1beta1/options/snapshots/%s?feed=indicative&limit=1000&page_token=%s&strike_price_gte=%v&strike_price_lte=%v&expiration_date_gte=%s&expiration_date_lte=%s&type=%s",
			optreq.Ticker,
			nextToken,
			optreq.StrikeRange[0],
			optreq.StrikeRange[1],
			optreq.DateRange[0],
			optreq.DateRange[1],
			optreq.Contract_type,
		)
	}

	if print {
		fmt.Printf("Total options processed: %d, Market data updates: %d\n", len(options), marketDataProcessed)
		if len(symbolsNotFound) > 0 {
			fmt.Println("\nSymbols from second API that weren't found in first API:")
			for symbol := range symbolsNotFound {
				fmt.Println(symbol)
			}
		}
	}

	return options, log, nil
}

// Helper function to build comma-separated list of option symbols
func buildSymbolsList(options []Option) string {
	var symbols []string
	for _, opt := range options {
		symbols = append(symbols, opt.Symbol)
	}
	return strings.Join(symbols, ",")
}

// Helper function to parse Bar data
func parseBar(data map[string]interface{}) *Bar {
	timestamp, _ := time.Parse(time.RFC3339Nano, getString(data["t"]))
	return &Bar{
		Close:          getFloat64(data["c"]),
		High:           getFloat64(data["h"]),
		Low:            getFloat64(data["l"]),
		NumberOfTrades: getInt(data["n"]),
		Open:           getFloat64(data["o"]),
		Timestamp:      timestamp,
		Volume:         getInt(data["v"]),
		VWAP:           getFloat64(data["vw"]),
	}
}

// Helper function to parse string to float64
func parseFloat64(s string) float64 {
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return f
}

// Helper function to parse string to int
func parseInt(s string) int {
	i, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return i
}

func APIRequest(url string, iteration int) (string, string, error) {
	debug := false

	if APIKeyID == "" || APISecretKey == "" {
		return "", "", fmt.Errorf("APIKeyID or APISecretKey is not set")
	}

	/*
		config, err := loadConfig()
		if err != nil {
			return "", "", fmt.Errorf("error loading config: %v", err)
		}
	*/

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", "", fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Add("accept", "application/json")
	req.Header.Add("APCA-API-KEY-ID", APIKeyID)
	req.Header.Add("APCA-API-SECRET-KEY", APISecretKey)

	var res *http.Response
	res, err = http.DefaultClient.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("error making request: %v", err)
	}

	// Retry logic for nil response or non-200 status
	retryNr := 1
	maxRetry := 12
	for res == nil || res.StatusCode != http.StatusOK {
		if res == nil {
			fmt.Printf("Response is nil (possibly due to connection loss), waiting for 5 seconds and retrying (%d)\n", retryNr)
		} else {
			fmt.Printf("Received status code %d, waiting for 5 seconds and retrying (%d)\n", res.StatusCode, retryNr)
			// Read error response body if available
			if res.Body != nil {
				errBody, _ := io.ReadAll(res.Body)
				res.Body.Close()
				fmt.Printf("Error response: %s\n", string(errBody))
			}
		}
		waitTime := 5 * time.Second
		time.Sleep(waitTime)
		res, err = http.DefaultClient.Do(req)
		if err != nil {
			return "", "", fmt.Errorf("error in retry attempt %d: %v", retryNr, err)
		}
		retryNr++
		if retryNr >= maxRetry {
			return "", "", fmt.Errorf("max retries reached (%d) with status code %d", maxRetry, res.StatusCode)
		}
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return "", "", fmt.Errorf("error reading response body: %v", err)
	}

	if debug {
		fmt.Println("Request URL:", url)
		fmt.Println("Response Status:", res.Status)
		fmt.Println("Response Length:", len(string(body)))
	}

	// Check if response is empty
	if len(body) == 0 {
		return "", "", fmt.Errorf("empty response received")
	}

	// Validate JSON response
	var jsonResponse map[string]interface{}
	if err := json.Unmarshal(body, &jsonResponse); err != nil {
		return "", "", fmt.Errorf("invalid JSON response: %v", err)
	}

	// Check for API error messages
	if errMsg, ok := jsonResponse["message"].(string); ok && errMsg != "" {
		return "", "", fmt.Errorf("API error: %s", errMsg)
	}

	// Check for option_contracts in response
	if contracts, ok := jsonResponse["option_contracts"]; !ok {
		// If no option_contracts field and no error message, might be a different type of response
		if debug {
			fmt.Println("Response does not contain option_contracts field")
		}
	} else if contracts == nil {
		return "", "", fmt.Errorf("option_contracts field is null")
	}

	if debug {
		fmt.Println("API Request successfully made")
	}

	return res.Status, string(body), nil
}

func stndrdth(n int) string {
	switch math.Mod(float64(n), 10) {
	case 1:
		return "st"
	case 2:
		return "nd"
	case 3:
		return "rd"
	default:
		return "th"
	}
}

func URLoption(req OptionURLReq) (string, error) {
	return "", nil
}

func MergeRequests(optreqs []OptionURLReq, nMax int) ([]Option, error) {

	if APIKeyID == "" || APISecretKey == "" {
		return []Option{}, fmt.Errorf("APIKeyID or APISecretKey is not set")
	}
	log.Println("APIKeyID and APISecretKey are set")

	var options []Option
	log := ""
	var msg string
	var options_tmp []Option
	var err error
	for _, optreq := range optreqs {
		options_tmp, msg, err = GetOptions(optreq, nMax)
		if err != nil {
			return []Option{}, fmt.Errorf("error getting options: %v", err)
		}
		for _, opt := range options_tmp {
			options = append(options, opt)
		}
		log += msg
	}
	return options, nil
}

func ProvideApiKey(apiKeyID, apiSecretKey string) {
	APIKeyID = apiKeyID
	APISecretKey = apiSecretKey
}

func handleApiKeyInit() {
	if APIKeyID == "" || APISecretKey == "" {
		log.Println("Warning: APIKeyID or APISecretKey is not set by environment variables. Trying to load from config.json")
		config, err := loadConfig()
		if err != nil {
			log.Println("Error loading config: %v", err)
			return
		}
		APIKeyID = config.APIKeyID
		APISecretKey = config.APISecretKey
	}
}

func SingleQuote(ticker string) (float64, error) {
	if APIKeyID == "" || APISecretKey == "" {
		return 0, fmt.Errorf("APIKeyID or APISecretKey is not set")
	}

	url := fmt.Sprintf("https://data.alpaca.markets/v2/stocks/quotes/latest?symbols=%s", ticker)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0, fmt.Errorf("Error creating request: %v", err)
	}

	req.Header.Add("accept", "application/json")
	req.Header.Add("APCA-API-KEY-ID", APIKeyID)
	req.Header.Add("APCA-API-SECRET-KEY", APISecretKey)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("Error making request: %v", err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return 0, fmt.Errorf("Error reading response: %v", err)
	}

	// Parse the JSON response
	var data map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		return 0, fmt.Errorf("Error parsing JSON: %v", err)
	}

	// Navigate through the JSON structure to get the ask price
	quotes, ok := data["quotes"].(map[string]interface{})
	if !ok {
		return 0, fmt.Errorf("No quotes data found in response")
	}

	tickerData, ok := quotes[ticker].(map[string]interface{})
	if !ok {
		return 0, fmt.Errorf("No data found for ticker %s", ticker)
	}

	ap, ok := tickerData["ap"].(float64)
	if !ok {
		return 0, fmt.Errorf("No ask price found for ticker %s", ticker)
	}

	return ap, nil
}

func getBool(v interface{}) bool {
	if b, ok := v.(bool); ok {
		return b
	}

	return false
}
