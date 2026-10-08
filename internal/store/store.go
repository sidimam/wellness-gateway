// Package store persiste lo stato in un file JSON (atomico) e cifra i segreti con AES-GCM.
package store

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/sidimam/wellness-gateway/internal/model"
)

// Store è thread-safe: ogni accesso passa da Read/Update.
type Store struct {
	mu    sync.RWMutex
	path  string
	key   []byte
	state model.State
}

// Open carica (o crea) lo stato nella cartella dati. La chiave di cifratura dei segreti è letta da
// WG_SECRET_KEY oppure generata e salvata in <dir>/secret.key.
func Open(dir string) (*Store, error) {
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return nil, err
	}
	key, err := loadOrCreateKey(dir)
	if err != nil {
		return nil, err
	}
	s := &Store{path: filepath.Join(dir, "state.json"), key: key}
	data, err := os.ReadFile(s.path)
	switch {
	case errors.Is(err, os.ErrNotExist):
		s.state = model.State{Settings: model.DefaultSettings()}
		return s, s.flush()
	case err != nil:
		return nil, err
	}
	if err := s.unmarshal(data); err != nil {
		return nil, err
	}
	if len(s.state.Settings.OpenRules) == 0 {
		s.state.Settings = model.DefaultSettings()
	}
	// migrazione v0.1.12: i dispositivi dell'app vedono solo il proprio profilo
	changed := false
	for i := range s.state.Devices {
		if s.state.Devices[i].Name != "Web UI" && !s.state.Devices[i].SelfOnly {
			s.state.Devices[i].SelfOnly = true
			changed = true
		}
	}
	if changed {
		_ = s.flush()
	}
	// migrazione v0.1.15: l'osservazione legge il calendario PUBBLICO (senza token), quindi può essere
	// più frequente senza esporre l'account: 60 s (vecchio default) → 15 s, e 3 s nelle ultime 4 ore.
	if s.state.Settings.PollSeconds == 60 || s.state.Settings.PollSeconds < 5 {
		s.state.Settings.PollSeconds = 15
		_ = s.flush()
	}
	if s.state.Settings.NearPollSeconds <= 0 {
		s.state.Settings.NearPollSeconds = 3
		_ = s.flush()
	}
	if s.state.Settings.NearHours <= 0 {
		s.state.Settings.NearHours = 4
		_ = s.flush()
	}
	return s, nil
}

func loadOrCreateKey(dir string) ([]byte, error) {
	if env := os.Getenv("WG_SECRET_KEY"); env != "" {
		sum := sha256.Sum256([]byte(env))
		return sum[:], nil
	}
	p := filepath.Join(dir, "secret.key")
	if b, err := os.ReadFile(p); err == nil {
		k, err := hex.DecodeString(string(b))
		if err == nil && len(k) == 32 {
			return k, nil
		}
	}
	k := make([]byte, 32)
	if _, err := rand.Read(k); err != nil {
		return nil, err
	}
	return k, os.WriteFile(p, []byte(hex.EncodeToString(k)), 0o600)
}

// Read esegue fn con accesso in sola lettura allo stato.
func (s *Store) Read(fn func(st *model.State)) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	fn(&s.state)
}

// Update esegue fn con accesso in scrittura e salva su disco.
func (s *Store) Update(fn func(st *model.State) error) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := fn(&s.state); err != nil {
		return err
	}
	return s.flush()
}

func (s *Store) flush() error {
	if len(s.state.Log) > 1000 {
		s.state.Log = s.state.Log[:1000]
	}
	data, err := json.MarshalIndent(s.statePersisted(), "", " ")
	if err != nil {
		return err
	}
	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, s.path)
}

// statePersisted include anche i campi `json:"-"` tramite una copia con alias.
func (s *Store) statePersisted() persisted {
	p := persisted{Settings: s.state.Settings, Items: s.state.Items, Log: s.state.Log}
	for _, u := range s.state.Users {
		p.Users = append(p.Users, userP{User: u, PasswordHash: u.PasswordHash, Salt: u.Salt})
	}
	for _, d := range s.state.Devices {
		p.Devices = append(p.Devices, deviceP{Device: d, Token: d.Token})
	}
	for _, pr := range s.state.Profiles {
		p.Profiles = append(p.Profiles, profileP{Profile: pr, PasswordEnc: pr.PasswordEnc, Token: pr.Token, MWUserID: pr.MWUserID})
	}
	return p
}

type userP struct {
	model.User
	PasswordHash string `json:"passwordHash"`
	Salt         string `json:"salt"`
}
type deviceP struct {
	model.Device
	Token string `json:"token"`
}
type profileP struct {
	model.Profile
	PasswordEnc string `json:"passwordEnc"`
	Token       string `json:"token"`
	MWUserID    string `json:"mwUserId"`
}
type persisted struct {
	Users    []userP         `json:"users"`
	Devices  []deviceP       `json:"devices"`
	Profiles []profileP      `json:"profiles"`
	Items    []model.Item    `json:"items"`
	Settings model.Settings  `json:"settings"`
	Log      []model.LogLine `json:"log"`
}

// UnmarshalJSON ricompone lo stato dai campi persistiti.
func (s *Store) unmarshal(data []byte) error {
	var p persisted
	if err := json.Unmarshal(data, &p); err != nil {
		return err
	}
	s.state = model.State{Settings: p.Settings, Items: p.Items, Log: p.Log}
	for _, u := range p.Users {
		u.User.PasswordHash, u.User.Salt = u.PasswordHash, u.Salt
		s.state.Users = append(s.state.Users, u.User)
	}
	for _, d := range p.Devices {
		d.Device.Token = d.Token
		s.state.Devices = append(s.state.Devices, d.Device)
	}
	for _, pr := range p.Profiles {
		pr.Profile.PasswordEnc, pr.Profile.Token, pr.Profile.MWUserID = pr.PasswordEnc, pr.Token, pr.MWUserID
		s.state.Profiles = append(s.state.Profiles, pr.Profile)
	}
	return nil
}

// Encrypt cifra un segreto (password mywellness) con AES-256-GCM.
func (s *Store) Encrypt(plain string) (string, error) {
	block, err := aes.NewCipher(s.key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(gcm.Seal(nonce, nonce, []byte(plain), nil)), nil
}

// Decrypt decifra un segreto.
func (s *Store) Decrypt(enc string) (string, error) {
	raw, err := base64.StdEncoding.DecodeString(enc)
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(s.key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	if len(raw) < gcm.NonceSize() {
		return "", errors.New("segreto troppo corto")
	}
	out, err := gcm.Open(nil, raw[:gcm.NonceSize()], raw[gcm.NonceSize():], nil)
	if err != nil {
		return "", err
	}
	return string(out), nil
}

// AddLog aggiunge una riga al registro (in testa).
func (s *Store) AddLog(level, profileID, text string) {
	_ = s.Update(func(st *model.State) error {
		st.Log = append([]model.LogLine{{Time: time.Now(), Level: level, ProfileID: profileID, Text: text}}, st.Log...)
		return nil
	})
}
