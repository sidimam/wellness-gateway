# Configuration

## Environment variables
| Variable | Default | Meaning |
|---|---|---|
| `LISTEN_ADDR` | `:8585` | Listen address |
| `CONFIG_DIR` | `/config` | State directory |
| `PUBLIC_URL` | | Address shown in the walkthrough |
| `TRUST_PROXY` | `false` | Trust `CF-Connecting-IP` / `X-Forwarded-For` |
| `APNS_KEY_PATH` / `APNS_KEY_ID` / `APNS_TEAM_ID` / `APNS_BUNDLE_ID` / `APNS_PRODUCTION` | team `X5SR67A8AL`, bundle `com.sdimambro.wellness-booking`, production `true` | Push |
| `WG_SECRET_KEY` | random in `secret.key` | Encryption passphrase |
| `TZ` | `Europe/Rome` | Timezone |

## Engine settings (web UI → Impostazioni, admin only)
- **Follow the club's opening time** (default on): the opening *day* comes from `bookingOpensOn` reported by mywellness; the *time of day* always comes from the matching rule. Off: day = class day − rule's days before.
- **Per-rule limit**: a named rule can carry `maxBookings` (e.g. Reformer 3): the gateway books that kind of class only while the profile has fewer active bookings of that kind, besides the profile-wide limit.
- **Opening rules**: text contained in the class name → days before + time; `*` is the default rule. Example for Wellness Town: `Reformer` → 3 days 05:00, `*` → 7 days 05:00. Rules without a name are ignored.
- **Lead** (ms, default 300), **burst** (s, default 120), **poll** (s, default 60, minimum 15; halved in the 4 hours before the class, ±15% jitter), **days ahead** (default 14), **priority notifications**.

## Per-profile
- **Max active bookings** (default 5) — the club's rule; bookings made on mywellness directly count too.
- **Visibility**: family or private.
