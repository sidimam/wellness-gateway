// Package model contiene i tipi condivisi tra store, engine e API.
package model

import "time"

// User è un utente locale del gateway (chi usa l'app iOS o la web UI).
type User struct {
	ID           string    `json:"id"`
	Username     string    `json:"username"`
	DisplayName  string    `json:"displayName"`
	PasswordHash string    `json:"-"`
	Salt         string    `json:"-"`
	IsAdmin      bool      `json:"isAdmin"`
	CreatedAt    time.Time `json:"createdAt"`
}

// Device è un dispositivo autenticato (token Bearer + token APNs).
type Device struct {
	Token     string    `json:"-"`
	UserID    string    `json:"userId"`
	Name      string    `json:"name"`
	APNSToken string    `json:"apnsToken,omitempty"`
	SelfOnly  bool      `json:"selfOnly"` // app iOS: vede solo il proprio profilo, anche se amministratore
	CreatedAt time.Time `json:"createdAt"`
	LastSeen  time.Time `json:"lastSeen"`
}

// Profile è un account Technogym mywellness gestito dal gateway.
type Profile struct {
	ID             string         `json:"id"`
	Label          string         `json:"label"`
	Username       string         `json:"username"`
	PasswordEnc    string         `json:"-"`
	Token          string         `json:"-"`
	MWUserID       string         `json:"-"`
	DisplayName    string         `json:"displayName"`
	FirstName      string         `json:"firstName,omitempty"`
	LastName       string         `json:"lastName,omitempty"`
	NickName       string         `json:"nickName,omitempty"`
	Email          string         `json:"email,omitempty"`
	PictureURL     string         `json:"pictureUrl,omitempty"`
	ThumbURL       string         `json:"thumbUrl,omitempty"`
	Identity       map[string]any `json:"identity,omitempty"` // tutto il userContext mywellness (senza token/password)
	FacilityURL    string         `json:"facilityUrl"`
	FacilityID     string         `json:"facilityId"`
	FacilityName   string         `json:"facilityName"`
	MaxBookings    int            `json:"maxBookings"`
	UserID         string         `json:"userId,omitempty"` // utente del gateway a cui appartiene (la "sua" persona)
	OwnerUserIDs   []string       `json:"ownerUserIds"`     // vuoto = visibile a tutti (famiglia)
	LastLoginAt    *time.Time     `json:"lastLoginAt,omitempty"`
	LastLoginErr   string         `json:"lastLoginError,omitempty"`
	ActiveBookings int            `json:"activeBookings"`
}

// OpenRule: le lezioni il cui nome contiene Pattern aprono DaysBefore giorni prima alle Hour:Minute.
// Pattern "*" = tutte le altre; pattern vuoto = ignorata.
type OpenRule struct {
	ID          string `json:"id"`
	Pattern     string `json:"pattern"`
	DaysBefore  int    `json:"daysBefore"`
	Hour        int    `json:"hour"`
	Minute      int    `json:"minute"`
	MaxBookings int    `json:"maxBookings"` // 0 = nessun limite specifico per questa regola
}

// Settings globali del motore.
type Settings struct {
	FollowServerOpenTime  bool       `json:"followServerOpenTime"`
	OpenRules             []OpenRule `json:"openRules"`
	LeadMilliseconds      int        `json:"leadMilliseconds"`
	BurstSeconds          int        `json:"burstSeconds"`
	PollSeconds           int        `json:"pollSeconds"`     // osservazione: lettura pubblica ogni N s (default 15)
	NearPollSeconds       int        `json:"nearPollSeconds"` // osservazione nelle ultime NearHours ore: ogni N s (default 3)
	NearHours             int        `json:"nearHours"`       // finestra "vicina" alla lezione (default 4)
	DaysAhead             int        `json:"daysAhead"`
	PriorityNotifications bool       `json:"priorityNotifications"`
}

// DefaultSettings restituisce i valori iniziali (Wellness Town).
func DefaultSettings() Settings {
	return Settings{
		FollowServerOpenTime:  true,
		OpenRules:             []OpenRule{{ID: "default", Pattern: "*", DaysBefore: 7, Hour: 5, Minute: 0}},
		LeadMilliseconds:      300,
		BurstSeconds:          120,
		PollSeconds:           15,
		NearPollSeconds:       3,
		NearHours:             4,
		DaysAhead:             14,
		PriorityNotifications: true,
	}
}

