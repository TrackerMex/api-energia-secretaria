// Command receptor captures one Wialon Retranslator unit on a loopback interface.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"net/netip"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"wialon-sener-pilot/internal/application"
	"wialon-sener-pilot/internal/domain"
	"wialon-sener-pilot/internal/historydb"
	"wialon-sener-pilot/internal/retranslator"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err := run(ctx, os.Args[1:], os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(ctx context.Context, args []string, stdout, stderr io.Writer) error {
	if len(args) == 0 {
		return errors.New("uso: receptor {listen|latest|history} -uid UID -db archivo")
	}
	mode := args[0]
	flags := flag.NewFlagSet(mode, flag.ContinueOnError)
	flags.SetOutput(stderr)
	uid := flags.String("uid", "", "UID ficticio autorizado del dispositivo CR-13")
	database := flags.String("db", "", "ruta del archivo SQLite del historial")
	var addr, from, to string
	switch mode {
	case "listen":
		flags.StringVar(&addr, "listen", "127.0.0.1:20163", "dirección y puerto de loopback")
	case "history":
		flags.StringVar(&from, "from", "", "inicio UTC incluido (RFC3339)")
		flags.StringVar(&to, "to", "", "final UTC excluido (RFC3339)")
	case "latest":
	default:
		return fmt.Errorf("subcomando desconocido %q", mode)
	}
	if err := flags.Parse(args[1:]); err != nil {
		return err
	}
	if flags.NArg() != 0 || strings.TrimSpace(*uid) == "" || strings.TrimSpace(*database) == "" {
		return errors.New("se requieren -uid y -db sin argumentos posicionales")
	}
	if mode == "listen" {
		local, err := netip.ParseAddrPort(addr)
		if err != nil || !local.Addr().IsLoopback() {
			return errors.New("-listen requiere una dirección IP de loopback y un puerto")
		}
		archive, err := historydb.OpenWriter(*database, *uid)
		if err != nil {
			return err
		}
		defer archive.Close()
		receiver := application.Receiver{UID: *uid, Archive: archive}
		listener, err := net.Listen("tcp", local.String())
		if err != nil {
			return err
		}
		fmt.Fprintln(stderr, "listening", listener.Addr())
		return retranslator.Serve(ctx, listener, receiver.Accept)
	}
	var start, end time.Time
	if mode == "history" {
		var err error
		if start, err = parseUTC(from); err != nil {
			return fmt.Errorf("-from: %w", err)
		}
		if end, err = parseUTC(to); err != nil {
			return fmt.Errorf("-to: %w", err)
		}
	}
	archive, err := historydb.OpenReader(*database, *uid)
	if err != nil {
		return err
	}
	defer archive.Close()
	receiver := application.Receiver{UID: *uid, Archive: archive}
	var fixes []domain.GPSFix
	if mode == "history" {
		fixes, err = receiver.Between(ctx, start, end)
	} else {
		var last *domain.GPSFix
		last, err = receiver.Latest(ctx)
		if last != nil {
			fixes = append(fixes, *last)
		}
	}
	if err != nil {
		return err
	}
	return printFixes(stdout, fixes)
}

func parseUTC(value string) (time.Time, error) {
	t, err := time.Parse(time.RFC3339Nano, value)
	if err != nil {
		return time.Time{}, err
	}
	_, offset := t.Zone()
	if offset != 0 {
		return time.Time{}, errors.New("se requiere una fecha UTC con zona Z o +00:00")
	}
	return t.UTC(), nil
}

type position struct {
	Unit      string    `json:"unit"`
	UID       string    `json:"uid"`
	GPS       time.Time `json:"gps_utc"`
	Latitude  float64   `json:"latitude"`
	Longitude float64   `json:"longitude"`
	Speed     uint16    `json:"speed"`
	Course    uint16    `json:"course"`
}

func printFixes(w io.Writer, fixes []domain.GPSFix) error {
	positions := make([]position, 0, len(fixes))
	for _, fix := range fixes {
		positions = append(positions, position{
			Unit: domain.ReceiverUnit, UID: fix.UID, GPS: fix.GPSTime.UTC(), Latitude: fix.Latitude,
			Longitude: fix.Longitude, Speed: fix.Speed, Course: fix.Course,
		})
	}
	return json.NewEncoder(w).Encode(positions)
}
