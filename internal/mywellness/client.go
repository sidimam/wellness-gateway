// Package mywellness è il client per le API usate dalla web app Technogym mywellness
// (endpoint ricavati dal bundle di widgets.mywellness.com).
package mywellness

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	AppID      = "EC1D38D7-D359-48D0-A60C-D8C0B8FB9DF9"
	AppName    = "enduserweb"
	AppVersion = "1.0.0"
	Culture    = "it-IT"
	coreURL    = "https://core.mywellness.com"
	calURL     = "https://calendar.mywellness.com"
	svcURL     = "https://services.mywellness.com"
)

// Rome è il fuso del centro.
var Rome = mustLoad("Europe/Rome")

func mustLoad(name string) *time.Location {
	l, err := time.LoadLocation(name)
	if err != nil {
		return time.FixedZone("CET", 3600)
	}
	return l
}

// Client è senza stato: token e userId sono passati a ogni chiamata (un profilo = una Session).
type Client struct {
	HTTP *http.Client
}

// New crea un client con timeout brevi (le chiamate devono essere rapide).
func New() *Client {
	return &Client{HTTP: &http.Client{Timeout: 15 * time.Second}}
}

// Session identifica un profilo autenticato.
type Session struct {
	Token  string
	UserID string
}

// APIError è un errore applicativo restituito dal server.
type APIError struct {
	Status  int
	Message string
	Field   string
}

func (e *APIError) Error() string { return fmt.Sprintf("mywellness %d: %s", e.Status, e.Message) }

// ErrUnauthorized indica token scaduto o credenziali rifiutate.
var ErrUnauthorized = errors.New("sessione mywellness non valida")

type apiErr struct {
	Type         string `json:"type"`
	ErrorMessage string `json:"errorMessage"`
	Message      string `json:"message"`
	Field        string `json:"field"`
	Details      string `json:"details"`
}

func (e apiErr) text() string {
	for _, s := range []string{e.ErrorMessage, e.Message, e.Details, e.Field, e.Type} {
		if s != "" {
			return s
		}
	}
	return "errore"
}

type errEnvelope struct {
	Errors []apiErr `json:"errors"`
}

func (c *Client) do(ctx context.Context, method, base, path string, query url.Values, body any, sess *Session) ([]byte, *http.Response, error) {
	if query == nil {
		query = url.Values{}
	}
	query.Set("_c", Culture)
	req, err := http.NewRequestWithContext(ctx, method, base+path+"?"+query.Encode(), nil)
	if err != nil {
		return nil, nil, err
	}
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, nil, err
		}
		req.Body = io.NopCloser(bytes.NewReader(b))
		req.ContentLength = int64(len(b))
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("X-MWAPPS-APPID", AppID)
	req.Header.Set("X-MWAPPS-CLIENT", AppName)
	req.Header.Set("X-MWAPPS-CLIENTVERSION", AppVersion+","+AppName)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Origin", "https://widgets.mywellness.com")
	req.Header.Set("Referer", "https://widgets.mywellness.com/")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.5 Safari/605.1.15")
	if sess != nil && sess.Token != "" {
		req.Header.Set("Authorization", "Bearer "+sess.Token)
	}
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(io.LimitReader(resp.Body, 20<<20))
	return data, resp, err
}

func errorsIn(data []byte) []apiErr {
	var env errEnvelope
	if json.Unmarshal(data, &env) == nil && len(env.Errors) > 0 {
		return env.Errors
	}
	return nil
}

// LoginResult è l'esito del login, con i dati del profilo mywellness.
type LoginResult struct {
	Session     Session
	DisplayName string
	FirstName   string
	LastName    string
	NickName    string
	Email       string
	PictureURL  string
	ThumbURL    string
}

