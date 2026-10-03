# Installation

## Unraid (Community Applications template)
1. Docker → **Add Container** → *Template repositories* → add `https://github.com/sidimam/wellness-gateway` → pick **wellness-gateway**.
2. Port `8585`, appdata `/mnt/user/appdata/wellness-gateway`, **Public URL** = the address the iPhone uses (e.g. `https://booking.example.com`).
3. Push: copy the APNs `.p8` key to the appdata folder as `AuthKey_APNS.p8` and set **APNs key id**.
4. Apply → open the Web UI → **walkthrough**: create the administrator, add the mywellness profiles (login is verified immediately), check push, connect the app.

## Docker Compose
See `docker-compose.yml` in the repository.

## Cloudflare Tunnel
Public hostname → service `HTTP` → `http://<nas-lan-ip>:8585`; keep `TRUST_PROXY=true`. The gateway has its own authentication; if you also enable Cloudflare Access on the hostname, add a bypass for `/api/*` (or give the app a service token) or the iPhone will be stopped by the Access login page.

## Connect the iOS app
Wellness Booking → Altro → **Server**: address + gateway username/password. The app registers its APNs token with the gateway on login.
