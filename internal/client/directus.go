package client

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

type Directus struct {
	baseURL string
	http    *http.Client
	token   string
}

func NewDirectus(baseURL, token string, timeout time.Duration) *Directus {
	return &Directus{
		baseURL: baseURL,
		http:    &http.Client{Timeout: timeout},
		token:   token,
	}
}

func (c *Directus) Request(ctx context.Context, method, path string, body any) (*http.Response, error) {
	var reqBody *strings.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reqBody = strings.NewReader(string(b))
	} else {
		reqBody = strings.NewReader("")
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, reqBody)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
	return c.http.Do(req)
}
