# Changelog

## v0.1.1 (2026-10-03)
- Web UI: theme (system/light/dark) and language (it/en/es/fr/de) selectors in the header; automatic refresh (20 s, and on tab focus) of bookings, classes and log.
- Gateway users can own their mywellness profile (`userId`): created with the user (setup / add user with optional mywellness credentials) or linked with `mine: true`; `GET /me` returns `myProfileId` and clients default to it.
- Web UI: profile selector in every section, "Ultime attività" card on the bookings page, log filters (profile, level), opening-rules table with matched-classes preview and explicit add/delete, stale-view fix.

## v0.1.0 (2026-10-03)
- First release: multi-profile engine (scheduler at opening time, waiting list, continuous watching of freed spots, weekly recurrences, per-club booking limit), external cancellation sync, APNs Time Sensitive push, web UI with walkthrough, REST API for the Wellness Booking iOS app, Unraid template, GHCR multi-arch image.
