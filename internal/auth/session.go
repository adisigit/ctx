package auth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type SessionClient struct {
	apiBase string
	http    *http.Client
}

func NewSessionClient(apiBase string) *SessionClient {
	return &SessionClient{
		apiBase: apiBase,
		http:    &http.Client{Timeout: 10 * time.Second},
	}
}

type createSessionResp struct {
	SessionID string `json:"session_id"`
	LoginURL  string `json:"login_url"`
}

func (c *SessionClient) CreateSession(publicKeyB64 string) (sessionID, loginURL string, err error) {
	body, _ := json.Marshal(map[string]string{"public_key": publicKeyB64})
	resp, err := c.http.Post(c.apiBase+"/auth/cli/session", "application/json", bytes.NewReader(body))
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("failed to load session, status %d", resp.StatusCode)
	}
	var out createSessionResp
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", "", fmt.Errorf("failed to decode session response: %w", err)
	}
	return out.SessionID, out.LoginURL, nil
}

type pollResp struct {
	Status         string `json:"status"`
	EncryptedToken string `json:"encrypted_token"`
}

func (c *SessionClient) PollSession(sessionID string, timeout time.Duration) (string, error) {
	deadline := time.Now().Add(timeout)
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	for {
		if time.Now().After(deadline) {
			return "", fmt.Errorf("polling session timed out")
		}
		resp, err := c.http.Get(c.apiBase + "/auth/cli/session/" + sessionID)
		if err != nil {
			<-ticker.C
			continue
		}
		var out pollResp
		decodeErr := json.NewDecoder(resp.Body).Decode(&out)
		resp.Body.Close()
		if decodeErr != nil {
			<-ticker.C
			continue
		}
		if out.Status == "completed" {
			return out.EncryptedToken, nil
		}
		<-ticker.C
	}
}

func VerifyToken(baseUrl, token string) (bool, error) {
	req, err := http.NewRequest("GET", baseUrl+"/api/auth/cli/verify", nil)
	if err != nil {
		return false, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusUnauthorized {
		return false, nil
	}
	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	var out struct {
		Valid bool `json:"valid"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return false, fmt.Errorf("failed to decode verify response: %w", err)
	}
	return out.Valid, nil
}
