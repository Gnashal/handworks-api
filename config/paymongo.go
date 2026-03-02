package config

import (
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

// func (c *PaymongoClient) do(ctx context.Context, method, path string, body any, v any) error {
// 	var buf *bytes.Buffer
// 	if body != nil {
// 		b, err := json.Marshal(body)
// 		if err != nil {
// 			return err
// 		}
// 		buf = bytes.NewBuffer(b)
// 	} else {
// 		buf = &bytes.Buffer{}
// 	}

// 	req, err := http.NewRequestWithContext(ctx, method, c.BaseURL+path, buf)
// 	if err != nil {
// 		return err
// 	}

// 	basic := base64.StdEncoding.EncodeToString([]byte(c.SecretKey + ":"))
// 	req.Header.Set("Authorization", "Basic "+basic)
// 	req.Header.Set("Content-Type", "application/json")
// 	req.Header.Set("Accept", "application/json")

// 	res, err := c.HTTP.Do(req)
// 	if err != nil {
// 		return err
// 	}
// 	defer res.Body.Close()

// 	if res.StatusCode >= 400 {
// 		// TODO: decode error body into a struct for better error messages.
// 		return fmt.Errorf("paymongo error: status %d", res.StatusCode)
// 	}

// 	if v != nil {
// 		return json.NewDecoder(res.Body).Decode(v)
// 	}
// 	return nil
// }