package alpacaApiClient

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"strings"
	"time"
)

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

	url = "https://data.alpaca.markets/v1beta1/options/snapshots/TSLA?feed=indicative&limit=1000&type=call&page_token="
	optreq := OptionURLReq{
		Ticker:        "TSLA",
		Contract_type: "call",
		StrikeRange:   []int{0, 10000},
		DateRange:     []string{"2025-05-23", "2027-01-23"},
	}

	options, _, err := GetOptions(optreq, -1)
	check(err)
	fmt.Println(len(options))
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
		fmt.Println(err)
		return ""
	}
	defer file.Close()

	// Decode the JSON data from the file
	var readStr string
	if err := json.NewDecoder(file).Decode(&readStr); err != nil {
		fmt.Println(err)
		return ""
	}

	// Return the decoded string
	return fmt.Sprint(readStr)
}

func JsonToOptions(path string) []Option {
	str := LoadJson(path)

	// Parse the JSON into a map
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(str), &data); err != nil {
		fmt.Println("Error parsing JSON:", err)
		return nil
	}

	// Get the snapshots
	snapshots, ok := data["snapshots"].(map[string]interface{})
	if !ok {
		fmt.Println("Error: snapshots not found in JSON")
		return nil
	}

	var options []Option
	for id, snapshot := range snapshots {
		snapshotMap, ok := snapshot.(map[string]interface{})
		if !ok {
			continue
		}

		// Get the latestQuote
		latestQuote, ok := snapshotMap["option_contracts"].(map[string]interface{})
		if !ok {
			continue
		}

		// Create new Option from latestQuote data
		newOption := Option{
			ID:                id,
			Symbol:            getString(latestQuote["symbol"]),
			Name:              getString(latestQuote["name"]),
			Status:            getString(latestQuote["status"]),
			Tradable:          getBool(latestQuote["tradable"]),
			ExpirationDate:    getString(latestQuote["expiration_date"]),
			RootSymbol:        getString(latestQuote["root_symbol"]),
			UnderlyingSymbol:  getString(latestQuote["underlying_symbol"]),
			UnderlyingAssetID: getString(latestQuote["underlying_asset_id"]),
			Type:              getString(latestQuote["type"]),
			Style:             getString(latestQuote["style"]),
			StrikePrice:       getFloat64(latestQuote["strike_price"]),
			Multiplier:        getInt(latestQuote["multiplier"]),
			Size:              getInt(latestQuote["size"]),
			OpenInterest:      getInt(latestQuote["open_interest"]),
			OpenInterestDate:  getString(latestQuote["open_interest_date"]),
			ClosePrice:        getFloat64(latestQuote["close_price"]),
			ClosePriceDate:    getString(latestQuote["close_price_date"]),
			PPIND:             getBool(latestQuote["ppind"]),
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

type Option struct {
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

	// Base URL for options chain with strike price and expiration date filtering
	// Using limit=1000 to minimize API requests (maximum allowed by API)
	baseURL := fmt.Sprintf("https://data.alpaca.markets/v1beta1/options/snapshots/%s", optreq.Ticker)
	url := fmt.Sprintf("%s?feed=indicative&limit=1000&type=%s&strike_price_gte=%d&strike_price_lte=%d&expiration_date_gte=%s&expiration_date_lte=%s",
		baseURL,
		optreq.Contract_type,
		optreq.StrikeRange[0],
		optreq.StrikeRange[1],
		optreq.DateRange[0],
		optreq.DateRange[1])

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

			// Get the snapshots
			snapshots, ok := data["snapshots"].(map[string]interface{})
			if !ok {
				return nil, log, fmt.Errorf("invalid response format: snapshots not found")
			}

			// Process each option in the current page
			for id, snapshot := range snapshots {
				// Skip if we've already processed this option
				if processedIDs[id] {
					continue
				}
				processedIDs[id] = true

				snapshotMap, ok := snapshot.(map[string]interface{})
				if !ok {
					continue
				}

				// Get the latestQuote
				latestQuote, ok := snapshotMap["option_contracts"].(map[string]interface{})
				if !ok {
					continue
				}

				// Create new Option
				newOption := Option{
					ID:                getString(latestQuote["id"]),
					Symbol:            getString(latestQuote["symbol"]),
					Name:              getString(latestQuote["name"]),
					Status:            getString(latestQuote["status"]),
					Tradable:          getBool(latestQuote["tradable"]),
					ExpirationDate:    getString(latestQuote["expiration_date"]),
					RootSymbol:        getString(latestQuote["root_symbol"]),
					UnderlyingSymbol:  getString(latestQuote["underlying_symbol"]),
					UnderlyingAssetID: getString(latestQuote["underlying_asset_id"]),
					Type:              getString(latestQuote["type"]),
					Style:             getString(latestQuote["style"]),
					StrikePrice:       getFloat64(latestQuote["strike_price"]),
					Multiplier:        getInt(latestQuote["multiplier"]),
					Size:              getInt(latestQuote["size"]),
					OpenInterest:      getInt(latestQuote["open_interest"]),
					OpenInterestDate:  getString(latestQuote["open_interest_date"]),
					ClosePrice:        getFloat64(latestQuote["close_price"]),
					ClosePriceDate:    getString(latestQuote["close_price_date"]),
					PPIND:             getBool(latestQuote["ppind"]),
				}

				options = append(options, newOption)
				optionCounter++

				if print {
					fmt.Printf("\r %v API requests successfully made - %v available options found", requestCounter+1, optionCounter)
				}

				// Check if we've reached the maximum number of options
				if nMax > 0 && len(options) >= nMax {
					if print {
						fmt.Println("")
					}
					return options, log, nil
				}
			}

			// Check for next page token
			nextToken, ok := data["next_page_token"].(string)
			if !ok || nextToken == "" {
				if print {
					fmt.Println("")
				}
				return options, log, nil
			}

			// Update URL for next page with strike price and expiration date filtering
			url = fmt.Sprintf("%s?feed=indicative&limit=1000&type=%s&strike_price_gte=%d&strike_price_lte=%d&expiration_date_gte=%s&expiration_date_lte=%s&page_token=%s",
				baseURL,
				optreq.Contract_type,
				optreq.StrikeRange[0],
				optreq.StrikeRange[1],
				optreq.DateRange[0],
				optreq.DateRange[1],
				nextToken)
			requestCounter++

			if print {
				fmt.Printf("\r %v API requests successfully made - %v available options found", requestCounter+1, optionCounter)
			}
		}
	}
}

