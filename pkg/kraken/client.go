package kraken

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	APIURL            = "https://api.kraken.com"
	WSS_URL           = "wss://ws-auth.kraken.com/v2"
	API_VERSION       = "0"
	REST_TIMEOUT      = 10 * time.Second
	HeartbeatInterval = 10 * time.Second
	ReconnectDelay    = 5 * time.Second
)

type ConnectionState int

const (
	Disconnected ConnectionState = iota
	Connecting
	Connected
)

type Client struct {
	apiKey     string
	apiSecret  string
	httpClient *http.Client
	ws         *websocket.Conn
	wsLock     sync.Mutex
	state      ConnectionState
	stateLock  sync.RWMutex
	done       chan struct{}
}

type TickerInfo struct {
	Last   float64
	Ask    float64
	Bid    float64
	Volume float64
}

type VolumeDistribution string

const (
	EvenDistribution   VolumeDistribution = "even"
	NormalDistribution VolumeDistribution = "normal" // More volume in middle
	CustomDistribution VolumeDistribution = "custom" // User-provided weights
)

func NewClient(apiKey, apiSecret string) *Client {
	return &Client{
		apiKey:    apiKey,
		apiSecret: apiSecret,
		httpClient: &http.Client{
			Timeout: REST_TIMEOUT,
		},
	}
}

// getSignature creates API authentication signature per Kraken documentation
func (c *Client) getSignature(path string, nonce string, postData string) string {
	// Create sha256 hash of nonce and post data
	sha := sha256.New()
	sha.Write([]byte(nonce + postData))
	shasum := sha.Sum(nil)

	// Decode base64 secret
	key, _ := base64.StdEncoding.DecodeString(c.apiSecret)

	// Create HMAC
	mac := hmac.New(sha512.New, key)
	mac.Write(append([]byte(path), shasum...))
	macsum := mac.Sum(nil)

	return base64.StdEncoding.EncodeToString(macsum)
}

