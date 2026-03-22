package api

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/umarbektokyo/matetra-engine/utils"
)

var jwtSecret = func() []byte {
	if s := os.Getenv("JWT_SECRET"); s != "" {
		return []byte(s)
	}
	return []byte("matetra-dev-secret-change-in-prod")
}()

type UserStore struct {
	mu    sync.RWMutex
	users map[string]string // name -> hash
}

func NewUserStore() *UserStore {
	return &UserStore{users: make(map[string]string)}
}

func (us *UserStore) Register(name, password string) (string, error) {
	us.mu.Lock()
	defer us.mu.Unlock()

	hash := utils.Hash(password)

	if existing, ok := us.users[name]; ok {
		// Allow re-login with same password
		if existing == hash {
			return generateJWT(name), nil
		}
		return "", fmt.Errorf("username '%s' already taken", name)
	}

	us.users[name] = hash
	return generateJWT(name), nil
}

func (us *UserStore) GetHash(name string) (string, bool) {
	us.mu.RLock()
	defer us.mu.RUnlock()
	h, ok := us.users[name]
	return h, ok
}

type jwtHeader struct {
	Alg string `json:"alg"`
	Typ string `json:"typ"`
}

type jwtPayload struct {
	Sub string `json:"sub"`
	Iat int64  `json:"iat"`
	Exp int64  `json:"exp"`
}

func generateJWT(username string) string {
	header := jwtHeader{Alg: "HS256", Typ: "JWT"}
	payload := jwtPayload{
		Sub: username,
		Iat: time.Now().Unix(),
		Exp: time.Now().Add(24 * time.Hour).Unix(),
	}

	headerBytes, _ := json.Marshal(header)
	payloadBytes, _ := json.Marshal(payload)

	headerB64 := base64.RawURLEncoding.EncodeToString(headerBytes)
	payloadB64 := base64.RawURLEncoding.EncodeToString(payloadBytes)

	sigInput := headerB64 + "." + payloadB64
	mac := hmac.New(sha256.New, jwtSecret)
	mac.Write([]byte(sigInput))
	sig := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))

	return headerB64 + "." + payloadB64 + "." + sig
}

func validateJWT(token string) (string, error) {
	parts := splitToken(token)
	if len(parts) != 3 {
		return "", fmt.Errorf("invalid token format")
	}

	// Verify signature
	sigInput := parts[0] + "." + parts[1]
	mac := hmac.New(sha256.New, jwtSecret)
	mac.Write([]byte(sigInput))
	expectedSig := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))

	if !hmac.Equal([]byte(parts[2]), []byte(expectedSig)) {
		return "", fmt.Errorf("invalid token signature")
	}

	// Decode payload
	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return "", fmt.Errorf("invalid token payload")
	}

	var payload jwtPayload
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		return "", fmt.Errorf("invalid token payload")
	}

	if time.Now().Unix() > payload.Exp {
		return "", fmt.Errorf("token expired")
	}

	return payload.Sub, nil
}

func splitToken(token string) []string {
	var parts []string
	start := 0
	for i := 0; i < len(token); i++ {
		if token[i] == '.' {
			parts = append(parts, token[start:i])
			start = i + 1
		}
	}
	parts = append(parts, token[start:])
	return parts
}
