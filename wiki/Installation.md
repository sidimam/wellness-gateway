# Installation

## Unraid (Community Applications template)
1. Docker → **Add Container** → *Template repositories* → add `https://github.com/sidimam/wellness-gateway` → pick **wellness-gateway**.
2. Port `8585`, appdata `/mnt/user/appdata/wellness-gateway`, **Public URL** = the address the iPhone uses (e.g. `https://booking.example.com`).
3. Push: copy the APNs `.p8` key to the appdata folder as `AuthKey_APNS.p8` and set **APNs key id**.
4. Apply → open the Web UI → **walkthrough**: create the administrator, add the mywellness profiles (login is verified immediately), check push, connect the app.

## Docker Compose
See `docker-compose.yml` in the repository.

## Cloudflare Tunnel
Public hostname → service `HTTP` → `http://<nas-lan-ip>:8585`; keep `TRUST_PROXY=true`. The gateway has its own authentication, so Access is optional. If you do protect the hostname with Cloudflare Access, create a **service token** (Zero Trust → Access → Service Auth) with a *Service Auth* policy on the application and enter its Client ID / Client Secret in the iOS app (Altro → Server di casa → *Connessione tramite Cloudflare Access*): the app sends `CF-Access-Client-Id` / `CF-Access-Client-Secret` on every request. Without the token the iPhone would be stopped by the Access login page.

## Connect the iOS app
Wellness Booking → Altro → **Server**: address + gateway username/password. The app registers its APNs token with the gateway on login.
