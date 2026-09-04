package store

import (
	"context"

	"github.com/gongcha-cup/backend/internal/domain"
)

// EventRepository persists events and their demand lines.
type EventRepository struct {
	db DBTX
}

// NewEventRepository creates an EventRepository bound to db.
func NewEventRepository(db DBTX) *EventRepository { return &EventRepository{db: db} }

// GetDemands returns all demand lines for an event.
func (r *EventRepository) GetDemands(ctx context.Context, eventID string) ([]domain.EventDemand, error) {
	rows, err := r.db.Query(ctx, `
		SELECT event_id, product_id, requested_quantity, unit
		FROM event_demands WHERE event_id=$1
		ORDER BY product_id`, eventID)
	if err != nil {
		return nil, mapError(err)
	}
	defer rows.Close()
	var out []domain.EventDemand
	for rows.Next() {
		var d domain.EventDemand
		if err := rows.Scan(&d.EventID, &d.ProductID, &d.RequestedQuantity, &d.Unit); err != nil {
			return nil, mapError(err)
		}
		out = append(out, d)
	}
	return out, mapError(rows.Err())
}
