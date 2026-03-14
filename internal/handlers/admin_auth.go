package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"gabe565.com/linx-server/internal/config"
	"gabe565.com/linx-server/internal/headers"
	"gabe565.com/linx-server/internal/template"
	"gabe565.com/linx-server/internal/util"
)

// AdminAuthPage serves the HTML for the admin authentication page.
func AdminAuthPage(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Vary", "Cookie")
	cookie, err := r.Cookie("admin_auth")
	if err == nil && cookie.Value == "authenticated" {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	ServeAsset(w, r, http.StatusOK, template.WithTitle("Login"))
}

// AdminAuthAPI handles requests for admin authentication.
func AdminAuthAPI(w http.ResponseWriter, r *http.Request) {
	var password string

	if r.Method == http.MethodPost {
		var req struct {
			Password string `json:"password"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err == nil {
			password = req.Password
		}
	} else if r.Method == http.MethodGet {
		// Try to deobfuscate from query parameter 'd'
		if d := r.URL.Query().Get("d"); d != "" {
			if decoded, err := util.Deobfuscate(d); err == nil {
				password = string(decoded)
			}
		}
		// Fallback to plain 'p' for backward compatibility or if 'd' fails
		if password == "" {
			password = r.URL.Query().Get("p")
		}
	}

	if password == "" {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Compare with the stored hash
	if password == config.Default.Auth.AdminPasswordHash {
		// Set a simple cookie for authentication.
		// In a real app, you'd use a more secure session management.
		cookie := &http.Cookie{
			Name:     "admin_auth",
			Value:    "authenticated",
			Path:     "/",
			HttpOnly: true,
			Secure:   headers.GetSiteURL(r).Scheme == "https",
			SameSite: http.SameSiteLaxMode,
			MaxAge:   int(config.Default.Auth.CookieExpiry.Seconds()),
		}
		http.SetCookie(w, cookie)

		w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate")

		if strings.EqualFold("application/json", r.Header.Get("Accept")) ||
			r.Header.Get("Content-Type") == "application/json" ||
			r.URL.Query().Get("json") == "true" {
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			_, _ = w.Write([]byte(util.Obfuscate([]byte(`{"message": "Login successful"}`))))
			return
		}

		http.Redirect(w, r, "/", http.StatusFound) // Redirect to root
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusUnauthorized)
	_, _ = w.Write([]byte(util.Obfuscate([]byte(`{"error": "Invalid password"}`))))
}

// AdminAuthMiddleware checks for admin authentication cookie.
func AdminAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Vary", "Cookie")
		cookie, err := r.Cookie("admin_auth")
		if err != nil || cookie.Value != "authenticated" {
			// If not authenticated, redirect to the login page
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}
		next.ServeHTTP(w, r)
	})
}