// AddOrder places a new order via REST API
func (c *Client) AddOrder(ctx context.Context, req OrderRequest) (*OrderResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid order: %w", err)
	}

	endpoint := "/0/private/AddOrder"

	// Create form data
	data := url.Values{}
	data.Set("nonce", strconv.FormatInt(time.Now().UnixNano(), 10))
	data.Set("ordertype", string(req.Type))
	data.Set("type", req.Side)
	data.Set("volume", req.Volume)
	data.Set("pair", req.Pair)

	if req.Price != "" {
		data.Set("price", req.Price)
	}
	if req.Leverage != "" {
		data.Set("leverage", req.Leverage)
	}

	// Create signature
	signature := c.getSignature(endpoint, data.Get("nonce"), data.Encode())

	// Create request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", APIURL+endpoint, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Add("API-Key", c.apiKey)
	httpReq.Header.Add("API-Sign", signature)
	httpReq.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	// Execute request
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var result struct {
		Error  []string      `json:"error"`
		Result OrderResponse `json:"result"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if len(result.Error) > 0 {
		return nil, fmt.Errorf("API error: %v", result.Error)
	}

	return &result.Result, nil
}

// ConnectWebSocket establishes WebSocket connection
func (c *Client) ConnectWebSocket(ctx context.Context) error {
	c.setState(Connecting)

	dialer := websocket.Dialer{
		HandshakeTimeout: 10 * time.Second,
	}

	conn, _, err := dialer.DialContext(ctx, WSS_URL, nil)
	if err != nil {
		c.setState(Disconnected)
		return fmt.Errorf("failed to connect: %w", err)
	}

	c.wsLock.Lock()
	c.ws = conn
	c.wsLock.Unlock()

	c.setState(Connected)
	c.done = make(chan struct{})

	// Start heartbeat
	go c.heartbeat(ctx)
	// Start reconnection monitor
	go c.monitorConnection(ctx)

	return nil
}

func (c *Client) heartbeat(ctx context.Context) {
	ticker := time.NewTicker(HeartbeatInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-c.done:
			return
		case <-ticker.C:
			c.wsLock.Lock()
			err := c.ws.WriteMessage(websocket.PingMessage, nil)
			c.wsLock.Unlock()

			if err != nil {
				fmt.Printf("heartbeat failed: %v\n", err)
				c.reconnect(ctx)
			}
		}
	}
}

func (c *Client) monitorConnection(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-c.done:
			return
		default:
			c.wsLock.Lock()
			_, _, err := c.ws.ReadMessage()
			c.wsLock.Unlock()

			if err != nil {
				fmt.Printf("connection error: %v\n", err)
				c.reconnect(ctx)
			}
		}
	}
}

func (c *Client) reconnect(ctx context.Context) {
	if c.getState() == Connecting {
		return
	}

	c.setState(Connecting)

	// Close existing connection
	c.Close()

	// Attempt to reconnect with backoff
	for {
		select {
		case <-ctx.Done():
			return
		default:
			if err := c.ConnectWebSocket(ctx); err == nil {
				return
			}
			time.Sleep(ReconnectDelay)
		}
	}
}

// AddOrderWS places a new order via WebSocket API
func (c *Client) AddOrderWS(ctx context.Context, req WSOrderRequest) error {
	if c.ws == nil {
		return fmt.Errorf("websocket connection not established")
	}

	message := map[string]interface{}{
		"method": "add_order",
		"params": req,
	}

	c.wsLock.Lock()
	err := c.ws.WriteJSON(message)
	c.wsLock.Unlock()

	if err != nil {
		return fmt.Errorf("failed to send order: %w", err)
	}

	// Read response
	_, msgBytes, err := c.ws.ReadMessage()
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	var response struct {
		Error  string `json:"error"`
		Result struct {
			OrderID string `json:"order_id"`
		} `json:"result"`
	}

	if err := json.Unmarshal(msgBytes, &response); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	if response.Error != "" {
		return fmt.Errorf("API error: %s", response.Error)
	}

	return nil
}

func (c *Client) Close() error {
	if c.ws != nil {
		return c.ws.Close()
	}
	return nil
}

// SubscribeToTicker subscribes to real-time price updates
func (c *Client) SubscribeToTicker(ctx context.Context, pair string, priceChan chan<- float64) error {
	sub := map[string]interface{}{
		"event": "subscribe",
		"pair":  []string{pair},
		"subscription": map[string]string{
			"name": "ticker",
		},
	}

	c.wsLock.Lock()
	err := c.ws.WriteJSON(sub)
	c.wsLock.Unlock()
	if err != nil {
		return fmt.Errorf("failed to subscribe: %w", err)
	}

	go func() {
		defer close(priceChan)
		for {
			select {
			case <-ctx.Done():
				return
			default:
				_, message, err := c.ws.ReadMessage()
				if err != nil {
					fmt.Printf("websocket read error: %v\n", err)
					return
				}

				var tickerData []interface{}
				if err := json.Unmarshal(message, &tickerData); err != nil {
					fmt.Printf("parse error: %v\n", err)
					continue
				}

				// Check if it's a ticker update (array with price data)
				if len(tickerData) < 4 || tickerData[2] != "ticker" {
					continue
				}

				data := tickerData[1].(map[string]interface{})
				if price, ok := data["c"].([]interface{}); ok && len(price) > 0 {
					if lastPrice, ok := price[0].(string); ok {
						if p, err := strconv.ParseFloat(lastPrice, 64); err == nil {
							priceChan <- p
						}
					}
				}
			}
		}
	}()

	return nil
}

func (c *Client) GetTickerPrice(ctx context.Context, pair string) (*TickerInfo, error) {
	// TODO: Implement actual Kraken API call
	// This is a mock implementation for now
	return &TickerInfo{
		Last:   50000.0, // Mock price
		Ask:    50001.0,
		Bid:    49999.0,
		Volume: 100.0,
	}, nil
}

type TrailingEntryConfig struct {
	Pair         string
	Side         string
	UpperBand    float64
	LowerBand    float64
	TotalVolume  float64
	NumOrders    int
	Interval     time.Duration
	Distribution VolumeDistribution
	Weights      []float64 // Optional custom weights
}

func calculateOrderVolumes(config TrailingEntryConfig) []float64 {
	volumes := make([]float64, config.NumOrders)

	switch config.Distribution {
	case NormalDistribution:
		// Approximate normal distribution weights
		// More volume in the middle, less at the edges
		middle := float64(config.NumOrders-1) / 2
		sum := 0.0

		for i := 0; i < config.NumOrders; i++ {
			// Calculate distance from middle (0 to 1)
			distance := math.Abs(float64(i)-middle) / middle
			// Convert to a weight (1 at middle, smaller at edges)
			weight := 1 - (distance * 0.6) // 0.6 controls the curve steepness
			volumes[i] = weight
			sum += weight
		}

		// Normalize to total volume
		for i := range volumes {
			volumes[i] = (volumes[i] / sum) * config.TotalVolume
		}

	case CustomDistribution:
		if len(config.Weights) != config.NumOrders {
			// Fall back to even distribution if weights are invalid
			for i := range volumes {
				volumes[i] = config.TotalVolume / float64(config.NumOrders)
			}
			return volumes
		}

		// Normalize custom weights to total volume
		sum := 0.0
		for _, w := range config.Weights {
			sum += w
		}
		for i, w := range config.Weights {
			volumes[i] = (w / sum) * config.TotalVolume
		}

	default: // EvenDistribution
		for i := range volumes {
			volumes[i] = config.TotalVolume / float64(config.NumOrders)
		}
	}

	return volumes
}

func (c *Client) ExecuteTrailingEntry(ctx context.Context, config TrailingEntryConfig) error {
	volumes := calculateOrderVolumes(config)
	priceStep := (config.UpperBand - config.LowerBand) / float64(config.NumOrders-1)

	priceChan := make(chan float64)
	if err := c.SubscribeToTicker(ctx, config.Pair, priceChan); err != nil {
		return fmt.Errorf("failed to subscribe to ticker: %w", err)
	}

	placedPrices := make(map[float64]bool)

	for price := range priceChan {
		if price >= config.LowerBand && price <= config.UpperBand {
			for i := 0; i < config.NumOrders; i++ {
				var orderPrice float64
				if config.Side == "buy" {
					orderPrice = config.UpperBand - (float64(i) * priceStep)
				} else {
					orderPrice = config.LowerBand + (float64(i) * priceStep)
				}

				if placedPrices[orderPrice] {
					continue
				}

				req := OrderRequest{
					Pair:   config.Pair,
					Type:   LimitOrder,
					Side:   config.Side,
					Volume: strconv.FormatFloat(volumes[i], 'f', 8, 64),
					Price:  strconv.FormatFloat(orderPrice, 'f', 2, 64),
				}

				if _, err := c.AddOrder(ctx, req); err != nil {
					return fmt.Errorf("failed to place order: %w", err)
				}

				placedPrices[orderPrice] = true
				fmt.Printf("Placed %s order: %v %v at %v\n",
					config.Side, volumes[i], config.Pair, orderPrice)
			}

			// All orders placed, we're done
			return nil
		}
	}

	return nil
}

func (c *Client) setState(state ConnectionState) {
	c.stateLock.Lock()
	defer c.stateLock.Unlock()
	c.state = state
}

func (c *Client) getState() ConnectionState {
	c.stateLock.RLock()
	defer c.stateLock.RUnlock()
	return c.state
}
