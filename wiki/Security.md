# Security

- mywellness passwords are stored AES-256-GCM encrypted; the key lives in `/config/secret.key` (0600) or comes from `WG_SECRET_KEY`. Session tokens from mywellness are stored too and refreshed by re-login when they expire.
- Gateway user passwords: PBKDF2-SHA256, 600 000 iterations, random salt.
- Device tokens: 32 random bytes; revocable by the admin (Devices). Login lockout: 5 failures per IP → 15 minutes.
- No CORS, strict CSP on the web UI, security headers on every response, `/api/*` is `no-store`.
- Only the gateway talks to Technogym; the iOS app holds a gateway token only.
- Expose it over HTTPS (Cloudflare Tunnel or a reverse proxy). The gateway itself speaks plain HTTP on the LAN.