// Login autentica con email/username e password.
func (c *Client) Login(ctx context.Context, username, password string) (*LoginResult, error) {
	body := map[string]any{"username": username, "password": password, "keepMeLoggedIn": true}
	data, resp, err := c.do(ctx, http.MethodPost, coreURL, "/v2/enduser/authentication/login", nil, body, nil)
	if err != nil {
		return nil, err
	}
	var out struct {
		Result      string `json:"result"`
		Token       string `json:"token"`
		UserContext *struct {
			ID              json.RawMessage `json:"id"`
			FirstName       string          `json:"firstName"`
			LastName        string          `json:"lastName"`
			NickName        string          `json:"nickName"`
			Email           string          `json:"email"`
			PictureURL      string          `json:"pictureUrl"`
			ThumbPictureURL string          `json:"thumbPictureUrl"`
		} `json:"userContext"`
		AccountLockedInfo *struct {
			BlockedFor int `json:"blockedFor"`
		} `json:"accountLockedInfo"`
	}
	_ = json.Unmarshal(data, &out)
	if errs := errorsIn(data); len(errs) > 0 {
		return nil, &APIError{Status: resp.StatusCode, Message: joinErrs(errs), Field: errs[0].Field}
	}
	if resp.StatusCode == http.StatusUnauthorized && strings.EqualFold(out.Result, "mfarequired") {
		return nil, &APIError{Status: 401, Message: "l'account richiede un codice MFA (verifica in due passaggi): non supportato"}
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, &APIError{Status: resp.StatusCode, Message: "login rifiutato (HTTP " + fmt.Sprint(resp.StatusCode) + ")"}
	}
	token := out.Token
	if token == "" {
		if h := resp.Header.Get("Authorization"); strings.HasPrefix(h, "Bearer ") {
			token = strings.TrimPrefix(h, "Bearer ")
		}
	}
	if out.UserContext == nil || token == "" {
		if out.AccountLockedInfo != nil {
			return nil, &APIError{Status: resp.StatusCode, Message: fmt.Sprintf("account bloccato per %d minuti (troppi tentativi)", out.AccountLockedInfo.BlockedFor)}
		}
		msg := "credenziali non valide"
		if out.Result != "" {
			msg = "login fallito: " + out.Result
		}
		return nil, &APIError{Status: resp.StatusCode, Message: msg}
	}
	id := strings.Trim(string(out.UserContext.ID), `"`)
	name := strings.TrimSpace(out.UserContext.FirstName + " " + out.UserContext.LastName)
	if name == "" {
		name = out.UserContext.NickName
	}
	if name == "" {
		name = username
	}
	uc := out.UserContext
	pic := strings.Replace(uc.PictureURL, "http://", "https://", 1)
	thumb := strings.Replace(uc.ThumbPictureURL, "http://", "https://", 1)
	return &LoginResult{Session: Session{Token: token, UserID: id}, DisplayName: name, FirstName: uc.FirstName, LastName: uc.LastName,
		NickName: uc.NickName, Email: uc.Email, PictureURL: pic, ThumbURL: thumb}, nil
}

func joinErrs(errs []apiErr) string {
	parts := make([]string, 0, len(errs))
	for _, e := range errs {
		parts = append(parts, e.text())
	}
	return strings.Join(parts, "; ")
}

// LoginStatus verifica se il token è ancora valido (il widget lo fa ogni 10 minuti).
// Nota: con token scaduto le altre API rispondono 200 "come anonimo", quindi questo controllo è indispensabile.
func (c *Client) LoginStatus(ctx context.Context, sess *Session) (bool, error) {
	if sess == nil || sess.Token == "" {
		return false, nil
	}
	data, resp, err := c.do(ctx, http.MethodPost, svcURL, "/application/"+AppID+"/GetLoginStatus", nil, map[string]any{}, sess)
	if err != nil {
		return false, err
	}
	if resp.StatusCode == http.StatusUnauthorized {
		return false, nil
	}
	for _, e := range errorsIn(data) {
		if e.Field == "TokenNotValid" || e.Details == "TokenNotValid" || e.ErrorMessage == "TokenNotValid" {
			return false, nil
		}
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return false, &APIError{Status: resp.StatusCode, Message: "stato sessione non disponibile"}
	}
	return true, nil
}

// Facility è un centro.
type Facility struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	City string `json:"city"`
	URL  string `json:"url"`
}

// FacilityDetail risolve l'URL del widget (es. "wellnesstown").
func (c *Client) FacilityDetail(ctx context.Context, facilityURL string) (*Facility, error) {
	q := url.Values{"facilityUrl": {facilityURL}}
	data, resp, err := c.do(ctx, http.MethodGet, coreURL, "/v2/enduser/facility/detail", q, nil, nil)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, &APIError{Status: resp.StatusCode, Message: "centro '" + facilityURL + "' non trovato"}
	}
	var f Facility
	if err := json.Unmarshal(data, &f); err != nil || f.ID == "" {
		return nil, &APIError{Status: resp.StatusCode, Message: "risposta centro non riconosciuta"}
	}
	return &f, nil
}

// BookingInfo sono le regole di prenotazione di una lezione.
type BookingInfo struct {
	BookingOpensOn               string `json:"bookingOpensOn"`
	CancellationMinutesInAdvance int    `json:"cancellationMinutesInAdvance"`
	BookingCloseMinutesInAdvance int    `json:"bookingCloseMinutesInAdvance"`
	BookingHasWaitingList        bool   `json:"bookingHasWaitingList"`
	BookingTimeInAdvanceValue    int    `json:"bookingTimeInAdvanceValue"`
	BookingUserStatus            string `json:"bookingUserStatus"`
	BookingAvailable             bool   `json:"bookingAvailable"`
	DayInAdvanceStartHour        int    `json:"dayInAdvanceStartHour"`
	DayInAdvanceStartMinutes     int    `json:"dayInAdvanceStartMinutes"`
}

