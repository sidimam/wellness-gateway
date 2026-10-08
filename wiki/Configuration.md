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
- **Per-rule limit (separate quota)**: a named rule can carry `maxBookings` (e.g. Reformer 3). Classes covered by such a rule have their **own quota**: they are checked only against the rule's maximum and are **not counted** in the profile-wide limit. A rule with `maxBookings` 0 has no quota of its own and its classes count in the profile limit.
- **Opening rules**: text contained in the class name → days before + time; `*` is the default rule. Example for Wellness Town: `Reformer` → 3 days 05:00, `*` → 7 days 05:00. Rules without a name are ignored.
- **Lead** (ms, default 300), **burst** (s, default 120), **days ahead** (default 14), **priority notifications**.
- **Watching cadence**: `pollSeconds` (default 15, minimum 5) between public schedule reads; `nearPollSeconds` (default 3, minimum 2) in the last `nearHours` hours (default 4); ±15% jitter. Reads are anonymous (no token), so the cadence does not expose the account; one authenticated read per class every 60 s keeps participation and waiting-list position up to date.

## Per-profile
- **Max active bookings** (default 5) — the club's rule; bookings made on mywellness directly count too. Classes covered by a rule with its own `maxBookings` (e.g. Reformer 3) are excluded from this count.
- **Visibility**: family or private.
