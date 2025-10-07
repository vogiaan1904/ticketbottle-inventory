package main

import (
	"context"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/vogiaan/ticketbottle-inventory/internal/models"
	"github.com/vogiaan/ticketbottle-inventory/internal/repository"
	dbgorm "github.com/vogiaan/ticketbottle-inventory/pkg/gorm"
	"gorm.io/gorm"
)

// This example demonstrates how to use the ticket inventory system
func main() {
	// 1. Connect to database
	config := dbgorm.DefaultConfig()
	db, err := dbgorm.Connect(config, 5, 2*time.Second)
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer db.Close()

	// 2. Auto-migrate models
	if err := db.AutoMigrate(&models.TicketClass{}, &models.Reservation{}); err != nil {
		log.Fatalf("Migration failed: %v", err)
	}

	// 3. Create repositories - db.DB is the embedded *gorm.DB
	ctx := context.Background()
	ticketClassRepo := repository.NewTicketClassRepository(db.DB)
	reservationRepo := repository.NewReservationRepository(db.DB)

	// 4. Demonstrate usage
	if err := demonstrateTicketInventory(ctx, db.DB, ticketClassRepo, reservationRepo); err != nil {
		log.Fatalf("Demo failed: %v", err)
	}

	log.Println("✓ All operations completed successfully!")
}

