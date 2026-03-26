package handlers

import (
	"context"
	"encoding/json"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"time"

	"gabe565.com/linx-server/internal/config"
	"gabe565.com/linx-server/internal/headers"
)

type contextKey int

const (
	ContextKeyTurnstileVerified contextKey = iota
)

type turnstileResponse struct {
	Success     bool      `json:"success"`
	ChallengeTS time.Time `json:"challenge_ts"`
	Hostname    string    `json:"hostname"`
	ErrorCodes  []string  `json:"error-codes"`
	Action      string    `json:"action"`
	CData       string    `json:"cdata"`
}

func VerifyTurnstile(token string, remoteAddr string) (bool, error) {
	if !config.Default.Turnstile.Enabled {
		return true, nil
	}

	if token == "" {
		slog.Warn("Turnstile token is empty")
		return false, nil
	}

	data := url.Values{}
	data.Set("secret", config.Default.Turnstile.SecretKey)
	data.Set("response", token)

	if host, _, err := net.SplitHostPort(remoteAddr); err == nil {
		data.Set("remoteip", host)
	} else {
		data.Set("remoteip", remoteAddr)
	}

	resp, err := http.PostForm("https://challenges.cloudflare.com/turnstile/v0/siteverify", data)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	var result turnstileResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return false, err
	}

	if !result.Success {
		slog.Warn("Turnstile verification failed", "error_codes", result.ErrorCodes, "hostname", result.Hostname)
	}

	return result.Success, nil
}

func TurnstileMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !config.Default.Turnstile.Enabled {
			next.ServeHTTP(w, r)
			return
		}

		// Skip critical paths
		path := r.URL.Path
		if path == "/ping" || path == "/favicon.ico" || path == "/api/config" {
			next.ServeHTTP(w, r)
			return
		}

		// Skip authenticated admins
		if cookie, err := r.Cookie("admin_auth"); err == nil && cookie.Value == "authenticated" {
			next.ServeHTTP(w, r)
			return
		}

		// Skip if Turnstile verified cookie exists
		if cookie, err := r.Cookie("turnstile_verified"); err == nil && cookie.Value == "true" {
			ctx := context.WithValue(r.Context(), ContextKeyTurnstileVerified, true)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		token := r.Header.Get("X-Turnstile-Token")
		if token == "" {
			token = r.URL.Query().Get("cf-turnstile-response")
		}

		if token != "" {
			if success, _ := VerifyTurnstile(token, r.RemoteAddr); success {
				// Set cookie for future requests
				cookie := &http.Cookie{
					Name:     "turnstile_verified",
					Value:    "true",
					Path:     "/",
					HttpOnly: true,
					Secure:   headers.GetSiteURL(r).Scheme == "https",
					SameSite: http.SameSiteLaxMode,
					MaxAge:   86400 * 30, // 30 days
				}
				http.SetCookie(w, cookie)

				ctx := context.WithValue(r.Context(), ContextKeyTurnstileVerified, true)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}
		}

		ClockHandler(w, r)
	})
}
