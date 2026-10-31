// wellness-gateway: prenota le lezioni Technogym mywellness per più profili e avvisa l'app iOS via APNs.
package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/sidimam/wellness-gateway/internal/apns"
	"github.com/sidimam/wellness-gateway/internal/engine"
	"github.com/sidimam/wellness-gateway/internal/people"
	"github.com/sidimam/wellness-gateway/internal/server"
	"github.com/sidimam/wellness-gateway/internal/store"
	"github.com/sidimam/wellness-gateway/internal/telegram"
)

func env(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func main() {
	log.SetFlags(log.Ldate | log.Ltime)
	dataDir := env("CONFIG_DIR", "/config")
	listen := env("LISTEN_ADDR", ":8585")
	trustProxy, _ := strconv.ParseBool(env("TRUST_PROXY", "false"))
	publicURL := strings.TrimRight(env("PUBLIC_URL", ""), "/")

	st, err := store.Open(dataDir)
	if err != nil {
		log.Fatalf("stato: %v", err)
	}

	var push *apns.Client
	cfg := apns.Config{
		KeyPath:  env("APNS_KEY_PATH", ""),
		KeyID:    env("APNS_KEY_ID", ""),
		TeamID:   env("APNS_TEAM_ID", "X5SR67A8AL"),
		BundleID: env("APNS_BUNDLE_ID", "com.sdimambro.wellness-booking"),
	}
	cfg.Production, _ = strconv.ParseBool(env("APNS_PRODUCTION", "true"))
	if cfg.Enabled() {
		if push, err = apns.New(cfg); err != nil {
			log.Printf("APNs disabilitato: %v", err)
			push = nil
		} else {
			log.Printf("APNs attivo (key %s, topic %s, production=%v)", cfg.KeyID, cfg.BundleID, cfg.Production)
		}
	} else {
		log.Printf("APNs non configurato: niente notifiche push (APNS_KEY_PATH, APNS_KEY_ID, APNS_TEAM_ID)")
	}

	eng := engine.New(st, push)
	eng.Version = server.Version
	eng.Telegram = telegram.New(env("TELEGRAM_TOKEN", ""), env("TELEGRAM_CHAT_ID", ""))
	eng.SummaryHour, _ = strconv.Atoi(env("TELEGRAM_SUMMARY_HOUR", "21"))
	if eng.Telegram != nil {
		log.Printf("Telegram attivo: avvisi di prenotazione e riepilogo alle %d:00", eng.SummaryHour)
	} else {
		log.Printf("Telegram non configurato (TELEGRAM_TOKEN, TELEGRAM_CHAT_ID)")
	}
	if pdb, err := people.Open(dataDir); err != nil {
		log.Printf("database soci non disponibile: %v", err)
	} else {
		eng.People = pdb
		defer pdb.Close()
	}
	srv := server.New(st, eng, trustProxy, push != nil, publicURL)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	go eng.Run(ctx)
	go eng.RunPeople(ctx)
	go eng.RunTelegramSummary(ctx)

	hs := &http.Server{Addr: listen, Handler: srv.Handler(), ReadHeaderTimeout: 10 * time.Second}
	go func() {
		log.Printf("wellness-gateway %s in ascolto su %s (dati in %s)", server.Version, listen, dataDir)
		if err := hs.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("http: %v", err)
		}
	}()
	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = hs.Shutdown(shutdownCtx)
}
