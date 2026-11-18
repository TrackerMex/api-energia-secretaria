package application

import (
	"context"
	"errors"
	"time"

	"wialon-sener-pilot/internal/domain"
)

// Receiver owns the unit selection used by both capture and local queries.
type Receiver struct {
	UID     string
	Archive domain.PositionArchive
}

func (r Receiver) Accept(ctx context.Context, fix domain.GPSFix) error {
	if fix.UID != r.UID {
		return nil
	}
	return r.Archive.Save(ctx, fix)
}

func (r Receiver) Latest(ctx context.Context) (*domain.GPSFix, error) {
	return r.Archive.Latest(ctx, r.UID)
}

func (r Receiver) Between(ctx context.Context, from, to time.Time) ([]domain.GPSFix, error) {
	if !from.Before(to) {
		return nil, errors.New("el inicio debe preceder al final del intervalo")
	}
	return r.Archive.Between(ctx, r.UID, from, to)
}
