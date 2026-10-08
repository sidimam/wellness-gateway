// Package engine esegue scheduler e osservazione per tutti i profili mywellness.
package engine

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/sidimam/wellness-gateway/internal/apns"
	"github.com/sidimam/wellness-gateway/internal/model"
	"github.com/sidimam/wellness-gateway/internal/mywellness"
	"github.com/sidimam/wellness-gateway/internal/store"
)

// Engine coordina profili, calendario e prenotazioni.
type Engine struct {
	Store *store.Store
	MW    *mywellness.Client
	Push  *apns.Client // può essere nil

	mu        sync.Mutex
	calendars map[string]calendar  // per profilo: calendario autenticato (daysAhead)
	dayCache  map[string]calendar  // per profilo|giorno: ultima lettura dell'osservazione (condivisa tra le lezioni del giorno)
	sessOK    map[string]time.Time // per profilo: ultima verifica positiva della sessione (GetLoginStatus)
	nextPoll  map[string]time.Time
	probes    map[string]calendar  // per struttura|giorno: ultima lettura PUBBLICA (senza token) dell'osservazione
	lastAuth  map[string]time.Time // per item: ultima lettura autenticata durante l'osservazione
	lastWarn  map[string]time.Time // per item: ultimo avviso di rete nel registro
	lastFull  time.Time
	wake      chan struct{}
	NextWake  time.Time
	Started   time.Time
}

type calendar struct {
	events []mywellness.ClassEvent
	at     time.Time
}

// New crea il motore.
func New(st *store.Store, push *apns.Client) *Engine {
	return &Engine{Store: st, MW: mywellness.New(), Push: push, calendars: map[string]calendar{}, dayCache: map[string]calendar{}, sessOK: map[string]time.Time{}, nextPoll: map[string]time.Time{}, probes: map[string]calendar{}, lastAuth: map[string]time.Time{}, lastWarn: map[string]time.Time{}, wake: make(chan struct{}, 1), Started: time.Now()}
}

// Kick sveglia il loop (dopo modifiche da API).
func (e *Engine) Kick() {
	select {
	case e.wake <- struct{}{}:
	default:
	}
}

func (e *Engine) settings() model.Settings {
	var s model.Settings
	e.Store.Read(func(st *model.State) { s = st.Settings })
	return s
}

func (e *Engine) logf(level, profileID, format string, args ...any) {
	text := fmt.Sprintf(format, args...)
	log.Printf("[%s] %s", level, text)
	e.Store.AddLog(level, profileID, text)
}

// Run esegue il loop finché il contesto non è annullato.
func (e *Engine) Run(ctx context.Context) {
	e.logf("info", "", "motore avviato")
	for ctx.Err() == nil {
		s := e.settings()
		if time.Since(e.lastFull) > 10*time.Minute {
			e.RefreshAll(ctx)
		}
		wake := time.Now().Add(5 * time.Second)
		var items []model.Item
		e.Store.Read(func(st *model.State) { items = append(items, st.Items...) })
		now := time.Now()
		for _, it := range items {
			if it.State.Terminal() || it.State == model.StateFailed {
				continue
			}
			if !now.Before(it.Start.Add(-5 * time.Minute)) {
				e.setItem(it.ID, func(x *model.Item) { x.State = model.StateExpired; x.LastMessage = "Prenotazioni chiuse" })
				e.logf("warn", it.ProfileID, "scaduta: %s %s", it.Name, fmtTime(it.Start))
				continue
			}
			switch it.State {
			case model.StatePending:
				fire := fireAt(it, s)
				if fire == nil {
					e.setItem(it.ID, func(x *model.Item) { x.LastMessage = "Orario di apertura sconosciuto" })
					continue
				}
				target := fire.Add(-time.Duration(s.LeadMilliseconds) * time.Millisecond)
				if !now.Before(target) {
					e.setItem(it.ID, func(x *model.Item) { x.State = model.StateBursting })
					e.attempt(ctx, it.ID, "apertura prenotazioni")
					wake = minT(wake, time.Now().Add(1500*time.Millisecond))
				} else {
					wake = minT(wake, target)
				}
			case model.StateBursting:
				fire := fireAt(it, s)
				if fire == nil {
					fire = &now
				}
				if !now.After(fire.Add(time.Duration(s.BurstSeconds) * time.Second)) {
					e.attempt(ctx, it.ID, "ritentativo")
					wake = minT(wake, time.Now().Add(1500*time.Millisecond))
				} else {
					e.setItem(it.ID, func(x *model.Item) {
						x.State = model.StateWatching
						x.LastMessage = "Finestra di apertura conclusa, passo all'osservazione"
					})
					e.logf("info", it.ProfileID, "%s %s: osservazione attiva", it.Name, fmtTime(it.Start))
				}
			case model.StateWatching, model.StateWaitingList:
				due, ok := e.nextPoll[it.ID]
				if !ok || !now.Before(due) {
					e.watch(ctx, it.ID)
					next := time.Now().Add(pollInterval(it, s))
					e.nextPoll[it.ID] = next
					wake = minT(wake, next)
				} else {
					wake = minT(wake, due)
				}
			}
		}
		e.NextWake = wake
		d := time.Until(wake)
		if d < 50*time.Millisecond {
			d = 50 * time.Millisecond
		}
		if d > 5*time.Second {
			d = 5 * time.Second
		}
		select {
		case <-ctx.Done():
		case <-time.After(d):
		case <-e.wake:
		}
	}
	e.logf("warn", "", "motore fermato")
}

func minT(a, b time.Time) time.Time {
	if b.Before(a) {
		return b
	}
	return a
}

