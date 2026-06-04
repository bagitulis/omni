package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// dev-login body: only tenant_id required

type devLoginResponse struct {
	Success     bool   `json:"success"`
	AccessToken string `json:"access_token"`
	Token       string `json:"token"` // fallback
	Message     string `json:"message"`
}

// login performs dev-login to obtain a JWT token
func (c *OmniClient) login() error {
	loginBody, _ := json.Marshal(map[string]string{
		"tenant_id": c.defaultTenant,
	})
	req, err := http.NewRequest("POST", c.baseURL+"/api/auth/dev-login", bytes.NewReader(loginBody))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Origin", "http://localhost") // REQUIRED: passes isDevModeAllowed()

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("login request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read login response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("login HTTP %d: %s", resp.StatusCode, string(respBody))
	}

	var result devLoginResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return fmt.Errorf("parse login response: %w", err)
	}

	token := result.AccessToken
	if token == "" {
		token = result.Token
	}
	if !result.Success || token == "" {
		msg := result.Message
		if msg == "" {
			msg = "no token returned"
		}
		return fmt.Errorf("login failed: %s", msg)
	}

	c.token = token
	return nil
}
