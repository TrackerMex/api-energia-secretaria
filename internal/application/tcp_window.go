package application

import (
	"context"
	"errors"
	"fmt"
	"time"

	"wialon-sener-pilot/internal/domain"
)

// TCPWindow rebuilds the operator's explicit historical selection on every pass.
type TCPWindow struct {
	History domain.PositionArchive
	Outbox  domain.Outbox
	Sender  domain.ReportSender
	Profile domain.Profile
}

func (w TCPWindow) Run(ctx context.Context, from, to time.Time) error {
	if !from.Before(to) || w.Profile.RetranslatorUID == "" {
		return errors.New("se requieren un intervalo y un UID de retransmisión")
	}
	fixes, err := w.History.Between(ctx, w.Profile.RetranslatorUID, from, to)
	if err != nil {
		return err
	}
	inputs := make([]domain.Input, 0, len(fixes))
	for _, fix := range fixes {
		if fix.UID != w.Profile.RetranslatorUID {
			return errors.New("el historial devolvió una unidad diferente")
		}
		key := fmt.Sprintf("tcp:%s:%x", fix.UID, fix.Fingerprint)
		inputs = append(inputs, domain.Input{Key: key, Build: func(id string) ([]byte, error) { return domain.TCPReport(fix, w.Profile, id) }})
	}
	if err = w.Outbox.SaveWindow(ctx, from, to, inputs); err != nil {
		return err
	}
	return deliverPending(ctx, w.Outbox, w.Sender, w.Profile.RFC)
}