func fmtTime(t time.Time) string { return t.In(mywellness.Rome).Format("Mon 2 Jan 15:04") }

// fireAt calcola quando tentare la prenotazione. L'ORA viene sempre dalla regola che corrisponde
// (specifica, altrimenti *). Il GIORNO viene dall'apertura comunicata dal centro se "segui il centro" è
// attivo e il centro la comunica, altrimenti da "giorni prima" della regola.
func fireAt(it model.Item, s model.Settings) *time.Time {
	r := s.Rule(it.Name)
	var day time.Time
	if s.FollowServerOpenTime && it.ServerOpensOn != nil {
		day = it.ServerOpensOn.In(mywellness.Rome)
	} else {
		st := it.Start.In(mywellness.Rome)
		day = time.Date(st.Year(), st.Month(), st.Day()-r.DaysBefore, 0, 0, 0, 0, mywellness.Rome)
	}
	t := time.Date(day.Year(), day.Month(), day.Day(), r.Hour, r.Minute, 0, 0, mywellness.Rome)
	return &t
}

// RuleCount conta le prenotazioni attive del profilo che ricadono in una regola (per il limite per regola).
func (e *Engine) RuleCount(profileID string, r model.OpenRule) int {
	pat := strings.TrimSpace(r.Pattern)
	now := time.Now()
	seen := map[string]bool{}
	e.mu.Lock()
	if c, ok := e.calendars[profileID]; ok {
		for _, ev := range c.events {
			if ev.IsParticipant && ev.Start().After(now) && (pat == "*" || strings.Contains(strings.ToLower(ev.Name), strings.ToLower(pat))) {
				seen[ev.Key()] = true
			}
		}
	}
	e.mu.Unlock()
	e.Store.Read(func(st *model.State) {
		for _, it := range st.Items {
			if it.ProfileID == profileID && it.State == model.StateBooked && it.Start.After(now) &&
				(pat == "*" || strings.Contains(strings.ToLower(it.Name), strings.ToLower(pat))) {
				seen[it.ClassID+"|"+itoa(it.PartitionDate)] = true
			}
		}
	})
	return len(seen)
}

// pollInterval: l'osservazione legge il calendario pubblico (senza token) ogni PollSeconds (default 15 s)
// e ogni NearPollSeconds (default 3 s) nelle ultime NearHours ore, con una variazione casuale del ±15%.
func pollInterval(it model.Item, s model.Settings) time.Duration {
	base := time.Duration(max(5, s.PollSeconds)) * time.Second
	nearHours := s.NearHours
	if nearHours <= 0 {
		nearHours = 4
	}
	if time.Until(it.Start) < time.Duration(nearHours)*time.Hour {
		base = time.Duration(max(2, s.NearPollSeconds)) * time.Second
	}
	j := 0.85 + rand.Float64()*0.30
	return time.Duration(float64(base) * j)
}

func (e *Engine) setItem(id string, fn func(*model.Item)) {
	_ = e.Store.Update(func(st *model.State) error {
		for i := range st.Items {
			if st.Items[i].ID == id {
				fn(&st.Items[i])
				s := st.Settings
				st.Items[i].FireAt = fireAt(st.Items[i], s)
			}
		}
		return nil
	})
}

func (e *Engine) item(id string) (model.Item, bool) {
	var out model.Item
	found := false
	e.Store.Read(func(st *model.State) {
		for _, it := range st.Items {
			if it.ID == id {
				out, found = it, true
			}
		}
	})
	return out, found
}

func (e *Engine) profile(id string) (model.Profile, bool) {
	var out model.Profile
	found := false
	e.Store.Read(func(st *model.State) {
		for _, p := range st.Profiles {
			if p.ID == id {
				out, found = p, true
			}
		}
	})
	return out, found
}

// Session restituisce la sessione mywellness del profilo, facendo login se serve.
func (e *Engine) Session(ctx context.Context, profileID string, force bool) (*mywellness.Session, error) {
	p, ok := e.profile(profileID)
	if !ok {
		return nil, errors.New("profilo inesistente")
	}
	if !force && p.Token != "" && p.MWUserID != "" {
		return &mywellness.Session{Token: p.Token, UserID: p.MWUserID}, nil
	}
	pw, err := e.Store.Decrypt(p.PasswordEnc)
	if err != nil {
		return nil, err
	}
	res, err := e.MW.Login(ctx, p.Username, pw)
	now := time.Now()
	if err != nil {
		_ = e.Store.Update(func(st *model.State) error {
			for i := range st.Profiles {
				if st.Profiles[i].ID == profileID {
					st.Profiles[i].LastLoginErr = err.Error()
				}
			}
			return nil
		})
		e.logf("error", profileID, "login mywellness fallito per %s: %v", p.Label, err)
		return nil, err
	}
	_ = e.Store.Update(func(st *model.State) error {
		for i := range st.Profiles {
			if st.Profiles[i].ID == profileID {
				st.Profiles[i].Token = res.Session.Token
				st.Profiles[i].MWUserID = res.Session.UserID
				st.Profiles[i].DisplayName = res.DisplayName
				st.Profiles[i].FirstName, st.Profiles[i].LastName, st.Profiles[i].NickName = res.FirstName, res.LastName, res.NickName
				st.Profiles[i].Email, st.Profiles[i].PictureURL, st.Profiles[i].ThumbURL = res.Email, res.PictureURL, res.ThumbURL
				if len(res.Identity) > 0 {
					st.Profiles[i].Identity = res.Identity
				}
				st.Profiles[i].LastLoginAt = &now
				st.Profiles[i].LastLoginErr = ""
			}
		}
		return nil
	})
	e.mu.Lock()
	e.sessOK[profileID] = time.Now()
	e.mu.Unlock()
	e.logf("success", profileID, "login mywellness riuscito per %s (%s)", p.Label, res.DisplayName)
	return &res.Session, nil
}

