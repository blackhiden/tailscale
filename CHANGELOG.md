# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### New Features

- **JSON QR Code for Headscale**: The `--qr` flag now generates a JSON payload QR code when authenticating with headscale or custom control servers, enabling better mobile app integration. Tailscale server authentication continues to use URL-only QR codes for backward compatibility.

### Technical Details

- JSON payload includes: `type`, `version`, `auth_id`, `server_url`
- The `expires_at` field is intentionally omitted as expiry data is not available at the CLI layer
- Automatic server type detection: JSON QR for headscale/custom, URL QR for Tailscale
- Backward compatible with existing mobile app implementations
- Implements FEAT-001 from ROADMAP.md

---

## Previous Versions

See git history for changes prior to this roadmap tracking.
