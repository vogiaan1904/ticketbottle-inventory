package grpc

import (
	"context"

	svc "github.com/vogiaan/ticketbottle-inventory/internal/services"
	invpb "github.com/vogiaan/ticketbottle-inventory/pkg/grpc/inventory"
	"github.com/vogiaan/ticketbottle-inventory/pkg/logger"
)

type grpcService struct {
	rSvc  svc.ReservationService
	tcSvc svc.TicketClassService
	l     logger.Logger
	invpb.UnimplementedInventoryServiceServer
}

func NewGrpcService(rSvc svc.ReservationService, tcSvc svc.TicketClassService, l logger.Logger) invpb.InventoryServiceServer {
	return &grpcService{
		rSvc:  rSvc,
		tcSvc: tcSvc,
		l:     l,
	}
}

func (s *grpcService) CreateTicketClass(ctx context.Context, req *invpb.CreateTicketClassRequest) (*invpb.CreateTicketClassResponse, error) {
	// Implementation goes here
}
