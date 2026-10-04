# API Reference

Base path `/api/v1`, JSON. Authentication: `Authorization: Bearer <token>` obtained from `POST /auth/login`. Errors: `{"error":"message"}` with a meaningful HTTP status (400, 401, 403, 404, 409, 429, 502).

## Setup and auth
- `GET /setup` → `{needsSetup, version, push, publicUrl}`
- `POST /setup` `{username, password, displayName, mywellness?: {username, password, facilityUrl, maxBookings}}` → `{token, user, profile?, profileError?}` (only while no users exist)
- `POST /auth/login` `{username, password, deviceName, selfOnly}` → `{token, user, version, push, selfOnly}`; with `selfOnly: true` the device sees only the person's own profile even if admin (the iOS app always sends it); 5 failures per IP → 15 min lockout (429)
- `POST /auth/logout`, `GET /me` → `{user, device, version, push, myProfileId}`, `POST /me/password` `{oldPassword, newPassword}`
- `POST /devices/apns` `{token}` registers the APNs token of the calling device; `POST /devices/test-notification`

## Profiles
- `GET /profiles` → visible profiles (`id, label, username, displayName, facilityUrl, facilityId, facilityName, maxBookings, ownerUserIds, lastLoginAt, lastLoginError, activeBookings`)
- `POST /profiles` `{label, username, password, facilityUrl, maxBookings, private, mine, userId}` — verifies the mywellness login before saving; `mine: true` makes it the caller's own profile, `userId` (admin) assigns it to another user
- `PUT /profiles/{id}` `{label?, password?, maxBookings?, private?}`, `DELETE /profiles/{id}`, `POST /profiles/{id}/relogin`
- `GET /profiles/{id}/classes?q=&refresh=1` → schedule for the next *daysAhead* days: upstream fields plus `start`, `end`, `opensOn`, `tracked` (item, if any)
- `GET /profiles/{id}/bookings?refresh=1` → future classes with `isParticipant=true` (booked by the gateway **or** on mywellness)
- `POST /profiles/{id}/unbook` `{classId, partitionDate}` → cancels on mywellness (409 with a message when too late, i.e. less than 2 hours before); a tracked item becomes `cancelled`
- `POST /profiles/{id}/leave-waiting-list` `{classId, partitionDate, removeItem}` → leaves the mywellness waiting list; the tracked item becomes `cancelled` (or is removed with `removeItem: true`)

## Items (tracked classes)
- `GET /items?profile=` ; `POST /items` `{profileId, classId, partitionDate, recurring}` ; `DELETE /items/{id}?rule=1` (also stop the weekly rule) ; `POST /items/{id}/retry`
- Item: `id, profileId, classId, partitionDate, name, start, end, room, trainer, pictureUrl, serverOpensOn, fireAt, recurring, state, lastMessage, lastCheck, attempts, bookedAt, availablePlaces, maxParticipants`
- States: `pending, bursting, watching, waitingList, booked, failed, expired, cancelled`

## Engine
- `GET /settings`, `PUT /settings` (admin), `GET /log?profile=&limit=`, `GET /status`
## Admin
- `GET /users` (with `profileId`/`profileLabel` of the person's own profile), `PUT /users/{id}` `{username?, displayName?, isAdmin?, password?}`, `POST /users` `{username, password, displayName, isAdmin, mywellness?}` (creates the person's own profile too), `DELETE /users/{id}`, `POST /users/{id}/password`, `GET /devices`, `DELETE /devices/{id}`
