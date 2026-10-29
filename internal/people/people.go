// Package people conserva in SQLite (/config/people.db) i dati PUBBLICI dei soci visti nelle lezioni
// (nickname, foto, profilo privato, presenze, nome e cognome dei profili pubblici) e le classifiche MOVEs del centro.
// Nessun dato privato di terzi: chi ha il profilo privato non ha nome, e dal profilo pubblico non si salvano peso, età,
// sesso né frequenza cardiaca.
package people

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/sidimam/wellness-gateway/internal/mywellness"
	_ "modernc.org/sqlite"
)

// DB è il database dei soci e delle classifiche.
type DB struct {
	db *sql.DB
}

// Open apre (o crea) people.db nella cartella dati.
func Open(dir string) (*DB, error) {
	dsn := "file:" + filepath.Join(dir, "people.db") + "?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)&_pragma=foreign_keys(ON)"
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	for _, q := range schema {
		if _, err := db.Exec(q); err != nil {
			db.Close()
			return nil, fmt.Errorf("schema: %w", err)
		}
	}
	// colonne aggiunte dopo la prima versione (ALTER fallisce se esistono già: si ignora)
	for _, q := range []string{
		`ALTER TABLE members ADD COLUMN first_name TEXT NOT NULL DEFAULT ''`,
		`ALTER TABLE members ADD COLUMN last_name TEXT NOT NULL DEFAULT ''`,
		`ALTER TABLE members ADD COLUMN name_checked TEXT NOT NULL DEFAULT ''`,
	} {
		if _, err := db.Exec(q); err != nil && !strings.Contains(err.Error(), "duplicate column") {
			db.Close()
			return nil, fmt.Errorf("schema: %w", err)
		}
	}
	return &DB{db: db}, nil
}

// NameMaxAge: dopo quanto si rilegge nome e cognome dal profilo pubblico (cambi di nome o profilo reso privato).
const NameMaxAge = 7 * 24 * time.Hour

// NeedName filtra gli id dei soci con profilo pubblico il cui nome non è mai stato letto o è più vecchio di NameMaxAge.
func (d *DB) NeedName(ctx context.Context, ids []string) []string {
	cut := time.Now().Add(-NameMaxAge).UTC().Format(time.RFC3339)
	out := []string{}
	for _, id := range ids {
		var priv int
		var checked string
		if err := d.db.QueryRowContext(ctx, `SELECT private, name_checked FROM members WHERE id=?`, id).Scan(&priv, &checked); err != nil {
			continue
		}
		if priv == 0 && (checked == "" || checked < cut) {
			out = append(out, id)
		}
	}
	return out
}

// MissingNames: fino a limit soci pubblici del centro senza nome letto (o con lettura più vecchia di NameMaxAge).
func (d *DB) MissingNames(ctx context.Context, facilityID string, limit int) []string {
	cut := time.Now().Add(-NameMaxAge).UTC().Format(time.RFC3339)
	rows, err := d.db.QueryContext(ctx, `SELECT id FROM members WHERE facility_id=? AND private=0 AND (name_checked='' OR name_checked<?)
		ORDER BY last_seen DESC LIMIT ?`, facilityID, cut, limit)
	if err != nil {
		return nil
	}
	defer rows.Close()
	out := []string{}
	for rows.Next() {
		var id string
		if rows.Scan(&id) == nil {
			out = append(out, id)
		}
	}
	return out
}

// SetName salva nome e cognome letti dal profilo pubblico; se il profilo è diventato privato li cancella.
func (d *DB) SetName(ctx context.Context, id, first, last string, private bool) error {
	if private {
		first, last = "", ""
	}
	_, err := d.db.ExecContext(ctx, `UPDATE members SET first_name=?, last_name=?, private=CASE WHEN ? THEN 1 ELSE private END, name_checked=? WHERE id=?`,
		first, last, boolInt(private), time.Now().UTC().Format(time.RFC3339), id)
	return err
}