// ensureSession restituisce una sessione VERIFICATA: controlla GetLoginStatus (al più ogni 5 minuti)
// e rifà il login se il token è scaduto. I token mywellness durano circa un'ora e le API con token
// scaduto rispondono come anonimo, quindi senza questo controllo isParticipant sarebbe sempre falso.
func (e *Engine) ensureSession(ctx context.Context, profileID string) (*mywellness.Session, error) {
	sess, err := e.Session(ctx, profileID, false)
	if err != nil {
		return nil, err
	}
	e.mu.Lock()
	last := e.sessOK[profileID]
	e.mu.Unlock()
	if time.Since(last) < 5*time.Minute {
		return sess, nil
	}
	ok, err := e.MW.LoginStatus(ctx, sess)
	if err != nil {
		return nil, err
	}
	if !ok {
		e.logf("info", profileID, "sessione mywellness scaduta: nuovo login")
		if sess, err = e.Session(ctx, profileID, true); err != nil {
			return nil, err
		}
	}
	e.mu.Lock()
	e.sessOK[profileID] = time.Now()
	e.mu.Unlock()
	return sess, nil
}

// Calendar restituisce il calendario autenticato del profilo (cache 60 s, o forzato).
func (e *Engine) Calendar(ctx context.Context, profileID string, force bool) ([]mywellness.ClassEvent, error) {
	e.mu.Lock()
	c, ok := e.calendars[profileID]
	e.mu.Unlock()
	if ok && !force && time.Since(c.at) < 60*time.Second {
		return c.events, nil
	}
	p, found := e.profile(profileID)
	if !found {
		return nil, errors.New("profilo inesistente")
	}
	s := e.settings()
	sess, err := e.ensureSession(ctx, profileID)
	if err != nil {
		return nil, err
	}
	from := time.Now()
	to := from.AddDate(0, 0, max(3, s.DaysAhead))
	events, err := e.MW.Search(ctx, p.FacilityID, from, to, sess)
	if errors.Is(err, mywellness.ErrUnauthorized) {
		e.mu.Lock()
		delete(e.sessOK, profileID)
		e.mu.Unlock()
		if sess, err = e.Session(ctx, profileID, true); err == nil {
			events, err = e.MW.Search(ctx, p.FacilityID, from, to, sess)
		}
	}
	if err != nil {
		return nil, err
	}
	// Se la risposta farebbe risultare disdetta una lezione prenotata, prima verifica DAVVERO la sessione:
	// un token scaduto fa rispondere mywellness come anonimo (isParticipant=false ovunque) e senza questo
	// controllo la prenotazione risulterebbe "disdetta" e poi di nuovo "prenotata" a ogni ora.
	if e.wouldCancel(profileID, events) {
		ok, lerr := e.MW.LoginStatus(ctx, sess)
		if lerr != nil {
			return nil, lerr
		}
		if !ok {
			e.logf("info", profileID, "risposta anonima da mywellness (token scaduto): nuovo login e rilettura")
			if sess, err = e.Session(ctx, profileID, true); err != nil {
				return nil, err
			}
			if events, err = e.MW.Search(ctx, p.FacilityID, from, to, sess); err != nil {
				return nil, err
			}
		}
	}
	sort.Slice(events, func(i, j int) bool { return events[i].Start().Before(events[j].Start()) })
	e.mu.Lock()
	e.calendars[profileID] = calendar{events: events, at: time.Now()}
	e.mu.Unlock()
	e.syncItems(profileID, events)
	e.attachRecurring(profileID, events)
	return events, nil
}

// wouldCancel: tra gli eventi letti c'è una lezione futura che il gateway ha come prenotata ma che
// risulta senza partecipazione (candidata a "disdetta rilevata").
func (e *Engine) wouldCancel(profileID string, events []mywellness.ClassEvent) bool {
	byKey := map[string]mywellness.ClassEvent{}
	for _, ev := range events {
		byKey[ev.Key()] = ev
	}
	now := time.Now()
	found := false
	e.Store.Read(func(st *model.State) {
		for _, it := range st.Items {
			if it.ProfileID != profileID || it.State != model.StateBooked || !it.Start.After(now) {
				continue
			}
			if ev, ok := byKey[it.ClassID+"|"+itoa(it.PartitionDate)]; ok && !ev.IsParticipant {
				found = true
			}
		}
	})
	return found
}

// RefreshAll aggiorna il calendario di ogni profilo.
func (e *Engine) RefreshAll(ctx context.Context) {
	var ids []string
	e.Store.Read(func(st *model.State) {
		for _, p := range st.Profiles {
			ids = append(ids, p.ID)
		}
	})
	for _, id := range ids {
		// profili creati prima della v0.1.6: ricarica l'identità mywellness (nome, foto) con un nuovo login
		if p, ok := e.profile(id); ok && p.FirstName == "" && p.PictureURL == "" && p.LastLoginAt != nil {
			_, _ = e.Session(ctx, id, true)
		}
		if _, err := e.Calendar(ctx, id, true); err != nil {
			e.logf("warn", id, "aggiornamento calendario fallito: %v", err)
		}
	}
	e.lastFull = time.Now()
}

