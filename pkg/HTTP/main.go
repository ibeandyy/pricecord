package http

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"reflect"
	"strings"
)

type Client struct {
	BaseURL    string
	HTTPClient *http.Client
	Logger     *log.Logger
}

func NewClient() *Client {
	return &Client{
		BaseURL:    "https://api.coingecko.com/api/v3",
		HTTPClient: &http.Client{},
		Logger:     log.New(log.Writer(), "HTTP Client", log.LstdFlags),
	}
}

type Token struct {
	ID     string `json:"id"`
	Symbol string `json:"symbol"`
	Name   string `json:"name"`
}

func (c *Client) GetTokens() ([]Token, error) {
	c.LogRequest("GetTokens")
	data, err := c.HTTPClient.Get(c.BaseURL + "/coins/list")
	if err != nil {
		c.LogError("Error getting data", err.Error())
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			c.LogRequest("Error closing body", err.Error())
		}
	}(data.Body)

	var res []Token
	err = json.NewDecoder(data.Body).Decode(&res)
	if err != nil {
		c.LogError("Error decoding json", err.Error())
		return nil, err
	}
	return res, nil
}

type TokenPriceResponse struct {
	Data map[string]USDValue `json:"-"`
}

type USDValue struct {
	USD float64 `json:"usd"`
}

// GetTokenPrices takes in a slice of token names and returns the price of the tokens in USD
func (c *Client) GetTokenPrices(tokenName []string) (TokenPriceResponse, error) {
	c.LogRequest("GetTokenPrices")
	tokens := strings.Join(tokenName, ",")
	data, err := c.HTTPClient.Get(c.BaseURL + "/simple/price?ids=" + tokens + "&vs_currencies=usd")
	if err != nil {
		c.LogError("Error getting data", err.Error())
		return TokenPriceResponse{}, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			c.LogRequest("Error closing body", err.Error())
		}
	}(data.Body)

	var res TokenPriceResponse
	err = json.NewDecoder(data.Body).Decode(&res)
	if err != nil {
		c.LogError("Error decoding json", err.Error())
		return TokenPriceResponse{}, err
	}

	return res, nil
}

func (c *Client) GetTokenPrice(tokenName string) (TokenPriceResponse, error) {
	c.LogRequest("GetTokenPrices")
	data, err := c.HTTPClient.Get(c.BaseURL + "/simple/price?ids=" + tokenName + "&vs_currencies=usd")
	if err != nil {
		return TokenPriceResponse{}, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			c.LogRequest("Error closing body", err.Error())
		}
	}(data.Body)

	var res TokenPriceResponse
	err = json.NewDecoder(data.Body).Decode(&res)
	if err != nil {
		return TokenPriceResponse{}, err
	}

	return res, nil
}

type DefiDataRes struct {
	Data struct {
		DefiMarketCap        string  `json:"defi_market_cap"`
		EthMarketCap         string  `json:"eth_market_cap"`
		DefiToEthRatio       string  `json:"defi_to_eth_ratio"`
		TradingVolume24H     string  `json:"trading_volume_24h"`
		DefiDominance        string  `json:"defi_dominance"`
		TopCoinName          string  `json:"top_coin_name"`
		TopCoinDefiDominance float64 `json:"top_coin_defi_dominance"`
	} `json:"data"`
}

func (c *Client) GetDeFiData() (DefiDataRes, error) {
	c.LogRequest("GetDeFiData")
	data, err := c.HTTPClient.Get(c.BaseURL + "/global/decentralized_finance_defi")
	if err != nil {
		c.LogError("Error getting data", err.Error())
		return DefiDataRes{}, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			c.LogError("Error closing body", err.Error())
		}
	}(data.Body)
	var res DefiDataRes
	err = json.NewDecoder(data.Body).Decode(&res)
	if err != nil {
		c.LogError("Error decoding json", err.Error())
		return DefiDataRes{}, err
	}
	return res, nil
}

func (c *Client) LogRequest(method ...string) {
	c.Logger.Printf("[I] v", method)
}

func (c *Client) LogError(error ...string) {
	c.Logger.Printf("[E] %v", error)

}

func (c *Client) GetDefiOther(key string) (string, string) {
	data, err := c.GetDeFiData()
	if err != nil {
		c.LogError("Error getting defi data", err.Error())
		return "", ""
	}
	r := reflect.ValueOf(data)
	field := reflect.Indirect(r).FieldByName(key)
	if field.IsValid() {
		return key, fmt.Sprintf("%v", field.Interface())
	}
	return "", ""
}
