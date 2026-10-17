// Copyright (c) Tailscale Inc & contributors
// SPDX-License-Identifier: BSD-3-Clause

package cli

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestParseAuthURL(t *testing.T) {
	tests := []struct {
		name       string
		url        string
		wantServer string
		wantAuthID string
		wantErr    bool
	}{
		// Valid Tailscale URLs
		{
			name:       "Tailscale URL standard",
			url:        "https://login.tailscale.com/a/0123456789abcdef",
			wantServer: "https://login.tailscale.com",
			wantAuthID: "0123456789abcdef",
			wantErr:    false,
		},
		{
			name:       "Tailscale URL with trailing slash",
			url:        "https://login.tailscale.com/a/0123456789abcdef/",
			wantServer: "https://login.tailscale.com",
			wantAuthID: "",
			wantErr:    true, // trailing slash creates empty authID
		},
		
		// Valid Headscale URLs
		{
			name:       "Headscale URL standard",
			url:        "https://headscale.example.com/register/hskey-authreq-abc123",
			wantServer: "https://headscale.example.com",
			wantAuthID: "hskey-authreq-abc123",
			wantErr:    false,
		},
		{
			name:       "Headscale URL with port",
			url:        "https://headscale.example.com:8443/register/hskey-authreq-xyz789",
			wantServer: "https://headscale.example.com:8443",
			wantAuthID: "hskey-authreq-xyz789",
			wantErr:    false,
		},
		{
			name:       "Headscale URL with nested path",
			url:        "https://headscale.example.com/api/v1/register/hskey-authreq-nested",
			wantServer: "https://headscale.example.com",
			wantAuthID: "hskey-authreq-nested",
			wantErr:    false,
		},
		
		// Valid custom server URLs
		{
			name:       "Custom server with /auth/ path",
			url:        "https://custom.server.net/auth/xyz789",
			wantServer: "https://custom.server.net",
			wantAuthID: "xyz789",
			wantErr:    false,
		},
		{
			name:       "Custom server with /a/ path",
			url:        "https://custom.server.net/a/customkey123",
			wantServer: "https://custom.server.net",
			wantAuthID: "customkey123",
			wantErr:    false,
		},
		{
			name:       "HTTP (non-HTTPS) server",
			url:        "http://localhost:8080/register/testkey",
			wantServer: "http://localhost:8080",
			wantAuthID: "testkey",
			wantErr:    false,
		},
		
		// Invalid URLs - malformed
		{
			name:    "Invalid URL - no scheme",
			url:     "not-a-url",
			wantErr: true,
		},
		{
			name:    "Empty URL",
			url:     "",
			wantErr: true,
		},
		{
			name:    "URL with scheme but no host",
			url:     "https:///path/to/auth",
			wantErr: true,
		},
		{
			name:    "URL with only host",
			url:     "https://example.com",
			wantErr: true, // path too short
		},
		{
			name:    "URL with single path component",
			url:     "https://example.com/register",
			wantErr: true, // path too short (needs at least 2 components)
		},
		{
			name:    "URL with only slash",
			url:     "https://example.com/",
			wantErr: true, // path too short
		},
		{
			name:    "Relative URL",
			url:     "/register/abc123",
			wantErr: true, // no scheme or host
		},
		
		// Edge cases
		{
			name:       "URL with query parameters",
			url:        "https://example.com/register/authkey?param=value",
			wantServer: "https://example.com",
			wantAuthID: "authkey", // Query params are stripped by url.Parse
			wantErr:    false,
		},
		{
			name:       "URL with fragment",
			url:        "https://example.com/register/authkey#fragment",
			wantServer: "https://example.com",
			wantAuthID: "authkey",
			wantErr:    false, // fragment is not part of path
		},
		{
			name:       "URL with username in host",
			url:        "https://user@example.com/register/authkey",
			wantServer: "https://example.com", // url.Parse strips user info from Host
			wantAuthID: "authkey",
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server, authID, err := parseAuthURL(tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseAuthURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if server != tt.wantServer {
					t.Errorf("parseAuthURL() server = %v, want %v", server, tt.wantServer)
				}
				if authID != tt.wantAuthID {
					t.Errorf("parseAuthURL() authID = %v, want %v", authID, tt.wantAuthID)
				}
			}
		})
	}
}

