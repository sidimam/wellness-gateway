package engine

import (
	"testing"
	"time"

	"github.com/sidimam/wellness-gateway/internal/model"
	"github.com/sidimam/wellness-gateway/internal/mywellness"
	"github.com/sidimam/wellness-gateway/internal/store"
)

// Il centro rigenera il calendario: stessa lezione e ora con un nuovo identificativo.
// La voce vecchia passa alla lezione nuova; se esiste già una voce per la nuova, la vecchia è un doppione.
func TestSyncItemsNewClassID(t *testing.T) {
	st, err := store.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	e := New(st, nil)
	start := time.Date(2030, 11, 5, 17, 30, 0, 0, mywellness.Rome)
	mk := func(classID string, pd int, day int) model.Item {
		s := start.AddDate(0, 0, day)
		return model.Item{ID: "p1|" + classID + "|" + itoa(pd), ProfileID: "p1", ClassID: classID, PartitionDate: pd, Name: "Group Reformer",
			Start: s, End: s.Add(45 * time.Minute), State: model.StateBooked, Recurring: true}
	}
	_ = st.Update(func(s *model.State) error {
		s.Profiles = []model.Profile{{ID: "p1", Label: "Daniela"}}
		s.Items = []model.Item{
			mk("old", 20301105, 0), mk("new", 20301105, 0), // doppione: la vecchia va tolta
			mk("old2", 20301107, 2), // solo la vecchia: va spostata sulla nuova
		}
		return nil
	})
	ev := func(id string, pd int, day int, participant bool) mywellness.ClassEvent {
		s := start.AddDate(0, 0, day)
		return mywellness.ClassEvent{ID: id, Name: "Group Reformer", PartitionDate: pd, StartDate: s.Format("2006-01-02T15:04:05"),
			EndDate: s.Add(45 * time.Minute).Format("2006-01-02T15:04:05"), IsParticipant: participant, MaxParticipants: 7}
	}
	e.syncItems("p1", []mywellness.ClassEvent{ev("new", 20301105, 0, true), ev("new2", 20301107, 2, true)})
	var ids []string
	st.Read(func(s *model.State) {
		for _, it := range s.Items {
			ids = append(ids, it.ID)
		}
	})
	want := map[string]bool{"p1|new|20301105": true, "p1|new2|20301107": true}
	if len(ids) != 2 || !want[ids[0]] || !want[ids[1]] {
		t.Fatalf("voci dopo la sincronizzazione: %v, attese %v", ids, want)
	}
	// e la ricorrenza non riaggancia la stessa lezione
	e.attachRecurring("p1", []mywellness.ClassEvent{ev("new2", 20301107, 2, true), ev("other", 20301107, 2, false)})
	n := 0
	st.Read(func(s *model.State) { n = len(s.Items) })
	if n != 2 {
		t.Fatalf("la ricorrenza ha aggiunto un doppione: %d voci", n)
	}
}
