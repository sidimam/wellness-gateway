package engine

import (
	"strings"
	"testing"
	"time"

	"github.com/sidimam/wellness-gateway/internal/model"
	"github.com/sidimam/wellness-gateway/internal/mywellness"
	"github.com/sidimam/wellness-gateway/internal/store"
)

// Il riepilogo dice di chi sono prenotazioni, osservazioni e tentativi di domani (con entrambi gli orari).
func TestSummaryNamesUsers(t *testing.T) {
	st, err := store.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	n := time.Now().In(mywellness.Rome)
	tomorrow := time.Date(n.Year(), n.Month(), n.Day()+1, 0, 0, 0, 0, mywellness.Rome)
	_ = st.Update(func(s *model.State) error {
		s.Profiles = []model.Profile{{ID: "p1", Label: "Daniela"}, {ID: "p2", Label: "Sofia"}}
		s.Settings = model.Settings{FollowServerOpenTime: false, BurstSeconds: 30, OpenRules: []model.OpenRule{
			{ID: "r", Pattern: "Reformer", DaysBefore: 3, Hour: 5, Minute: 1, AlsoAt: "07:01"},
			{ID: "default", Pattern: "*", DaysBefore: 7, Hour: 5, Minute: 1},
		}}
		s.Items = []model.Item{
			{ID: "a", ProfileID: "p1", Name: "Group Reformer", Start: tomorrow.AddDate(0, 0, 3).Add(17*time.Hour + 30*time.Minute), State: model.StatePending},
			{ID: "b", ProfileID: "p2", Name: "PILATES", Start: tomorrow.Add(18 * time.Hour), State: model.StateWatching},
		}
		return nil
	})
	st.AddLog("success", "p1", "✅ ACQUA GYM Sat 7 Nov 11:30: Prenotata (apertura prenotazioni)")
	e := New(st, nil)
	e.Version = "test"
	out := e.Summary()
	for _, want := range []string{"Daniela — ACQUA GYM", "Daniela — Group Reformer", "tentativi alle 05:01 e 07:01", "Sofia — PILATES", "osservazione"} {
		if !strings.Contains(out, want) {
			t.Errorf("manca %q in:\n%s", want, out)
		}
	}
}

// Su Telegram vanno prenotazioni ed eventi utili, col nome dell'utente; mai i rinnovi di sessione.
func TestTelegramText(t *testing.T) {
	cases := []struct {
		level, text, want string
		ok                bool
	}{
		{"info", "sessione mywellness scaduta: nuovo login", "", false},
		{"success", "login mywellness riuscito per Daniela (Daniela M.)", "", false},
		{"success", "✅ Group Reformer Thu 5 Nov 17:30: Prenotata (posto liberato)", "Daniela · ✅ Group Reformer Thu 5 Nov 17:30: Prenotata (posto liberato)", true},
		{"warn", "scaduta: PILATES Mon 9 Nov 18:00", "❌ Daniela · non prenotata, prenotazioni chiuse — PILATES Mon 9 Nov 18:00", true},
		{"error", "login mywellness fallito per Daniela: password", "⚠️ Daniela · login mywellness fallito per Daniela: password", true},
	}
	for _, c := range cases {
		got, ok := telegramText(c.level, "Daniela", c.text)
		if ok != c.ok || got != c.want {
			t.Errorf("%q: %q %v, atteso %q %v", c.text, got, ok, c.want, c.ok)
		}
	}
}
