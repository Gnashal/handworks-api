package config

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"handworks-api/types"
	"io"
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

func (c *PaymongoClient) CreateQRPHCode(
	ctx context.Context,
	payload any,
) (*types.QRPHCodeResponse, error) {
	url := c.BaseURL + "/qrph/generate"

	jsonBody, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	encoded := base64.StdEncoding.EncodeToString([]byte(c.SecretKey + ":"))
	req.Header.Set("Authorization", "Basic "+encoded)

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("paymongo create qrph code failed: status=%d body=%s", resp.StatusCode, string(body))
	}

	var result types.QRPHCodeResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}
