// Package telegram invia i messaggi del gateway al bot Telegram di casa (solo invio: niente polling, il bot
// è letto da altri servizi).
package telegram

import (
	"context"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

// Bot invia messaggi a una chat. Un Bot nil non fa nulla.
type Bot struct {
	token, chat string
	client      *http.Client
	base        string

	mu   sync.Mutex
	sent map[string]time.Time // testo → ultimo invio (niente doppioni ravvicinati)
}

// New restituisce nil se token o chat mancano.
func New(token, chat string) *Bot {
	token, chat = strings.TrimSpace(token), strings.TrimSpace(chat)
	if token == "" || chat == "" {
		return nil
	}
	return &Bot{token: token, chat: chat, client: &http.Client{Timeout: 15 * time.Second}, base: "https://api.telegram.org", sent: map[string]time.Time{}}
}

// Send invia il testo in background (3 tentativi). Lo stesso testo non viene ripetuto entro dedupe.
func (b *Bot) Send(text string, dedupe time.Duration) {
	if b == nil || strings.TrimSpace(text) == "" {
		return
	}
	b.mu.Lock()
	if last, ok := b.sent[text]; ok && dedupe > 0 && time.Since(last) < dedupe {
		b.mu.Unlock()
		return
	}
	b.sent[text] = time.Now()
	for k, t := range b.sent {
		if time.Since(t) > 24*time.Hour {
			delete(b.sent, k)
		}
	}
	b.mu.Unlock()
	go func() {
		for i := 0; i < 3; i++ {
			if err := b.SendNow(context.Background(), text); err == nil {
				return
			} else if i == 2 {
				log.Printf("telegram: invio fallito: %v", err)
			}
			time.Sleep(time.Duration(5*(i+1)) * time.Second)
		}
	}()
}

// SendNow invia subito e restituisce l'errore.
func (b *Bot) SendNow(ctx context.Context, text string) error {
	if b == nil {
		return nil
	}
	form := url.Values{"chat_id": {b.chat}, "text": {text}, "disable_web_page_preview": {"true"}}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, b.base+"/bot"+b.token+"/sendMessage", strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := b.client.Do(req)
	if err != nil {
		return err
	}
	resp.Body.Close()
	if resp.StatusCode/100 != 2 {
		return &httpError{resp.StatusCode}
	}
	return nil
}

type httpError struct{ code int }

func (e *httpError) Error() string { return "telegram HTTP " + http.StatusText(e.code) }
