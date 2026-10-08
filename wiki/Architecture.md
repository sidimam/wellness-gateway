# Architecture

```
iPhone (Wellness Booking) ──HTTPS/Bearer──▶ wellness-gateway (Unraid container) ──HTTPS──▶ mywellness API
        ▲                                        │
        └────────────── APNs push ◀──────────────┘
```

- **Users** are local gateway accounts (Simone, Daniela, Sofia…). Each login from a device issues a long-lived bearer token stored in the iPhone Keychain.
- **Profiles** are mywellness accounts. Each user sees their own profile only; administrators see every profile (the family) and can act on any of them. Only the gateway talks to Technogym; the phone never holds mywellness credentials.
- **Engine** loop (one goroutine): every tick it evaluates all tracked items: `pending` → at `fireAt − lead` becomes `bursting` (Book every 1.5 s for `burstSeconds`) → `booked`, or `waitingList`/`watching` when full → reads the day's **public** schedule (no token; one request per club and day shared by every profile and class) every `pollSeconds` (default 15 s) and every `nearPollSeconds` (default 3 s) in the last `nearHours` hours (default 4), ±15% jitter, plus one authenticated read per class every 60 s for participation and waiting-list position; it books the moment a place shows up (`availablePlaces > 0`, participants < max, or `bookingUserStatus = CanBook`), retrying 3 times 0.7 s apart. Items expire 5 minutes before the class starts.
- **Calendar refresh** every 10 minutes per profile (authenticated, so `isParticipant` is known): updates places, marks items booked/cancelled from outside, attaches weekly recurrences, recounts active bookings.
- **Push** fan-out to every device of the users who can see the profile; bad tokens are dropped automatically.
- **State** is a single `state.json` written atomically; mywellness passwords are AES-256-GCM encrypted with a key in `secret.key` (or `WG_SECRET_KEY`).
