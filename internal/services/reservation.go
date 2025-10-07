package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/vogiaan/ticketbottle-inventory/internal/models"
	pkgGorm "github.com/vogiaan/ticketbottle-inventory/pkg/gorm"
	pkgLog "github.com/vogiaan/ticketbottle-inventory/pkg/logger"
	"gorm.io/gorm"
)

type implReservationService struct {
	l    pkgLog.Logger
	repo *pkgGorm.Repository
}

type ReservationService interface {
	Create(ctx context.Context, in CreateReservationInput) (models.Reservation, error)
	GetByID(ctx context.Context, id uint) (models.Reservation, error)
	GetByOrderID(ctx context.Context, orderID uuid.UUID) ([]models.Reservation, error)
	GetByOrderIDAndTicketClass(ctx context.Context, orderID uuid.UUID, ticketClassID uint) (*models.Reservation, error)
	GetActiveByTicketClassID(ctx context.Context, ticketClassID uint) ([]models.Reservation, error)
	GetExpired(ctx context.Context, limit int) ([]models.Reservation, error)
	GetExpiredByTicketClassID(ctx context.Context, ticketClassID uint) ([]models.Reservation, error)
	UpdateStatus(ctx context.Context, id uint, status models.ReservationStatus) error
	UpdateStatusByOrderID(ctx context.Context, orderID uuid.UUID, status models.ReservationStatus) error
	CancelReservation(ctx context.Context, id uint) error
	ExpireReservation(ctx context.Context, id uint) error
	ExpireReservations(ctx context.Context, ids []uint) error
	Delete(ctx context.Context, id uint) error
	GetTotalReservedQuantity(ctx context.Context, ticketClassID uint) (int, error)
}

func NewReservationService(l pkgLog.Logger, repo *pkgGorm.Repository) ReservationService {
	return &implReservationService{
		l:    l,
		repo: repo,
	}
}

func (s implReservationService) Create(ctx context.Context, in CreateReservationInput) (models.Reservation, error) {
	r := s.buildModel(in)
	if err := s.repo.Create(ctx, &r); err != nil {
		s.l.Errorf(ctx, "service.reservation.Create: %v", err)
		return models.Reservation{}, err
	}

	return r, nil
}

func (s implReservationService) GetByID(ctx context.Context, id uint) (models.Reservation, error) {
	var r models.Reservation
	if err := s.repo.FindByID(ctx, &r, id); err != nil {
		if err == gorm.ErrRecordNotFound {
			s.l.Warnf(ctx, "service.reservation.GetByID: %v", err)
		}
		s.l.Errorf(ctx, "service.reservation.GetByID: %v", err)
		return models.Reservation{}, err
	}
	return r, nil
}

func (s implReservationService) GetByOrderID(ctx context.Context, orderID uuid.UUID) ([]models.Reservation, error) {
	var rs []models.Reservation
	if err := s.repo.WithContext(ctx).Where("order_id = ?", orderID).Find(&rs).Error; err != nil {
		s.l.Errorf(ctx, "service.reservation.GetByOrderID: %v", err)
		return nil, err
	}

	return rs, nil
}

