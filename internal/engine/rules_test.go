package engine

import (
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
	// con "segui l'orario del centro" il giorno viene da mywellness, l'ora dalla regola valida in quel giorno
	opens := at("2026-11-01 07:00")
	it := model.Item{Name: "Group Reformer", Start: at("2026-11-04 17:30"), ServerOpensOn: &opens}
	if got := fireAt(it, wellnessTown(true)); !got.Equal(at("2026-11-01 07:01")) {
		t.Errorf("follow server: fireAt %s, atteso 2026-11-01 07:01", got.In(mywellness.Rome).Format("2006-01-02 15:04"))
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
