# Initial Prompt for Implementation Agent

**Worktree:** `/home/denny/Project/tailscale-project/tailscale-qr-json-implementation`  
**Branch:** `feature/qr-json-implementation`  
**Task:** Core JSON QR Code Implementation

---

## Your Mission

Implement JSON QR code generation for headscale/custom control servers in the Tailscale CLI.

## Context

You are working on **FEAT-001: CLI Headscale Registration QR Code with JSON Payload** from ROADMAP.md.

**Investigation completed - Key findings:**
- Expiry data (`expires_at`) is NOT available at CLI layer
- Must make JSON payload work WITHOUT expiry field
- Current QR code location: `cmd/tailscale/cli/up.go` lines 804-810
- Strategy: Auto-detect server type and display appropriate QR

## Your Tasks

### 1. Create `cmd/tailscale/cli/qr.go`

```go
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
// Returns empty strings if parsing fails.
func parseAuthURL(authURL string) (serverURL, authID string, err error) {
	// TODO: Implement parsing logic for:
	// - Tailscale: https://login.tailscale.com/a/0123456789
	// - Headscale: https://headscale.example.com/register/hskey-authreq-abc123
	// - Custom: various formats
}

// buildRegistrationQR creates JSON payload for QR code.
// Returns JSON string ready for QR encoding.
func buildRegistrationQR(authURL string) (string, error) {
	// TODO: Parse URL, build JSON, marshal
}

// isHeadscaleOrCustomServer detects non-Tailscale control servers.
func isHeadscaleOrCustomServer(authURL string) bool {
	// TODO: Return true if not login.tailscale.com
}
```

### 2. Modify `cmd/tailscale/cli/up.go`

Find lines 804-810:
```go
fmt.Fprintf(Stderr, "\nTo authenticate, visit:\n\n\t%s\n\n", authURL)
if upArgs.qr && buildfeatures.HasQRCodes {
	_, err := qrcodes.Fprintln(Stderr, qrcodes.Format(upArgs.qrFormat), authURL)
	if err != nil {
		log.Print(err)
	}
}
```

Replace with:
```go
fmt.Fprintf(Stderr, "\nTo authenticate, visit:\n\n\t%s\n\n", authURL)
if upArgs.qr && buildfeatures.HasQRCodes {
	// Detect server type and display appropriate QR
	var qrContent string
	if isHeadscaleOrCustomServer(authURL) {
		// Display JSON QR for headscale/custom servers
		jsonQR, err := buildRegistrationQR(authURL)
		if err != nil {
			// Fallback to URL QR if JSON generation fails
			qrContent = authURL
		} else {
			qrContent = jsonQR
		}
	} else {
		// Display URL QR for Tailscale (backward compatible)
		qrContent = authURL
	}
	
	_, err := qrcodes.Fprintln(Stderr, qrcodes.Format(upArgs.qrFormat), qrContent)
	if err != nil {
		log.Print(err)
	}
}
```

### 3. JSON Payload Structure

```json
{
  "type": "headscale-registration",
  "version": "1",
  "auth_id": "hskey-authreq-<id>",
  "server_url": "https://<control-server-url>"
}
```

**Note:** `expires_at` is omitted because it's not available at the CLI layer.

## Implementation Guidelines

- Read ROADMAP.md and AGENT_INSTRUCTIONS.md first
- Keep backward compatibility for Tailscale URLs
- Handle parsing errors gracefully (fallback to URL QR)
- Use existing imports where possible
- Follow project coding style

## Testing Locally

```bash
# Build
./tool/go build ./cmd/tailscale

# Manual test
./tailscale login --login-server=https://headscale.example.com --qr
```

## Commit Message

```
cli: add JSON QR code support for headscale/custom servers

Extends the --qr flag to display a JSON payload QR code when
authenticating with headscale or custom control servers, while
preserving URL-only QR codes for Tailscale servers.

The JSON payload includes type, version, auth_id, and server_url.
The expires_at field is omitted as expiry data is not available
at the CLI layer.

Implements FEAT-001-004 from ROADMAP.md.

Signed-off-by: Your Name <your.email@example.com>
```

## When Complete

1. Verify code compiles: `./tool/go build ./cmd/tailscale`
2. Test manually if possible
3. Commit your changes
4. Report completion with commit SHA

---

**Start by reading ROADMAP.md and AGENT_INSTRUCTIONS.md, then implement the solution.**
