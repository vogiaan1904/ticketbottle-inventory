package service

import (
	"github.com/vogiaan/ticketbottle-inventory/internal/models"
)

func (s implReservationService) buildModel(in CreateReservationInput) models.Reservation {
	return models.Reservation{
		OrderID:       in.OrderID,
		TicketClassID: in.TicketClassID,
		Qty:           in.Qty,
		Status:        models.ReservationStatusActive,
		ExpiresAt:     in.ExpiresAt,
	}
}
