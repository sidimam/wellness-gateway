package mywellness

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

// Il login deve conservare tutto il userContext (identity) senza token/password e con le foto in https.
func TestLoginKeepsIdentity(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/enduser/authentication/login" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"result":"Success","token":"tok123","userContext":{"id":"u1","firstName":"Daniela","lastName":"Manieri","nickName":"dani","email":"d@example.com","pictureUrl":"http://img/p.jpg","thumbPictureUrl":"http://img/t.jpg","gender":"F","birthDate":"1975-01-02T00:00:00","culture":"it-IT","measurementSystem":"Metric","refreshToken":"secret-rt","somePassword":"x","currentFacilityId":"fac1"}}`))
	}))
	defer srv.Close()
	old := coreURL
	coreURL = srv.URL
	defer func() { coreURL = old }()

	res, err := New().Login(context.Background(), "d@example.com", "pw")
	if err != nil {
		t.Fatalf("login: %v", err)
	}
	if res.Session.Token != "tok123" || res.Session.UserID != "u1" || res.DisplayName != "Daniela Manieri" || res.PictureURL != "https://img/p.jpg" {
		t.Fatalf("campi base errati: %+v", res)
	}
	if res.Identity["gender"] != "F" || res.Identity["culture"] != "it-IT" || res.Identity["currentFacilityId"] != "fac1" || res.Identity["pictureUrl"] != "https://img/p.jpg" {
		t.Fatalf("identity incompleta: %v", res.Identity)
	}
	for _, k := range []string{"refreshToken", "somePassword"} {
		if _, ok := res.Identity[k]; ok {
			t.Fatalf("identity contiene %s", k)
		}
	}
}

func TestHasPlace(t *testing.T) {
	cases := []struct {
		ev   ClassEvent
		want bool
	}{
		{ClassEvent{AvailablePlaces: 0, MaxParticipants: 7, NumberOfParticipants: 7}, false},
		{ClassEvent{AvailablePlaces: 1, MaxParticipants: 7, NumberOfParticipants: 7}, true},
		{ClassEvent{AvailablePlaces: 0, MaxParticipants: 7, NumberOfParticipants: 6}, true},
		{ClassEvent{MaxParticipants: 7, NumberOfParticipants: 7, BookingInfo: &BookingInfo{BookingUserStatus: "CanBook"}}, true},
		{ClassEvent{MaxParticipants: 7, NumberOfParticipants: 7, BookingInfo: &BookingInfo{BookingUserStatus: "CannotBook"}}, false},
	}
	for i, c := range cases {
		if got := c.ev.HasPlace(); got != c.want {
			t.Errorf("case %d: HasPlace=%v want %v", i, got, c.want)
		}
	}
}

// Dal profilo pubblico si prendono solo nome e cognome; con profilo privato niente nome.
func TestPublicProfileName(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.Header.Get("Authorization") != "Bearer tok" {
			http.Error(w, "no", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/Core/User/priv/PublicProfile" {
			_, _ = w.Write([]byte(`{"data":{"userId":"priv","firstName":"X","lastName":"Y","hasPrivateProfile":true}}`))
			return
		}
		_, _ = w.Write([]byte(`{"data":{"userId":"u1","nickName":"dani","firstName":" Daniela ","lastName":"Manieri","weight":70,"age":50,"fcMax":170,"hasPrivateProfile":false}}`))
	}))
	defer srv.Close()
	old := svcURL
	svcURL = srv.URL
	defer func() { svcURL = old }()

	c, sess := New(), &Session{Token: "tok"}
	pn, err := c.PublicProfile(context.Background(), sess, "u1")
	if err != nil || pn.FirstName != "Daniela" || pn.LastName != "Manieri" || pn.Private {
		t.Fatalf("pubblico: %+v %v", pn, err)
	}
	pn, err = c.PublicProfile(context.Background(), sess, "priv")
	if err != nil || pn.FirstName != "" || pn.LastName != "" || !pn.Private {
		t.Fatalf("privato: %+v %v", pn, err)
	}
	if _, err := c.PublicProfile(context.Background(), nil, "u1"); err != ErrUnauthorized {
		t.Fatalf("senza sessione: %v", err)
	}
}
