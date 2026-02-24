package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"gabe565.com/linx-server/internal/config"
	"gabe565.com/linx-server/internal/template"
)

// AdminAuthPage serves the HTML for the admin authentication page.
func AdminAuthPage(w http.ResponseWriter, r *http.Request) {
	ServeAsset(w, r, http.StatusOK, template.WithTitle("Admin Login"))
}

// AdminAuthAPI handles POST requests for admin authentication.
func AdminAuthAPI(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Compare with the stored hash
	if req.Password == config.Default.Auth.AdminPasswordHash {
		// Set a simple cookie for authentication.
		// In a real app, you'd use a more secure session management.
		cookie := &http.Cookie{
			Name:     "admin_auth",
			Value:    "authenticated",
			Path:     "/",
			HttpOnly: true,
			Secure:   r.URL.Scheme == "https",
			SameSite: http.SameSiteLaxMode,
			MaxAge:   0, // Session cookie, deleted on browser close
		}
		http.SetCookie(w, cookie)

		if strings.EqualFold("application/json", r.Header.Get("Accept")) || r.Header.Get("Content-Type") == "application/json" {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"message": "Login successful"}`))
			return
		}

		http.Redirect(w, r, "/", http.StatusFound) // Redirect to root
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	_, _ = w.Write([]byte(`{"error": "Invalid password"}`))
}

// AdminAuthMiddleware checks for admin authentication cookie.
func AdminAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("admin_auth")
		if err != nil || cookie.Value != "authenticated" {
			// If not authenticated, redirect to the admin login page
			http.Redirect(w, r, "/admin", http.StatusFound)
			return
		}
		next.ServeHTTP(w, r)
	})
}