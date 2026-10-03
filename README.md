<p align="center"><img src="assets/icon.png" width="96" alt=""></p>

# wellness-gateway

**Always-on booking engine for Technogym mywellness classes**, built to run as a small container on an Unraid NAS (or any Docker host). It keeps several mywellness accounts (the whole family), books each class the instant booking opens, joins the waiting list when a class is full and keeps **watching** it to grab a freed spot immediately, 24/7. It detects cancellations made in the official app or website, enforces the club's active-bookings limit and sends **Time Sensitive push notifications** to the companion iOS app, [Wellness Booking](https://github.com/sidimam/wellness-booking).

Why a server? iOS never runs an app at an exact time in the background, and the mywellness waiting list only *notifies* when a spot frees up: somebody still has to click "book" first. The gateway is that somebody, every few seconds, for everyone in the house.

## Features

- **Multi-profile**: one entry per mywellness account (email + password, encrypted at rest with AES-256-GCM). The gateway logs in, keeps the session token and renews it when it expires. Profiles are visible to the whole family or to one gateway user only.
- **Scheduler**: books at the opening time reported by the club for each class (`bookingOpensOn`), with a configurable lead (ms) and a burst of retries after opening. Per-class **opening rules** (text contained in the class name → days before + time, `*` for everything else) are used when the club reports nothing or when you turn "follow the club" off.
- **Watching**: when a class is full the gateway joins the waiting list and polls the schedule every N seconds (tighter in the last useful hours before the 2-hour cancellation deadline). As soon as `availablePlaces > 0` it calls `Book`.
- **Recurrences**: "every Tuesday at 17:30" rules attach new occurrences as they appear in the schedule.
- **Limit**: default 5 active bookings per profile (club rule), counting bookings made directly on mywellness too.
- **Two-way sync**: bookings made in the official app/website show up (and can be cancelled from here); a class cancelled there is marked *cancelled* and is **not** rebooked.
- **Push**: APNs HTTP/2 with JWT (ES256), `interruption-level: time-sensitive`, no third-party libraries.
- **Web UI** with a first-run walkthrough (admin → profiles → push → connect the app), bookings, schedule with class pictures, profiles, users, settings and activity log.
- **Users and devices**: local accounts with PBKDF2-SHA256 passwords, per-device bearer tokens, login rate limiting, admin role.
- Standard library only, single static binary, ~10 MB image, runs as `nobody:users` (99:100).

## Install on Unraid

1. Docker → Add Container → *Template repositories* → add `https://github.com/sidimam/wellness-gateway` → choose **wellness-gateway**.
2. Keep port `8585` and appdata `/mnt/user/appdata/wellness-gateway`. Set **Public URL** to the address the iPhone will use (e.g. `https://booking.example.com`).
3. For push notifications copy your APNs key to `/mnt/user/appdata/wellness-gateway/AuthKey_APNS.p8` and fill **APNs key id** (team id and bundle id are pre-filled for the Wellness Booking app).
4. Apply, open the Web UI (`http://<nas>:8585`) and follow the walkthrough.

With Docker Compose see [docker-compose.yml](docker-compose.yml).

### Remote access (Cloudflare Tunnel)

Add a public hostname to your tunnel → service `HTTP`, URL `http://<nas-lan-ip>:8585`. Keep `TRUST_PROXY=true`. Optionally protect it with Cloudflare Access: allow by e-mail for the browser and add a *Service Auth* policy with a service token; the Wellness Booking app can send the token's `CF-Access-Client-Id`/`CF-Access-Client-Secret` headers (Altro → Server di casa).

## Configuration (environment)

| Variable | Default | Meaning |
|---|---|---|
| `LISTEN_ADDR` | `:8585` | Listen address |
| `CONFIG_DIR` | `/config` | State dir: `state.json`, `secret.key` |
| `PUBLIC_URL` | | Address shown in the walkthrough for the app |
| `TRUST_PROXY` | `false` | Use `CF-Connecting-IP` / `X-Forwarded-For` for login rate limiting |
| `APNS_KEY_PATH`, `APNS_KEY_ID`, `APNS_TEAM_ID`, `APNS_BUNDLE_ID` | team `X5SR67A8AL`, bundle `com.sdimambro.wellness-booking` | Push configuration; push is disabled when the key is missing |
| `APNS_PRODUCTION` | `true` | Production APNs host (TestFlight/App Store); `false` for Xcode debug builds |
| `WG_SECRET_KEY` | | Passphrase for encrypting mywellness passwords; otherwise a random key is stored in `/config/secret.key` |
| `TZ` | `Europe/Rome` | Timezone |

## API (for the iOS app)

All endpoints are under `/api/v1`, JSON, `Authorization: Bearer <device token>`. Errors are `{"error":"…"}`.

| Method | Path | Purpose |
|---|---|---|
| GET/POST | `/setup` | First admin (only when no users exist) |
| POST | `/auth/login` `{username,password,deviceName}` → `{token,user}` | Device token |
| POST | `/auth/logout` · GET `/me` · POST `/me/password` | Session |
| POST | `/devices/apns` `{token}` · POST `/devices/test-notification` | Push registration |
| GET | `/profiles` · POST `/profiles` · PUT/DELETE `/profiles/{id}` · POST `/profiles/{id}/relogin` | mywellness accounts |
| GET | `/profiles/{id}/classes[?q=&refresh=1]` | Schedule (with `isParticipant`, `pictureUrl`, `tracked`) |
| GET | `/profiles/{id}/bookings` · POST `/profiles/{id}/unbook` `{classId,partitionDate}` | Active bookings (gateway + mywellness) and cancellation |
| GET | `/items[?profile=]` · POST `/items` `{profileId,classId,partitionDate,recurring}` · DELETE `/items/{id}[?rule=1]` · POST `/items/{id}/retry` | Tracked classes |
| GET/PUT | `/settings` · GET `/log` · GET `/status` | Engine |
| GET/POST/DELETE | `/users…` · GET/DELETE `/devices…` | Admin |
| GET | `/healthz` | No auth |

## mywellness endpoints used

Derived from the official widget bundle (`widgets.mywellness.com`): `POST core.mywellness.com/v2/enduser/authentication/login`, `GET core.mywellness.com/v2/enduser/facility/detail`, `GET calendar.mywellness.com/v2/enduser/class/Search`, `POST calendar.mywellness.com/v2/enduser/class/Book`, `POST …/class/Unbook`, with headers `X-MWAPPS-APPID`, `X-MWAPPS-CLIENT: enduserweb`, `Authorization: Bearer`. Full table in the [wiki](https://github.com/sidimam/wellness-gateway/wiki/API-mywellness).

## Development

```bash
go build ./... && go vet ./... && gofmt -l .
CONFIG_DIR=/tmp/wg LISTEN_ADDR=:8585 go run ./cmd/wellness-gateway
docker build --build-arg VERSION=dev -t wellness-gateway:dev .
```

Images are built by GitHub Actions for `linux/amd64` and `linux/arm64` and pushed to `ghcr.io/sidimam/wellness-gateway` (`latest` from `main`, `X.Y.Z` from tags).

## License

MIT © 2026 Simone Di Mambro. Not affiliated with Technogym.
