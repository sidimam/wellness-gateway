// Package server espone l'API HTTP (/api/v1) e la web UI embedded.
package server

import (
	"context"
	"embed"
	"encoding/json"
	"errors"
	"io/fs"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/sidimam/wellness-gateway/internal/engine"
	"github.com/sidimam/wellness-gateway/internal/model"
	"github.com/sidimam/wellness-gateway/internal/mywellness"
	"github.com/sidimam/wellness-gateway/internal/store"
)

// Version è impostata a build time con -ldflags.
var Version = "dev"

//go:embed static/*
var static embed.FS

// Server è il server HTTP.
type Server struct {
	Store      *store.Store
	Engine     *engine.Engine
	TrustProxy bool
	PushOn     bool
	PublicURL  string

	mu       sync.Mutex
	failures map[string]*loginFail
}

type loginFail struct {
	count int
	until time.Time
}

// New crea il server.
func New(st *store.Store, eng *engine.Engine, trustProxy, pushOn bool, publicURL string) *Server {
	return &Server{Store: st, Engine: eng, TrustProxy: trustProxy, PushOn: pushOn, PublicURL: publicURL, failures: map[string]*loginFail{}}
}

type ctxKey int

const principalKey ctxKey = 1

type principal struct {
	user   model.User
	device model.Device
}

// Handler costruisce il router.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	sub, _ := fs.Sub(static, "static")
	mux.Handle("GET /ui/", http.StripPrefix("/ui/", http.FileServer(http.FS(sub))))
	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Security-Policy", "default-src 'self'; img-src 'self' https: data:; style-src 'self' 'unsafe-inline'; frame-ancestors 'none'")
		w.Header().Set("Cache-Control", "no-cache")
		http.ServeFileFS(w, r, sub, "index.html")
	})
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, 200, map[string]any{"status": "ok", "version": Version})
	})
	mux.HandleFunc("GET /api/v1/setup", s.setupStatus)
	mux.HandleFunc("POST /api/v1/setup", s.setup)
	mux.HandleFunc("POST /api/v1/auth/login", s.login)

	auth := func(h func(http.ResponseWriter, *http.Request, principal)) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			p, ok := s.authenticate(r)
			if !ok {
				w.Header().Set("WWW-Authenticate", `Bearer realm="wellness-gateway"`)
				writeErr(w, 401, "non autenticato")
				return
			}
			h(w, r, p)
		}
	}
	admin := func(h func(http.ResponseWriter, *http.Request, principal)) http.HandlerFunc {
		return auth(func(w http.ResponseWriter, r *http.Request, p principal) {
			if !p.user.IsAdmin {
				writeErr(w, 403, "riservato all'amministratore")
				return
			}
			h(w, r, p)
		})
	}

	mux.HandleFunc("POST /api/v1/auth/logout", auth(s.logout))
	mux.HandleFunc("GET /api/v1/me", auth(s.me))
	mux.HandleFunc("POST /api/v1/me/password", auth(s.changePassword))
	mux.HandleFunc("POST /api/v1/devices/apns", auth(s.registerAPNS))
	mux.HandleFunc("POST /api/v1/devices/test-notification", auth(s.testNotification))
	mux.HandleFunc("GET /api/v1/status", auth(s.status))

	mux.HandleFunc("GET /api/v1/profiles", auth(s.listProfiles))
	mux.HandleFunc("POST /api/v1/profiles", auth(s.addProfile))
	mux.HandleFunc("PUT /api/v1/profiles/{id}", auth(s.updateProfile))
	mux.HandleFunc("DELETE /api/v1/profiles/{id}", auth(s.deleteProfile))
	mux.HandleFunc("POST /api/v1/profiles/{id}/relogin", auth(s.reloginProfile))
	mux.HandleFunc("GET /api/v1/profiles/{id}/classes", auth(s.classes))
	mux.HandleFunc("GET /api/v1/profiles/{id}/bookings", auth(s.bookings))
	mux.HandleFunc("POST /api/v1/profiles/{id}/unbook", auth(s.unbook))
	mux.HandleFunc("POST /api/v1/profiles/{id}/leave-waiting-list", auth(s.leaveWaitingList))

	mux.HandleFunc("GET /api/v1/items", auth(s.listItems))
	mux.HandleFunc("POST /api/v1/items", auth(s.addItem))
	mux.HandleFunc("DELETE /api/v1/items/{id}", auth(s.deleteItem))
	mux.HandleFunc("POST /api/v1/items/{id}/retry", auth(s.retryItem))

	mux.HandleFunc("GET /api/v1/settings", auth(s.getSettings))
	mux.HandleFunc("PUT /api/v1/settings", admin(s.putSettings))
	mux.HandleFunc("GET /api/v1/log", auth(s.getLog))

	mux.HandleFunc("GET /api/v1/users", admin(s.listUsers))
	mux.HandleFunc("POST /api/v1/users", admin(s.addUser))
	mux.HandleFunc("PUT /api/v1/users/{id}", admin(s.updateUser))
	mux.HandleFunc("DELETE /api/v1/users/{id}", admin(s.deleteUser))
	mux.HandleFunc("POST /api/v1/users/{id}/password", admin(s.setUserPassword))
	mux.HandleFunc("GET /api/v1/devices", admin(s.listDevices))
	mux.HandleFunc("DELETE /api/v1/devices/{id}", admin(s.deleteDevice))

	return recoverer(securityHeaders(accessLog(mux)))
}