// ClassEvent è un'occorrenza di lezione.
type ClassEvent struct {
	ID                   string       `json:"id"`
	Name                 string       `json:"name"`
	StartDate            string       `json:"startDate"`
	EndDate              string       `json:"endDate"`
	PartitionDate        int          `json:"partitionDate"`
	Room                 string       `json:"room"`
	AssignedTo           string       `json:"assignedTo"`
	MaxParticipants      int          `json:"maxParticipants"`
	NumberOfParticipants int          `json:"numberOfParticipants"`
	AvailablePlaces      int          `json:"availablePlaces"`
	IsParticipant        bool         `json:"isParticipant"`
	IsInWaitingList      bool         `json:"isInWaitingList"`
	WaitingListPosition  int          `json:"waitingListPosition"`
	PictureURL           string       `json:"pictureUrl"`
	BookingInfo          *BookingInfo `json:"bookingInfo"`
}

// Key identifica l'occorrenza.
func (e ClassEvent) Key() string { return fmt.Sprintf("%s|%d", e.ID, e.PartitionDate) }

// Start restituisce l'inizio (ora del centro).
func (e ClassEvent) Start() time.Time { return parseLocal(e.StartDate) }

// End restituisce la fine.
func (e ClassEvent) End() time.Time { return parseLocal(e.EndDate) }

// OpensOn restituisce l'apertura prenotazioni comunicata dal server, se presente.
func (e ClassEvent) OpensOn() *time.Time {
	if e.BookingInfo == nil || e.BookingInfo.BookingOpensOn == "" {
		return nil
	}
	if t, err := time.Parse(time.RFC3339, e.BookingInfo.BookingOpensOn); err == nil {
		return &t
	}
	t := parseLocal(e.BookingInfo.BookingOpensOn)
	if t.IsZero() {
		return nil
	}
	return &t
}

func parseLocal(s string) time.Time {
	if len(s) >= 19 {
		s = s[:19]
	}
	t, err := time.ParseInLocation("2006-01-02T15:04:05", s, Rome)
	if err != nil {
		return time.Time{}
	}
	return t
}

// Search restituisce le lezioni tra due date (incluse). Con sessione popola IsParticipant.
func (c *Client) Search(ctx context.Context, facilityID string, from, to time.Time, sess *Session) ([]ClassEvent, error) {
	q := url.Values{
		"eventTypes": {"Class"},
		"facilityId": {facilityID},
		"fromDate":   {from.In(Rome).Format("2006-01-02")},
		"toDate":     {to.In(Rome).Format("2006-01-02")},
	}
	data, resp, err := c.do(ctx, http.MethodGet, calURL, "/v2/enduser/class/Search", q, nil, sess)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == http.StatusUnauthorized {
		return nil, ErrUnauthorized
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, &APIError{Status: resp.StatusCode, Message: "errore calendario"}
	}
	var list []ClassEvent
	if err := json.Unmarshal(data, &list); err == nil {
		return list, nil
	}
	var wrapped struct {
		Data []ClassEvent `json:"data"`
	}
	if err := json.Unmarshal(data, &wrapped); err == nil && wrapped.Data != nil {
		return wrapped.Data, nil
	}
	if errs := errorsIn(data); len(errs) > 0 {
		return nil, &APIError{Status: resp.StatusCode, Message: joinErrs(errs)}
	}
	return nil, &APIError{Status: resp.StatusCode, Message: "formato calendario non riconosciuto"}
}

// Outcome è l'esito di una prenotazione.
type Outcome string

const (
	Booked       Outcome = "booked"
	WaitingList  Outcome = "waitingList"
	Full         Outcome = "full"
	NotOpenYet   Outcome = "notOpenYet"
	Unauthorized Outcome = "unauthorized"
	NoPermission Outcome = "noPermission"
	Failed       Outcome = "failed"
)

// BookResult è esito + messaggio del server.
type BookResult struct {
	Outcome Outcome
	Message string
}

