// internal/handlers/handlers.go
package handlers

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/aankitroy/oauth-sample/backend/internal/auth"
	"github.com/aankitroy/oauth-sample/backend/internal/rbac"
	"github.com/aankitroy/oauth-sample/backend/internal/session"
)

type Server struct {
	OIDCConfig *auth.OIDCConfig
	SessionMgr *session.Manager
	RBACStore  *rbac.RBACStore
	// Possibly more config (logout URL, etc.)
}

// Example: Token Exchange
func (s *Server) TokenExchangeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var payload struct {
		Code         string `json:"code"`
		CodeVerifier string `json:"codeVerifier"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	tr, err := auth.ExchangeCodeForTokens(s.OIDCConfig, payload.Code, payload.CodeVerifier)
	if err != nil {
		http.Error(w, "Token exchange failed", http.StatusInternalServerError)
		return
	}

	// Fetch user details from the user info endpoint
	req, err := http.NewRequest("GET", s.OIDCConfig.UserInfoURL, nil)
	if err != nil {
		http.Error(w, "Failed to create request", http.StatusInternalServerError)
		return
	}
	req.Header.Set("Authorization", "Bearer "+tr.AccessToken)
	resp, err := http.DefaultClient.Do(req)
	fmt.Println("resp: ", resp)
	if err != nil {
		http.Error(w, "Failed to fetch user info", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		http.Error(w, "Failed to fetch user info", http.StatusUnauthorized)
		return
	}

	var userInfo struct {
		Email  string `json:"email"`
		UserID string `json:"sub"` // Assuming 'sub' is the user ID field
	}
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		http.Error(w, "Failed to decode user info", http.StatusInternalServerError)
		return
	}

	email := userInfo.Email
	//_ := userInfo.UserID
	role, _ := s.RBACStore.GetUserRole(email)

	sessionID := randomString(32)
	now := time.Now().Unix()

	// Access token expiry in epoch seconds
	expiresAt := now + int64(tr.ExpiresIn)

	sessData := &session.SessionData{
		AccessToken:  tr.AccessToken,
		RefreshToken: tr.RefreshToken,
		Role:         role,
		ExpiresAt:    expiresAt,
		LastActivity: now,
	}

	// TTL for the entire session in Redis. You might set a max-lifetime
	// (like 30 minutes) or we can rely only on inactivity logic. For example:
	if err := s.SessionMgr.CreateSession(r.Context(), sessionID, sessData, 30*time.Minute); err != nil {
		http.Error(w, "Failed to create session", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    sessionID,
		Path:     "/",
		HttpOnly: true,
		// Secure: true in production
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// Protected route
func (s *Server) ProtectedHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_id")
	if err != nil {
		http.Error(w, "Unauthorized (no cookie)", http.StatusUnauthorized)
		return
	}
	sessionID := cookie.Value

	sessData, err := s.SessionMgr.GetSession(r.Context(), sessionID)
	if err != nil {
		http.Error(w, "Invalid session", http.StatusUnauthorized)
		return
	}

	// Check inactivity
	now := time.Now().Unix()
	if s.SessionMgr.CheckInactivityTimeout(sessData, now, 60) { // 60s for demo
		// Session expired
		s.SessionMgr.DeleteSession(r.Context(), sessionID)
		http.Error(w, "Session timed out", http.StatusUnauthorized)
		return
	}

	// Rolling session: update lastActivity
	s.SessionMgr.UpdateLastActivity(r.Context(), sessionID, now, 30*time.Minute)

	// (Optional) check access token expiry and refresh if needed
	// if now >= sessData.ExpiresAt { refresh logic... }

	// RBAC example: only admin can see this endpoint
	if sessData.Role != "admin" {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Protected content for admin"})
}

// Logout
func (s *Server) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_id")
	if err == nil {
		s.SessionMgr.DeleteSession(r.Context(), cookie.Value)
	}
	// Optionally call the IdP's end_session_endpoint here

	// Expire the cookie in the browser
	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    "",
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
		HttpOnly: true,
		Path:     "/",
		// Secure: true in production
	})

	w.WriteHeader(http.StatusOK)
}

func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	for i := 0; i < n; i++ {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