func (d *DB) Close() error { return d.db.Close() }

var schema = []string{
	`CREATE TABLE IF NOT EXISTS members (
		id TEXT PRIMARY KEY, facility_id TEXT NOT NULL DEFAULT '', nick_name TEXT NOT NULL DEFAULT '', picture_url TEXT NOT NULL DEFAULT '',
		private INTEGER NOT NULL DEFAULT 0, first_seen TEXT NOT NULL, last_seen TEXT NOT NULL, extra TEXT NOT NULL DEFAULT '')`,
	`CREATE TABLE IF NOT EXISTS classes (
		class_id TEXT NOT NULL, partition_date INTEGER NOT NULL, facility_id TEXT NOT NULL DEFAULT '', name TEXT NOT NULL DEFAULT '',
		start TEXT NOT NULL DEFAULT '', room TEXT NOT NULL DEFAULT '', trainer TEXT NOT NULL DEFAULT '', participants INTEGER NOT NULL DEFAULT 0,
		max_participants INTEGER NOT NULL DEFAULT 0, waiting INTEGER NOT NULL DEFAULT 0, observed_at TEXT NOT NULL,
		PRIMARY KEY (class_id, partition_date))`,
	`CREATE TABLE IF NOT EXISTS attendance (
		class_id TEXT NOT NULL, partition_date INTEGER NOT NULL, member_id TEXT NOT NULL, joined_on TEXT NOT NULL DEFAULT '',
		station TEXT NOT NULL DEFAULT '', equipment TEXT NOT NULL DEFAULT '', extra TEXT NOT NULL DEFAULT '', observed_at TEXT NOT NULL,
		PRIMARY KEY (class_id, partition_date, member_id))`,
	`CREATE INDEX IF NOT EXISTS attendance_member ON attendance(member_id)`,
	`CREATE TABLE IF NOT EXISTS leaderboard (
		facility_id TEXT NOT NULL, trend TEXT NOT NULL, period TEXT NOT NULL, position INTEGER NOT NULL,
		nick_name TEXT NOT NULL DEFAULT '', picture_url TEXT NOT NULL DEFAULT '', moves INTEGER NOT NULL DEFAULT 0, observed_at TEXT NOT NULL,
		PRIMARY KEY (facility_id, trend, period, position))`,
}

// ClassMeta descrive la lezione di cui si registrano i partecipanti.
type ClassMeta struct {
	ClassID         string
	PartitionDate   int
	FacilityID      string
	Name            string
	Start           time.Time
	Room, Trainer   string
	Participants    int
	MaxParticipants int
	Waiting         int
}

