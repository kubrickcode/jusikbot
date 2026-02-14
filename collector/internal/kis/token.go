package kis

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

const (
	renewBeforeExpiry = 30 * time.Minute
	tokenPath         = "/oauth2/tokenP"
)

// tokenResponse represents the KIS OAuth2 token endpoint response.
type tokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int64  `json:"expires_in"`
	TokenType   string `json:"token_type"`
}

// TokenProvider manages KIS OAuth2 tokens with lazy init and pre-expiry renewal.
// Why net/http.Client instead of httpclient.Client: token endpoint requires POST,
// but httpclient.Client only supports GET. Acceptable since token responses are small
// and token issuance is infrequent (~once per 24h).
// Why sync.Mutex over sync.RWMutex: token reads always check expiry, which may trigger
// a renewal write. RWMutex adds complexity without benefit for this access pattern.
type TokenProvider struct {
	appKey     string
	appSecret  string
	baseURL    string
	expiresAt  time.Time
	httpClient *http.Client
	mu         sync.Mutex
	token      string
}

func NewTokenProvider(baseURL, appKey, appSecret string, httpClient *http.Client) *TokenProvider {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 10 * time.Second}
	}
	return &TokenProvider{
		appKey:     appKey,
		appSecret:  appSecret,
		baseURL:    baseURL,
		httpClient: httpClient,
	}
}

// Token returns a valid access token, fetching or renewing as needed.
// Thread-safe via sync.Mutex.
func (p *TokenProvider) Token(ctx context.Context) (string, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.isValid() {
		return p.token, nil
	}

	return p.fetchToken(ctx)
}

func (p *TokenProvider) isValid() bool {
	return p.token != "" && time.Now().Before(p.expiresAt.Add(-renewBeforeExpiry))
}

func (p *TokenProvider) fetchToken(ctx context.Context) (string, error) {
	reqBody, err := json.Marshal(map[string]string{
		"appkey":     p.appKey,
		"appsecret":  p.appSecret,
		"grant_type": "client_credentials",
	})
	if err != nil {
		return "", fmt.Errorf("marshal token request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.baseURL+tokenPath, bytes.NewReader(reqBody))
	if err != nil {
		return "", fmt.Errorf("create token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("token request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read token response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("token request failed (HTTP %d): %s", resp.StatusCode, string(body))
	}

	var tok tokenResponse
	if err := json.Unmarshal(body, &tok); err != nil {
		return "", fmt.Errorf("parse token response: %w", err)
	}

	if tok.AccessToken == "" {
		return "", fmt.Errorf("empty access token in response")
	}

	p.token = tok.AccessToken
	p.expiresAt = time.Now().Add(time.Duration(tok.ExpiresIn) * time.Second)

	return p.token, nil
}
