package service

import (
	"time"

	"github.com/google/uuid"
)

type CreateTicketClassInput struct {
	EventID     uuid.UUID
	Name        string
	PriceCents  int64
	Currency    string
	Total       int
	SaleStartAt *time.Time
	SaleEndAt   *time.Time
}

type UpdateTicketClassInput struct {
	Name        string
	PriceCents  *int64
	Currency    string
	Total       int
	SaleStartAt *time.Time
	SaleEndAt   *time.Time
	Status      *string
}