// ---- middleware

func recoverer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				log.Printf("panic: %v", rec)
				writeErr(w, 500, "errore interno")
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("X-Frame-Options", "DENY")
		h.Set("Referrer-Policy", "no-referrer")
		if strings.HasPrefix(r.URL.Path, "/api/") {
			h.Set("Cache-Control", "no-store")
		}
		next.ServeHTTP(w, r)
	})
}

type statusWriter struct {
	http.ResponseWriter
	code int
}

func (w *statusWriter) WriteHeader(c int) { w.code = c; w.ResponseWriter.WriteHeader(c) }

func accessLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/healthz" {
			next.ServeHTTP(w, r)
			return
		}
		sw := &statusWriter{ResponseWriter: w, code: 200}
		t := time.Now()
		next.ServeHTTP(sw, r)
		if strings.HasPrefix(r.URL.Path, "/api/") {
			log.Printf("%s %s → %d in %s", r.Method, r.URL.Path, sw.code, time.Since(t).Round(time.Millisecond))
		}
	})
}

func (s *Server) clientIP(r *http.Request) string {
	if s.TrustProxy {
		if v := r.Header.Get("CF-Connecting-IP"); v != "" {
			return v
		}
		if v := r.Header.Get("X-Forwarded-For"); v != "" {
			return strings.TrimSpace(strings.Split(v, ",")[0])
		}
		if v := r.Header.Get("X-Real-IP"); v != "" {
			return v
		}
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

// ---- helpers

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func writeErr(w http.ResponseWriter, code int, msg string) {
	writeJSON(w, code, map[string]string{"error": msg})
}

func decode(r *http.Request, v any) error {
	dec := json.NewDecoder(http.MaxBytesReader(nil, r.Body, 1<<20))
	return dec.Decode(v)
}

func (s *Server) authenticate(r *http.Request) (principal, bool) {
	h := r.Header.Get("Authorization")
	if !strings.HasPrefix(h, "Bearer ") {
		return principal{}, false
	}
	tok := strings.TrimSpace(strings.TrimPrefix(h, "Bearer "))
	var p principal
	found := false
	s.Store.Read(func(st *model.State) {
		for _, d := range st.Devices {
			if d.Token == tok {
				for _, u := range st.Users {
					if u.ID == d.UserID {
						p = principal{user: u, device: d}
						found = true
					}
				}
			}
		}
	})
	if found && time.Since(p.device.LastSeen) > time.Minute {
		_ = s.Store.Update(func(st *model.State) error {
			for i := range st.Devices {
				if st.Devices[i].Token == tok {
					st.Devices[i].LastSeen = time.Now()
				}
			}
			return nil
		})
	}
	return p, found
}

func publicUser(u model.User) map[string]any {
	return map[string]any{"id": u.ID, "username": u.Username, "displayName": u.DisplayName, "isAdmin": u.IsAdmin}
}

// ---- setup & auth

func (s *Server) setupStatus(w http.ResponseWriter, r *http.Request) {
	n := 0
	s.Store.Read(func(st *model.State) { n = len(st.Users) })
	writeJSON(w, 200, map[string]any{"needsSetup": n == 0, "version": Version, "push": s.PushOn, "publicUrl": s.PublicURL})
}

func (s *Server) setup(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Username, Password, DisplayName string
		Mywellness                      *profileInput
	}
	if err := decode(r, &in); err != nil || strings.TrimSpace(in.Username) == "" || len(in.Password) < 6 {
		writeErr(w, 400, "nome utente e password (almeno 6 caratteri) obbligatori")
		return
	}
	var created *model.User
	err := s.Store.Update(func(st *model.State) error {
		if len(st.Users) > 0 {
			return errors.New("configurazione iniziale già eseguita")
		}
		hash, salt, err := store.HashPassword(in.Password)
		if err != nil {
			return err
		}
		u := model.User{ID: store.NewID(), Username: strings.ToLower(strings.TrimSpace(in.Username)), DisplayName: in.DisplayName, PasswordHash: hash, Salt: salt, IsAdmin: true, CreatedAt: time.Now()}
		if u.DisplayName == "" {
			u.DisplayName = u.Username
		}
		st.Users = append(st.Users, u)
		created = &u
		return nil
	})
	if err != nil {
		writeErr(w, 409, err.Error())
		return
	}
	tok := s.issueDevice(created.ID, "Web UI")
	out := map[string]any{"token": tok, "user": publicUser(*created)}
	if in.Mywellness != nil && in.Mywellness.Username != "" {
		if in.Mywellness.Label == "" {
			in.Mywellness.Label = created.DisplayName
		}
		if pr, _, err := s.createProfile(r.Context(), *in.Mywellness, *created, created.ID); err != nil {
			out["profileError"] = err.Error()
		} else {
			out["profile"] = pr
		}
	}
	writeJSON(w, 201, out)
}

func (s *Server) issueDevice(userID, name string) string {
	tok := store.NewToken()
	_ = s.Store.Update(func(st *model.State) error {
		st.Devices = append(st.Devices, model.Device{Token: tok, UserID: userID, Name: name, CreatedAt: time.Now(), LastSeen: time.Now()})
		return nil
	})
	return tok
}

func (s *Server) login(w http.ResponseWriter, r *http.Request) {
	ip := s.clientIP(r)
	s.mu.Lock()
	if f, ok := s.failures[ip]; ok && time.Now().Before(f.until) {
		s.mu.Unlock()
		w.Header().Set("Retry-After", "900")
		writeErr(w, 429, "troppi tentativi: riprova tra 15 minuti")
		return
	}
	s.mu.Unlock()
	var in struct{ Username, Password, DeviceName string }
	if err := decode(r, &in); err != nil {
		writeErr(w, 400, "richiesta non valida")
		return
	}
	var user *model.User
	s.Store.Read(func(st *model.State) {
		for _, u := range st.Users {
			if strings.EqualFold(u.Username, strings.TrimSpace(in.Username)) && store.VerifyPassword(in.Password, u.PasswordHash, u.Salt) {
				uu := u
				user = &uu
			}
		}
	})
	if user == nil {
		s.mu.Lock()
		f := s.failures[ip]
		if f == nil {
			f = &loginFail{}
			s.failures[ip] = f
		}
		f.count++
		if f.count >= 5 {
			f.until = time.Now().Add(15 * time.Minute)
			f.count = 0
		}
		s.mu.Unlock()
		writeErr(w, 401, "credenziali non valide")
		return
	}
	s.mu.Lock()
	delete(s.failures, ip)
	s.mu.Unlock()
	name := in.DeviceName
	if name == "" {
		name = "Dispositivo"
	}
	tok := s.issueDevice(user.ID, name)
	writeJSON(w, 200, map[string]any{"token": tok, "user": publicUser(*user), "version": Version, "push": s.PushOn})
}

func (s *Server) logout(w http.ResponseWriter, r *http.Request, p principal) {
	_ = s.Store.Update(func(st *model.State) error {
		kept := st.Devices[:0]
		for _, d := range st.Devices {
			if d.Token != p.device.Token {
				kept = append(kept, d)
			}
		}
		st.Devices = kept
		return nil
	})
	writeJSON(w, 200, map[string]bool{"ok": true})
}

func (s *Server) me(w http.ResponseWriter, r *http.Request, p principal) {
	var mine string
	s.Store.Read(func(st *model.State) {
		for _, pr := range st.Profiles {
			if pr.UserID == p.user.ID {
				mine = pr.ID
			}
		}
	})
	writeJSON(w, 200, map[string]any{"user": publicUser(p.user), "device": p.device, "version": Version, "push": s.PushOn, "myProfileId": mine})
}

func (s *Server) changePassword(w http.ResponseWriter, r *http.Request, p principal) {
	var in struct{ OldPassword, NewPassword string }
	if err := decode(r, &in); err != nil || len(in.NewPassword) < 6 {
		writeErr(w, 400, "nuova password di almeno 6 caratteri")
		return
	}
	if !store.VerifyPassword(in.OldPassword, p.user.PasswordHash, p.user.Salt) {
		writeErr(w, 403, "password attuale errata")
		return
	}
	hash, salt, _ := store.HashPassword(in.NewPassword)
	_ = s.Store.Update(func(st *model.State) error {
		for i := range st.Users {
			if st.Users[i].ID == p.user.ID {
				st.Users[i].PasswordHash, st.Users[i].Salt = hash, salt
			}
		}
		return nil
	})
	writeJSON(w, 200, map[string]bool{"ok": true})
}

func (s *Server) registerAPNS(w http.ResponseWriter, r *http.Request, p principal) {
	var in struct{ Token string }
	if err := decode(r, &in); err != nil || in.Token == "" {
		writeErr(w, 400, "token mancante")
		return
	}
	_ = s.Store.Update(func(st *model.State) error {
		for i := range st.Devices {
			if st.Devices[i].Token == p.device.Token {
				st.Devices[i].APNSToken = in.Token
			}
		}
		return nil
	})
	writeJSON(w, 200, map[string]bool{"ok": true, "push": s.PushOn})
}

func (s *Server) testNotification(w http.ResponseWriter, r *http.Request, p principal) {
	n := s.Engine.TestNotification(p.user.ID)
	writeJSON(w, 200, map[string]any{"sent": n, "push": s.PushOn})
}

func (s *Server) status(w http.ResponseWriter, r *http.Request, p principal) {
	var profiles, items, active int
	s.Store.Read(func(st *model.State) {
		for _, pr := range st.Profiles {
			if engine.Visible(pr, p.user.ID, p.user.IsAdmin) {
				profiles++
			}
		}
		for _, it := range st.Items {
			items++
			if !it.State.Terminal() {
				active++
			}
		}
	})
	writeJSON(w, 200, map[string]any{"version": Version, "startedAt": s.Engine.Started, "nextWake": s.Engine.NextWake,
		"profiles": profiles, "items": items, "activeItems": active, "push": s.PushOn, "publicUrl": s.PublicURL})
}

// ---- profili

func (s *Server) visibleProfile(p principal, id string) (model.Profile, bool) {
	var out model.Profile
	ok := false
	s.Store.Read(func(st *model.State) {
		for _, pr := range st.Profiles {
			if pr.ID == id && engine.Visible(pr, p.user.ID, p.user.IsAdmin) {
				out, ok = pr, true
			}
		}
	})
	return out, ok
}

type limitInfo struct {
	Pattern string `json:"pattern"`
	Active  int    `json:"active"`
	Max     int    `json:"max"`
}

type profileOut struct {
	model.Profile
	Limits []limitInfo `json:"limits"`
}

func (s *Server) profileOut(pr model.Profile, settings model.Settings) profileOut {
	o := profileOut{Profile: pr, Limits: []limitInfo{}}
	for _, r := range settings.OpenRules {
		pat := strings.TrimSpace(r.Pattern)
		if r.MaxBookings > 0 && pat != "" && pat != "*" {
			o.Limits = append(o.Limits, limitInfo{Pattern: pat, Active: s.Engine.RuleCount(pr.ID, r), Max: r.MaxBookings})
		}
	}
	return o
}

func (s *Server) listProfiles(w http.ResponseWriter, r *http.Request, p principal) {
	var profiles []model.Profile
	var settings model.Settings
	s.Store.Read(func(st *model.State) {
		settings = st.Settings
		for _, pr := range st.Profiles {
			if engine.Visible(pr, p.user.ID, p.user.IsAdmin) {
				profiles = append(profiles, pr)
			}
		}
	})
	out := []profileOut{}
	for _, pr := range profiles {
		out = append(out, s.profileOut(pr, settings))
	}
	writeJSON(w, 200, out)
}

type profileInput struct {
	Label, Username, Password, FacilityURL string
	MaxBookings                            int
	Private                                bool
	Mine                                   bool   // il profilo appartiene all'utente chiamante
	UserID                                 string // (admin) il profilo appartiene a questo utente
}

// createProfile verifica il login mywellness e salva il profilo. owner = utente a cui appartiene (può essere vuoto).
func (s *Server) createProfile(ctx context.Context, in profileInput, by model.User, owner string) (model.Profile, int, error) {
	if in.Username == "" || in.Password == "" {
		return model.Profile{}, 400, errors.New("email e password mywellness obbligatorie")
	}
	if in.FacilityURL == "" {
		in.FacilityURL = "wellnesstown"
	}
	if in.MaxBookings <= 0 {
		in.MaxBookings = 5
	}
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	mw := mywellness.New()
	fac, err := mw.FacilityDetail(ctx, strings.ToLower(strings.TrimSpace(in.FacilityURL)))
	if err != nil {
		return model.Profile{}, 400, err
	}
	res, err := mw.Login(ctx, strings.TrimSpace(in.Username), in.Password)
	if err != nil {
		return model.Profile{}, 400, errors.New("login mywellness: " + err.Error())
	}
	enc, err := s.Store.Encrypt(in.Password)
	if err != nil {
		return model.Profile{}, 500, err
	}
	now := time.Now()
	pr := model.Profile{ID: store.NewID(), Label: in.Label, Username: strings.TrimSpace(in.Username), PasswordEnc: enc,
		Token: res.Session.Token, MWUserID: res.Session.UserID, DisplayName: res.DisplayName, UserID: owner,
		FirstName: res.FirstName, LastName: res.LastName, NickName: res.NickName, Email: res.Email, PictureURL: res.PictureURL, ThumbURL: res.ThumbURL,
		FacilityURL: fac.URL, FacilityID: fac.ID, FacilityName: fac.Name, MaxBookings: in.MaxBookings, LastLoginAt: &now}
	if pr.Label == "" {
		pr.Label = res.DisplayName
	}
	if in.Private && owner != "" {
		pr.OwnerUserIDs = []string{owner}
	} else if in.Private {
		pr.OwnerUserIDs = []string{by.ID}
	}
	err = s.Store.Update(func(st *model.State) error {
		for _, x := range st.Profiles {
			if strings.EqualFold(x.Username, pr.Username) && x.FacilityID == pr.FacilityID {
				return errors.New("profilo già presente")
			}
		}
		st.Profiles = append(st.Profiles, pr)
		return nil
	})
	if err != nil {
		return model.Profile{}, 409, err
	}
	s.Store.AddLog("success", pr.ID, "profilo aggiunto: "+pr.Label+" ("+pr.Username+") da "+by.Username)
	go func() { _, _ = s.Engine.Calendar(context.Background(), pr.ID, true) }()
	return pr, 201, nil
}

func (s *Server) addProfile(w http.ResponseWriter, r *http.Request, p principal) {
	var in profileInput
	if err := decode(r, &in); err != nil {
		writeErr(w, 400, "richiesta non valida")
		return
	}
	owner := ""
	if in.Mine {
		owner = p.user.ID
	} else if in.UserID != "" && p.user.IsAdmin {
		owner = in.UserID
	}
	pr, code, err := s.createProfile(r.Context(), in, p.user, owner)
	if err != nil {
		writeErr(w, code, err.Error())
		return
	}
	writeJSON(w, code, pr)
}

func (s *Server) updateProfile(w http.ResponseWriter, r *http.Request, p principal) {
	id := r.PathValue("id")
	if _, ok := s.visibleProfile(p, id); !ok {
		writeErr(w, 404, "profilo non trovato")
		return
	}
	var in struct {
		Label       *string
		Password    *string
		MaxBookings *int
		Private     *bool
		UserID      *string
	}
	if err := decode(r, &in); err != nil {
		writeErr(w, 400, "richiesta non valida")
		return
	}
	var out model.Profile
	err := s.Store.Update(func(st *model.State) error {
		for i := range st.Profiles {
			pr := &st.Profiles[i]
			if pr.ID != id {
				continue
			}
			if in.Label != nil {
				pr.Label = *in.Label
			}
			if in.UserID != nil && (p.user.IsAdmin || *in.UserID == p.user.ID) {
				pr.UserID = *in.UserID
			}
			if in.MaxBookings != nil && *in.MaxBookings > 0 {
				pr.MaxBookings = *in.MaxBookings
			}
			if in.Private != nil {
				if *in.Private {
					pr.OwnerUserIDs = []string{p.user.ID}
				} else {
					pr.OwnerUserIDs = nil
				}
			}
			if in.Password != nil && *in.Password != "" {
				enc, err := s.Store.Encrypt(*in.Password)
				if err != nil {
					return err
				}
				pr.PasswordEnc, pr.Token, pr.MWUserID = enc, "", ""
			}
			out = *pr
		}
		return nil
	})
	if err != nil {
		writeErr(w, 500, err.Error())
		return
	}
	if in.Password != nil {
		go func() { _, _ = s.Engine.Session(context.Background(), id, true) }()
	}
	writeJSON(w, 200, out)
}

func (s *Server) deleteProfile(w http.ResponseWriter, r *http.Request, p principal) {
	id := r.PathValue("id")
	pr, ok := s.visibleProfile(p, id)
	if !ok {
		writeErr(w, 404, "profilo non trovato")
		return
	}
	_ = s.Store.Update(func(st *model.State) error {
		kept := st.Profiles[:0]
		for _, x := range st.Profiles {
			if x.ID != id {
				kept = append(kept, x)
			}
		}
		st.Profiles = kept
		items := st.Items[:0]
		for _, it := range st.Items {
			if it.ProfileID != id {
				items = append(items, it)
			}
		}
		st.Items = items
		return nil
	})
	s.Store.AddLog("warn", "", "profilo rimosso: "+pr.Label+" da "+p.user.Username)
	writeJSON(w, 200, map[string]bool{"ok": true})
}

func (s *Server) reloginProfile(w http.ResponseWriter, r *http.Request, p principal) {
	id := r.PathValue("id")
	if _, ok := s.visibleProfile(p, id); !ok {
		writeErr(w, 404, "profilo non trovato")
		return
	}
	if _, err := s.Engine.Session(r.Context(), id, true); err != nil {
		writeErr(w, 400, err.Error())
		return
	}
	pr, _ := s.visibleProfile(p, id)
	writeJSON(w, 200, pr)
}

type classOut struct {
	mywellness.ClassEvent
	Start   time.Time   `json:"start"`
	End     time.Time   `json:"end"`
	OpensOn *time.Time  `json:"opensOn,omitempty"`
	Tracked *model.Item `json:"tracked,omitempty"`
}

func (s *Server) classes(w http.ResponseWriter, r *http.Request, p principal) {
	id := r.PathValue("id")
	if _, ok := s.visibleProfile(p, id); !ok {
		writeErr(w, 404, "profilo non trovato")
		return
	}
	force := r.URL.Query().Get("refresh") == "1"
	events, err := s.Engine.Calendar(r.Context(), id, force)
	if err != nil {
		writeErr(w, 502, err.Error())
		return
	}
	items := map[string]model.Item{}
	s.Store.Read(func(st *model.State) {
		for _, it := range st.Items {
			if it.ProfileID == id {
				items[it.ClassID+"|"+strconv.Itoa(it.PartitionDate)] = it
			}
		}
	})
	filter := strings.ToLower(r.URL.Query().Get("q"))
	out := make([]classOut, 0, len(events))
	for _, ev := range events {
		if filter != "" && !strings.Contains(strings.ToLower(ev.Name+" "+ev.AssignedTo+" "+ev.Room), filter) {
			continue
		}
		c := classOut{ClassEvent: ev, Start: ev.Start(), End: ev.End(), OpensOn: ev.OpensOn()}
		if it, ok := items[ev.Key()]; ok {
			itc := it
			c.Tracked = &itc
		}
		out = append(out, c)
	}
	writeJSON(w, 200, out)
}

// bookings: tutte le lezioni future a cui il profilo è iscritto (fatte dal gateway o da mywellness).
func (s *Server) bookings(w http.ResponseWriter, r *http.Request, p principal) {
	id := r.PathValue("id")
	if _, ok := s.visibleProfile(p, id); !ok {
		writeErr(w, 404, "profilo non trovato")
		return
	}
	events, err := s.Engine.Calendar(r.Context(), id, r.URL.Query().Get("refresh") == "1")
	if err != nil {
		writeErr(w, 502, err.Error())
		return
	}
	items := map[string]model.Item{}
	s.Store.Read(func(st *model.State) {
		for _, it := range st.Items {
			if it.ProfileID == id {
				items[it.ClassID+"|"+strconv.Itoa(it.PartitionDate)] = it
			}
		}
	})
	out := []classOut{}
	now := time.Now()
	for _, ev := range events {
		if !ev.IsParticipant || !ev.Start().After(now) {
			continue
		}
		c := classOut{ClassEvent: ev, Start: ev.Start(), End: ev.End(), OpensOn: ev.OpensOn()}
		if it, ok := items[ev.Key()]; ok {
			itc := it
			c.Tracked = &itc
		}
		out = append(out, c)
	}
	writeJSON(w, 200, out)
}

func (s *Server) unbook(w http.ResponseWriter, r *http.Request, p principal) {
	id := r.PathValue("id")
	if _, ok := s.visibleProfile(p, id); !ok {
		writeErr(w, 404, "profilo non trovato")
		return
	}
	var in struct {
		ClassID       string
		PartitionDate int
	}
	if err := decode(r, &in); err != nil || in.ClassID == "" || in.PartitionDate == 0 {
		writeErr(w, 400, "classId e partitionDate obbligatori")
		return
	}
	if err := s.Engine.Unbook(r.Context(), id, in.ClassID, in.PartitionDate, p.user.Username); err != nil {
		writeErr(w, 502, err.Error())
		return
	}
	writeJSON(w, 200, map[string]bool{"ok": true})
}

func (s *Server) leaveWaitingList(w http.ResponseWriter, r *http.Request, p principal) {
	id := r.PathValue("id")
	if _, ok := s.visibleProfile(p, id); !ok {
		writeErr(w, 404, "profilo non trovato")
		return
	}
	var in struct {
		ClassID       string
		PartitionDate int
		RemoveItem    bool
	}
	if err := decode(r, &in); err != nil || in.ClassID == "" || in.PartitionDate == 0 {
		writeErr(w, 400, "classId e partitionDate obbligatori")
		return
	}
	if err := s.Engine.LeaveWaitingList(r.Context(), id, in.ClassID, in.PartitionDate, p.user.Username); err != nil {
		writeErr(w, 502, err.Error())
		return
	}
	if in.RemoveItem {
		s.Engine.RemoveItem(id+"|"+in.ClassID+"|"+strconv.Itoa(in.PartitionDate), false)
	}
	writeJSON(w, 200, map[string]bool{"ok": true})
}

// ---- items

func (s *Server) listItems(w http.ResponseWriter, r *http.Request, p principal) {
	pid := r.URL.Query().Get("profile")
	out := []model.Item{}
	s.Store.Read(func(st *model.State) {
		vis := map[string]bool{}
		for _, pr := range st.Profiles {
			vis[pr.ID] = engine.Visible(pr, p.user.ID, p.user.IsAdmin)
		}
		for _, it := range st.Items {
			if vis[it.ProfileID] && (pid == "" || it.ProfileID == pid) {
				out = append(out, it)
			}
		}
	})
	writeJSON(w, 200, out)
}

func (s *Server) addItem(w http.ResponseWriter, r *http.Request, p principal) {
	var in struct {
		ProfileID, ClassID string
		PartitionDate      int
		Recurring          bool
	}
	if err := decode(r, &in); err != nil || in.ProfileID == "" || in.ClassID == "" || in.PartitionDate == 0 {
		writeErr(w, 400, "profileId, classId e partitionDate obbligatori")
		return
	}
	if _, ok := s.visibleProfile(p, in.ProfileID); !ok {
		writeErr(w, 404, "profilo non trovato")
		return
	}
	it, err := s.Engine.AddItem(r.Context(), in.ProfileID, in.ClassID, in.PartitionDate, in.Recurring, p.user.Username)
	if err != nil {
		writeErr(w, 400, err.Error())
		return
	}
	writeJSON(w, 201, it)
}

func (s *Server) itemVisible(p principal, id string) (model.Item, bool) {
	var out model.Item
	ok := false
	s.Store.Read(func(st *model.State) {
		for _, it := range st.Items {
			if it.ID != id {
				continue
			}
			for _, pr := range st.Profiles {
				if pr.ID == it.ProfileID && engine.Visible(pr, p.user.ID, p.user.IsAdmin) {
					out, ok = it, true
				}
			}
		}
	})
	return out, ok
}

func (s *Server) deleteItem(w http.ResponseWriter, r *http.Request, p principal) {
	id := r.PathValue("id")
	if _, ok := s.itemVisible(p, id); !ok {
		writeErr(w, 404, "lezione non trovata")
		return
	}
	s.Engine.RemoveItem(id, r.URL.Query().Get("rule") == "1")
	writeJSON(w, 200, map[string]bool{"ok": true})
}

func (s *Server) retryItem(w http.ResponseWriter, r *http.Request, p principal) {
	id := r.PathValue("id")
	if _, ok := s.itemVisible(p, id); !ok {
		writeErr(w, 404, "lezione non trovata")
		return
	}
	s.Engine.Retry(id)
	it, _ := s.itemVisible(p, id)
	writeJSON(w, 200, it)
}

// ---- impostazioni, log, utenti, dispositivi

func (s *Server) getSettings(w http.ResponseWriter, r *http.Request, p principal) {
	var out model.Settings
	s.Store.Read(func(st *model.State) { out = st.Settings })
	writeJSON(w, 200, out)
}

func (s *Server) putSettings(w http.ResponseWriter, r *http.Request, p principal) {
	var in model.Settings
	if err := decode(r, &in); err != nil {
		writeErr(w, 400, "richiesta non valida")
		return
	}
	engine.Normalize(&in)
	_ = s.Store.Update(func(st *model.State) error { st.Settings = in; return nil })
	s.Engine.Kick()
	writeJSON(w, 200, in)
}

func (s *Server) getLog(w http.ResponseWriter, r *http.Request, p principal) {
	pid := r.URL.Query().Get("profile")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 || limit > 1000 {
		limit = 200
	}
	out := []model.LogLine{}
	s.Store.Read(func(st *model.State) {
		vis := map[string]bool{}
		for _, pr := range st.Profiles {
			vis[pr.ID] = engine.Visible(pr, p.user.ID, p.user.IsAdmin)
		}
		for _, l := range st.Log {
			if l.ProfileID != "" && !vis[l.ProfileID] {
				continue
			}
			if pid != "" && l.ProfileID != pid {
				continue
			}
			out = append(out, l)
			if len(out) >= limit {
				break
			}
		}
	})
	writeJSON(w, 200, out)
}

func (s *Server) listUsers(w http.ResponseWriter, r *http.Request, p principal) {
	out := []map[string]any{}
	s.Store.Read(func(st *model.State) {
		for _, u := range st.Users {
			m := publicUser(u)
			for _, pr := range st.Profiles {
				if pr.UserID == u.ID {
					m["profileId"] = pr.ID
					m["profileLabel"] = pr.Label
				}
			}
			out = append(out, m)
		}
	})
	writeJSON(w, 200, out)
}

func (s *Server) addUser(w http.ResponseWriter, r *http.Request, p principal) {
	var in struct {
		Username, Password, DisplayName string
		IsAdmin                         bool
		Mywellness                      *profileInput
	}
	if err := decode(r, &in); err != nil || strings.TrimSpace(in.Username) == "" || len(in.Password) < 6 {
		writeErr(w, 400, "nome utente e password (almeno 6 caratteri) obbligatori")
		return
	}
	hash, salt, _ := store.HashPassword(in.Password)
	u := model.User{ID: store.NewID(), Username: strings.ToLower(strings.TrimSpace(in.Username)), DisplayName: in.DisplayName, PasswordHash: hash, Salt: salt, IsAdmin: in.IsAdmin, CreatedAt: time.Now()}
	if u.DisplayName == "" {
		u.DisplayName = u.Username
	}
	err := s.Store.Update(func(st *model.State) error {
		for _, x := range st.Users {
			if x.Username == u.Username {
				return errors.New("nome utente già in uso")
			}
		}
		st.Users = append(st.Users, u)
		return nil
	})
	if err != nil {
		writeErr(w, 409, err.Error())
		return
	}
	out := publicUser(u)
	if in.Mywellness != nil && in.Mywellness.Username != "" {
		if in.Mywellness.Label == "" {
			in.Mywellness.Label = u.DisplayName
		}
		if pr, _, err := s.createProfile(r.Context(), *in.Mywellness, p.user, u.ID); err != nil {
			out["profileError"] = err.Error()
		} else {
			out["profile"] = pr
		}
	}
	writeJSON(w, 201, out)
}

func (s *Server) updateUser(w http.ResponseWriter, r *http.Request, p principal) {
	id := r.PathValue("id")
	var in struct {
		DisplayName *string
		IsAdmin     *bool
		Password    *string
		Username    *string
	}
	if err := decode(r, &in); err != nil {
		writeErr(w, 400, "richiesta non valida")
		return
	}
	if in.Password != nil && len(*in.Password) < 6 {
		writeErr(w, 400, "password di almeno 6 caratteri")
		return
	}
	var out map[string]any
	err := s.Store.Update(func(st *model.State) error {
		for i := range st.Users {
			u := &st.Users[i]
			if u.ID != id {
				continue
			}
			if in.Username != nil && strings.TrimSpace(*in.Username) != "" {
				nu := strings.ToLower(strings.TrimSpace(*in.Username))
				for _, x := range st.Users {
					if x.ID != id && x.Username == nu {
						return errors.New("nome utente già in uso")
					}
				}
				u.Username = nu
			}
			if in.DisplayName != nil && *in.DisplayName != "" {
				u.DisplayName = *in.DisplayName
			}
			if in.IsAdmin != nil {
				if !*in.IsAdmin {
					admins := 0
					for _, x := range st.Users {
						if x.IsAdmin && x.ID != id {
							admins++
						}
					}
					if admins == 0 {
						return errors.New("deve restare almeno un amministratore")
					}
				}
				u.IsAdmin = *in.IsAdmin
			}
			if in.Password != nil {
				hash, salt, err := store.HashPassword(*in.Password)
				if err != nil {
					return err
				}
				u.PasswordHash, u.Salt = hash, salt
			}
			out = publicUser(*u)
		}
		if out == nil {
			return errors.New("utente non trovato")
		}
		return nil
	})
	if err != nil {
		writeErr(w, 409, err.Error())
		return
	}
	writeJSON(w, 200, out)
}

func (s *Server) deleteUser(w http.ResponseWriter, r *http.Request, p principal) {
	id := r.PathValue("id")
	if id == p.user.ID {
		writeErr(w, 400, "non puoi eliminare l'utente con cui sei collegato: entra con un altro amministratore")
		return
	}
	err := s.Store.Update(func(st *model.State) error {
		admins := 0
		for _, u := range st.Users {
			if u.IsAdmin && u.ID != id {
				admins++
			}
		}
		if admins == 0 {
			return errors.New("deve restare almeno un amministratore")
		}
		users := st.Users[:0]
		for _, u := range st.Users {
			if u.ID != id {
				users = append(users, u)
			}
		}
		st.Users = users
		for i := range st.Profiles {
			if st.Profiles[i].UserID == id {
				st.Profiles[i].UserID = "" // il profilo mywellness resta, senza utente
			}
		}
		devs := st.Devices[:0]
		for _, d := range st.Devices {
			if d.UserID != id {
				devs = append(devs, d)
			}
		}
		st.Devices = devs
		return nil
	})
	if err != nil {
		writeErr(w, 409, err.Error())
		return
	}
	writeJSON(w, 200, map[string]bool{"ok": true})
}

func (s *Server) setUserPassword(w http.ResponseWriter, r *http.Request, p principal) {
	var in struct{ Password string }
	if err := decode(r, &in); err != nil || len(in.Password) < 6 {
		writeErr(w, 400, "password di almeno 6 caratteri")
		return
	}
	hash, salt, _ := store.HashPassword(in.Password)
	id := r.PathValue("id")
	_ = s.Store.Update(func(st *model.State) error {
		for i := range st.Users {
			if st.Users[i].ID == id {
				st.Users[i].PasswordHash, st.Users[i].Salt = hash, salt
			}
		}
		return nil
	})
	writeJSON(w, 200, map[string]bool{"ok": true})
}

func (s *Server) listDevices(w http.ResponseWriter, r *http.Request, p principal) {
	out := []map[string]any{}
	s.Store.Read(func(st *model.State) {
		for _, d := range st.Devices {
			out = append(out, map[string]any{"id": d.Token[:12], "userId": d.UserID, "name": d.Name, "push": d.APNSToken != "", "createdAt": d.CreatedAt, "lastSeen": d.LastSeen})
		}
	})
	writeJSON(w, 200, out)
}

func (s *Server) deleteDevice(w http.ResponseWriter, r *http.Request, p principal) {
	id := r.PathValue("id")
	_ = s.Store.Update(func(st *model.State) error {
		kept := st.Devices[:0]
		for _, d := range st.Devices {
			if !strings.HasPrefix(d.Token, id) {
				kept = append(kept, d)
			}
		}
		st.Devices = kept
		return nil
	})
	writeJSON(w, 200, map[string]bool{"ok": true})
}