// RecordParticipants salva lezione, soci e presenze (upsert). Le righe di presenza di soci non più iscritti vengono tolte.
func (d *DB) RecordParticipants(ctx context.Context, meta ClassMeta, parts []mywellness.Participant) error {
	tx, err := d.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	now := time.Now().UTC().Format(time.RFC3339)
	if _, err := tx.Exec(`INSERT INTO classes(class_id,partition_date,facility_id,name,start,room,trainer,participants,max_participants,waiting,observed_at)
		VALUES(?,?,?,?,?,?,?,?,?,?,?) ON CONFLICT(class_id,partition_date) DO UPDATE SET facility_id=excluded.facility_id,name=excluded.name,start=excluded.start,
		room=excluded.room,trainer=excluded.trainer,participants=excluded.participants,max_participants=excluded.max_participants,waiting=excluded.waiting,observed_at=excluded.observed_at`,
		meta.ClassID, meta.PartitionDate, meta.FacilityID, meta.Name, meta.Start.Format(time.RFC3339), meta.Room, meta.Trainer, len(parts), meta.MaxParticipants, meta.Waiting, now); err != nil {
		return err
	}
	ids := make([]any, 0, len(parts))
	for _, p := range parts {
		if p.UserID == "" {
			continue
		}
		ids = append(ids, p.UserID)
		extra := ""
		if len(p.Extra) > 0 {
			if b, err := json.Marshal(p.Extra); err == nil {
				extra = string(b)
			}
		}
		if _, err := tx.Exec(`INSERT INTO members(id,facility_id,nick_name,picture_url,private,first_seen,last_seen,extra) VALUES(?,?,?,?,?,?,?,?)
			ON CONFLICT(id) DO UPDATE SET facility_id=CASE WHEN excluded.facility_id<>'' THEN excluded.facility_id ELSE members.facility_id END,
			nick_name=CASE WHEN excluded.nick_name<>'' THEN excluded.nick_name ELSE members.nick_name END,
			picture_url=CASE WHEN excluded.picture_url<>'' THEN excluded.picture_url ELSE members.picture_url END,
			private=excluded.private, last_seen=excluded.last_seen,
			first_name=CASE WHEN excluded.private=1 THEN '' ELSE members.first_name END,
			last_name=CASE WHEN excluded.private=1 THEN '' ELSE members.last_name END, extra=CASE WHEN excluded.extra<>'' THEN excluded.extra ELSE members.extra END`,
			p.UserID, meta.FacilityID, p.NickName, p.PictureURL, boolInt(p.Private), now, now, extra); err != nil {
			return err
		}
		if _, err := tx.Exec(`INSERT INTO attendance(class_id,partition_date,member_id,joined_on,station,equipment,extra,observed_at) VALUES(?,?,?,?,?,?,?,?)
			ON CONFLICT(class_id,partition_date,member_id) DO UPDATE SET joined_on=excluded.joined_on,station=excluded.station,equipment=excluded.equipment,observed_at=excluded.observed_at`,
			meta.ClassID, meta.PartitionDate, p.UserID, p.JoinedOn, p.Station, p.Equipment, extra, now); err != nil {
			return err
		}
	}
	// chi non è più iscritto (ha disdetto) non deve risultare presente
	q := `DELETE FROM attendance WHERE class_id=? AND partition_date=?`
	args := []any{meta.ClassID, meta.PartitionDate}
	if len(ids) > 0 {
		q += ` AND member_id NOT IN (?` + strings.Repeat(",?", len(ids)-1) + `)`
		args = append(args, ids...)
	}
	if _, err := tx.Exec(q, args...); err != nil {
		return err
	}
	return tx.Commit()
}

// RecordLeaderboard salva la classifica di un periodo (sostituendo le posizioni precedenti dello stesso periodo).
func (d *DB) RecordLeaderboard(ctx context.Context, facilityID, trend, period string, entries []mywellness.LeaderboardEntry) error {
	tx, err := d.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.Exec(`DELETE FROM leaderboard WHERE facility_id=? AND trend=? AND period=?`, facilityID, trend, period); err != nil {
		return err
	}
	now := time.Now().UTC().Format(time.RFC3339)
	for _, e := range entries {
		if _, err := tx.Exec(`INSERT INTO leaderboard(facility_id,trend,period,position,nick_name,picture_url,moves,observed_at) VALUES(?,?,?,?,?,?,?,?)`,
			facilityID, trend, period, e.Position, e.NickName, e.PictureURL, e.Moves, now); err != nil {
			return err
		}
	}
	return tx.Commit()
}

// MemberView è un socio con le statistiche di presenza ricavate dal database.
type MemberView struct {
	ID         string         `json:"id"`
	NickName   string         `json:"nickName"`
	FirstName  string         `json:"firstName,omitempty"` // dal profilo pubblico (vuoto se privato o non ancora letto)
	LastName   string         `json:"lastName,omitempty"`
	PictureURL string         `json:"pictureUrl,omitempty"`
	Private    bool           `json:"private"`
	FirstSeen  string         `json:"firstSeen"`
	LastSeen   string         `json:"lastSeen"`
	SeenCount  int            `json:"seenCount"`          // lezioni distinte in cui è stato visto
	Classes    []string       `json:"classes,omitempty"`  // tipi di lezione frequentati (più frequenti prima)
	JoinedOn   string         `json:"joinedOn,omitempty"` // solo nella lista partecipanti di una lezione
	Station    string         `json:"station,omitempty"`
	Equipment  string         `json:"equipment,omitempty"`
	Extra      map[string]any `json:"extra,omitempty"`
	BestWeek   *int           `json:"bestWeekPosition,omitempty"`  // miglior posizione in classifica settimanale (per nickname)
	BestMonth  *int           `json:"bestMonthPosition,omitempty"` // miglior posizione in classifica mensile (per nickname)
}

