package retranslator

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sync"

	"wialon-sener-pilot/internal/domain"
)

// Serve owns listener. Context cancellation closes idle peers and waits for all
// handlers. A storage failure stops the server and is returned to its caller.
func Serve(parent context.Context, listener net.Listener, handle func(context.Context, domain.GPSFix) error) error {
	return ServeWithProfile(parent, listener, handle, domain.Profile{})
}

// ServeWithProfile captures the selected measurements from each individual frame.
// A zero profile preserves the receiver's position-only operation.
func ServeWithProfile(parent context.Context, listener net.Listener, handle func(context.Context, domain.GPSFix) error, profile domain.Profile) error {
	if listener == nil || handle == nil {
		return errors.New("falta listener o manejador de posiciones")
	}
	ctx, cancel := context.WithCancelCause(parent)
	defer cancel(nil)
	defer listener.Close()
	stopClose := context.AfterFunc(ctx, func() { listener.Close() })
	defer stopClose()
	var workers sync.WaitGroup
	for {
		conn, err := listener.Accept()
		if err != nil {
			cancel(fmt.Errorf("aceptar conexión Wialon: %w", err))
			break
		}
		workers.Add(1)
		go func() {
			defer workers.Done()
			if err := receive(ctx, conn, handle, profile); err != nil {
				cancel(fmt.Errorf("guardar posición Wialon: %w", err))
			}
		}()
	}
	workers.Wait()
	cause := context.Cause(ctx)
	if parent.Err() != nil && errors.Is(cause, context.Cause(parent)) {
		return nil
	}
	return cause
}

func receive(ctx context.Context, conn net.Conn, handle func(context.Context, domain.GPSFix) error, profile domain.Profile) error {
	defer conn.Close()
	stopClose := context.AfterFunc(ctx, func() { conn.Close() })
	defer stopClose()
	for ctx.Err() == nil {
		frame, err := readFrame(conn)
		if err != nil {
			return nil
		} // The stream ended or the next boundary is unknown.
		fix, present, err := decodeWithProfile(frame, profile)
		if err != nil {
			continue
		} // Complete frame: the following one starts at a known boundary.
		if present {
			if err := handle(ctx, fix); err != nil {
				return err
			}
		}
	}
	return nil
}