// syncItems aggiorna posti, apertura e stato partecipante dagli eventi freschi; ricalcola le prenotazioni attive.
func (e *Engine) syncItems(profileID string, events []mywellness.ClassEvent) {
	byKey := map[string]mywellness.ClassEvent{}
	active := 0
	now := time.Now()
	settings := e.settings()
	for _, ev := range events {
		byKey[ev.Key()] = ev
		// le lezioni con una quota propria (es. Reformer 3) non contano nel limite generale
		if _, own := settings.OwnQuota(ev.Name); ev.IsParticipant && ev.Start().After(now) && !own {
			active++
		}
	}
	var booked, cancelled []model.Item
	_ = e.Store.Update(func(st *model.State) error {
		for i := range st.Items {
			it := &st.Items[i]
			if it.ProfileID != profileID {
				continue
			}
			ev, ok := byKey[it.ClassID+"|"+itoa(it.PartitionDate)]
			if !ok {
				continue
			}
			ap, mp := ev.AvailablePlaces, ev.MaxParticipants
			it.Available, it.MaxPlaces = &ap, &mp
			if o := ev.OpensOn(); o != nil {
				it.ServerOpensOn = o
			}
			it.FireAt = fireAt(*it, st.Settings)
			switch {
			case ev.IsParticipant && it.State != model.StateBooked:
				it.State = model.StateBooked
				it.LastMessage = "Risulti già iscritta/o sul calendario"
				it.BookedAt = &now
				booked = append(booked, *it)
			case !ev.IsParticipant && it.State == model.StateBooked && it.Start.After(now) &&
				(it.BookedAt == nil || now.Sub(*it.BookedAt) > 2*time.Minute):
				// disdetta fatta dall'app ufficiale o dal sito: non riprenotare
				it.State = model.StateCancelled
				it.LastMessage = "Disdetta da mywellness (app o web)"
				cancelled = append(cancelled, *it)
			case ev.IsInWaitingList && (it.State == model.StatePending || it.State == model.StateWatching):
				it.State = model.StateWaitingList
			}
		}
		for i := range st.Profiles {
			if st.Profiles[i].ID == profileID {
				st.Profiles[i].ActiveBookings = active
			}
		}
		return nil
	})
	for _, it := range booked {
		e.logf("success", profileID, "%s %s: già prenotata sul calendario", it.Name, fmtTime(it.Start))
	}
	for _, it := range cancelled {
		e.logf("warn", profileID, "%s %s: disdetta rilevata da mywellness", it.Name, fmtTime(it.Start))
		e.notify(profileID, "Disdetta rilevata", fmt.Sprintf("%s %s è stata disdetta da mywellness.", it.Name, fmtTime(it.Start)), false)
	}
}

// attachRecurring aggancia le occorrenze future alle regole settimanali del profilo.
func (e *Engine) attachRecurring(profileID string, events []mywellness.ClassEvent) {
	var added []model.Item
	_ = e.Store.Update(func(st *model.State) error {
		rules := map[string]model.Item{} // regola → item ricorrente più vecchio (la ricorrenza parte da lì)
		known := map[string]bool{}
		for _, it := range st.Items {
			if it.ProfileID != profileID {
				continue
			}
			known[it.ID] = true
			if it.Recurring {
				k := it.RuleKey(mywellness.Rome)
				if cur, ok := rules[k]; !ok || it.Start.Before(cur.Start) {
					rules[k] = it
				}
			}
		}
		if len(rules) == 0 {
			return nil
		}
		now := time.Now()
		for _, ev := range events {
			if !ev.Start().After(now) {
				continue
			}
			probe := itemFrom(profileID, ev, true, "")
			if known[probe.ID] {
				continue
			}
			if src, ok := rules[probe.RuleKey(mywellness.Rome)]; ok {
				if ev.Start().Before(src.Start) {
					continue // la ricorrenza vale dalla data scelta in poi
				}
				probe.CreatedBy = src.CreatedBy
				if ev.IsParticipant {
					probe.State = model.StateBooked
					probe.LastMessage = "Già prenotata"
				}
				probe.FireAt = fireAt(probe, st.Settings)
				st.Items = append(st.Items, probe)
				known[probe.ID] = true
				added = append(added, probe)
			}
		}
		sortItems(st.Items)
		return nil
	})
	for _, it := range added {
		e.logf("info", profileID, "ricorrenza: aggiunta %s %s", it.Name, fmtTime(it.Start))
	}
}

func sortItems(items []model.Item) {
	sort.Slice(items, func(i, j int) bool { return items[i].Start.Before(items[j].Start) })
}

func itoa(i int) string { return fmt.Sprint(i) }

// itemFrom costruisce un Item da un evento.
func itemFrom(profileID string, ev mywellness.ClassEvent, recurring bool, createdBy string) model.Item {
	ap, mp := ev.AvailablePlaces, ev.MaxParticipants
	return model.Item{
		ID: profileID + "|" + ev.Key(), ProfileID: profileID, ClassID: ev.ID, PartitionDate: ev.PartitionDate,
		Name: ev.Name, Start: ev.Start(), End: ev.End(), Room: ev.Room, Trainer: ev.AssignedTo, PictureURL: ev.PictureURL,
		ServerOpensOn: ev.OpensOn(), Recurring: recurring, State: model.StatePending,
		Available: &ap, MaxPlaces: &mp, CreatedBy: createdBy, CreatedAt: time.Now(),
	}
}

