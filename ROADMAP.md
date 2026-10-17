# Tailscale Roadmap

**Last Updated:** 2026-10-17  
**Purpose:** Track features, enhancements, and technical debt across the Tailscale codebase.

---

## Table of Contents

- [Project Status](#project-status)
- [Active Features](#active-features)
- [Implementation Phases](#implementation-phases)
- [Completed Work](#completed-work)
- [Future Work](#future-work)

---

## Project Status

**Current Version:** See `VERSION.txt`  
**Go Version:** 1.27.2  
**Build System:** Makefile + `./tool/go`  
**Test Command:** `./tool/go test ./...`  
**Lint Command:** `make check && make vet && make lint`

---

## Active Features

### FEAT-001: CLI Headscale Registration QR Code with JSON Payload

**Status:** Pending  
**Priority:** Medium  
**Created:** 2026-10-17

#### Overview

Enable the Tailscale CLI to generate and display a QR code containing structured JSON metadata for headscale/control server registration when running in console-only (no GUI) environments.

#### Problem Statement

When a user runs `tailscale login --login-server=<url>` on a CLI-only system (e.g., server without GUI), the client prints an authentication URL but cannot easily transfer this to a mobile device for scanning. The current `--qr` flag encodes only the URL, but for headscale compatibility and better mobile app integration, a structured JSON payload is preferred.

#### Requirements

**Confirmed:**
- When `tailscale login --login-server=<url>` is executed, print authentication URL as text (existing behavior)
- Generate and display a terminal-based QR code containing a JSON payload:
  ```json
  {
    "type": "headscale-registration",
    "version": "1",
    "auth_id": "hskey-authreq-<id>",
    "server_url": "https://<control-server-url>",
    "expires_at": "2026-10-17T13:53:34.38788541+07:00"
  }
  ```
- QR code should be printed to terminal using ASCII/Unicode characters
- Must work without GUI or browser

**Implementation Decisions:**
- Extend existing `--qr` flag functionality in `cmd/tailscale/cli/up.go`
- Parse authentication URL to extract `auth_id` and `server_url`
- Use existing `util/qrcodes` package for QR generation
- Make JSON QR display conditional on flag or auto-detect headscale

**Open Questions:**
- **Q1:** Is `expires_at` currently available at the CLI layer? (See TASK FEAT-001-003)
- **Q2:** Should JSON QR be opt-in via new flag (e.g., `--qr-json`), or auto-enabled for non-Tailscale servers?
- **Q3:** Should we print both URL-only QR (backward compat) and JSON QR, or only JSON?
- **Q4:** What should happen if `expires_at` is unavailable? (omit field, use placeholder, or estimate?)

#### Risks

- **Expiry Data Unavailable:** `expires_at` may not be exposed in current control protocol at CLI layer
- **Backward Compatibility:** Existing mobile apps/scanners expect URL format
- **URL Parsing Complexity:** Auth URLs may have varied formats (Tailscale vs Headscale vs custom)
- **Mobile App Support:** Mobile apps may need updates to support JSON QR payload

---

## Implementation Phases

### Phase 1: Investigation and Preparation

**Objective:** Understand current architecture and determine feasibility of obtaining all required data.

#### TASK FEAT-001-001: Analyze Authentication URL Structure

**Status:** Pending  
**Objective:** Document the structure and variants of authentication URLs from different control servers.

**Files Affected:**
- Research only (no code changes)

**Implementation Details:**
1. Examine authentication URL generation in `control/controlclient/direct.go`
2. Test with Tailscale control server URLs (format: `https://login.tailscale.com/...`)
3. Test with Headscale URLs (format: `https://<server>/register/hskey-authreq-...`)
4. Document URL structure: base server URL, registration path, auth request ID, query parameters

**Acceptance Criteria:**
- Document listing Tailscale and Headscale URL formats with examples
- Parser logic to extract `auth_id` and `server_url`
- Edge cases documented (custom servers, URL encoding)

**Verification:**
```bash
./tailscale login
./tailscale login --login-server=https://headscale.example.com
```

**Risks:** URL formats may vary; custom servers may use non-standard formats

---

#### TASK FEAT-001-002: Inventory Existing QR Code Implementation

**Status:** Pending  
**Objective:** Document the current QR code implementation and identify reusable components.

**Files Affected:**
- `util/qrcodes/qrcodes.go` (read-only)
- `util/qrcodes/format.go` (read-only)
- `cmd/tailscale/cli/up.go` lines 98-100, 790-810 (read-only)

**Implementation Details:**
1. Review `util/qrcodes.Fprintln` function capabilities
2. Verify supported formats (ASCII, Large, Small, Auto)
3. Examine current CLI usage
4. Test JSON payload size (~200 bytes) against QR capacity

**Acceptance Criteria:**
- Documentation of available QR functions and format options
- Confirmation JSON payload fits within QR code capacity

**Verification:**
```bash
grep -r "qrcodes\\.Fprintln" cmd/tailscale/cli/
./tool/go doc tailscale.com/util/qrcodes
```

**Risks:** Terminal rendering issues on some platforms

---

#### TASK FEAT-001-003: Investigate Expiry Data Availability

**Status:** Pending  
**Objective:** Determine if `expires_at` timestamp is available at CLI layer.

**Files Affected:**
- `ipn/backend.go` (lines 311-450) (read-only)
- `tailcfg/tailcfg.go` (lines 1372-1387) (read-only)
- `control/controlclient/direct.go` (investigation)
- `ipn/ipnlocal/local.go` (investigation)

**Implementation Details:**
1. Check if `RegisterResponse.AuthURL` includes expiry metadata
2. Search for expiry fields in control protocol messages
3. Check if `ipn.Notify` can carry expiry information
4. Options if unavailable:
   - Option A: Request backend changes to expose expiry
   - Option B: Use default (+1 hour)
   - Option C: Omit `expires_at` when unavailable
   - Option D: Fetch via separate LocalAPI call

**Acceptance Criteria:**
- Document whether `expires_at` is available, needs backend changes, or should be omitted
- Recommendation on approach

**Verification:**
```bash
grep -r "AuthURL\|expires\|Expiry" tailcfg/ ipn/ control/
```

---

### Phase 2: Core Implementation

**Objective:** Implement JSON QR code generation and display in the CLI.

**Dependencies:** Phase 1 complete, decision made on Open Questions Q1-Q4

#### TASK FEAT-001-004: Add URL Parsing Logic

**Status:** Pending  
**Objective:** Implement parser to extract `auth_id` and `server_url` from authentication URLs.

**Dependencies:** FEAT-001-001 complete

**Files Affected:**
- `cmd/tailscale/cli/up.go` (new helper function)

**Implementation Details:**
Create `parseAuthURL(url string) (serverURL, authID string, err error)` function that:
- Handles Tailscale format: `https://login.tailscale.com/...`
- Handles Headscale format: `https://<server>/register/hskey-authreq-<id>`
- Uses `net/url` package for robust parsing
- Extracts base server URL (scheme + host) and auth request ID from path

**Acceptance Criteria:**
- Parser correctly extracts `server_url` and `auth_id` from Tailscale, Headscale, and custom URLs
- Returns descriptive errors for malformed URLs
- Handles URL encoding correctly

**Tests:**
Unit tests in `cmd/tailscale/cli/up_test.go`:
```go
func TestParseAuthURL(t *testing.T) {
    tests := []struct {
        name      string
        url       string
        wantURL   string
        wantID    string
        wantError bool
    }{
        {
            name:    "headscale format",
            url:     "https://headscale.example.com/register/hskey-authreq-6bfbec80d7cb",
            wantURL: "https://headscale.example.com",
            wantID:  "hskey-authreq-6bfbec80d7cb",
        },
        // ... more test cases
    }
}
```

**Verification:**
```bash
./tool/go test ./cmd/tailscale/cli -run TestParseAuthURL -v
```

**Risks:** URL format variations not covered in tests

---

#### TASK FEAT-001-005: Implement JSON QR Payload Generation

**Status:** Pending  
**Objective:** Create function to build and serialize the JSON QR payload.

**Dependencies:** FEAT-001-003 (expiry decision), FEAT-001-004 complete

**Files Affected:**
- `cmd/tailscale/cli/up.go` (new helper function and struct)

**Implementation Details:**
1. Define struct:
```go
type headscaleRegistrationQR struct {
    Type      string    `json:"type"`
    Version   string    `json:"version"`
    AuthID    string    `json:"auth_id"`
    ServerURL string    `json:"server_url"`
    ExpiresAt time.Time `json:"expires_at,omitempty"`
}
```
2. Create `buildRegistrationQR(authURL string, expiresAt *time.Time) (string, error)`
3. Call `parseAuthURL` to extract components
4. Marshal to JSON and return

**Acceptance Criteria:**
- JSON output matches specification
- `expires_at` included if available, omitted otherwise
- JSON is valid and compact

**Tests:**
Unit tests verifying JSON output with and without expiry

**Verification:**
```bash
./tool/go test ./cmd/tailscale/cli -run TestBuildRegistrationQR -v
```

**Risks:** JSON serialization edge cases (timezone formatting)

---

#### TASK FEAT-001-006: Integrate JSON QR into CLI Login Flow

**Status:** Pending  
**Objective:** Modify CLI login flow to display JSON QR code alongside authentication URL.

**Dependencies:** FEAT-001-004, FEAT-001-005 complete

**Files Affected:**
- `cmd/tailscale/cli/up.go` (lines 780-810, modify QR display logic)

**Implementation Details:**
Modify the notification handling in `runUp` function where auth URL is printed:
1. Add flag (if needed): `--qr-json` or auto-detect headscale URLs
2. Generate JSON payload using `buildRegistrationQR`
3. Display JSON QR using `qrcodes.Fprintln`
4. Optionally display URL-only QR for backward compatibility
5. Update help text

**Acceptance Criteria:**
- JSON QR code displayed when `--qr` flag used during login
- Authentication URL still printed as text
- QR renders correctly in terminal
- Error handling doesn't break login flow

**Tests:**
Manual testing:
```bash
./tailscale login --qr
./tailscale login --login-server=https://headscale.example.com --qr
```

**Verification:**
1. Run login with various flag combinations
2. Scan QR code with mobile reader, verify JSON payload
3. Confirm login completes successfully

---

### Phase 3: Testing and Documentation

**Objective:** Comprehensive testing and documentation updates.

**Dependencies:** Phase 2 complete

#### TASK FEAT-001-007: Add Integration Tests

**Status:** Pending  
**Objective:** Create comprehensive tests for the entire feature.

**Files Affected:**
- `cmd/tailscale/cli/up_test.go` (new integration tests)

**Implementation Details:**
- Create mock IPN bus watcher
- Test full flow from notification to QR display
- Test all flag combinations
- Test error cases

**Acceptance Criteria:**
- Tests cover Tailscale and Headscale login with QR
- Tests cover with/without expiry data
- All QR format options tested

**Verification:**
```bash
./tool/go test ./cmd/tailscale/cli -v
```

---

#### TASK FEAT-001-008: Update Documentation

**Status:** Pending  
**Objective:** Document the new QR code behavior.

**Files Affected:**
- `cmd/tailscale/cli/up.go` (help text)
- `docs/cli.md` (if exists)

**Implementation Details:**
1. Update `--qr` flag help text
2. Add documentation explaining JSON payload structure
3. Add troubleshooting section

**Acceptance Criteria:**
- Help text clearly explains QR feature
- Documentation includes usage examples and JSON schema

**Verification:**
```bash
./tailscale login --help | grep -A5 "qr"
```

---

#### TASK FEAT-001-009: Regression Testing

**Status:** Pending  
**Objective:** Ensure existing functionality is not broken.

**Implementation Details:**
1. Run full test suite: `./tool/go test ./...`
2. Run linters: `make check && make vet && make lint`
3. Test existing login workflows
4. Test on multiple platforms (Linux, macOS, Windows if available)

**Acceptance Criteria:**
- All existing tests pass
- No linter warnings
- Existing login workflows work as before

**Verification:**
```bash
make test
make check
./tailscale login  # without --qr
```

**Risks:** Platform-specific regressions (Windows terminal encoding)

---

## Completed Work

*(No items yet - this is a new roadmap)*

---

## Future Work

### Potential Enhancements (Not Scheduled)

- Mobile app QR scanner integration (coordinate with mobile teams)
- Extended metadata in QR payload (node info, network tags)
- QR code customization (error correction level, colors)
- Control server protocol improvements (standardize expiry field)

---

## Risk Mitigation

**High-Impact Risk:** Control server doesn't provide expiry data  
**Mitigation:** Make `expires_at` optional; document limitation  
**Contingency:** Ship without expiry field, add in future update

**Medium-Impact Risk:** Mobile apps don't support JSON QR payload  
**Mitigation:** Coordinate with mobile teams; provide URL-only fallback  
**Contingency:** Document that feature requires updated mobile app

**Medium-Impact Risk:** URL parsing fails for edge cases  
**Mitigation:** Comprehensive test cases; graceful error handling  
**Contingency:** User can still copy URL manually

---

## Implementation Checklist

Before marking feature complete:

- [ ] All Phase 1 investigation tasks completed
- [ ] URL parsing handles Tailscale, Headscale, and custom formats
- [ ] JSON payload structure matches specification
- [ ] QR code renders correctly in terminal
- [ ] Expiry data handled appropriately
- [ ] All unit tests pass
- [ ] Integration tests cover main scenarios
- [ ] Manual testing successful
- [ ] Documentation updated
- [ ] Regression tests pass
- [ ] No new linter warnings
- [ ] Feature validated on Linux

---

**END OF ROADMAP**


