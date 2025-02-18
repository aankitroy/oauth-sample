// internal/session/session.go
package session

import (
	"context"
	"encoding/json"
	"time"

	"github.com/go-redis/redis/v8"
)

type SessionData struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	Role         string `json:"role"`
	ExpiresAt    int64  `json:"expires_at"`
	LastActivity int64  `json:"last_activity"`
}

type Manager struct {
	redisClient *redis.Client
	// Possibly an OIDC config or method to refresh tokens
}

// NewManager is an Fx-provided constructor
func NewManager() *Manager {
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		// Password: "",
		// DB: 0,
	})
	return &Manager{redisClient: rdb}
}

func (m *Manager) CreateSession(ctx context.Context, sessionID string, data *SessionData, ttl time.Duration) error {
	b, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return m.redisClient.Set(ctx, sessionID, b, ttl).Err()
}

func (m *Manager) GetSession(ctx context.Context, sessionID string) (*SessionData, error) {
	val, err := m.redisClient.Get(ctx, sessionID).Result()
	if err != nil {
		return nil, err
	}
	var s SessionData
	if err := json.Unmarshal([]byte(val), &s); err != nil {
		return nil, err
	}
	return &s, nil
}

func (m *Manager) DeleteSession(ctx context.Context, sessionID string) error {
	return m.redisClient.Del(ctx, sessionID).Err()
}

// For "rolling" session, call this each time the user makes a request
func (m *Manager) UpdateLastActivity(ctx context.Context, sessionID string, newLastActivity int64, newTTL time.Duration) error {
	sess, err := m.GetSession(ctx, sessionID)
	if err != nil {
		return err
	}
	sess.LastActivity = newLastActivity

	b, _ := json.Marshal(sess)
	return m.redisClient.Set(ctx, sessionID, b, newTTL).Err()
}

// Check if session is expired due to inactivity
func (m *Manager) CheckInactivityTimeout(sess *SessionData, now int64, inactivityLimitSecs int64) bool {
	return (now - sess.LastActivity) > inactivityLimitSecs
}

// etc.