// Book prenota una lezione (o entra in lista d'attesa se piena).
func (c *Client) Book(ctx context.Context, sess *Session, classID string, partitionDate int) (BookResult, error) {
	if sess == nil || sess.Token == "" || sess.UserID == "" {
		return BookResult{Outcome: Unauthorized}, nil
	}
	body := map[string]any{"partitionDate": partitionDate, "userId": sess.UserID, "classId": classID, "station": nil}
	data, resp, err := c.do(ctx, http.MethodPost, calURL, "/v2/enduser/class/Book", nil, body, sess)
	if err != nil {
		return BookResult{}, err
	}
	if resp.StatusCode == http.StatusUnauthorized {
		return BookResult{Outcome: Unauthorized}, nil
	}
	if errs := errorsIn(data); len(errs) > 0 {
		for _, e := range errs {
			if e.Type == "Security" && (e.Field == "TokenNotValid" || e.Details == "TokenNotValid") {
				return BookResult{Outcome: Unauthorized}, nil
			}
		}
		field := errs[0].Field + " " + errs[0].Details
		msg := joinErrs(errs)
		if strings.Contains(field, "NoPermissionsForUserException") {
			return BookResult{Outcome: NoPermission, Message: msg}, nil
		}
		l := strings.ToLower(msg + " " + field)
		if strings.Contains(l, "not open") || strings.Contains(l, "non ancora") || strings.Contains(l, "notopen") ||
			strings.Contains(l, "tooearly") || strings.Contains(l, "inadvance") || strings.Contains(l, "aperta") {
			return BookResult{Outcome: NotOpenYet, Message: msg}, nil
		}
		return BookResult{Outcome: Failed, Message: msg}, nil
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return BookResult{Outcome: Failed, Message: fmt.Sprintf("HTTP %d", resp.StatusCode)}, nil
	}
	var out struct {
		Result  string `json:"result"`
		Message string `json:"message"`
	}
	_ = json.Unmarshal(data, &out)
	switch out.Result {
	case "Booked":
		return BookResult{Outcome: Booked}, nil
	case "UserAddedToWaitingList":
		return BookResult{Outcome: WaitingList}, nil
	case "PlaceNotAvailable", "ToMuchParticipants":
		return BookResult{Outcome: Full, Message: out.Result}, nil
	default:
		return BookResult{Outcome: Failed, Message: strings.TrimSpace(out.Result + " " + out.Message)}, nil
	}
}

// Unbook cancella una prenotazione. Esiti del server: UnBooked, TooLate, BookingNotAvailable, EventNotExists, UserNotExists, Failed.
func (c *Client) Unbook(ctx context.Context, sess *Session, classID string, partitionDate int) error {
	body := map[string]any{"partitionDate": partitionDate, "userId": sess.UserID, "classId": classID}
	data, resp, err := c.do(ctx, http.MethodPost, calURL, "/v2/enduser/class/Unbook", nil, body, sess)
	if err != nil {
		return err
	}
	if resp.StatusCode == http.StatusUnauthorized {
		return ErrUnauthorized
	}
	if errs := errorsIn(data); len(errs) > 0 {
		return &APIError{Status: resp.StatusCode, Message: joinErrs(errs)}
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return &APIError{Status: resp.StatusCode, Message: "disdetta rifiutata"}
	}
	var out struct {
		Result  string `json:"result"`
		Message string `json:"message"`
	}
	_ = json.Unmarshal(data, &out)
	switch out.Result {
	case "", "UnBooked":
		return nil
	case "TooLate":
		return &APIError{Status: 409, Message: "troppo tardi per disdire (la cancellazione è consentita fino a 2 ore prima)"}
	case "BookingNotAvailable":
		return &APIError{Status: 409, Message: "disdetta non disponibile per questa lezione"}
	default:
		return &APIError{Status: 409, Message: "disdetta rifiutata: " + strings.TrimSpace(out.Result+" "+out.Message)}
	}
}

// LeaveWaitingList esce dalla lista d'attesa. Esiti: Removed, UserNotInWaitingList, Failed.
func (c *Client) LeaveWaitingList(ctx context.Context, sess *Session, classID string, partitionDate int) error {
	body := map[string]any{"partitionDate": fmt.Sprint(partitionDate), "userId": sess.UserID}
	data, resp, err := c.do(ctx, http.MethodPost, svcURL, "/core/calendarevent/"+classID+"/RemoveFromWaitingList", nil, body, sess)
	if err != nil {
		return err
	}
	if resp.StatusCode == http.StatusUnauthorized {
		return ErrUnauthorized
	}
	if errs := errorsIn(data); len(errs) > 0 {
		return &APIError{Status: resp.StatusCode, Message: joinErrs(errs)}
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return &APIError{Status: resp.StatusCode, Message: "uscita dalla lista d'attesa rifiutata"}
	}
	var wrapped struct {
		Data json.RawMessage `json:"data"`
	}
	raw := strings.Trim(string(data), "\" \n")
	if json.Unmarshal(data, &wrapped) == nil && len(wrapped.Data) > 0 {
		raw = strings.Trim(string(wrapped.Data), "\"")
	}
	switch raw {
	case "Removed", "UserNotInWaitingList", "":
		return nil
	default:
		return &APIError{Status: 409, Message: "uscita dalla lista d'attesa fallita: " + raw}
	}
}