func (d *DB) decorate(ctx context.Context, m *MemberView) {
	_ = d.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM attendance WHERE member_id=?`, m.ID).Scan(&m.SeenCount)
	rows, err := d.db.QueryContext(ctx, `SELECT c.name, COUNT(*) n FROM attendance a JOIN classes c ON c.class_id=a.class_id AND c.partition_date=a.partition_date
		WHERE a.member_id=? GROUP BY c.name ORDER BY n DESC LIMIT 6`, m.ID)
	if err == nil {
		for rows.Next() {
			var name string
			var n int
			if rows.Scan(&name, &n) == nil {
				m.Classes = append(m.Classes, name)
			}
		}
		rows.Close()
	}
	if m.NickName != "" {
		var w, mo sql.NullInt64
		_ = d.db.QueryRowContext(ctx, `SELECT MIN(position) FROM leaderboard WHERE trend='week' AND nick_name=?`, m.NickName).Scan(&w)
		_ = d.db.QueryRowContext(ctx, `SELECT MIN(position) FROM leaderboard WHERE trend='month' AND nick_name=?`, m.NickName).Scan(&mo)
		if w.Valid {
			v := int(w.Int64)
			m.BestWeek = &v
		}
		if mo.Valid {
			v := int(mo.Int64)
			m.BestMonth = &v
		}
	}
}

func scanMember(sc interface{ Scan(...any) error }, m *MemberView) error {
	var priv int
	var extra string
	if err := sc.Scan(&m.ID, &m.NickName, &m.FirstName, &m.LastName, &m.PictureURL, &priv, &m.FirstSeen, &m.LastSeen, &extra); err != nil {
		return err
	}
	m.Private = priv == 1
	if extra != "" {
		_ = json.Unmarshal([]byte(extra), &m.Extra)
	}
	return nil
}

// ClassParticipants: i partecipanti registrati di una lezione, con statistiche.
func (d *DB) ClassParticipants(ctx context.Context, classID string, partitionDate int) ([]MemberView, string, error) {
	var observed string
	_ = d.db.QueryRowContext(ctx, `SELECT observed_at FROM classes WHERE class_id=? AND partition_date=?`, classID, partitionDate).Scan(&observed)
	rows, err := d.db.QueryContext(ctx, `SELECT m.id,m.nick_name,m.first_name,m.last_name,m.picture_url,m.private,m.first_seen,m.last_seen,m.extra,a.joined_on,a.station,a.equipment
		FROM attendance a JOIN members m ON m.id=a.member_id WHERE a.class_id=? AND a.partition_date=? ORDER BY a.joined_on, m.nick_name`, classID, partitionDate)
	if err != nil {
		return nil, observed, err
	}
	defer rows.Close()
	var out []MemberView
	for rows.Next() {
		var m MemberView
		var priv int
		var extra string
		if err := rows.Scan(&m.ID, &m.NickName, &m.FirstName, &m.LastName, &m.PictureURL, &priv, &m.FirstSeen, &m.LastSeen, &extra, &m.JoinedOn, &m.Station, &m.Equipment); err != nil {
			return nil, observed, err
		}
		m.Private = priv == 1
		if extra != "" {
			_ = json.Unmarshal([]byte(extra), &m.Extra)
		}
		out = append(out, m)
	}
	for i := range out {
		d.decorate(ctx, &out[i])
	}
	if out == nil {
		out = []MemberView{}
	}
	return out, observed, nil
}

// Members: i soci visti in un centro, i più frequenti prima (filtro q su nickname, nome e cognome).
func (d *DB) Members(ctx context.Context, facilityIDs []string, q string, limit int) ([]MemberView, error) {
	if limit <= 0 || limit > 500 {
		limit = 200
	}
	where := "1=1"
	args := []any{}
	if len(facilityIDs) > 0 {
		where += " AND m.facility_id IN (?" + strings.Repeat(",?", len(facilityIDs)-1) + ")"
		for _, f := range facilityIDs {
			args = append(args, f)
		}
	}
	if q = strings.TrimSpace(q); q != "" {
		where += " AND (m.nick_name LIKE ? COLLATE NOCASE OR (m.first_name || ' ' || m.last_name) LIKE ? COLLATE NOCASE OR (m.last_name || ' ' || m.first_name) LIKE ? COLLATE NOCASE)"
		args = append(args, "%"+q+"%", "%"+q+"%", "%"+q+"%")
	}
	args = append(args, limit)
	rows, err := d.db.QueryContext(ctx, `SELECT m.id,m.nick_name,m.first_name,m.last_name,m.picture_url,m.private,m.first_seen,m.last_seen,m.extra FROM members m
		WHERE `+where+` ORDER BY (SELECT COUNT(*) FROM attendance a WHERE a.member_id=m.id) DESC, m.last_seen DESC LIMIT ?`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []MemberView{}
	for rows.Next() {
		var m MemberView
		if err := scanMember(rows, &m); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	for i := range out {
		d.decorate(ctx, &out[i])
	}
	return out, nil
}

// Attendance è una presenza nello storico di un socio.
type Attendance struct {
	ClassID       string `json:"classId"`
	PartitionDate int    `json:"partitionDate"`
	Name          string `json:"name"`
	Start         string `json:"start"`
	Room          string `json:"room,omitempty"`
	Trainer       string `json:"trainer,omitempty"`
	JoinedOn      string `json:"joinedOn,omitempty"`
	Station       string `json:"station,omitempty"`
}

// Member: scheda di un socio con storico presenze e posizioni in classifica.
func (d *DB) Member(ctx context.Context, id string) (*MemberView, []Attendance, []LeaderboardRow, error) {
	var m MemberView
	if err := scanMember(d.db.QueryRowContext(ctx, `SELECT id,nick_name,first_name,last_name,picture_url,private,first_seen,last_seen,extra FROM members WHERE id=?`, id), &m); err != nil {
		return nil, nil, nil, err
	}
	d.decorate(ctx, &m)
	rows, err := d.db.QueryContext(ctx, `SELECT a.class_id,a.partition_date,c.name,c.start,c.room,c.trainer,a.joined_on,a.station FROM attendance a
		JOIN classes c ON c.class_id=a.class_id AND c.partition_date=a.partition_date WHERE a.member_id=? ORDER BY c.start DESC LIMIT 200`, id)
	if err != nil {
		return nil, nil, nil, err
	}
	defer rows.Close()
	att := []Attendance{}
	for rows.Next() {
		var a Attendance
		if err := rows.Scan(&a.ClassID, &a.PartitionDate, &a.Name, &a.Start, &a.Room, &a.Trainer, &a.JoinedOn, &a.Station); err != nil {
			return nil, nil, nil, err
		}
		att = append(att, a)
	}
	lb := []LeaderboardRow{}
	if m.NickName != "" {
		r2, err := d.db.QueryContext(ctx, `SELECT facility_id,trend,period,position,nick_name,picture_url,moves,observed_at FROM leaderboard WHERE nick_name=? ORDER BY period DESC, trend`, m.NickName)
		if err == nil {
			for r2.Next() {
				var l LeaderboardRow
				if r2.Scan(&l.FacilityID, &l.Trend, &l.Period, &l.Position, &l.NickName, &l.PictureURL, &l.Moves, &l.ObservedAt) == nil {
					lb = append(lb, l)
				}
			}
			r2.Close()
		}
	}
	return &m, att, lb, nil
}

// LeaderboardRow è una posizione in classifica salvata.
type LeaderboardRow struct {
	FacilityID string `json:"facilityId"`
	Trend      string `json:"trend"`
	Period     string `json:"period"`
	Position   int    `json:"position"`
	NickName   string `json:"nickName"`
	PictureURL string `json:"pictureUrl,omitempty"`
	Moves      int    `json:"moves"`
	ObservedAt string `json:"observedAt"`
	MemberID   string `json:"memberId,omitempty"` // socio con lo stesso nickname, se visto in una lezione
}

// Leaderboard: l'ultimo periodo salvato (o quello richiesto) per centro e trend.
func (d *DB) Leaderboard(ctx context.Context, facilityID, trend, period string) ([]LeaderboardRow, string, error) {
	if period == "" {
		_ = d.db.QueryRowContext(ctx, `SELECT period FROM leaderboard WHERE facility_id=? AND trend=? ORDER BY period DESC LIMIT 1`, facilityID, trend).Scan(&period)
	}
	if period == "" {
		return []LeaderboardRow{}, "", nil
	}
	rows, err := d.db.QueryContext(ctx, `SELECT l.facility_id,l.trend,l.period,l.position,l.nick_name,l.picture_url,l.moves,l.observed_at,
		COALESCE((SELECT m.id FROM members m WHERE m.nick_name=l.nick_name AND m.nick_name<>'' LIMIT 1),'')
		FROM leaderboard l WHERE l.facility_id=? AND l.trend=? AND l.period=? ORDER BY l.position`, facilityID, trend, period)
	if err != nil {
		return nil, period, err
	}
	defer rows.Close()
	out := []LeaderboardRow{}
	for rows.Next() {
		var l LeaderboardRow
		if err := rows.Scan(&l.FacilityID, &l.Trend, &l.Period, &l.Position, &l.NickName, &l.PictureURL, &l.Moves, &l.ObservedAt, &l.MemberID); err != nil {
			return nil, period, err
		}
		out = append(out, l)
	}
	return out, period, nil
}

// Periods: i periodi salvati per centro e trend (per sfogliare le classifiche passate).
func (d *DB) Periods(ctx context.Context, facilityID, trend string) ([]string, error) {
	rows, err := d.db.QueryContext(ctx, `SELECT DISTINCT period FROM leaderboard WHERE facility_id=? AND trend=? ORDER BY period DESC LIMIT 60`, facilityID, trend)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []string{}
	for rows.Next() {
		var p string
		if rows.Scan(&p) == nil {
			out = append(out, p)
		}
	}
	return out, nil
}

// Stats: numeri per la pagina Persone.
type Stats struct {
	Members    int    `json:"members"`
	Classes    int    `json:"classes"`
	Attendance int    `json:"attendance"`
	LastUpdate string `json:"lastUpdate,omitempty"`
}

func (d *DB) Stats(ctx context.Context, facilityIDs []string) Stats {
	var s Stats
	where := "1=1"
	args := []any{}
	if len(facilityIDs) > 0 {
		where = "facility_id IN (?" + strings.Repeat(",?", len(facilityIDs)-1) + ")"
		for _, f := range facilityIDs {
			args = append(args, f)
		}
	}
	_ = d.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM members WHERE `+where, args...).Scan(&s.Members)
	_ = d.db.QueryRowContext(ctx, `SELECT COUNT(*), COALESCE(MAX(observed_at),'') FROM classes WHERE `+where, args...).Scan(&s.Classes, &s.LastUpdate)
	_ = d.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM attendance a JOIN classes c ON c.class_id=a.class_id AND c.partition_date=a.partition_date WHERE c.`+where, args...).Scan(&s.Attendance)
	return s
}

func boolInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
