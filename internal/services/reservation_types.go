package service

import (
	"time"

	"github.com/google/uuid"
)

type CreateReservationInput struct {
	OrderID       uuid.UUID
	TicketClassID uint
	Qty           int
	ExpiresAt     time.Time
}
