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