// AddItem crea (o restituisce) l'item per una lezione del calendario del profilo.
func (e *Engine) AddItem(ctx context.Context, profileID, classID string, partitionDate int, recurring bool, createdBy string) (model.Item, error) {
	events, err := e.Calendar(ctx, profileID, false)
	if err != nil {
		return model.Item{}, err
	}
	var ev *mywellness.ClassEvent
	for i := range events {
		if events[i].ID == classID && events[i].PartitionDate == partitionDate {
			ev = &events[i]
		}
	}
	if ev == nil {
		return model.Item{}, errors.New("lezione non trovata in calendario")
	}
	var out model.Item
	_ = e.Store.Update(func(st *model.State) error {
		id := profileID + "|" + ev.Key()
		for i := range st.Items {
			if st.Items[i].ID == id {
				st.Items[i].Recurring = st.Items[i].Recurring || recurring
				out = st.Items[i]
				return nil
			}
		}
		it := itemFrom(profileID, *ev, recurring, createdBy)
		if ev.IsParticipant {
			it.State, it.LastMessage = model.StateBooked, "Già prenotata"
		} else if ev.IsInWaitingList {
			it.State = model.StateWaitingList
		}
		it.FireAt = fireAt(it, st.Settings)
		st.Items = append(st.Items, it)
		sortItems(st.Items)
		out = it
		return nil
	})
	e.logf("success", profileID, "aggiunta %s %s%s", out.Name, fmtTime(out.Start), map[bool]string{true: " (ogni settimana)", false: ""}[recurring])
	if recurring {
		e.attachRecurring(profileID, events)
	}
	e.Kick()
	return out, nil
}

// RemoveItem elimina un item; con rule=true toglie anche la ricorrenza dalle occorrenze non concluse.
func (e *Engine) RemoveItem(id string, rule bool) {
	_ = e.Store.Update(func(st *model.State) error {
		var key string
		for _, it := range st.Items {
			if it.ID == id {
				key = it.ProfileID + "|" + it.RuleKey(mywellness.Rome)
			}
		}
		kept := st.Items[:0]
		for _, it := range st.Items {
			if it.ID == id {
				continue
			}
			if rule && key != "" && it.ProfileID+"|"+it.RuleKey(mywellness.Rome) == key {
				if !it.State.Terminal() {
					continue
				}
				it.Recurring = false
			}
			kept = append(kept, it)
		}
		st.Items = kept
		return nil
	})
	delete(e.nextPoll, id)
}

// Retry riporta un item fallito in attesa.
func (e *Engine) Retry(id string) {
	e.setItem(id, func(x *model.Item) { x.State = model.StatePending; x.LastMessage = ""; x.Attempts = 0 })
	delete(e.nextPoll, id)
	e.Kick()
}

// Unbook disdice una prenotazione (dell'item o fatta direttamente su mywellness) e segna l'item come disdetto.
func (e *Engine) Unbook(ctx context.Context, profileID, classID string, partitionDate int, by string) error {
	sess, err := e.ensureSession(ctx, profileID)
	if err != nil {
		return err
	}
	if err := e.MW.Unbook(ctx, sess, classID, partitionDate); err != nil {
		if errors.Is(err, mywellness.ErrUnauthorized) {
			if sess, err = e.Session(ctx, profileID, true); err == nil {
				err = e.MW.Unbook(ctx, sess, classID, partitionDate)
			}
		}
		if err != nil {
			return err
		}
	}
	id := profileID + "|" + classID + "|" + itoa(partitionDate)
	name := classID
	if it, ok := e.item(id); ok {
		name = it.Name + " " + fmtTime(it.Start)
		e.setItem(id, func(x *model.Item) { x.State = model.StateCancelled; x.LastMessage = "Disdetta da " + by })
		delete(e.nextPoll, id)
	}
	e.logf("warn", profileID, "disdetta (%s): %s", by, name)
	e.mu.Lock()
	delete(e.calendars, profileID)
	e.mu.Unlock()
	e.Kick()
	return nil
}

// LeaveWaitingList esce dalla lista d'attesa su mywellness e segna l'item come disdetto.
func (e *Engine) LeaveWaitingList(ctx context.Context, profileID, classID string, partitionDate int, by string) error {
	sess, err := e.ensureSession(ctx, profileID)
	if err != nil {
		return err
	}
	if err := e.MW.LeaveWaitingList(ctx, sess, classID, partitionDate); err != nil {
		if errors.Is(err, mywellness.ErrUnauthorized) {
			if sess, err = e.Session(ctx, profileID, true); err == nil {
				err = e.MW.LeaveWaitingList(ctx, sess, classID, partitionDate)
			}
		}
		if err != nil {
			return err
		}
	}
	id := profileID + "|" + classID + "|" + itoa(partitionDate)
	name := classID
	if it, ok := e.item(id); ok {
		name = it.Name + " " + fmtTime(it.Start)
		e.setItem(id, func(x *model.Item) {
			x.State = model.StateCancelled
			x.LastMessage = "Uscita dalla lista d'attesa (" + by + ")"
		})
		delete(e.nextPoll, id)
	}
	e.logf("warn", profileID, "uscita dalla lista d'attesa (%s): %s", by, name)
	e.mu.Lock()
	delete(e.calendars, profileID)
	e.mu.Unlock()
	e.Kick()
	return nil
}

