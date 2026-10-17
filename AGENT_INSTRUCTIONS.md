# Agent Instructions for FEAT-001 Parallel Implementation

**Feature:** CLI Headscale Registration QR Code with JSON Payload  
**Base Commit:** 0590a4ca9e7bfa5fbafbde57460ab18f775f7b4f  
**Date:** 2026-10-17

---

## Investigation Summary (Completed)

### Key Findings
1. **Expiry data NOT available** - `ipn.Notify.BrowseToURL` is just `*string`, no timestamp
2. **Solution:** Make `expires_at` optional in JSON payload (omit when unavailable)
3. **QR display location:** `cmd/tailscale/cli/up.go` lines 804-810
4. **Strategy:** Extend existing `--qr` flag; auto-detect server type

### Architecture
- Current: `--qr` flag displays URL-only QR via `qrcodes.Fprintln`
- Target: Display JSON QR for headscale/custom servers, URL QR for Tailscale

---

## Parallel Worktrees

### Worktree 1: Core Implementation
**Branch:** `feature/qr-json-implementation`  
**Path:** `../tailscale-qr-json-implementation`  
**Agent:** Implementation Agent  
**Focus:** Core JSON QR logic

**Tasks:**
1. Create helper functions in new file `cmd/tailscale/cli/qr.go`:
   - `parseAuthURL(url string) (serverURL, authID string, err error)`
   - `buildRegistrationQR(authURL string) (string, error)`
   - `isHeadscaleOrCustomServer(url string) bool`

2. Modify `cmd/tailscale/cli/up.go` lines 804-810:
   - Detect server type from authURL
   - Build JSON payload for headscale/custom servers
   - Display JSON QR for non-Tailscale, URL QR for Tailscale

3. JSON payload structure:
```json
{
  "type": "headscale-registration",
  "version": "1",
  "auth_id": "hskey-authreq-<id>",
  "server_url": "https://<control-server-url>"
}
```
Note: `expires_at` omitted (not available at CLI layer)

**Dependencies:** None (first implementation)  
**Deliverables:**
- New file: `cmd/tailscale/cli/qr.go`
- Modified: `cmd/tailscale/cli/up.go`
- Commit with working implementation

---

### Worktree 2: Tests
**Branch:** `feature/qr-json-tests`  
**Path:** `../tailscale-qr-json-tests`  
**Agent:** Testing Agent  
**Focus:** Unit and integration tests

**Tasks:**
1. Create/extend `cmd/tailscale/cli/qr_test.go`:
   - Test `parseAuthURL` with Tailscale, Headscale, custom URLs
   - Test `buildRegistrationQR` JSON output
   - Test `isHeadscaleOrCustomServer` detection

2. Extend `cmd/tailscale/cli/up_test.go`:
   - Integration test for QR display with mock IPN bus
   - Test both JSON QR and URL QR paths

**Expected Test Cases:**
- Tailscale URL: `https://login.tailscale.com/a/0123456789`
- Headscale URL: `https://headscale.example.com/register/hskey-authreq-abc123`
- Custom URL: `https://custom.server.net/auth/xyz`
- Edge cases: malformed URLs, missing components

**Dependencies:** Can start in parallel; interfaces defined in AGENT_INSTRUCTIONS  
**Deliverables:**
- New file: `cmd/tailscale/cli/qr_test.go`
- Modified: `cmd/tailscale/cli/up_test.go`
- All tests passing

---

### Worktree 3: Documentation
**Branch:** `feature/qr-json-docs`  
**Path:** `../tailscale-qr-json-docs`  
**Agent:** Documentation Agent  
**Focus:** Update documentation

**Tasks:**
1. Update `cmd/tailscale/cli/up.go` help text for `--qr` flag:
   - Explain JSON QR for headscale/custom servers
   - Note URL QR preserved for Tailscale

2. Update `docs/cli.md`:
   - Add section on QR code behavior
   - Document JSON payload structure
   - Add usage examples
   - Add troubleshooting section

3. Example usage to document:
```bash
# Tailscale server (URL QR - backward compatible)
./tailscale login --qr

# Headscale server (JSON QR)
./tailscale login --login-server=https://headscale.example.com --qr

# Custom server (JSON QR)
./tailscale login --login-server=https://custom.server.net --qr
```

**Dependencies:** Can start in parallel; spec defined in ROADMAP.md  
**Deliverables:**
- Modified: `cmd/tailscale/cli/up.go` (help text)
- Modified: `docs/cli.md`
- Clear, accurate documentation

---

## Implementation Order

1. **First:** Core Implementation (Worktree 1) - Sequential
2. **Parallel:** Tests (Worktree 2) + Documentation (Worktree 3)

---

## Shared Interfaces (for parallel work)

### Function Signatures (agreed contract)

```go
// parseAuthURL extracts server URL and auth ID from authentication URL.
// Returns empty strings if parsing fails.
func parseAuthURL(url string) (serverURL, authID string, err error)

// buildRegistrationQR creates JSON payload for QR code.
// Returns JSON string ready for QR encoding.
func buildRegistrationQR(authURL string) (string, error)

// isHeadscaleOrCustomServer detects non-Tailscale control servers.
// Returns true for headscale and custom servers.
func isHeadscaleOrCustomServer(url string) bool
```

### JSON Payload Structure

```go
type RegistrationQR struct {
    Type      string `json:"type"`       // "headscale-registration"
    Version   string `json:"version"`    // "1"
    AuthID    string `json:"auth_id"`    // "hskey-authreq-<id>"
    ServerURL string `json:"server_url"` // "https://<server>"
    // ExpiresAt omitted - not available at CLI layer
}
```

---

## Testing Commands

```bash
# Build
./tool/go build ./cmd/tailscale

# Test specific package
./tool/go test ./cmd/tailscale/cli -v

# Test with coverage
./tool/go test ./cmd/tailscale/cli -coverprofile=coverage.out

# Full test suite
./tool/go test ./...

# Lint
make check && make vet && make lint
```

---

## Integration After Parallel Work

1. Merge Worktree 1 (core implementation) first
2. Rebase Worktree 2 (tests) on updated main
3. Rebase Worktree 3 (docs) on updated main
4. Run full regression tests
5. Merge tests and docs

---

## Success Criteria

- [ ] Core implementation complete and working
- [ ] All unit tests pass
- [ ] Integration tests pass
- [ ] Full regression suite passes (`./tool/go test ./...`)
- [ ] Documentation clear and accurate
- [ ] Backward compatibility preserved (Tailscale URLs still work)
- [ ] No linter warnings

---

**END OF AGENT_INSTRUCTIONS.md**
