# Initial Prompt for Testing Agent

**Worktree:** `/home/denny/Project/tailscale-project/tailscale-qr-json-tests`  
**Branch:** `feature/qr-json-tests`  
**Task:** Unit and Integration Tests for JSON QR Code

---

## Your Mission

Write comprehensive tests for the JSON QR code feature implemented in FEAT-001.

## Context

You are working on **FEAT-001: CLI Headscale Registration QR Code with JSON Payload** from ROADMAP.md.

The core implementation defines these functions (see AGENT_INSTRUCTIONS.md):
- `parseAuthURL(url string) (serverURL, authID string, err error)`
- `buildRegistrationQR(authURL string) (string, error)`
- `isHeadscaleOrCustomServer(url string) bool`

## Your Tasks

### 1. Create `cmd/tailscale/cli/qr_test.go`

```go
// Copyright (c) Tailscale Inc & contributors
// SPDX-License-Identifier: BSD-3-Clause

package cli

import (
	"encoding/json"
	"testing"
)

func TestParseAuthURL(t *testing.T) {
	tests := []struct {
		name      string
		url       string
		wantServer string
		wantAuthID string
		wantErr   bool
	}{
		{
			name:       "Tailscale URL",
			url:        "https://login.tailscale.com/a/0123456789abcdef",
			wantServer: "https://login.tailscale.com",
			wantAuthID: "0123456789abcdef",
			wantErr:    false,
		},
		{
			name:       "Headscale URL",
			url:        "https://headscale.example.com/register/hskey-authreq-abc123",
			wantServer: "https://headscale.example.com",
			wantAuthID: "hskey-authreq-abc123",
			wantErr:    false,
		},
		{
			name:       "Custom control server",
			url:        "https://custom.server.net/auth/xyz789",
			wantServer: "https://custom.server.net",
			wantAuthID: "xyz789",
			wantErr:    false,
		},
		{
			name:    "Invalid URL",
			url:     "not-a-url",
			wantErr: true,
		},
		{
			name:    "Empty URL",
			url:     "",
			wantErr: true,
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
		name    string
		authURL string
		wantErr bool
		validate func(t *testing.T, jsonStr string)
	}{
		{
			name:    "Headscale URL",
			authURL: "https://headscale.example.com/register/hskey-authreq-abc123",
			wantErr: false,
			validate: func(t *testing.T, jsonStr string) {
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
			},
		},
		{
			name:    "Invalid URL",
			authURL: "not-a-url",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonStr, err := buildRegistrationQR(tt.authURL)
			if (err != nil) != tt.wantErr {
				t.Errorf("buildRegistrationQR() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && tt.validate != nil {
				tt.validate(t, jsonStr)
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
```

### 2. Run Tests

```bash
# Test the qr package
./tool/go test ./cmd/tailscale/cli -run TestParseAuthURL -v
./tool/go test ./cmd/tailscale/cli -run TestBuildRegistrationQR -v
./tool/go test ./cmd/tailscale/cli -run TestIsHeadscaleOrCustomServer -v

# All CLI tests
./tool/go test ./cmd/tailscale/cli -v
```

## Implementation Guidelines

- Read ROADMAP.md and AGENT_INSTRUCTIONS.md first
- Use table-driven tests
- Cover edge cases (empty, malformed, various formats)
- Validate JSON structure and content
- Tests should be independent and repeatable

## Commit Message

```
cli: add tests for JSON QR code generation

Adds comprehensive unit tests for parseAuthURL, buildRegistrationQR,
and isHeadscaleOrCustomServer functions.

Tests cover Tailscale, Headscale, and custom control server URLs,
plus edge cases for error handling.

Implements FEAT-001-007 from ROADMAP.md.

Signed-off-by: Your Name <your.email@example.com>
```

## When Complete

1. Verify all tests pass: `./tool/go test ./cmd/tailscale/cli -v`
2. Check coverage: `./tool/go test ./cmd/tailscale/cli -coverprofile=coverage.out`
3. Commit your changes
4. Report completion with commit SHA

---

**Start by reading ROADMAP.md and AGENT_INSTRUCTIONS.md, then write comprehensive tests.**
