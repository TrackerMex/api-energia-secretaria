// Command piloto reads one Wialon unit and delivers position reports to SENER development.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"wialon-sener-pilot/internal/application"
	"wialon-sener-pilot/internal/delivery"
	"wialon-sener-pilot/internal/domain"
	"wialon-sener-pilot/internal/ingest"
)

func loadProfile(path string) (domain.Profile, error) {
	var profile domain.Profile
	file, err := os.Open(path)
	if err != nil {
		return profile, err
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	decoder.DisallowUnknownFields()
	if err = decoder.Decode(&profile); err != nil {
		return profile, errors.New("invalid unit profile JSON")
	}
	return profile, nil
}

func utcInstant(raw string) (time.Time, error) {
	t, err := time.Parse(time.RFC3339, raw)
	_, offset := t.Zone()
	if err != nil || offset != 0 || t.Unix() < 0 || t.Nanosecond() != 0 {
		return time.Time{}, errors.New("provide a UTC timestamp in whole seconds, e.g. 2026-11-18T12:00:00Z")
	}
	return t.UTC(), nil
}

func run(ctx context.Context) error {
	environment := flag.String("environment", "", "live requests require development")
	profilePath := flag.String("profile", "", "path to the private single-unit profile")
	dbPath := flag.String("db", "/var/lib/wialon-sener/pilot.db", "SQLite state file")
	start := flag.String("from", "", "first UTC timestamp (RFC3339)")
	end := flag.String("to", "", "last UTC timestamp for one pass (inclusive)")
	watch := flag.Bool("watch", false, "poll through now using the stored cursor")
	flag.Parse()
	if *environment != "development" || *profilePath == "" || *dbPath == "" || flag.NArg() != 0 {
		return errors.New("specify -environment development, -profile and an explicit UTC window")
	}
	from, err := utcInstant(*start)
	if err != nil {
		return err
	}
	var to time.Time
	var interval time.Duration
	if *watch {
		if *end != "" {
			return errors.New("-watch uses the current time; omit -to")
		}
		interval, err = time.ParseDuration(os.Getenv("WIALON_POLL_INTERVAL"))
		if err != nil || interval <= 0 {
			return errors.New("WIALON_POLL_INTERVAL must be a positive Go duration for -watch")
		}
		to = time.Now().UTC().Truncate(time.Second)
	} else {
		to, err = utcInstant(*end)
		if err != nil {
			return err
		}
	}
	if to.Before(from) || to.After(time.Now()) {
		return errors.New("the UTC window must be ordered and end at or before now")
	}
	unit, err := strconv.ParseInt(os.Getenv("WIALON_UNIT_ID"), 10, 64)
	if err != nil || unit <= 0 {
		return errors.New("invalid WIALON_UNIT_ID")
	}
	profile, err := loadProfile(*profilePath)
	if err != nil {
		return err
	}
	if profile.UnitID != unit {
		return errors.New("the unit in the profile must match WIALON_UNIT_ID")
	}
	base := strings.TrimSpace(os.Getenv("WIALON_BASE_URL"))
	address, err := url.Parse(base)
	if err != nil || address.Scheme != "https" || address.Hostname() == "" || address.User != nil || (address.Path != "" && address.Path != "/") || address.RawQuery != "" || address.Fragment != "" {
		return errors.New("WIALON_BASE_URL requires the HTTPS origin of the Wialon data center")
	}
	token, devToken := os.Getenv("WIALON_TOKEN"), os.Getenv("SENER_DEV_TOKEN")
	if token == "" || devToken == "" {
		return errors.New("WIALON_TOKEN and SENER_DEV_TOKEN are required for development requests")
	}
	client := &http.Client{Timeout: 30 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	source := ingest.Source{BaseURL: base, Token: token, Client: client}
	sender := delivery.Sender{URL: delivery.DevelopmentURL, Token: devToken, Client: client}
	store, err := delivery.Open(*dbPath, unit, profile.RFC)
	if err != nil {
		return err
	}
	defer store.Close()
	pilot := application.Pilot{Outbox: store, Source: source, Sender: sender, Profile: profile}
	initial := from
	for {
		if *watch {
			cursor, err := store.Cursor(ctx)
			if err != nil {
				return err
			}
			if !cursor.IsZero() {
				from = cursor
				if from.Before(initial) {
					from = initial
				}
			}
			to = time.Now().UTC().Truncate(time.Second)
		}
		// Limit each API load and its transaction while catching up a long window.
		for !from.After(to) {
			stop := from.Add(time.Hour)
			if stop.After(to) {
				stop = to
			}
			err = pilot.RunWindow(ctx, from, stop)
			if err != nil {
				if !*watch || ctx.Err() != nil {
					return err
				}
				fmt.Fprintf(os.Stderr, "pilot cycle: %v\n", err)
				break
			}
			if stop.Equal(to) {
				break
			}
			from = stop
		}
		counts, err := store.Counts(ctx)
		if err != nil {
			return err
		}
		fmt.Printf("accepted=%d pending=%d rejected=%d invalid=%d uncertain=%d\n",
			counts[domain.Accepted], counts[domain.Pending], counts[domain.Rejected], counts[domain.Invalid], counts[domain.Uncertain])
		if !*watch {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(interval):
		}
	}
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err := run(ctx); err != nil && !errors.Is(err, context.Canceled) {
		fmt.Fprintln(os.Stderr, "piloto:", err)
		os.Exit(1)
	}
}
