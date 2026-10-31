package telegram

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSendNow(t *testing.T) {
	var got string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		if r.URL.Path != "/botTOK/sendMessage" || r.Form.Get("chat_id") != "42" {
			http.Error(w, "no", 400)
			return
		}
		got = r.Form.Get("text")
	}))
	defer srv.Close()
	b := New("TOK", "42")
	b.base = srv.URL
	if err := b.SendNow(context.Background(), "✅ Daniela: Reformer"); err != nil || got != "✅ Daniela: Reformer" {
		t.Fatalf("invio: %v %q", err, got)
	}
	if New("", "42") != nil || New("TOK", "") != nil {
		t.Fatal("senza token o chat il bot deve essere nil")
	}
	var nilBot *Bot
	nilBot.Send("x", 0) // non deve andare in panic
}
