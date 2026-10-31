#!/usr/bin/env python3
"""Server di prova della web UI: serve internal/server/static e risponde alle API con dati finti.
Serve ai test di non regressione nel browser senza toccare il gateway vero (niente mywellness, niente prenotazioni).
Uso: python3 scripts/mockui.py [porta]   → http://localhost:8599 (login con qualunque utente/password)."""
import json, sys, os, datetime as dt
from http.server import ThreadingHTTPServer, SimpleHTTPRequestHandler
from urllib.parse import urlparse

STATIC = os.path.join(os.path.dirname(os.path.abspath(__file__)), "..", "internal", "server", "static")
PORT = int(sys.argv[1]) if len(sys.argv) > 1 else 8599
now = dt.datetime.now().astimezone()
now = now.replace(minute=0, second=0, microsecond=0)
def iso(h): return (now + dt.timedelta(hours=h)).isoformat(timespec="seconds")
def local(h): return (now + dt.timedelta(hours=h)).strftime("%Y-%m-%dT%H:%M:%S")   # come mywellness (ora locale)
PHOTO = "data:image/svg+xml;utf8," + "<svg xmlns='http://www.w3.org/2000/svg' width='400' height='400'><rect width='400' height='400' fill='%23e83e8c'/><circle cx='200' cy='160' r='90' fill='%23fff'/></svg>"
MEMBERS = {
    "m1": {"id": "m1", "nickName": "dani", "firstName": "Daniela", "lastName": "Prova", "pictureUrl": PHOTO, "private": False, "firstSeen": iso(-500), "lastSeen": iso(-1), "seenCount": 5, "classes": ["Group Reformer", "PILATES"]},
    "m2": {"id": "m2", "nickName": "giusi", "firstName": "Giusi", "lastName": "Esempio", "pictureUrl": "", "private": False, "firstSeen": iso(-300), "lastSeen": iso(-1), "seenCount": 2, "classes": ["PILATES"]},
    "m3": {"id": "m3", "nickName": "*******", "private": True, "firstSeen": iso(-30), "lastSeen": iso(-1), "seenCount": 1},
}
CLASSES = [
    {"id": "c1", "partitionDate": 20261031, "name": "Group Reformer", "start": iso(-3), "end": iso(-2), "startDate": local(-3), "endDate": local(-2), "numberOfParticipants": 3, "maxParticipants": 7, "availablePlaces": 4, "room": "Studio Reformer", "assignedTo": "ISTRUTTORE"},
    {"id": "c2", "partitionDate": 20261101, "name": "PILATES", "start": iso(20), "end": iso(21), "startDate": local(20), "endDate": local(21), "numberOfParticipants": 2, "maxParticipants": 12, "availablePlaces": 10, "tracked": {"state": "pending"}},
]
ATT = {"m1": [("c2", 20261101, "PILATES", iso(20)), ("c1", 20261031, "Group Reformer", iso(-3))], "m2": [("c2", 20261101, "PILATES", iso(20))], "m3": [("c1", 20261031, "Group Reformer", iso(-3))]}
def participants(cid):
    found = [m for k, m in MEMBERS.items() if any(a[0] == cid for a in ATT[k])]
    return found or list(MEMBERS.values())   # lezioni del calendario vero: tutti i soci di prova
RULES = [{"id": "r", "pattern": "Reformer", "daysBefore": 3, "hour": 5, "minute": 1, "alsoAt": "07:01", "maxBookings": 3, "maxPerWeek": 2},
         {"id": "default", "pattern": "*", "daysBefore": 7, "hour": 5, "minute": 1, "maxBookings": 0}]

