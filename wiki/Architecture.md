# Architecture

```
iPhone (Wellness Booking) ──HTTPS/Bearer──▶ wellness-gateway (Unraid container) ──HTTPS──▶ mywellness API
        ▲                                        │
        └────────────── APNs push ◀──────────────┘
```

- **Users** are local gateway accounts (Simone, Daniela, Sofia…). Each login from a device issues a long-lived bearer token stored in the iPhone Keychain.
- **Profiles** are mywellness accounts. A profile is either *family* (visible to every user) or *private* (owner only). Only the gateway talks to Technogym; the phone never holds mywellness credentials.
- **Engine** loop (one goroutine): every tick it evaluates all tracked items: `pending` → at `fireAt − lead` becomes `bursting` (Book every 1.5 s for `burstSeconds`) → `booked`, or `waitingList`/`watching` when full → polls the day's schedule every `pollSeconds` (÷2 within 6 h of the class, ÷3 in the last 45 min before the cancellation deadline) and books the moment `availablePlaces > 0`. Items expire 5 minutes before the class starts.
- **Calendar refresh** every 10 minutes per profile (authenticated, so `isParticipant` is known): updates places, marks items booked/cancelled from outside, attaches weekly recurrences, recounts active bookings.
- **Push** fan-out to every device of the users who can see the profile; bad tokens are dropped automatically.
- **State** is a single `state.json` written atomically; mywellness passwords are AES-256-GCM encrypted with a key in `secret.key` (or `WG_SECRET_KEY`).
