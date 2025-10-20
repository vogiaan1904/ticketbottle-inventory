package service

import (
	"time"
)

type ReserveInput struct {
	OrderID   string
	Items     []ReserveItem
	ExpiresAt time.Time
}

type ReserveItem struct {
	TicketClassID int64
	Qty           int
}
