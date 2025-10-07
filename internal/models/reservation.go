package models

import (
	"time"

	"github.com/google/uuid"
)

// Reservation represents short-lived holds created by Order service
type Reservation struct {
	ID            uint              `gorm:"primarykey"`
	OrderID       uuid.UUID         `gorm:"type:uuid;not null;uniqueIndex:idx_order_ticket"`
	TicketClassID uint              `gorm:"not null;uniqueIndex:idx_order_ticket;index:idx_ticket_status_expires"`
	Qty           int               `gorm:"not null"`
	ExpiresAt     time.Time         `gorm:"not null;index:idx_ticket_status_expires"`
	Status        ReservationStatus `gorm:"not null;index:idx_ticket_status_expires"`
	CreatedAt     time.Time
	UpdatedAt     time.Time

	// Relations
	TicketClass TicketClass `gorm:"constraint:OnDelete:RESTRICT"`
}

func (Reservation) TableName() string {
	return "reservation"
}

func (r *Reservation) IsActive() bool {
	return r.Status == ReservationStatusActive && time.Now().Before(r.ExpiresAt)
}

func (r *Reservation) IsExpired() bool {
	return time.Now().After(r.ExpiresAt) && r.Status == ReservationStatusActive
}

type ReservationStatus string

const (
	ReservationStatusActive    ReservationStatus = "ACTIVE"
	ReservationStatusConfirmed ReservationStatus = "CONFIRMED"
	ReservationStatusExpired   ReservationStatus = "EXPIRED"
	ReservationStatusCancelled ReservationStatus = "CANCELLED"
)