func APIRequest(url string, iteration int) (string, string, error) {
	debug := false

	config, err := loadConfig()
	if err != nil {
		return "", "", fmt.Errorf("error loading config: %v", err)
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", "", err
	}

	req.Header.Add("accept", "application/json")
	req.Header.Add("APCA-API-KEY-ID", config.APIKeyID)
	req.Header.Add("APCA-API-SECRET-KEY", config.APISecretKey)

	var res *http.Response
	res, err = http.DefaultClient.Do(req)
	if err != nil {
		return "", "", err
	}

	// Retry logic for nil response or non-200 status
	retryNr := 1
	maxRetry := 12
	for res == nil || res.StatusCode != http.StatusOK {
		if res == nil {
			fmt.Printf("Response is nil (possibly due to connection loss), waiting for 5 seconds and retrying (%d)\n", retryNr)
		} else {
			fmt.Printf("Received status code %d, waiting for 5 seconds and retrying (%d)\n", res.StatusCode, retryNr)
		}
		waitTime := 5 * time.Second
		time.Sleep(waitTime)
		res, err = http.DefaultClient.Do(req)
		if err != nil {
			return "", "", err
		}
		retryNr++
		if retryNr >= maxRetry {
			fmt.Printf("API Request failed after %d retries. Last status: %d\n", maxRetry, res.StatusCode)
			return "", "", fmt.Errorf("max retries reached with status code %d", res.StatusCode)
		}
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return "", "", err
	}

	if debug {
		fmt.Println("Request made:")
		fmt.Println(url)
		fmt.Println("Response and body:")
		fmt.Println(res, "\n", string(body))
	}

	// Check for error in response
	if strings.Contains(string(body), "ERROR") {
		errormsg := strings.Split(string(body), "\"error\":")[1]
		errormsg = strings.Split(errormsg, "}")[0]
		fmt.Printf("An error occurred: \n%s\nWaiting 60 seconds and retrying...\n", errormsg)
		time.Sleep(60 * time.Second)
		return APIRequest(url, 1)
	}

	// Check if response is empty or invalid
	if len(strings.Split(string(body), "\"snapshots\":{")) < 2 {
		if debug {
			fmt.Println("no result")
		}
		if iteration < 3 {
			if debug {
				fmt.Printf("ReRequesting in 500 milliseconds. That will be the %v%v reRequest.\n", iteration, stndrdth(iteration))
			}
			time.Sleep(500 * time.Millisecond)
			return APIRequest(url, iteration+1)
		}
		return "", "", fmt.Errorf("no results")
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

func singleQuote(ticker string) (float64, error) {
	config, err := loadConfig()
	if err != nil {
		return 0, fmt.Errorf("Error loading config: %v", err)
	}

	url := fmt.Sprintf("https://data.alpaca.markets/v2/stocks/quotes/latest?symbols=%s", ticker)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0, fmt.Errorf("Error creating request: %v", err)
	}

	req.Header.Add("accept", "application/json")
	req.Header.Add("APCA-API-KEY-ID", config.APIKeyID)
	req.Header.Add("APCA-API-SECRET-KEY", config.APISecretKey)

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