func (s implReservationService) GetByOrderIDAndTicketClass(ctx context.Context, orderID uuid.UUID, ticketClassID uint) (*models.Reservation, error) {
	var r models.Reservation
	err := s.repo.WithContext(ctx).
		Where("order_id = ? AND ticket_class_id = ?", orderID, ticketClassID).
		First(&r).Error
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func (s implReservationService) GetActiveByTicketClassID(ctx context.Context, ticketClassID uint) ([]models.Reservation, error) {
	var rs []models.Reservation
	if err := s.repo.WithContext(ctx).
		Where("ticket_class_id = ? AND status = ?", ticketClassID, models.ReservationStatusActive).
		Where("expires_at > ?", time.Now()).
		Find(&rs).Error; err != nil {
		s.l.Errorf(ctx, "service.reservation.GetActiveByTicketClassID: %v", err)
		return nil, err
	}

	return rs, nil
}

func (s implReservationService) GetExpired(ctx context.Context, limit int) ([]models.Reservation, error) {
	var rs []models.Reservation
	if err := s.repo.WithContext(ctx).
		Where("status = ? AND expires_at <= ?", models.ReservationStatusActive, time.Now()).
		Limit(limit).
		Find(&rs).Error; err != nil {
		s.l.Errorf(ctx, "service.reservation.GetExpired: %v", err)
		return nil, err
	}

	return rs, nil
}

func (s implReservationService) GetExpiredByTicketClassID(ctx context.Context, ticketClassID uint) ([]models.Reservation, error) {
	var rs []models.Reservation
	if err := s.repo.WithContext(ctx).
		Where("ticket_class_id = ? AND status = ? AND expires_at <= ?",
			ticketClassID, models.ReservationStatusActive, time.Now()).
		Find(&rs).Error; err != nil {
		s.l.Errorf(ctx, "service.reservation.GetExpiredByTicketClassID: %v", err)
		return nil, err
	}

	return rs, nil
}

func (s implReservationService) UpdateStatus(ctx context.Context, id uint, status models.ReservationStatus) error {
	var r models.Reservation
	if err := s.repo.FindByID(ctx, &r, id); err != nil {
		if err == gorm.ErrRecordNotFound {
			s.l.Warnf(ctx, "service.reservation.UpdateStatus: %v", err)
		}
		s.l.Errorf(ctx, "service.reservation.UpdateStatus: %v", err)
		return err
	}

	r.Status = status
	if err := s.repo.Update(ctx, &r); err != nil {
		s.l.Errorf(ctx, "service.reservation.UpdateStatus: %v", err)
		return err
	}

	return nil
}

func (s implReservationService) UpdateStatusByOrderID(ctx context.Context, orderID uuid.UUID, status models.ReservationStatus) error {
	if err := s.repo.WithContext(ctx).Model(&models.Reservation{}).
		Where("order_id = ?", orderID).
		Update("status", status).Error; err != nil {
		s.l.Errorf(ctx, "service.reservation.UpdateStatusByOrderID: %v", err)
		return err
	}

	return nil
}

func (s implReservationService) CancelReservation(ctx context.Context, id uint) error {
	return s.UpdateStatus(ctx, id, models.ReservationStatusCancelled)
}

func (s implReservationService) ExpireReservation(ctx context.Context, id uint) error {
	return s.UpdateStatus(ctx, id, models.ReservationStatusExpired)
}

func (s implReservationService) ExpireReservations(ctx context.Context, ids []uint) error {
	return s.repo.WithContext(ctx).Model(&models.Reservation{}).
		Where("id IN ?", ids).
		Update("status", models.ReservationStatusExpired).Error
}

func (s implReservationService) Delete(ctx context.Context, id uint) error {
	var r models.Reservation
	if err := s.repo.FindByID(ctx, &r, id); err != nil {
		if err == gorm.ErrRecordNotFound {
			s.l.Warnf(ctx, "service.reservation.Delete: %v", err)
		}
		s.l.Errorf(ctx, "service.reservation.Delete: %v", err)
		return err
	}

	if err := s.repo.Delete(ctx, &r); err != nil {
		s.l.Errorf(ctx, "service.reservation.Delete: %v", err)
		return err
	}

	return nil
}

func (s implReservationService) GetTotalReservedQuantity(ctx context.Context, ticketClassID uint) (int, error) {
	var result struct {
		Total int
	}

	if err := s.repo.WithContext(ctx).Model(&models.Reservation{}).
		Select("COALESCE(SUM(qty), 0) as total").
		Where("ticket_class_id = ? AND status = ? AND expires_at > ?",
			ticketClassID, models.ReservationStatusActive, time.Now()).
		Scan(&result).Error; err != nil {
		s.l.Errorf(ctx, "service.reservation.GetTotalReservedQuantity: %v", err)
		return 0, err
	}
	return result.Total, nil
}
