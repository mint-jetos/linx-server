package handlers

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"path"
	"strconv"
	"strings"
	"time"

	"gabe565.com/linx-server/internal/config"
	"gabe565.com/linx-server/internal/headers"
)

type AltchaChallenge struct {
	Algorithm string `json:"algorithm"`
	Challenge string `json:"challenge"`
	Salt      string `json:"salt"`
	Signature string `json:"signature"`
	MaxNumber int    `json:"maxnumber"`
}

func GetAltchaChallenge(w http.ResponseWriter, r *http.Request) {
	if !config.Default.Altcha.Enabled {
		http.Error(w, "ALTCHA disabled", http.StatusForbidden)
		return
	}

	salt, err := generateSalt()
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Append expiry to salt
	expires := time.Now().Add(config.Default.Altcha.Expires.Duration).Unix()
	salt = fmt.Sprintf("%s?expires=%d", salt, expires)

	number, err := generateRandomNumber(config.Default.Altcha.MaxNumber)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	challenge := sha256.Sum256([]byte(salt + strconv.Itoa(number)))
	challengeHex := hex.EncodeToString(challenge[:])

	signature := hmacSha256(config.Default.Altcha.HMACKey, salt+challengeHex)

	resp := AltchaChallenge{
		Algorithm: "SHA-256",
		Challenge: challengeHex,
		Salt:      salt,
		Signature: signature,
		MaxNumber: config.Default.Altcha.MaxNumber,
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

func VerifyAltcha(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Payload string `json:"payload"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if verifyAltchaPayload(req.Payload) {
		siteURL := headers.GetSiteURL(r)
		isSecure := siteURL.Scheme == "https" || r.Header.Get("X-Forwarded-Proto") == "https"

		cookie := &http.Cookie{
			Name:     "altcha_verified",
			Value:    "true",
			Path:     "/",
			HttpOnly: true,
			Secure:   isSecure,
			SameSite: http.SameSiteLaxMode,
			MaxAge:   86400, // 1 day
			Expires:  time.Now().Add(24 * time.Hour),
		}
		http.SetCookie(w, cookie)
		slog.Info("ALTCHA verified, setting cookie", "secure", isSecure, "host", r.Host)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("Verified"))
	} else {
		http.Error(w, "Verification failed", http.StatusUnauthorized)
	}
}

func verifyAltchaPayload(payloadB64 string) bool {
	payloadJson, err := base64.StdEncoding.DecodeString(payloadB64)
	if err != nil {
		return false
	}

	var payload struct {
		Algorithm string `json:"algorithm"`
		Challenge string `json:"challenge"`
		Number    int    `json:"number"`
		Salt      string `json:"salt"`
		Signature string `json:"signature"`
	}
	if err := json.Unmarshal(payloadJson, &payload); err != nil {
		return false
	}

	// 1. Check signature
	expectedSignature := hmacSha256(config.Default.Altcha.HMACKey, payload.Salt+payload.Challenge)
	if payload.Signature != expectedSignature {
		return false
	}

	// 2. Check challenge
	hash := sha256.Sum256([]byte(payload.Salt + strconv.Itoa(payload.Number)))
	expectedChallenge := hex.EncodeToString(hash[:])
	if payload.Challenge != expectedChallenge {
		return false
	}

	// 3. Check expiry
	parts := strings.Split(payload.Salt, "?expires=")
	if len(parts) > 1 {
		expires, err := strconv.ParseInt(parts[1], 10, 64)
		if err == nil && time.Now().Unix() > expires {
			return false
		}
	}

	return true
}

// AltchaMiddleware checks for ALTCHA verification
func AltchaMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !config.Default.Altcha.Enabled {
			next.ServeHTTP(w, r)
			return
		}

		cleanPath := path.Clean(r.URL.Path)
		ext := path.Ext(cleanPath)

		// Bypass ALTCHA for API, scripts, and static assets
		if strings.HasPrefix(cleanPath, "/api/altcha") ||
			cleanPath == "/altcha.min.js" ||
			cleanPath == "/altcha_gatekeeper.js" ||
			cleanPath == "/favicon.ico" ||
			cleanPath == "/robots.txt" ||
			cleanPath == "/api/config" ||
			ext == ".js" || ext == ".css" || ext == ".png" || ext == ".jpg" || ext == ".webp" || ext == ".svg" || ext == ".map" ||
			strings.HasPrefix(r.URL.Path, "/api/altcha") ||
			r.URL.Path == "/altcha.min.js" ||
			r.URL.Path == "/altcha_gatekeeper.js" {
			next.ServeHTTP(w, r)
			return
		}

		cookie, err := r.Cookie("altcha_verified")
		if err == nil && cookie.Value == "true" {
			next.ServeHTTP(w, r)
			return
		}

		// Log if cookie is missing for debugging
		if !strings.HasPrefix(r.URL.Path, "/api/") {
			slog.Info("ALTCHA cookie check failed", "path", r.URL.Path, "err", err)
		}

		// If it's a GET request for a page, show the gatekeeper
		if r.Method == http.MethodGet && !strings.HasPrefix(r.URL.Path, "/api/") {
			ServeAltchaGatekeeper(w, r)
			return
		}

		// For other requests (POST upload, etc.), if not verified, return error
		http.Error(w, "ALTCHA verification required", http.StatusForbidden)
	})
}

func ServeAltchaGatekeeper(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
	w.Header().Set("Pragma", "no-cache")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`<!DOCTYPE html>
<html>
<head>
    <title>Verifying Visitor...</title>
    <script src="/altcha.min.js" type="module"></script>
    <script src="/altcha_gatekeeper.js" defer></script>
    <style>
        body { font-family: sans-serif; display: flex; flex-direction: column; align-items: center; justify-content: center; height: 100vh; margin: 0; background: #1a1a1a; color: #e0e0e0; }
        .loading { margin-bottom: 20px; font-size: 1.2rem; }
    </style>
</head>
<body>
    <div class="loading">One moment, verifying your connection...</div>
    
    <altcha-widget 
        id="gatekeeper"
        challengeurl="/api/altcha/challenge" 
        auto="onload" 
        hide-logo 
        hide-footer>
    </altcha-widget>
</body>
</html>`))
}

func generateSalt() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func generateRandomNumber(max int) (int, error) {
	b := make([]byte, 4)
	if _, err := rand.Read(b); err != nil {
		return 0, err
	}
	// Simple conversion to int within range
	num := int(uint32(b[0])<<24 | uint32(b[1])<<16 | uint32(b[2])<<8 | uint32(b[3]))
	if num < 0 {
		num = -num
	}
	return num % max, nil
}

func hmacSha256(key, data string) string {
	h := hmac.New(sha256.New, []byte(key))
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}
