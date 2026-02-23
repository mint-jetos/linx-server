package handlers

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"


	"gabe565.com/linx-server/internal/config"
)

// AdminAuthPage serves the HTML for the admin authentication page.
func AdminAuthPage(w http.ResponseWriter, r *http.Request) {
	// This will be replaced by the Vue frontend later
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Admin Login</title>
</head>
<body>
    <h1>Admin Login</h1>
    <form id="admin-login-form">
        <input type="password" id="password" placeholder="Password" required>
        <button type="submit">Login</button>
    </form>
    <div id="message"></div>

    <script>
        document.getElementById('admin-login-form').addEventListener('submit', async (e) => {
            e.preventDefault();
            const password = document.getElementById('password').value;
            const messageDiv = document.getElementById('message');

            try {
                const response = await fetch('/api/admin/auth', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json',
                    },
                    body: JSON.stringify({ password: password }),
                });

                if (response.ok) {
                    messageDiv.textContent = 'Login successful! Redirecting...';
                    window.location.href = '/'; // Redirect to the root page
                } else {
                    const errorData = await response.json();
                    messageDiv.textContent = 'Login failed: ' + errorData.error;
                }
            } catch (error) {
                messageDiv.textContent = 'An error occurred: ' + error.message;
            }
        });
    </script>
</body>
</html>`))
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

	// Hash the submitted password
	hasher := sha256.New()
	hasher.Write([]byte(req.Password))
	submittedHash := hex.EncodeToString(hasher.Sum(nil))

	// Compare with the stored hash
	if submittedHash == config.Default.Auth.AdminPasswordHash {
		// Set a simple cookie for authentication.
		// In a real app, you'd use a more secure session management.
		cookie := &http.Cookie{
			Name:     "admin_auth",
			Value:    "authenticated",
			Path:     "/",
			HttpOnly: true,
			Secure:   true, // Should be true in production with HTTPS
			SameSite: http.SameSiteLaxMode,
		}
		http.SetCookie(w, cookie)

		http.Redirect(w, r, "/", http.StatusFound) // Redirect to root
		return
	}

	http.Error(w, `{"error": "Invalid password"}`, http.StatusUnauthorized)
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