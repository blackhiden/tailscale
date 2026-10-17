// Copyright (c) Tailscale Inc & contributors
// SPDX-License-Identifier: BSD-3-Clause

package cli

import (
	"encoding/json"
	"errors"
	"net/url"
	"strings"
)

// registrationQR represents the JSON payload for headscale registration QR codes.
type registrationQR struct {
	Type      string `json:"type"`
	Version   string `json:"version"`
	AuthID    string `json:"auth_id"`
	ServerURL string `json:"server_url"`
}

// parseAuthURL extracts server URL and auth ID from authentication URL.
// It handles various formats:
// - Tailscale: https://login.tailscale.com/a/0123456789
// - Headscale: https://headscale.example.com/register/hskey-authreq-abc123
// - Custom: various formats with /auth/, /a/, /register/ paths
func parseAuthURL(authURL string) (serverURL, authID string, err error) {
	if authURL == "" {
		return "", "", errors.New("empty auth URL")
	}

	u, err := url.Parse(authURL)
	if err != nil {
		return "", "", err
	}

	if u.Scheme == "" || u.Host == "" {
		return "", "", errors.New("invalid URL format")
	}

	// Server URL is scheme + host
	serverURL = u.Scheme + "://" + u.Host

	// Extract auth ID from path
	// Common patterns:
	// /a/<id>
	// /auth/<id>
	// /register/<id>
	// /register/hskey-authreq-<id>
	path := strings.TrimPrefix(u.Path, "/")
	parts := strings.Split(path, "/")
	
	if len(parts) < 2 {
		return "", "", errors.New("auth URL path too short")
	}

	// Auth ID is typically the last component after /a/, /auth/, or /register/
	authID = parts[len(parts)-1]
	if authID == "" {
		return "", "", errors.New("empty auth ID")
	}

	return serverURL, authID, nil
}

// buildRegistrationQR creates JSON payload for QR code.
// Returns JSON string ready for QR encoding.
func buildRegistrationQR(authURL string) (string, error) {
	serverURL, authID, err := parseAuthURL(authURL)
	if err != nil {
		return "", err
	}

	qr := registrationQR{
		Type:      "headscale-registration",
		Version:   "1",
		AuthID:    authID,
		ServerURL: serverURL,
	}

	jsonBytes, err := json.Marshal(qr)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

// isHeadscaleOrCustomServer detects non-Tailscale control servers.
// Returns true for headscale and custom servers, false for Tailscale servers.
func isHeadscaleOrCustomServer(authURL string) bool {
	// Check if URL contains login.tailscale.com
	return !strings.Contains(authURL, "login.tailscale.com")
}
