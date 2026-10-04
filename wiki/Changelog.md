# Changelog

## v0.1.13 (2026-10-04)
- **Fix**: mywellness session tokens expire after about an hour and the schedule API then answers as an anonymous user (no `isParticipant`), so active bookings dropped to 0 and limits/cancellation detection were wrong. The gateway now verifies the session with `GetLoginStatus` (at most every 5 minutes) before authenticated reads, books, cancellations and waiting-list operations, and logs in again automatically when the token is invalid.

## v0.1.12 (2026-10-04)
- User isolation: devices that log in with `selfOnly: true` (the iOS app) see only the signed-in person's profile, items, bookings and log, even for administrators; user management is web-only. Existing app devices are migrated. Regular users are isolated on the web too; administrators keep the family view in the web UI.

## v0.1.11 (2026-10-04)
- Human-like traffic: watching every 60 s by default (minimum 15), 30 s in the 4 hours before the class, ±15% jitter; one schedule request per profile and day shared by all tracked classes of that day (10 s cache); browser User-Agent. Existing settings below 30 s are migrated to 60 s.

## v0.1.10 (2026-10-04)
- Booking time always comes from the matching rule (hour:minute); "follow the club" now decides only the opening day. Fixes rules set to 05:01 still showing 05:00.
- Per-rule active-bookings limit (`maxBookings` on a rule, e.g. Reformer 3) in addition to the per-profile limit; `GET /profiles` returns `limits` counters.

## v0.1.9 (2026-10-03)
- Opening rules: a named rule always decides the booking time (even with "follow the club" on); the club's reported opening time applies only to classes covered by `*`.

## v0.1.8 (2026-10-03)
- Local-only administrators (no mywellness account) can be created; the mywellness account is required only for regular users.

## v0.1.7 (2026-10-03)
- Profiles created before 0.1.6 reload their mywellness identity (name, picture) automatically at the first refresh.

## v0.1.6 (2026-10-03)
- Visibility: non-admin users see only their own profile (and profiles explicitly shared with them); administrators see every profile and can switch between them in the app and the web UI.
- Profile identity from mywellness (`firstName`, `lastName`, `nickName`, `email`, `pictureUrl`, `thumbUrl`) stored at login and exposed in `GET /profiles`; avatars in the web UI.

## v0.1.5 (2026-10-03)
- Recurrences attach only occurrences on or after the chosen class.
- Users: per-user dialog (gateway access, role, mywellness account with link/re-login/unlink/remove, assign an orphan account); role changes and deletion allowed for any admin as long as one admin remains; deleting a user keeps their mywellness account as orphan.

## v0.1.4 (2026-10-03)
- Leave the mywellness waiting list from the gateway (`POST /profiles/{id}/leave-waiting-list` `{classId, partitionDate, removeItem}` → item becomes `cancelled`); Unbook result codes handled (`TooLate` = less than 2 hours before, `BookingNotAvailable`); outline eye icon on password fields.

## v0.1.3 (2026-10-03)
- Web UI: "Utenti e account mywellness" merges users and profiles: adding a user requires their mywellness email/password and creates the person's profile; every profile has *Modifica* (label, mywellness password, max bookings, visibility, owner user); link an account to an existing user.

## v0.1.2 (2026-10-03)
- Theme and language selectors moved into Settings; eye toggle on password fields.
- Edit existing users (`PUT /users/{id}`: username, displayName, isAdmin, password) and link a mywellness account to an existing user from the users table; `GET /users` returns `profileId`/`profileLabel`; hint on the Profiles page when the signed-in person has no own profile.

## v0.1.1 (2026-10-03)
- Web UI: theme (system/light/dark) and language (it/en/es/fr/de) selectors in the header; automatic refresh (20 s, and on tab focus) of bookings, classes and log.
- Gateway users can own their mywellness profile (`userId`): created with the user (setup / add user with optional mywellness credentials) or linked with `mine: true`; `GET /me` returns `myProfileId` and clients default to it.
- Web UI: profile selector in every section, "Ultime attività" card on the bookings page, log filters (profile, level), opening-rules table with matched-classes preview and explicit add/delete, stale-view fix.

## v0.1.0 (2026-10-03)
- First release: multi-profile engine (scheduler at opening time, waiting list, continuous watching of freed spots, weekly recurrences, per-club booking limit), external cancellation sync, APNs Time Sensitive push, web UI with walkthrough, REST API for the Wellness Booking iOS app, Unraid template, GHCR multi-arch image.
