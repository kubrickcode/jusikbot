package kis

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func newTokenServer(t *testing.T, handler http.HandlerFunc) *httptest.Server {
	t.Helper()
	return httptest.NewServer(handler)
}

func validTokenHandler(token string, expiresIn int64) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		if r.URL.Path != tokenPath {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		var reqBody map[string]string
		if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if reqBody["grant_type"] != "client_credentials" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(tokenResponse{
			AccessToken: token,
			ExpiresIn:   expiresIn,
			TokenType:   "Bearer",
		})
	}
}

func TestTokenProvider(t *testing.T) {
	t.Run("fetches token on first call", func(t *testing.T) {
		srv := newTokenServer(t, validTokenHandler("test-token-123", 86400))
		defer srv.Close()

		provider := NewTokenProvider(srv.URL, "app-key", "app-secret", srv.Client())

		token, err := provider.Token(context.Background())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if token != "test-token-123" {
			t.Errorf("token = %q, want test-token-123", token)
		}
	})

	t.Run("caches token on subsequent calls", func(t *testing.T) {
		var callCount atomic.Int32
		srv := newTokenServer(t, func(w http.ResponseWriter, r *http.Request) {
			callCount.Add(1)
			validTokenHandler("cached-token", 86400)(w, r)
		})
		defer srv.Close()

		provider := NewTokenProvider(srv.URL, "app-key", "app-secret", srv.Client())

		token1, err := provider.Token(context.Background())
		if err != nil {
			t.Fatalf("first call error: %v", err)
		}

		token2, err := provider.Token(context.Background())
		if err != nil {
			t.Fatalf("second call error: %v", err)
		}

		if token1 != token2 {
			t.Errorf("tokens differ: %q vs %q", token1, token2)
		}
		if callCount.Load() != 1 {
			t.Errorf("server called %d times, want 1 (cached)", callCount.Load())
		}
	})

	t.Run("renews token near expiry", func(t *testing.T) {
		var callCount atomic.Int32
		srv := newTokenServer(t, func(w http.ResponseWriter, r *http.Request) {
			count := callCount.Add(1)
			token := "token-v1"
			if count > 1 {
				token = "token-v2"
			}
			// Why 1 second: forces immediate expiry for test purposes.
			validTokenHandler(token, 1)(w, r)
		})
		defer srv.Close()

		provider := NewTokenProvider(srv.URL, "app-key", "app-secret", srv.Client())

		token1, err := provider.Token(context.Background())
		if err != nil {
			t.Fatalf("first call error: %v", err)
		}
		if token1 != "token-v1" {
			t.Errorf("first token = %q, want token-v1", token1)
		}

		// Why sleep not needed: expiresIn=1s with 30min renewBeforeExpiry means
		// the token is already considered expired on the next call.
		token2, err := provider.Token(context.Background())
		if err != nil {
			t.Fatalf("second call error: %v", err)
		}
		if token2 != "token-v2" {
			t.Errorf("renewed token = %q, want token-v2", token2)
		}
		if callCount.Load() != 2 {
			t.Errorf("server called %d times, want 2 (initial + renewal)", callCount.Load())
		}
	})

	t.Run("sends correct credentials in request body", func(t *testing.T) {
		srv := newTokenServer(t, func(w http.ResponseWriter, r *http.Request) {
			var reqBody map[string]string
			if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
				t.Errorf("decode request body: %v", err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			if reqBody["appkey"] != "my-app-key" {
				t.Errorf("appkey = %q, want my-app-key", reqBody["appkey"])
			}
			if reqBody["appsecret"] != "my-app-secret" {
				t.Errorf("appsecret = %q, want my-app-secret", reqBody["appsecret"])
			}
			if reqBody["grant_type"] != "client_credentials" {
				t.Errorf("grant_type = %q, want client_credentials", reqBody["grant_type"])
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(tokenResponse{
				AccessToken: "ok",
				ExpiresIn:   86400,
				TokenType:   "Bearer",
			})
		})
		defer srv.Close()

		provider := NewTokenProvider(srv.URL, "my-app-key", "my-app-secret", srv.Client())
		if _, err := provider.Token(context.Background()); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("returns error on HTTP failure", func(t *testing.T) {
		srv := newTokenServer(t, func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("internal server error"))
		})
		defer srv.Close()

		provider := NewTokenProvider(srv.URL, "key", "secret", srv.Client())

		_, err := provider.Token(context.Background())
		if err == nil {
			t.Fatal("expected error for HTTP 500")
		}
	})

	t.Run("returns error on empty token", func(t *testing.T) {
		srv := newTokenServer(t, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(tokenResponse{
				AccessToken: "",
				ExpiresIn:   86400,
				TokenType:   "Bearer",
			})
		})
		defer srv.Close()

		provider := NewTokenProvider(srv.URL, "key", "secret", srv.Client())

		_, err := provider.Token(context.Background())
		if err == nil {
			t.Fatal("expected error for empty token")
		}
	})

	t.Run("respects context cancellation", func(t *testing.T) {
		srv := newTokenServer(t, func(w http.ResponseWriter, r *http.Request) {
			select {
			case <-r.Context().Done():
				return
			case <-time.After(500 * time.Millisecond):
				validTokenHandler("late-token", 86400)(w, r)
			}
		})
		defer srv.Close()

		provider := NewTokenProvider(srv.URL, "key", "secret", &http.Client{Timeout: 100 * time.Millisecond})

		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
		defer cancel()

		_, err := provider.Token(ctx)
		if err == nil {
			t.Fatal("expected error for cancelled context")
		}
	})
}