def route(method, path):
    p = path
    if p == "/api/v1/setup": return {"version": "mock", "needsSetup": False}
    USER = {"id": "u1", "username": "admin", "displayName": "Prova", "isAdmin": True}
    if p in ("/healthz", "/api/v1/healthz"): return {"status": "ok", "version": "mock"}
    if p == "/api/v1/auth/login": return {"token": "mock", "user": USER, "version": "mock", "push": False}
    if p == "/api/v1/me": return {"user": USER, "version": "mock", "push": False, "myProfileId": "p1"}
    if p == "/api/v1/profiles": return [{"id": "p1", "userId": "u1", "label": "Daniela", "username": "prova@example.com", "facilityUrl": "https://example.com", "facilityId": "f1", "facilityName": "Wellness Town", "activeBookings": 1, "maxBookings": 5, "limits": [{"pattern": "Reformer", "active": 1, "max": 3, "perWeek": 2, "thisWeek": 1}]}]
    if p == "/api/v1/status": return {"version": "mock", "startedAt": iso(-5), "nextWake": iso(0), "profiles": 1, "items": 2, "activeItems": 1, "push": False, "telegram": True}
    if p.startswith("/api/v1/log"): return []
    if p.startswith("/api/v1/items"):
        return [{"id": "p1|c9|1", "profileId": "p1", "classId": "c9", "partitionDate": 20261105, "name": "Group Reformer", "start": iso(80), "end": iso(81), "state": "pending", "fireAt": iso(10), "fireTimes": [iso(10), iso(12)], "recurring": True, "attempts": 0, "lastMessage": ""},
                {"id": "p1|c1|20261031", "profileId": "p1", "classId": "c1", "partitionDate": 20261031, "name": "Group Reformer", "start": iso(-3), "end": iso(-2), "state": "booked", "recurring": True, "attempts": 1, "lastMessage": "Prenotata"}]
    if p.endswith("/bookings"): return []
    if "/classes/" in p and p.endswith("/participants"):
        cid = p.split("/classes/")[1].split("/")[0]
        return {"participants": participants(cid), "observedAt": iso(-1), "liveError": ""}
    if p.endswith("/classes"): return CLASSES
    if p.startswith("/api/v1/people/members/"):
        mid = p.rsplit("/", 1)[1]
        return {"member": MEMBERS[mid], "attendance": [{"classId": a[0], "partitionDate": a[1], "name": a[2], "start": a[3]} for a in ATT[mid]], "leaderboard": []}
    if p.startswith("/api/v1/people/members"): return list(MEMBERS.values())
    if p.startswith("/api/v1/people/leaderboard"): return {"entries": [], "periods": [], "period": ""}
    if p.startswith("/api/v1/people/stats"): return {"members": 3, "classes": 2, "attendance": 4}
    if p == "/api/v1/settings": return {"followServerOpenTime": True, "openRules": RULES, "leadMilliseconds": 300, "burstSeconds": 120, "pollSeconds": 15, "nearPollSeconds": 3, "nearHours": 4, "daysAhead": 14, "custom": False}
    if p == "/api/v1/users": return []
    if p == "/api/v1/telegram/test": return {"ok": True}
    return {}

class H(SimpleHTTPRequestHandler):
    def __init__(self, *a, **k): super().__init__(*a, directory=STATIC, **k)
    def log_message(self, *a): pass
    def _api(self, method):
        body = json.dumps(route(method, urlparse(self.path).path)).encode()
        self.send_response(200); self.send_header("Content-Type", "application/json"); self.send_header("Content-Length", str(len(body))); self.end_headers(); self.wfile.write(body)
    def do_GET(self):
        u = urlparse(self.path)
        if u.path.startswith("/api/") or u.path == "/healthz": return self._api("GET")
        if u.path.startswith("/ui/"): self.path = u.path[3:]
        elif u.path == "/": self.path = "/index.html"
        return super().do_GET()
    def end_headers(self):
        self.send_header("Cache-Control", "no-store"); super().end_headers()
    def do_POST(self): self._api("POST")
    def do_PUT(self): self._api("PUT")
    def do_DELETE(self): self._api("DELETE")

print(f"mock web UI su http://localhost:{PORT}")
ThreadingHTTPServer(("127.0.0.1", PORT), H).serve_forever()