// limitReached verifica il limite generale del profilo (default 5). Le lezioni coperte da una regola
// con un massimo proprio (es. Reformer 3) hanno una quota separata e sono escluse dal conteggio.
func (e *Engine) limitReached(profileID string) (bool, int, int) {
	var active, maxB int
	e.Store.Read(func(st *model.State) {
		for _, p := range st.Profiles {
			if p.ID == profileID {
				maxB = p.MaxBookings
			}
		}
		now := time.Now()
		seen := map[string]bool{}
		e.mu.Lock()
		if c, ok := e.calendars[profileID]; ok {
			for _, ev := range c.events {
				if _, own := st.Settings.OwnQuota(ev.Name); ev.IsParticipant && ev.Start().After(now) && !own {
					seen[ev.Key()] = true
				}
			}
		}
		e.mu.Unlock()
		for _, it := range st.Items {
			if _, own := st.Settings.OwnQuota(it.Name); it.ProfileID == profileID && it.State == model.StateBooked && it.Start.After(now) && !own {
				seen[it.ClassID+"|"+itoa(it.PartitionDate)] = true
			}
		}
		active = len(seen)
	})
	if maxB <= 0 {
		maxB = 5
	}
	return active >= maxB, active, maxB
}

// attempt prova a prenotare un item.
func (e *Engine) attempt(ctx context.Context, id, reason string) {
	it, ok := e.item(id)
	if !ok {
		return
	}
	// Una lezione con quota propria (es. Reformer: massimo 3) è soggetta SOLO a quella quota;
	// tutte le altre contano nel limite generale del profilo (default 5).
	if r, own := e.settings().OwnQuota(it.Name); own {
		if n := e.RuleCount(it.ProfileID, r); n >= r.MaxBookings {
			first := it.Attempts == 0
			e.setItem(id, func(x *model.Item) {
				x.Attempts++
				x.LastMessage = fmt.Sprintf("Limite \"%s\" raggiunto (%d/%d): riprovo quando se ne libera una", r.Pattern, n, r.MaxBookings)
			})
			if first {
				e.logf("warn", it.ProfileID, "%s %s: limite regola %s (%d/%d) raggiunto", it.Name, fmtTime(it.Start), r.Pattern, n, r.MaxBookings)
			}
			return
		}
	} else if reached, active, maxB := e.limitReached(it.ProfileID); reached {
		first := it.Attempts == 0
		e.setItem(id, func(x *model.Item) {
			x.Attempts++
			x.LastMessage = fmt.Sprintf("Limite di %d prenotazioni attive raggiunto (%d): riprovo quando se ne libera una", maxB, active)
		})
		if first {
			e.logf("warn", it.ProfileID, "%s %s: limite prenotazioni (%d/%d) raggiunto", it.Name, fmtTime(it.Start), active, maxB)
		}
		return
	}
	now := time.Now()
	e.setItem(id, func(x *model.Item) { x.Attempts++; x.LastCheck = &now })
	sess, err := e.ensureSession(ctx, it.ProfileID)
	if err != nil {
		e.setItem(id, func(x *model.Item) { x.LastMessage = "Login mywellness non riuscito" })
		return
	}
	res, err := e.MW.Book(ctx, sess, it.ClassID, it.PartitionDate)
	if err == nil && res.Outcome == mywellness.Unauthorized {
		if sess, err = e.Session(ctx, it.ProfileID, true); err == nil {
			res, err = e.MW.Book(ctx, sess, it.ClassID, it.PartitionDate)
		}
	}
	if err != nil {
		e.setItem(id, func(x *model.Item) { x.LastMessage = "Rete: " + err.Error() })
		e.logf("warn", it.ProfileID, "%s: errore di rete (%v)", it.Name, err)
		return
	}
	e.handle(it, res, reason)
}

func (e *Engine) handle(it model.Item, res mywellness.BookResult, reason string) {
	switch res.Outcome {
	case mywellness.Booked:
		e.markBooked(it, "Prenotata ("+reason+")")
	case mywellness.WaitingList:
		if it.State != model.StateWaitingList {
			e.setItem(it.ID, func(x *model.Item) { x.State = model.StateWaitingList; x.LastMessage = "In lista d'attesa" })
			e.logf("warn", it.ProfileID, "%s %s: classe piena, in lista d'attesa. Osservazione attiva.", it.Name, fmtTime(it.Start))
			e.notify(it.ProfileID, "Lista d'attesa", fmt.Sprintf("%s %s è piena: osservazione attiva, prenoto appena si libera un posto.", it.Name, fmtTime(it.Start)), false)
		} else {
			e.setItem(it.ID, func(x *model.Item) { x.LastMessage = "In lista d'attesa" })
		}
	case mywellness.Full:
		if it.State != model.StateWatching && it.State != model.StateWaitingList {
			e.setItem(it.ID, func(x *model.Item) { x.State = model.StateWatching; x.LastMessage = "Classe piena" })
			e.logf("warn", it.ProfileID, "%s %s: piena, osservazione attiva", it.Name, fmtTime(it.Start))
		}
	case mywellness.NotOpenYet:
		e.setItem(it.ID, func(x *model.Item) { x.LastMessage = "Non ancora aperta (" + res.Message + ")" })
	case mywellness.Unauthorized:
		e.setItem(it.ID, func(x *model.Item) { x.LastMessage = "Login non riuscito" })
	case mywellness.NoPermission:
		e.setItem(it.ID, func(x *model.Item) { x.State = model.StateFailed; x.LastMessage = "Non autorizzato: " + res.Message })
		e.logf("error", it.ProfileID, "%s: non autorizzato (%s)", it.Name, res.Message)
		e.notify(it.ProfileID, "Prenotazione rifiutata", it.Name+": "+res.Message, true)
	default:
		e.setItem(it.ID, func(x *model.Item) { x.LastMessage = "Tentativo fallito: " + res.Message })
		if it.State != model.StateBursting {
			e.logf("warn", it.ProfileID, "%s: %s", it.Name, res.Message)
		}
	}
}