func demonstrateTicketInventory(
	ctx context.Context,
	db *gorm.DB,
	ticketClassRepo *repository.TicketClassRepository,
	reservationRepo *repository.ReservationRepository,
) error {
	eventID := uint(123) // Example event ID
	orderID := uuid.New()

	// ===== STEP 1: Create Ticket Classes =====
	log.Println("\n===== Creating Ticket Classes =====")

	saleStart := time.Now()
	saleEnd := time.Now().Add(30 * 24 * time.Hour) // 30 days from now

	ticketClasses := []*models.TicketClass{
		{
			EventID:     eventID,
			Name:        "General Admission",
			PriceCents:  5000, // $50.00
			Currency:    "USD",
			Total:       1000,
			Reserved:    0,
			Sold:        0,
			SaleStartAt: &saleStart,
			SaleEndAt:   &saleEnd,
			Status:      "ACTIVE",
		},
		{
			EventID:     eventID,
			Name:        "VIP",
			PriceCents:  15000, // $150.00
			Currency:    "USD",
			Total:       100,
			Reserved:    0,
			Sold:        0,
			SaleStartAt: &saleStart,
			SaleEndAt:   &saleEnd,
			Status:      "ACTIVE",
		},
		{
			EventID:     eventID,
			Name:        "Seat Zone A",
			PriceCents:  8000, // $80.00
			Currency:    "USD",
			Total:       500,
			Reserved:    0,
			Sold:        0,
			SaleStartAt: &saleStart,
			SaleEndAt:   &saleEnd,
			Status:      "ACTIVE",
		},
	}

	for _, tc := range ticketClasses {
		if err := ticketClassRepo.Create(ctx, tc); err != nil {
			return err
		}
		log.Printf("✓ Created ticket class: %s (ID: %d, Price: $%.2f, Total: %d)",
			tc.Name, tc.ID, float64(tc.PriceCents)/100, tc.Total)
	}

	// ===== STEP 2: Query Available Tickets =====
	log.Println("\n===== Querying Available Tickets =====")

	availableTickets, err := ticketClassRepo.GetAvailableByEventID(ctx, eventID)
	if err != nil {
		return err
	}
	log.Printf("✓ Found %d available ticket classes for event %d", len(availableTickets), eventID)

	for _, tc := range availableTickets {
		available := tc.Total - tc.Reserved - tc.Sold
		log.Printf("  - %s: %d available (Total: %d, Reserved: %d, Sold: %d)",
			tc.Name, available, tc.Total, tc.Reserved, tc.Sold)
	}

	// ===== STEP 3: Create Reservations (Transaction) =====
	log.Println("\n===== Creating Reservations =====")

	gaTicket := ticketClasses[0]  // General Admission
	vipTicket := ticketClasses[1] // VIP
	requestedQtyGA := 5
	requestedQtyVIP := 2

	// Use transaction to ensure atomicity
	err = db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txTicketRepo := repository.NewTicketClassRepository(tx)
		txReservationRepo := repository.NewReservationRepository(tx)

		// Check availability for GA tickets
		available, err := txTicketRepo.CheckAvailability(ctx, gaTicket.ID, requestedQtyGA)
		if err != nil {
			return err
		}
		if !available {
			log.Printf("✗ Not enough GA tickets available")
			return nil
		}

		// Check availability for VIP tickets
		available, err = txTicketRepo.CheckAvailability(ctx, vipTicket.ID, requestedQtyVIP)
		if err != nil {
			return err
		}
		if !available {
			log.Printf("✗ Not enough VIP tickets available")
			return nil
		}

		// Reserve GA tickets
		if err := txTicketRepo.IncrementReserved(ctx, gaTicket.ID, requestedQtyGA); err != nil {
			return err
		}

		// Create GA reservation
		gaReservation := &models.Reservation{
			OrderID:       orderID,
			TicketClassID: gaTicket.ID,
			Qty:           requestedQtyGA,
			ExpiresAt:     time.Now().Add(15 * time.Minute), // 15-minute hold
			Status:        models.ReservationStatusActive,
		}
		if err := txReservationRepo.Create(ctx, gaReservation); err != nil {
			return err
		}
		log.Printf("✓ Reserved %d x %s (Expires: %s)", requestedQtyGA, gaTicket.Name, gaReservation.ExpiresAt.Format("15:04:05"))

		// Reserve VIP tickets
		if err := txTicketRepo.IncrementReserved(ctx, vipTicket.ID, requestedQtyVIP); err != nil {
			return err
		}

		// Create VIP reservation
		vipReservation := &models.Reservation{
			OrderID:       orderID,
			TicketClassID: vipTicket.ID,
			Qty:           requestedQtyVIP,
			ExpiresAt:     time.Now().Add(15 * time.Minute),
			Status:        models.ReservationStatusActive,
		}
		if err := txReservationRepo.Create(ctx, vipReservation); err != nil {
			return err
		}
		log.Printf("✓ Reserved %d x %s (Expires: %s)", requestedQtyVIP, vipTicket.Name, vipReservation.ExpiresAt.Format("15:04:05"))

		return nil
	})
	if err != nil {
		return err
	}

	// ===== STEP 4: Verify Reservations =====
	log.Println("\n===== Verifying Reservations =====")

	reservations, err := reservationRepo.GetByOrderID(ctx, orderID)
	if err != nil {
		return err
	}
	log.Printf("✓ Order %s has %d reservations", orderID.String()[:8], len(reservations))

	for _, r := range reservations {
		log.Printf("  - Reservation ID: %d, Qty: %d, Status: %s, IsActive: %v",
			r.ID, r.Qty, r.Status, r.IsActive())
	}

	// Check updated ticket class stats
	updatedGA, err := ticketClassRepo.GetByID(ctx, gaTicket.ID)
	if err != nil {
		return err
	}
	log.Printf("✓ GA Ticket updated - Reserved: %d, Sold: %d, Available: %d",
		updatedGA.Reserved, updatedGA.Sold, updatedGA.Total-updatedGA.Reserved-updatedGA.Sold)

	// ===== STEP 5: Confirm Purchase =====
	log.Println("\n===== Confirming Purchase =====")

	err = db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txTicketRepo := repository.NewTicketClassRepository(tx)
		txReservationRepo := repository.NewReservationRepository(tx)

		// Get all reservations for the order
		orderReservations, err := txReservationRepo.GetByOrderID(ctx, orderID)
		if err != nil {
			return err
		}

		// Convert reservations to sales
		for _, res := range orderReservations {
			// Increment sold count (this also decrements reserved)
			if err := txTicketRepo.IncrementSold(ctx, res.TicketClassID, res.Qty); err != nil {
				return err
			}
		}

		// Mark reservations as confirmed
		if err := txReservationRepo.ConfirmReservationsByOrderID(ctx, orderID); err != nil {
			return err
		}

		log.Printf("✓ Order %s confirmed - Reservations converted to sales", orderID.String()[:8])
		return nil
	})
	if err != nil {
		return err
	}

	// ===== STEP 6: Verify Final State =====
	log.Println("\n===== Final State =====")

	finalGA, err := ticketClassRepo.GetByID(ctx, gaTicket.ID)
	if err != nil {
		return err
	}
	log.Printf("✓ GA Ticket final - Reserved: %d, Sold: %d, Available: %d",
		finalGA.Reserved, finalGA.Sold, finalGA.Total-finalGA.Reserved-finalGA.Sold)

	finalVIP, err := ticketClassRepo.GetByID(ctx, vipTicket.ID)
	if err != nil {
		return err
	}
	log.Printf("✓ VIP Ticket final - Reserved: %d, Sold: %d, Available: %d",
		finalVIP.Reserved, finalVIP.Sold, finalVIP.Total-finalVIP.Reserved-finalVIP.Sold)

	// ===== STEP 7: Demonstrate Expiration =====
	log.Println("\n===== Demonstrating Expiration Handling =====")

	// Create a reservation that's already expired
	expiredOrderID := uuid.New()
	expiredReservation := &models.Reservation{
		OrderID:       expiredOrderID,
		TicketClassID: gaTicket.ID,
		Qty:           3,
		ExpiresAt:     time.Now().Add(-5 * time.Minute), // Already expired
		Status:        models.ReservationStatusActive,
	}

	// First reserve the tickets
	if err := ticketClassRepo.IncrementReserved(ctx, gaTicket.ID, 3); err != nil {
		return err
	}

	if err := reservationRepo.Create(ctx, expiredReservation); err != nil {
		return err
	}
	log.Printf("✓ Created expired reservation (ID: %d) for testing", expiredReservation.ID)

	// Find and expire old reservations
	expiredReservations, err := reservationRepo.GetExpired(ctx, 100)
	if err != nil {
		return err
	}
	log.Printf("✓ Found %d expired reservations", len(expiredReservations))

	for _, expired := range expiredReservations {
		// Mark as expired
		if err := reservationRepo.ExpireReservation(ctx, expired.ID); err != nil {
			return err
		}
		// Release the reserved tickets
		if err := ticketClassRepo.DecrementReserved(ctx, expired.TicketClassID, expired.Qty); err != nil {
			return err
		}
		log.Printf("✓ Expired reservation ID %d and released %d tickets", expired.ID, expired.Qty)
	}

	return nil
}
