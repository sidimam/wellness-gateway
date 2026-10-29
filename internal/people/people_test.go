package people

import (
	"context"
	"testing"
	"time"

	"github.com/sidimam/wellness-gateway/internal/mywellness"
)

// Nome e cognome: salvati per i profili pubblici, cercabili, cancellati quando il socio rende privato il profilo.
func TestNames(t *testing.T) {
	ctx := context.Background()
	d, err := Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer d.Close()
	meta := ClassMeta{ClassID: "c1", PartitionDate: 20261030, FacilityID: "f1", Name: "Reformer", Start: time.Now()}
	parts := []mywellness.Participant{{UserID: "a", NickName: "dani"}, {UserID: "b", NickName: "", Private: true}}
	if err := d.RecordParticipants(ctx, meta, parts); err != nil {
		t.Fatal(err)
	}
	if got := d.NeedName(ctx, []string{"a", "b"}); len(got) != 1 || got[0] != "a" {
		t.Fatalf("NeedName: %v", got)
	}
	if got := d.MissingNames(ctx, "f1", 10); len(got) != 1 || got[0] != "a" {
		t.Fatalf("MissingNames: %v", got)
	}
	if err := d.SetName(ctx, "a", "Daniela", "Manieri", false); err != nil {
		t.Fatal(err)
	}
	if got := d.NeedName(ctx, []string{"a"}); len(got) != 0 {
		t.Fatalf("nome appena letto non va riletto: %v", got)
	}
	ms, err := d.Members(ctx, []string{"f1"}, "manieri dan", 10)
	if err == nil && len(ms) == 0 {
		ms, err = d.Members(ctx, []string{"f1"}, "Manieri Daniela", 10)
	}
	if err != nil || len(ms) != 1 || ms[0].FirstName != "Daniela" || ms[0].LastName != "Manieri" {
		t.Fatalf("ricerca per nome: %+v %v", ms, err)
	}
	// il socio rende privato il profilo: alla lettura successiva dei partecipanti il nome sparisce
	parts[0].Private = true
	if err := d.RecordParticipants(ctx, meta, parts); err != nil {
		t.Fatal(err)
	}
	m, _, _, err := d.Member(ctx, "a")
	if err != nil || m.FirstName != "" || m.LastName != "" || !m.Private {
		t.Fatalf("privato: %+v %v", m, err)
	}
}