func (e *Engine) markBooked(it model.Item, note string) {
	now := time.Now()
	e.setItem(it.ID, func(x *model.Item) { x.State = model.StateBooked; x.LastMessage = note; x.BookedAt = &now })
	delete(e.nextPoll, it.ID)
	delete(e.lastAuth, it.ID)
	e.mu.Lock()
	delete(e.calendars, it.ProfileID)
	e.mu.Unlock()
	e.logf("success", it.ProfileID, "✅ %s %s: %s", it.Name, fmtTime(it.Start), note)
	p, _ := e.profile(it.ProfileID)
	e.notify(it.ProfileID, "Prenotata ✅ · "+p.Label, fmt.Sprintf("%s · %s", it.Name, fmtTime(it.Start)), true)
}

// watch (osservazione) controlla i posti liberi di un item e prenota al volo.
// La lettura dei posti è PUBBLICA (senza token, condivisa tra tutti i profili e le lezioni dello stesso
// giorno) e quindi può essere frequente senza esporre l'account; ogni 60 s si fa anche una lettura
// autenticata per aggiornare partecipazione, lista d'attesa e stato "CanBook" dell'utente.
func (e *Engine) watch(ctx context.Context, id string) {
	it, ok := e.item(id)
	if !ok {
		return
	}
	now := time.Now()
	e.setItem(id, func(x *model.Item) { x.LastCheck = &now })
	p, found := e.profile(it.ProfileID)
	if !found {
		return
	}
	var ev *mywellness.ClassEvent
	authenticated := false
	if la := e.lastAuth[id]; time.Since(la) >= 60*time.Second {
		if events, err := e.dayEvents(ctx, it.ProfileID, it.Start); err == nil {
			e.lastAuth[id] = time.Now()
			authenticated = true
			ev = findEvent(events, it)
		} else {
			e.netWarn(it, err)
		}
	}
	if ev == nil {
		events, err := e.probe(ctx, p.FacilityID, it.Start)
		if err != nil {
			e.setItem(id, func(x *model.Item) { x.LastMessage = "Rete: " + err.Error() })
			e.netWarn(it, err)
			return
		}
		ev = findEvent(events, it)
	}
	if ev == nil {
		e.setItem(id, func(x *model.Item) { x.LastMessage = "Lezione non più in calendario" })
		return
	}
	ap, mp := ev.AvailablePlaces, ev.MaxParticipants
	e.setItem(id, func(x *model.Item) { x.Available, x.MaxPlaces = &ap, &mp })
	if authenticated && ev.IsParticipant {
		e.markBooked(it, "Posto confermato dal calendario")
		return
	}
	if ev.HasPlace() {
		e.logf("success", it.ProfileID, "🔔 posto libero per %s %s (%d/%d): prenoto subito", it.Name, fmtTime(it.Start), ev.NumberOfParticipants, ev.MaxParticipants)
		// fino a 3 tentativi ravvicinati: il posto può sparire tra una lettura e la prenotazione
		for i := 0; i < 3; i++ {
			e.attempt(ctx, id, "posto liberato")
			cur, ok := e.item(id)
			if !ok || cur.State == model.StateBooked || cur.State == model.StateFailed {
				break
			}
			if i < 2 {
				time.Sleep(700 * time.Millisecond)
			}
		}
		if cur, ok := e.item(id); ok && cur.State != model.StateBooked {
			e.nextPoll[id] = time.Now().Add(2 * time.Second)
			e.lastAuth[id] = time.Time{} // alla prossima lettura verifica subito se la prenotazione è passata
		}
		return
	}
	msg := fmt.Sprintf("Piena (%d/%d)", ev.NumberOfParticipants, ev.MaxParticipants)
	if ev.WaitingListCounter > 0 {
		msg += fmt.Sprintf(" · %d in lista d'attesa", ev.WaitingListCounter)
	}
	if authenticated && ev.IsInWaitingList {
		msg += " · tu in lista"
		if ev.WaitingListPosition > 0 {
			msg += fmt.Sprintf(" (posizione %d)", ev.WaitingListPosition)
		}
	} else if it.State == model.StateWaitingList {
		msg += " · tu in lista"
	}
	e.setItem(id, func(x *model.Item) { x.LastMessage = msg })
}

func findEvent(events []mywellness.ClassEvent, it model.Item) *mywellness.ClassEvent {
	for i := range events {
		if events[i].ID == it.ClassID && events[i].PartitionDate == it.PartitionDate {
			return &events[i]
		}
	}
	return nil
}

// netWarn scrive nel registro un errore di rete dell'osservazione al più ogni 10 minuti per lezione.
func (e *Engine) netWarn(it model.Item, err error) {
	if time.Since(e.lastWarn[it.ID]) < 10*time.Minute {
		return
	}
	e.lastWarn[it.ID] = time.Now()
	e.logf("warn", it.ProfileID, "%s %s: osservazione, errore di rete (%v)", it.Name, fmtTime(it.Start), err)
}

// probe legge il calendario PUBBLICO di un giorno (senza token), riusando per 2 s la lettura tra tutti i
// profili e le lezioni dello stesso giorno della stessa struttura.
func (e *Engine) probe(ctx context.Context, facilityID string, day time.Time) ([]mywellness.ClassEvent, error) {
	key := facilityID + "|" + day.In(mywellness.Rome).Format("2006-01-02")
	e.mu.Lock()
	c, ok := e.probes[key]
	e.mu.Unlock()
	if ok && time.Since(c.at) < 2*time.Second {
		return c.events, nil
	}
	events, err := e.MW.Search(ctx, facilityID, day, day, nil)
	if err != nil {
		return nil, err
	}
	e.mu.Lock()
	e.probes[key] = calendar{events: events, at: time.Now()}
	if len(e.probes) > 200 {
		for k, v := range e.probes {
			if time.Since(v.at) > time.Minute {
				delete(e.probes, k)
			}
		}
	}
	e.mu.Unlock()
	return events, nil
}

