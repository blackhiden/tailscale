// Copyright (c) Tailscale Inc & contributors
// SPDX-License-Identifier: BSD-3-Clause

package cli

import (
	"encoding/json"
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
		{name: "Tailscale URL", url: "https://login.tailscale.com/a/0123456789abcdef", wantServer: "https://login.tailscale.com", wantAuthID: "0123456789abcdef", wantErr: false},
		{name: "Headscale URL", url: "https://headscale.example.com/register/hskey-authreq-abc123", wantServer: "https://headscale.example.com", wantAuthID: "hskey-authreq-abc123", wantErr: false},
		{name: "Custom server", url: "https://custom.server.net/auth/xyz789", wantServer: "https://custom.server.net", wantAuthID: "xyz789", wantErr: false},
		{name: "Invalid URL", url: "not-a-url", wantErr: true},
		{name: "Empty URL", url: "", wantErr: true},
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
	authURL := "https://headscale.example.com/register/hskey-authreq-abc123"
	jsonStr, err := buildRegistrationQR(authURL)
	if err != nil {
		t.Fatalf("buildRegistrationQR() failed: %v", err)
	}

	var qr registrationQR
	if err := json.Unmarshal([]byte(jsonStr), &qr); err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}
	if qr.Type != "headscale-registration" {
		t.Errorf("Type = %v, want headscale-registration", qr.Type)
	}
	if qr.Version != "1" {
		t.Errorf("Version = %v, want 1", qr.Version)
	}
	if qr.ServerURL != "https://headscale.example.com" {
		t.Errorf("ServerURL = %v, want https://headscale.example.com", qr.ServerURL)
	}
	if qr.AuthID != "hskey-authreq-abc123" {
		t.Errorf("AuthID = %v, want hskey-authreq-abc123", qr.AuthID)
	}
}

func TestIsHeadscaleOrCustomServer(t *testing.T) {
	tests := []struct {
		name string
		url  string
		want bool
	}{
		{"Tailscale", "https://login.tailscale.com/a/123", false},
		{"Headscale", "https://headscale.example.com/register/abc", true},
		{"Custom", "https://custom.server.net/auth/xyz", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isHeadscaleOrCustomServer(tt.url); got != tt.want {
				t.Errorf("isHeadscaleOrCustomServer() = %v, want %v", got, tt.want)
			}
		})
	}
}
