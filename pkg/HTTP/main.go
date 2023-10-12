package http

import (
	"encoding/json"
	"log"
	"net/http"
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

type Coin struct {
	ID     string `json:"id"`
	Symbol string `json:"symbol"`
	Name   string `json:"name"`
}

func (c *Client) GetCoins() ([]Coin, error) {
	c.LogRequest("GetCoins")
	data, err := c.HTTPClient.Get(c.BaseURL + "/coins/list")
	if err != nil {
		return nil, err
	}
	defer data.Body.Close()

	var res []Coin
	err = json.NewDecoder(data.Body).Decode(&res)
	if err != nil {
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
		return TokenPriceResponse{}, err
	}
	defer data.Body.Close()

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
		return DefiDataRes{}, err
	}
	defer data.Body.Close()
	var res DefiDataRes
	err = json.NewDecoder(data.Body).Decode(&res)
	if err != nil {
		return DefiDataRes{}, err
	}
	return res, nil
}

func (c *Client) LogRequest(method string) {
	c.Logger.Printf("%v", method)
}