// dayEvents legge il calendario di un giorno per un profilo, riusando per 10 s la lettura
// tra tutte le lezioni dello stesso giorno (una sola richiesta a mywellness).
func (e *Engine) dayEvents(ctx context.Context, profileID string, day time.Time) ([]mywellness.ClassEvent, error) {
	key := profileID + "|" + day.In(mywellness.Rome).Format("2006-01-02")
	e.mu.Lock()
	c, ok := e.dayCache[key]
	e.mu.Unlock()
	if ok && time.Since(c.at) < 10*time.Second {
		return c.events, nil
	}
	p, found := e.profile(profileID)
	if !found {
		return nil, errors.New("profilo inesistente")
	}
	sess, err := e.ensureSession(ctx, profileID)
	if err != nil {
		return nil, err
	}
	events, err := e.MW.Search(ctx, p.FacilityID, day, day, sess)
	if errors.Is(err, mywellness.ErrUnauthorized) {
		if sess, err = e.Session(ctx, profileID, true); err == nil {
			events, err = e.MW.Search(ctx, p.FacilityID, day, day, sess)
		}
	}
	if err != nil {
		return nil, err
	}
	e.mu.Lock()
	e.dayCache[key] = calendar{events: events, at: time.Now()}
	if len(e.dayCache) > 200 {
		for k, v := range e.dayCache {
			if time.Since(v.at) > time.Minute {
				delete(e.dayCache, k)
			}
		}
	}
	e.mu.Unlock()
	return events, nil
}

// notify manda una push a tutti i dispositivi degli utenti che vedono il profilo.
func (e *Engine) notify(profileID, title, body string, priority bool) {
	if e.Push == nil {
		return
	}
	var tokens []string
	prio := priority
	e.Store.Read(func(st *model.State) {
		if !st.Settings.PriorityNotifications {
			prio = false
		}
		var owners []string
		for _, p := range st.Profiles {
			if p.ID == profileID {
				owners = p.OwnerUserIDs
			}
		}
		for _, d := range st.Devices {
			if d.APNSToken == "" {
				continue
			}
			if len(owners) == 0 || contains(owners, d.UserID) {
				tokens = append(tokens, d.APNSToken)
			}
		}
	})
	for _, tok := range tokens {
		go func(tok string) {
			ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer cancel()
			err := e.Push.Send(ctx, tok, apns.Notification{Title: title, Body: body, Priority: prio, Thread: "booking", Payload: map[string]any{"profileId": profileID}})
			if errors.Is(err, apns.ErrBadDeviceToken) {
				_ = e.Store.Update(func(st *model.State) error {
					for i := range st.Devices {
						if st.Devices[i].APNSToken == tok {
							st.Devices[i].APNSToken = ""
						}
					}
					return nil
				})
			} else if err != nil {
				log.Printf("APNs: %v", err)
			}
		}(tok)
	}
}

func contains(list []string, v string) bool {
	for _, x := range list {
		if x == v {
			return true
		}
	}
	return false
}

// TestNotification invia una notifica di prova ai dispositivi di un utente.
func (e *Engine) TestNotification(userID string) int {
	if e.Push == nil {
		return 0
	}
	var tokens []string
	e.Store.Read(func(st *model.State) {
		for _, d := range st.Devices {
			if d.UserID == userID && d.APNSToken != "" {
				tokens = append(tokens, d.APNSToken)
			}
		}
	})
	for _, tok := range tokens {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		_ = e.Push.Send(ctx, tok, apns.Notification{Title: "Notifica di prova", Body: "Così ti avviso quando prenoto o si libera un posto.", Priority: true, Thread: "booking"})
		cancel()
	}
	return len(tokens)
}

// Visible indica se l'utente può vedere il profilo: gli amministratori vedono tutti,
// gli altri utenti solo il proprio profilo (e quelli esplicitamente condivisi con loro).
func Visible(p model.Profile, userID string, isAdmin bool) bool {
	return isAdmin || p.UserID == userID || contains(p.OwnerUserIDs, userID)
}

// Normalize ripulisce le impostazioni ricevute dall'API.
func Normalize(s *model.Settings) {
	if s.PollSeconds < 5 {
		s.PollSeconds = 5
	}
	if s.NearPollSeconds <= 0 {
		s.NearPollSeconds = 3
	}
	if s.NearPollSeconds < 2 {
		s.NearPollSeconds = 2
	}
	if s.NearHours <= 0 {
		s.NearHours = 4
	}
	if s.BurstSeconds < 10 {
		s.BurstSeconds = 10
	}
	if s.LeadMilliseconds < 0 {
		s.LeadMilliseconds = 0
	}
	if s.DaysAhead < 3 {
		s.DaysAhead = 3
	}
	hasDefault := false
	for i := range s.OpenRules {
		s.OpenRules[i].Pattern = strings.TrimSpace(s.OpenRules[i].Pattern)
		if s.OpenRules[i].Pattern == "*" {
			hasDefault = true
		}
		if s.OpenRules[i].ID == "" {
			s.OpenRules[i].ID = store.NewID()
		}
	}
	if !hasDefault {
		s.OpenRules = append(s.OpenRules, model.OpenRule{ID: "default", Pattern: "*", DaysBefore: 7, Hour: 5})
	}
}