// SpecificRule restituisce la regola non predefinita che corrisponde al nome, se esiste.
func (s Settings) SpecificRule(className string) (OpenRule, bool) {
	for _, r := range s.OpenRules {
		p := trim(r.Pattern)
		if p == "" || p == "*" {
			continue
		}
		if containsFold(className, p) {
			return r, true
		}
	}
	return OpenRule{}, false
}

// OwnQuota restituisce la regola con un massimo proprio (es. Reformer 3) che copre la lezione:
// queste lezioni hanno una quota separata e NON contano nel limite generale del profilo.
func (s Settings) OwnQuota(className string) (OpenRule, bool) {
	r, ok := s.SpecificRule(className)
	if ok && r.MaxBookings > 0 {
		return r, true
	}
	return OpenRule{}, false
}

// Rule restituisce la regola applicabile a una lezione.
func (s Settings) Rule(className string) OpenRule {
	var def *OpenRule
	for i := range s.OpenRules {
		r := s.OpenRules[i]
		p := trim(r.Pattern)
		if p == "*" {
			if def == nil {
				def = &s.OpenRules[i]
			}
			continue
		}
		if p == "" {
			continue
		}
		if containsFold(className, p) {
			return r
		}
	}
	if def != nil {
		return *def
	}
	return OpenRule{Pattern: "*", DaysBefore: 7, Hour: 5}
}

// WatchState è lo stato di una lezione seguita.
type WatchState string

const (
	StatePending     WatchState = "pending"
	StateBursting    WatchState = "bursting"
	StateWatching    WatchState = "watching"
	StateWaitingList WatchState = "waitingList"
	StateBooked      WatchState = "booked"
	StateFailed      WatchState = "failed"
	StateExpired     WatchState = "expired"
	StateCancelled   WatchState = "cancelled" // disdetta dall'app ufficiale/web mywellness o dal gateway
)

// Terminal indica se lo stato è finale.
func (s WatchState) Terminal() bool {
	return s == StateBooked || s == StateExpired || s == StateCancelled
}

// Item è una lezione che il gateway deve prenotare per un profilo.
type Item struct {
	ID            string     `json:"id"` // profileID|classID|partitionDate
	ProfileID     string     `json:"profileId"`
	ClassID       string     `json:"classId"`
	PartitionDate int        `json:"partitionDate"`
	Name          string     `json:"name"`
	Start         time.Time  `json:"start"`
	End           time.Time  `json:"end"`
	Room          string     `json:"room,omitempty"`
	Trainer       string     `json:"trainer,omitempty"`
	PictureURL    string     `json:"pictureUrl,omitempty"`
	ServerOpensOn *time.Time `json:"serverOpensOn,omitempty"`
	FireAt        *time.Time `json:"fireAt,omitempty"`
	Recurring     bool       `json:"recurring"`
	State         WatchState `json:"state"`
	LastMessage   string     `json:"lastMessage"`
	LastCheck     *time.Time `json:"lastCheck,omitempty"`
	Attempts      int        `json:"attempts"`
	BookedAt      *time.Time `json:"bookedAt,omitempty"`
	Available     *int       `json:"availablePlaces,omitempty"`
	MaxPlaces     *int       `json:"maxParticipants,omitempty"`
	CreatedBy     string     `json:"createdBy"`
	CreatedAt     time.Time  `json:"createdAt"`
}

// RuleKey identifica la ricorrenza settimanale (nome|giorno|ora).
func (it Item) RuleKey(loc *time.Location) string {
	t := it.Start.In(loc)
	return lower(it.Name) + "|" + itoa(int(t.Weekday())) + "|" + itoa(t.Hour()) + ":" + itoa(t.Minute())
}

// LogLine è una riga del registro attività.
type LogLine struct {
	Time      time.Time `json:"time"`
	Level     string    `json:"level"` // info | success | warn | error
	ProfileID string    `json:"profileId,omitempty"`
	Text      string    `json:"text"`
}

// State è l'intero stato persistito.
type State struct {
	Users    []User    `json:"users"`
	Devices  []Device  `json:"devices"`
	Profiles []Profile `json:"profiles"`
	Items    []Item    `json:"items"`
	Settings Settings  `json:"settings"`
	Log      []LogLine `json:"log"`
}
