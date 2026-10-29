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

## Engine settings (web UI → Impostazioni; personal for every user, defaults set by the admin)
- Each gateway user has **their own** scheduler settings, applied to the classes of their profiles; the values below are the defaults used until a user saves their own. The administrator edits the defaults with the scope selector ("Le impostazioni predefinite"); any user can go back to the defaults with "Torna alle predefinite" (web) or "Torna alle predefinite" (iOS).
- **Follow the club's opening time** (default on): the opening *day* comes from `bookingOpensOn` reported by mywellness; the *time of day* always comes from the matching rule. Off: day = class day − rule's days before.
- **Never before the opening**: with "follow the club" on, when the opening reported by mywellness is later than the rule's time, the gateway fires one minute after the reported opening.
- **Valid from** (`validFrom`): a rule can start on a given opening day; keep the old rule and add a new one with the same name and the date (e.g. Reformer 05:01 always + Reformer 07:01 from 2026-11-01). The time of day comes from the rule valid on the opening day.
- **Max per week** (`maxPerWeek`): at most N bookings of that kind in the same Monday–Sunday week, bookings made directly on mywellness included (e.g. Reformer 2).
- **Per-rule limit (separate quota)**: a named rule can carry `maxBookings` (e.g. Reformer 3). Classes covered by such a rule have their **own quota**: they are checked only against the rule's maximum and are **not counted** in the profile-wide limit. A rule with `maxBookings` 0 has no quota of its own and its classes count in the profile limit.
- **Opening rules**: text contained in the class name → days before + time; `*` is the default rule. Example for Wellness Town: `Reformer` → 3 days 05:00, `*` → 7 days 05:00. Rules without a name are ignored.
- **Lead** (ms, default 300), **burst** (s, default 120), **days ahead** (default 14), **priority notifications**.
- **Watching cadence**: `pollSeconds` (default 15, minimum 5) between public schedule reads; `nearPollSeconds` (default 3, minimum 2) in the last `nearHours` hours (default 4); ±15% jitter. Reads are anonymous (no token), so the cadence does not expose the account; one authenticated read per class every 60 s keeps participation and waiting-list position up to date.

## People database
- `people.db` (SQLite) in the data dir, created automatically; tables `members`, `classes`, `attendance`, `leaderboard`. Back it up with the data dir. Delete it to forget everything: it is rebuilt from the next reads.

## Per-profile
- A profile is one mywellness account **at one club**. The same account can be linked to several clubs: add them from the user dialog (web) or the profile page (iOS); each club keeps its own schedule, tracked classes and limits.
- **Max active bookings** (default 5) — the club's rule; bookings made on mywellness directly count too. Classes covered by a rule with its own `maxBookings` (e.g. Reformer 3) are excluded from this count.
- **Visibility**: family or private.
