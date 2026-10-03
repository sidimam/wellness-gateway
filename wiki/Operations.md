# Operations

- **Update**: the template pins `:latest`; use Unraid's *Check for Updates* or the Auto Update plugin. State survives (it is in `/config`). Devices stay logged in.
- **Logs**: container log (`docker logs wellness-gateway`) and the in-app *Registro* (also `GET /api/v1/log`).
- **Backup**: copy `/mnt/user/appdata/wellness-gateway` (`state.json` + `secret.key`; without the key the stored passwords cannot be decrypted).
- **Health**: `GET /healthz` → `{"status":"ok","version":…}`.
- **Troubleshooting**
  - *login mywellness fallito*: wrong password or MFA enabled on the account → fix in Profili → Rifai login.
  - *Limite prenotazioni raggiunto*: the profile already has N active bookings (including ones made on mywellness).
  - *Non ancora aperta*: the server refused the Book before opening; the burst keeps retrying for `burstSeconds`.
  - No push: check `APNS_*` variables and that the iOS build matches `APNS_PRODUCTION` (TestFlight = true).
