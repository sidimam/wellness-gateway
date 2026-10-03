# wellness-gateway

Always-on booking engine for **Technogym mywellness** classes, packaged as a Docker container for Unraid. It books for several family accounts the instant booking opens, keeps watching full classes to take freed spots, syncs with cancellations made in the official app, and pushes Time Sensitive notifications to the **[Wellness Booking](https://github.com/sidimam/wellness-booking)** iOS app.

- [Architecture](Architecture) — profiles, engine, API, app, push
- [Installation](Installation) — Unraid template, Compose, first-run walkthrough
- [Configuration](Configuration) — environment variables, opening rules, limits
- [API-Reference](API-Reference) — REST API used by the iOS app
- [API-mywellness](API-mywellness) — upstream endpoints and fields
- [Security](Security) — credentials at rest, tokens, Cloudflare
- [Operations](Operations) — updates, logs, backup, troubleshooting
- [Changelog](Changelog)
