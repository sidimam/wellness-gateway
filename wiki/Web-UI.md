# Web UI

Served at `http://<nas>:8585/` (and through the tunnel). Plain HTML/JS embedded in the binary, strict CSP, session token in `sessionStorage`.

- **Walkthrough** on first run: administrator (optionally with their own mywellness account), mywellness profiles, push status, how to connect the app.
- **Prenotazioni**: per-profile selector (defaults to the signed-in person's profile), active bookings x/max, tracked classes with state and actions (retry, remove, stop recurrence, cancel), *Prenotate su mywellness* (all active bookings of the profile, from the gateway or the official app/website, with cancel), latest activity. Auto-refresh every 20 s.
- **Lezioni**: schedule of the selected profile with pictures, free places, opening time, waiting-list state; "Prenota questa" / "Ogni settimana". Search box, auto-refresh.
- **Profili** ("Utenti e account mywellness"): users table (username, name, role, mywellness account) with *Modifica*/*Elimina* and *Collega account mywellness*; add user (gateway username/password, role, **required** mywellness email/password, club, max bookings, visibility) which creates the person's profile; accounts table with *Modifica* (label, mywellness password, max, visibility, owner), *Rifai login*, *Rimuovi*.
- **Impostazioni** (admin): follow club opening time, booking rules table (text in class name → days before + time, `*` default, matched-classes preview), lead, burst, poll, days ahead, priority notifications; status; test push.
- **Registro**: activity log with profile and level filters, auto-refresh.
- **Impostazioni → Aspetto e lingua** (every user): theme (system/light/dark) and language (it, en, es, fr, de), stored in the browser. Every password field has a show/hide toggle.
- **Profili → Utenti** (admin): *Modifica* changes username, name, role (user/administrator) and password; *Collega account mywellness* creates the person's own profile.
