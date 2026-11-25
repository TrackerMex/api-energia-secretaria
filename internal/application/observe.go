package application

import (
	"context"
	"time"

	"wialon-sener-pilot/internal/domain"
)

// ObserveWindow returns an aggregate only after all pages of the interval arrive.
func ObserveWindow(ctx context.Context, source domain.PositionSource, unitID int64, from, to time.Time) (domain.ObservationSummary, error) {
	positions, err := source.Positions(ctx, unitID, from, to)
	if err != nil {
		return domain.ObservationSummary{}, err
	}
	return domain.SummarizeObservation(positions), nil
}
