package grpc

import (
	svc "github.com/vogiaan/ticketbottle-inventory/internal/services"
	"github.com/vogiaan/ticketbottle-inventory/pkg/logger"
	inventorypb "github.com/vogiaan/ticketbottle-inventory/protogen/inventory"
)

type InventoryGrpcService struct {
	rSvc  svc.ReservationService
	tcSvc svc.TicketClassService
	l     logger.Logger
	inventorypb.UnimplementedInventoryServiceServer
}

func NewInventoryGrpcService(rSvc svc.ReservationService, tcSvc svc.TicketClassService, l logger.Logger) inventorypb.InventoryServiceServer {
	return &InventoryGrpcService{
		rSvc:  rSvc,
		tcSvc: tcSvc,
		l:     l,
	}
}
