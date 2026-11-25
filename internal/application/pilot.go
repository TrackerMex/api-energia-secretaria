// Package application coordinates the single-unit pilot using domain ports.
package application

import (
	"context"
	"time"

	"wialon-sener-pilot/internal/domain"
)

type Pilot struct {
	Source  domain.PositionSource
	Outbox  domain.Outbox
	Sender  domain.ReportSender
	Profile domain.Profile
}

func (p Pilot) RunWindow(ctx context.Context, from, to time.Time) error {
	// Previously captured reports can progress even if Wialon is unavailable.
	if err := p.DeliverPending(ctx); err != nil {
		return err
	}
	positions, err := p.Source.Positions(ctx, p.Profile.UnitID, from, to)
	if err != nil {
		return err
	}
	inputs := make([]domain.Input, 0, len(positions))
	for _, position := range positions {
		inputs = append(inputs, domain.Input{Key: position.Key, Build: func(id string) ([]byte, error) { return domain.Report(position, p.Profile, id) }})
	}
	if err = p.Outbox.SaveWindow(ctx, from, to, inputs); err != nil {
		return err
	}
	return p.DeliverPending(ctx)
}

func (p Pilot) DeliverPending(ctx context.Context) error {
	return deliverPending(ctx, p.Outbox, p.Sender, p.Profile.RFC)
}

func deliverPending(ctx context.Context, outbox domain.Outbox, sender domain.ReportSender, rfc string) error {
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		row, err := outbox.Claim(ctx, time.Now())
		if err != nil || row == nil {
			return err
		}
		out := sender.Send(ctx, rfc, row.Body)
		if err = outbox.Finish(ctx, row.ID, out); err != nil {
			return err
		}
		if out.State == domain.Pending {
			return nil
		}
	}
}
