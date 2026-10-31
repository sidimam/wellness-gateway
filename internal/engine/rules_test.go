package engine

import (
	"strings"
	"testing"
	"time"

	"github.com/sidimam/wellness-gateway/internal/model"
	"github.com/sidimam/wellness-gateway/internal/mywellness"
)

// Regole Wellness Town dal 1/11/2026: Reformer 3 giorni prima, 05:01 fino al 31/10 e 07:01 dal 1/11, max 2 a settimana.
func wellnessTown(follow bool) model.Settings {
	return model.Settings{FollowServerOpenTime: follow, OpenRules: []model.OpenRule{
		{ID: "r1", Pattern: "Reformer", DaysBefore: 3, Hour: 5, Minute: 1, MaxBookings: 3, MaxPerWeek: 2},
		{ID: "r2", Pattern: "Reformer", DaysBefore: 3, Hour: 7, Minute: 1, MaxBookings: 3, MaxPerWeek: 2, ValidFrom: "2026-11-01"},
		{ID: "default", Pattern: "*", DaysBefore: 7, Hour: 5, Minute: 1},
	}}
}

func at(s string) time.Time {
	t, err := time.ParseInLocation("2006-01-02 15:04", s, mywellness.Rome)
	if err != nil {
		panic(err)
	}
	return t
}

func TestFireAtValidFrom(t *testing.T) {
	cases := []struct {
		name, start, want string
		follow            bool
	}{
		{"Group Reformer", "2026-11-03 17:30", "2026-10-31 05:01", false}, // apertura il 31/10: regola vecchia
		{"Group Reformer", "2026-11-04 17:30", "2026-11-01 07:01", false}, // apertura il 1/11: regola nuova
		{"Group Reformer", "2026-11-05 17:30", "2026-11-02 07:01", false},
		{"ACQUA GYM_acqua bassa", "2026-11-07 11:30", "2026-10-31 05:01", false}, // le altre lezioni non cambiano
	}
	for _, c := range cases {
		it := model.Item{Name: c.name, Start: at(c.start)}
		got := fireAt(it, wellnessTown(c.follow))
		if !got.Equal(at(c.want)) {
			t.Errorf("%s %s: fireAt %s, atteso %s", c.name, c.start, got.In(mywellness.Rome).Format("2006-01-02 15:04"), c.want)
		}
	}
	// con "segui l'orario del centro" il giorno viene da mywellness e l'apertura comunicata diventa un tentativo
	// IN PIÙ (non sostituisce l'orario della regola: il 31/10 il centro diceva 07:00 ma si prenotava alle 05:00)
	follow := []struct {
		start, opens string
		want         []string
	}{
		{"2026-11-04 17:30", "2026-11-01 07:00", []string{"2026-11-01 07:01"}},
		{"2026-11-03 17:30", "2026-10-31 07:00", []string{"2026-10-31 05:01", "2026-10-31 07:01"}},
		{"2026-11-03 17:30", "2026-10-31 05:00", []string{"2026-10-31 05:01"}},
		{"2026-11-06 17:30", "2026-11-03 05:00", []string{"2026-11-03 05:01", "2026-11-03 07:01"}},
	}
	for _, c := range follow {
		opens := at(c.opens)
		it := model.Item{Name: "Group Reformer", Start: at(c.start), ServerOpensOn: &opens}
		checkTimes(t, "follow "+c.start, fireTimes(it, wellnessTown(true)), c.want)
	}
}

func checkTimes(t *testing.T, label string, got []time.Time, want []string) {
	t.Helper()
	g := []string{}
	for _, x := range got {
		g = append(g, x.In(mywellness.Rome).Format("2006-01-02 15:04"))
	}
	if strings.Join(g, ",") != strings.Join(want, ",") {
		t.Errorf("%s: tentativi %v, attesi %v", label, g, want)
	}
}

// Due orari di primo tentativo nella stessa regola (05:01 e 07:01) e passaggio al secondo dopo la raffica.
func TestFireTimesAlsoAt(t *testing.T) {
	s := model.Settings{BurstSeconds: 30, OpenRules: []model.OpenRule{
		{ID: "r", Pattern: "Reformer", DaysBefore: 3, Hour: 5, Minute: 1, AlsoAt: "07:01", MaxBookings: 3, MaxPerWeek: 2},
		{ID: "default", Pattern: "*", DaysBefore: 7, Hour: 5, Minute: 1},
	}}
	it := model.Item{Name: "Group Reformer", Start: at("2026-11-05 17:30")}
	checkTimes(t, "reformer", fireTimes(it, s), []string{"2026-11-02 05:01", "2026-11-02 07:01"})
	checkTimes(t, "altre", fireTimes(model.Item{Name: "PILATES", Start: at("2026-11-09 18:00")}, s), []string{"2026-11-02 05:01"})
	// dopo la raffica delle 05:01 (round 0) il prossimo è alle 07:01
	if next, round, ok := nextRound(it, s, at("2026-11-02 05:10")); !ok || round != 1 || !next.Equal(at("2026-11-02 07:01")) {
		t.Errorf("nextRound 05:10: %v %d %v", next, round, ok)
	}
	// fatto anche il secondo: niente altri tentativi
	it.Round = 1
	if _, _, ok := nextRound(it, s, at("2026-11-02 07:10")); ok {
		t.Error("nessun terzo tentativo atteso")
	}
	// a gateway spento il secondo orario è passato da ore: non si recupera
	it.Round = 0
	if _, _, ok := nextRound(it, s, at("2026-11-02 12:00")); ok {
		t.Error("tentativo scaduto non va recuperato")
	}
}

func TestRuleOnAndWeekBounds(t *testing.T) {
	s := wellnessTown(true)
	if r := s.RuleOn("Group Reformer", "2026-10-31"); r.ID != "r1" {
		t.Errorf("31/10: regola %s, attesa r1", r.ID)
	}
	if r := s.RuleOn("Group Reformer", "2026-11-01"); r.ID != "r2" {
		t.Errorf("1/11: regola %s, attesa r2", r.ID)
	}
	if r := s.RuleOn("Pilates", "2026-11-01"); r.ID != "default" {
		t.Errorf("Pilates: regola %s, attesa default", r.ID)
	}
	from, to := weekBounds(at("2026-11-05 17:30")) // giovedì
	if !from.Equal(at("2026-11-02 00:00")) || !to.Equal(at("2026-11-09 00:00")) {
		t.Errorf("settimana %s–%s, attesa 2/11–9/11", from, to)
	}
	from, _ = weekBounds(at("2026-11-08 10:00")) // domenica: stessa settimana
	if !from.Equal(at("2026-11-02 00:00")) {
		t.Errorf("domenica: inizio settimana %s, atteso 2/11", from)
	}
}
