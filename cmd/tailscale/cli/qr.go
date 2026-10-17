// Copyright (c) Tailscale Inc & contributors
// SPDX-License-Identifier: BSD-3-Clause

package cli

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// buildRegistrationQR attempts to fetch JSON payload from headscale API.
// If the API endpoint is not available (404 or similar), returns empty string.
// Returns JSON string ready for QR encoding, or empty string if API not available.
func buildRegistrationQR(authURL string) (string, error) {
	serverURL, authID, err := parseAuthURL(authURL)
	if err != nil {
		return "", err
	}

	// Build the headscale API endpoint URL
	// Format: <HEADSCALE_URL>/api/v1/registration/qr/<AUTH_ID>
	// Note: This is a custom endpoint, not all servers implement it
	apiURL := fmt.Sprintf("%s/api/v1/registration/qr/%s", serverURL, authID)

	// Create HTTP client with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Make the API call
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to call headscale API: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	// If 404 or similar, the server doesn't implement this endpoint
	// Return empty string to signal "no QR support"
	if resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusMethodNotAllowed {
		return "", nil // Empty string = no QR support, just show URL
	}
	
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("headscale API returned status %d", resp.StatusCode)
	}

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	// Validate it's valid JSON
	var js json.RawMessage
	if err := json.Unmarshal(body, &js); err != nil {
		return "", fmt.Errorf("invalid JSON response: %w", err)
	}

	// Return the JSON string from headscale
	return string(body), nil
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

// isHeadscaleOrCustomServer detects non-Tailscale control servers.
// Returns true for headscale and custom servers, false for Tailscale servers.
func isHeadscaleOrCustomServer(authURL string) bool {
	// Check if URL contains login.tailscale.com
	return !strings.Contains(authURL, "login.tailscale.com")
}

// isRegistrationURL detects if the URL is a device registration URL (not OAuth/OIDC).
// Returns true for headscale/custom registration URLs that should use JSON QR format.
// Returns false for OAuth/OIDC authentication URLs that should use URL QR format.
func isRegistrationURL(authURL string) bool {
	u, err := url.Parse(authURL)
	if err != nil {
		return false
	}
	
	path := strings.ToLower(u.Path)
	
	// Headscale registration URLs typically have /register/ in the path
	// Examples:
	// - https://headscale.example.com/register/hskey-authreq-abc123
	// - https://custom.server.net/register/nodekey-xyz
	if strings.Contains(path, "/register") {
		return true
	}
	
	// OAuth/OIDC URLs typically have /oauth/, /oidc/, /auth/, or /authorize/
	// These should use URL QR (browser-based authentication)
	if strings.Contains(path, "/oauth") || 
	   strings.Contains(path, "/oidc") ||
	   strings.Contains(path, "/authorize") {
		return false
	}
	
	// Tailscale's /a/ path is for interactive login (OAuth)
	if strings.Contains(path, "/a/") {
		return false
	}
	
	// Default: if we can't determine, fall back to URL QR for safety
	// (OAuth flows will still work with URL QR)
	return false
}
