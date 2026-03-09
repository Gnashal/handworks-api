package config

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"handworks-api/types"
	"net/http"
	"time"
)

type PaymongoClient struct {
	SecretKey string
	BaseURL   string
	HTTP      *http.Client
}

func NewPaymongoClient(secretKey string) *PaymongoClient {
	return &PaymongoClient{
		SecretKey: secretKey,
		BaseURL:   "https://api.paymongo.com/v1",
		HTTP:      &http.Client{Timeout: 15 * time.Second},
	}
}

func (c *PaymongoClient) CreatePaymentIntent(
	ctx context.Context,
	payload any,
) (*types.PaymentIntentResponse, error) {

	url := c.BaseURL + "/payment_intents"

	jsonBody, _ := json.Marshal(payload)

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	// Basic auth
	encoded := base64.StdEncoding.EncodeToString([]byte(c.SecretKey + ":"))
	req.Header.Set("Authorization", "Basic "+encoded)

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result types.PaymentIntentResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}