func TestBuildRegistrationQR(t *testing.T) {
	tests := []struct {
		name           string
		authURL        string
		mockStatusCode int
		mockResponse   string
		wantErr        bool
		wantEmpty      bool // for 404/405 cases
		validateJSON   bool
	}{
		{
			name:           "Successful API response",
			authURL:        "/register/hskey-authreq-abc123",
			mockStatusCode: http.StatusOK,
			mockResponse:   `{"type":"headscale-registration","version":"1","auth_id":"hskey-authreq-abc123","server_url":"https://headscale.example.com"}`,
			wantErr:        false,
			validateJSON:   true,
		},
		{
			name:           "API endpoint not found (404)",
			authURL:        "/register/hskey-authreq-xyz",
			mockStatusCode: http.StatusNotFound,
			mockResponse:   "",
			wantErr:        false,
			wantEmpty:      true, // Should return empty string
		},
		{
			name:           "API method not allowed (405)",
			authURL:        "/register/hskey-authreq-test",
			mockStatusCode: http.StatusMethodNotAllowed,
			mockResponse:   "",
			wantErr:        false,
			wantEmpty:      true, // Should return empty string
		},
		{
			name:           "API server error (500)",
			authURL:        "/register/hskey-authreq-err",
			mockStatusCode: http.StatusInternalServerError,
			mockResponse:   "",
			wantErr:        true,
		},
		{
			name:           "Invalid JSON response",
			authURL:        "/register/hskey-authreq-bad",
			mockStatusCode: http.StatusOK,
			mockResponse:   "not valid json",
			wantErr:        true,
		},
		{
			name:    "Invalid URL",
			authURL: "not-a-url",
			wantErr: true,
		},
		{
			name:    "Empty URL",
			authURL: "",
			wantErr: true,
		},
		{
			name:    "URL with incomplete path",
			authURL: "/register",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var authURL string
			
			// Only create mock server for valid URL test cases
			if tt.mockStatusCode > 0 {
				// Create mock HTTP server
				mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(tt.mockStatusCode)
					if tt.mockResponse != "" {
						w.Write([]byte(tt.mockResponse))
					}
				}))
				defer mockServer.Close()
				
				// Build authURL using mock server
				authURL = mockServer.URL + tt.authURL
			} else {
				authURL = tt.authURL
			}

			jsonStr, err := buildRegistrationQR(authURL)
			
			if (err != nil) != tt.wantErr {
				t.Errorf("buildRegistrationQR() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			
			if tt.wantEmpty && jsonStr != "" {
				t.Errorf("buildRegistrationQR() expected empty string, got %v", jsonStr)
			}
			
			if tt.validateJSON && jsonStr != "" {
				// Validate it's valid JSON
				var js map[string]interface{}
				if err := json.Unmarshal([]byte(jsonStr), &js); err != nil {
					t.Errorf("buildRegistrationQR() returned invalid JSON: %v", err)
				}
			}
		})
	}
}

func TestIsHeadscaleOrCustomServer(t *testing.T) {
	tests := []struct {
		name string
		url  string
		want bool
	}{
		{
			name: "Tailscale official",
			url:  "https://login.tailscale.com/a/123",
			want: false,
		},
		{
			name: "Tailscale with subdomain",
			url:  "https://login.tailscale.com/register/456",
			want: false,
		},
		{
			name: "Headscale",
			url:  "https://headscale.example.com/register/abc",
			want: true,
		},
		{
			name: "Custom server",
			url:  "https://custom.server.net/auth/xyz",
			want: true,
		},
		{
			name: "Localhost",
			url:  "http://localhost:8080/register/test",
			want: true,
		},
		{
			name: "IP address server",
			url:  "https://192.168.1.100/register/key",
			want: true,
		},
		{
			name: "Empty string",
			url:  "",
			want: true, // empty string doesn't contain "login.tailscale.com"
		},
		{
			name: "Tailscale look-alike domain",
			url:  "https://login.tailscale.com.evil.net/a/123",
			want: false, // Contains "login.tailscale.com" substring, so detected as Tailscale
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isHeadscaleOrCustomServer(tt.url); got != tt.want {
				t.Errorf("isHeadscaleOrCustomServer() = %v, want %v", got, tt.want)
			}
		})
	}
}
