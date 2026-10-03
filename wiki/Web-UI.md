# Web UI

Served at `http://<nas>:8585/` (and through the tunnel). Plain HTML/JS embedded in the binary, strict CSP, session token in `sessionStorage`.

- **Walkthrough** on first run: administrator (optionally with their own mywellness account), mywellness profiles, push status, how to connect the app.
- **Prenotazioni**: per-profile selector (defaults to the signed-in person's profile), active bookings x/max, tracked classes with state and actions (retry, remove, stop recurrence, cancel), *Prenotate su mywellness* (all active bookings of the profile, from the gateway or the official app/website, with cancel), latest activity. Auto-refresh every 20 s.
- **Lezioni**: schedule of the selected profile with pictures, free places, opening time, waiting-list state; "Prenota questa" / "Ogni settimana". Search box, auto-refresh.
- **Profili**: mywellness accounts (label, account, club, max, active, last login) with re-login/remove; add profile (family/private, mine/family member); gateway users (admin) with optional mywellness credentials that create the person's own profile.
- **Impostazioni** (admin): follow club opening time, booking rules table (text in class name → days before + time, `*` default, matched-classes preview), lead, burst, poll, days ahead, priority notifications; status; test push.
- **Registro**: activity log with profile and level filters, auto-refresh.
- **Header**: theme (system/light/dark) and language (it, en, es, fr, de) selectors, stored in the browser.
