// Command tcp-sener processes a local receiver history through a loopback simulator.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/netip"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"wialon-sener-pilot/internal/application"
	"wialon-sener-pilot/internal/delivery"
	"wialon-sener-pilot/internal/domain"
	"wialon-sener-pilot/internal/historydb"
	"wialon-sener-pilot/internal/unitprofile"
)

func run(ctx context.Context, args []string, output io.Writer) error {
	flags := flag.NewFlagSet("tcp-sener", flag.ContinueOnError)
	uid := flags.String("uid", "", "UID retransmitido de la unidad de prueba")
	profilePath := flags.String("profile", "", "ficha de prueba del receptor")
	historyPath := flags.String("history", "", "SQLite del historial")
	outboxPath := flags.String("outbox", "", "SQLite de la cola del simulador local")
	start := flags.String("from", "", "inicio incluido del intervalo UTC")
	end := flags.String("to", "", "fin excluido del intervalo UTC")
	endpoint := flags.String("endpoint", "", "URL HTTP del simulador con IP de loopback")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if flags.NArg() != 0 || *uid == "" || *profilePath == "" || *historyPath == "" || *outboxPath == "" {
		return errors.New("se requieren -uid, -profile, -history y -outbox")
	}
	address, err := url.Parse(*endpoint)
	if err != nil || address.Scheme != "http" || address.User != nil || address.RawQuery != "" || address.Fragment != "" {
		return errors.New("el simulador requiere una URL HTTP local sin credenciales, consulta o fragmento")
	}
	ip, err := netip.ParseAddr(address.Hostname())
	if err != nil || !ip.IsLoopback() || ip.Zone() != "" {
		return errors.New("el simulador requiere una dirección IP literal de loopback")
	}
	from, err := utc(*start)
	if err != nil {
		return err
	}
	to, err := utc(*end)
	if err != nil {
		return err
	}
	if !from.Before(to) {
		return errors.New("el inicio debe preceder al final del intervalo")
	}
	profile, err := unitprofile.LoadTCP(*profilePath, *uid)
	if err != nil {
		return err
	}
	history, err := historydb.OpenReader(*historyPath, *uid)
	if err != nil {
		return err
	}
	defer history.Close()
	outbox, err := delivery.OpenTCP(*outboxPath, profile.RFC, *uid)
	if err != nil {
		return err
	}
	defer outbox.Close()
	// No environmental proxy or real development credentials enter this adapter.
	transport := &http.Transport{Proxy: nil}
	defer transport.CloseIdleConnections()
	sender := delivery.Sender{URL: address.String(), Token: "local-demo", Client: &http.Client{Transport: transport, Timeout: 10 * time.Second}}
	window := application.TCPWindow{History: history, Outbox: outbox, Sender: sender, Profile: profile}
	if err = window.Run(ctx, from, to); err != nil {
		return err
	}
	counts, err := outbox.Counts(ctx)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(output, "accepted=%d pending=%d rejected=%d invalid=%d uncertain=%d\n", counts[domain.Accepted], counts[domain.Pending], counts[domain.Rejected], counts[domain.Invalid], counts[domain.Uncertain])
	return err
}

func utc(value string) (time.Time, error) {
	t, err := time.Parse(time.RFC3339Nano, value)
	_, offset := t.Zone()
	if err != nil || offset != 0 || t.Unix() < 0 {
		return time.Time{}, errors.New("se requiere una fecha UTC RFC3339 desde 1970")
	}
	return t.UTC(), nil
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err := run(ctx, os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
