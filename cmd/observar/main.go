// Command observar summarizes engine signals from one authorized Wialon unit.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"wialon-sener-pilot/internal/application"
	"wialon-sener-pilot/internal/ingest"
)

const usage = "uso: observar -from UTC -to UTC (RFC3339 en segundos, ambos extremos incluidos)"

func utcSecond(text string) (time.Time, error) {
	instant, err := time.Parse(time.RFC3339, text)
	_, offset := instant.Zone()
	if err != nil || offset != 0 || instant.Nanosecond() != 0 || instant.Unix() < 0 || instant.Unix() > math.MaxUint32 {
		return time.Time{}, errors.New("se requiere una fecha UTC RFC3339 en segundos dentro del rango de Wialon")
	}
	return instant.UTC(), nil
}

func run(ctx context.Context, args []string, getenv func(string) string, client *http.Client, output io.Writer) error {
	flags := flag.NewFlagSet("observar", flag.ContinueOnError)
	// The flag package may echo unrecognized command-line arguments.
	flags.SetOutput(io.Discard)
	start := flags.String("from", "", "inicio UTC incluido")
	end := flags.String("to", "", "fin UTC incluido")
	if err := flags.Parse(args); errors.Is(err, flag.ErrHelp) {
		_, err = fmt.Fprintln(output, usage)
		return err
	} else if err != nil || flags.NArg() != 0 {
		return errors.New(usage)
	}
	from, err := utcSecond(*start)
	if err != nil {
		return err
	}
	to, err := utcSecond(*end)
	if err != nil {
		return err
	}
	if to.Before(from) {
		return errors.New("el fin del intervalo debe ser igual o posterior al inicio")
	}
	unitID, err := strconv.ParseInt(getenv("WIALON_UNIT_ID"), 10, 64)
	if err != nil || unitID <= 0 {
		return errors.New("WIALON_UNIT_ID requiere un ID interno válido")
	}
	origin, err := url.Parse(strings.TrimSpace(getenv("WIALON_BASE_URL")))
	if err != nil || origin.Scheme != "https" || origin.Hostname() == "" || origin.User != nil || (origin.Path != "" && origin.Path != "/") || origin.RawQuery != "" || origin.ForceQuery || origin.Fragment != "" {
		return errors.New("WIALON_BASE_URL requiere el origen HTTPS del centro de datos")
	}
	token := getenv("WIALON_TOKEN")
	if strings.TrimSpace(token) == "" {
		return errors.New("falta WIALON_TOKEN")
	}
	source := ingest.Source{BaseURL: origin.String(), Token: token, Client: client}
	summary, err := application.ObserveWindow(ctx, source, unitID, from, to)
	if err != nil {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		return errors.New("no se pudo completar la consulta de posiciones de Wialon")
	}
	return json.NewEncoder(output).Encode(summary)
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	transport := &http.Transport{}
	defer transport.CloseIdleConnections()
	client := &http.Client{Transport: transport, Timeout: 30 * time.Second}
	if err := run(ctx, os.Args[1:], os.Getenv, client, os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
