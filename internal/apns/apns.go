// Package apns invia notifiche push ad Apple (HTTP/2 + JWT ES256) con la sola libreria standard.
package apns

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"time"
)

// Config della connessione APNs.
type Config struct {
	KeyPath    string // AuthKey_XXXXXXXXXX.p8
	KeyID      string
	TeamID     string
	BundleID   string // topic
	Production bool
}

// Client APNs riutilizzabile (il JWT vale 60 minuti, rinnovato ogni 50).
type Client struct {
	cfg  Config
	key  *ecdsa.PrivateKey
	http *http.Client
	mu   sync.Mutex
	jwt  string
	jwtT time.Time
}

// Enabled indica se la configurazione è completa.
func (c Config) Enabled() bool {
	return c.KeyPath != "" && c.KeyID != "" && c.TeamID != "" && c.BundleID != ""
}

// New carica la chiave .p8.
func New(cfg Config) (*Client, error) {
	raw, err := os.ReadFile(cfg.KeyPath)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(raw)
	if block == nil {
		return nil, errors.New("chiave APNs: PEM non valido")
	}
	k, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	ec, ok := k.(*ecdsa.PrivateKey)
	if !ok {
		return nil, errors.New("chiave APNs: non è una chiave EC")
	}
	return &Client{cfg: cfg, key: ec, http: &http.Client{Timeout: 15 * time.Second}}, nil
}

func (c *Client) token() (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.jwt != "" && time.Since(c.jwtT) < 50*time.Minute {
		return c.jwt, nil
	}
	hdr, _ := json.Marshal(map[string]string{"alg": "ES256", "kid": c.cfg.KeyID, "typ": "JWT"})
	claims, _ := json.Marshal(map[string]any{"iss": c.cfg.TeamID, "iat": time.Now().Unix()})
	enc := base64.RawURLEncoding
	signing := enc.EncodeToString(hdr) + "." + enc.EncodeToString(claims)
	sum := sha256.Sum256([]byte(signing))
	r, s, err := ecdsa.Sign(rand.Reader, c.key, sum[:])
	if err != nil {
		return "", err
	}
	sig := make([]byte, 64)
	r.FillBytes(sig[:32])
	s.FillBytes(sig[32:])
	c.jwt = signing + "." + enc.EncodeToString(sig)
	c.jwtT = time.Now()
	return c.jwt, nil
}

// Notification è una notifica da inviare.
type Notification struct {
	Title    string
	Body     string
	Priority bool // interruption-level time-sensitive
	Thread   string
	Payload  map[string]any
}

// ErrBadDeviceToken indica un token da rimuovere.
var ErrBadDeviceToken = errors.New("token dispositivo non valido")

// Send invia la notifica a un token dispositivo.
func (c *Client) Send(ctx context.Context, deviceToken string, n Notification) error {
	jwt, err := c.token()
	if err != nil {
		return err
	}
	aps := map[string]any{
		"alert": map[string]string{"title": n.Title, "body": n.Body},
		"sound": "default",
	}
	if n.Priority {
		aps["interruption-level"] = "time-sensitive"
		aps["sound"] = map[string]any{"critical": 0, "name": "default", "volume": 1}
	}
	if n.Thread != "" {
		aps["thread-id"] = n.Thread
	}
	payload := map[string]any{"aps": aps}
	for k, v := range n.Payload {
		payload[k] = v
	}
	body, _ := json.Marshal(payload)
	host := "https://api.sandbox.push.apple.com"
	if c.cfg.Production {
		host = "https://api.push.apple.com"
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, host+"/3/device/"+deviceToken, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("authorization", "bearer "+jwt)
	req.Header.Set("apns-topic", c.cfg.BundleID)
	req.Header.Set("apns-push-type", "alert")
	req.Header.Set("apns-priority", "10")
	req.Header.Set("content-type", "application/json")
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		return nil
	}
	data, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	var e struct {
		Reason string `json:"reason"`
	}
	_ = json.Unmarshal(data, &e)
	if resp.StatusCode == http.StatusGone || e.Reason == "BadDeviceToken" || e.Reason == "Unregistered" || e.Reason == "DeviceTokenNotForTopic" {
		return ErrBadDeviceToken
	}
	return fmt.Errorf("APNs %d: %s", resp.StatusCode, e.Reason)
}
