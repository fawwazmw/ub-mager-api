package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/wardayadev/ub-mager-api/internal/config"
	"github.com/wardayadev/ub-mager-api/internal/database"
	"github.com/wardayadev/ub-mager-api/internal/model"
	"gorm.io/gorm"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	db, err := database.NewPostgres(cfg)
	if err != nil {
		log.Fatalf("failed to connect to postgres: %v", err)
	}

	fmt.Println("🌱 Seeding database...")

	ctx := context.Background()

	// Seed users
	adminUser := seedUser(ctx, db, "Admin UB-Mager", "+6280000000000", "admin@ubmager.local", "ADMIN")
	passenger1 := seedUser(ctx, db, "Budi Santoso", "+6281234567001", "budi@example.com", "PASSENGER")
	passenger2 := seedUser(ctx, db, "Siti Rahayu", "+6281234567002", "siti@example.com", "PASSENGER")
	driver1User := seedUser(ctx, db, "Agus Prasetyo", "+6289876543001", "agus@example.com", "DRIVER")
	driver2User := seedUser(ctx, db, "Dedi Kurniawan", "+6289876543002", "dedi@example.com", "DRIVER")
	driver3User := seedUser(ctx, db, "Eko Wijaya", "+6289876543003", "eko@example.com", "DRIVER")

	// Seed driver profiles
	driver1 := seedDriverProfile(ctx, db, driver1User.ID, model.VehicleMotorcycle, "Honda", "Vario 160", 2023, "Hitam", "N 1234 AB", true)
	driver2 := seedDriverProfile(ctx, db, driver2User.ID, model.VehicleMotorcycle, "Yamaha", "NMAX 155", 2024, "Biru", "N 5678 CD", true)
	driver3 := seedDriverProfile(ctx, db, driver3User.ID, model.VehicleCar, "Toyota", "Avanza", 2022, "Putih", "N 9012 EF", true)

	// Seed completed rides with ratings
	seedCompletedRide(ctx, db, passenger1.ID, driver1.ID, -7.9526, 112.6060, "Universitas Brawijaya", -7.9666, 112.6326, "Alun-Alun Kota Malang", 21500, 5)
	seedCompletedRide(ctx, db, passenger1.ID, driver2.ID, -7.9666, 112.6326, "Alun-Alun Kota Malang", -7.9780, 112.6340, "Stasiun Malang", 18000, 4)
	seedCompletedRide(ctx, db, passenger2.ID, driver1.ID, -7.9780, 112.6340, "Stasiun Malang", -7.9526, 112.6060, "Universitas Brawijaya", 22000, 5)
	seedCompletedRide(ctx, db, passenger2.ID, driver3.ID, -7.9526, 112.6060, "Universitas Brawijaya", -7.9270, 112.6530, "Mall Dinoyo City", 35000, 4)
	seedCompletedRide(ctx, db, passenger1.ID, driver1.ID, -7.9270, 112.6530, "Mall Dinoyo City", -7.9666, 112.6326, "Alun-Alun Kota Malang", 19500, 5)

	fmt.Println("")
	fmt.Println("✅ Seeding complete!")
	fmt.Println("")
	fmt.Println("📋 Test Accounts:")
	fmt.Println("┌──────────────────┬──────────────────────┬────────────────┐")
	fmt.Println("│ Role             │ Phone                │ Password       │")
	fmt.Println("├──────────────────┼──────────────────────┼────────────────┤")
	fmt.Printf("│ Admin            │ %s │ Password123!   │\n", adminUser.Phone)
	fmt.Printf("│ Passenger        │ %s │ Password123!   │\n", passenger1.Phone)
	fmt.Printf("│ Passenger        │ %s │ Password123!   │\n", passenger2.Phone)
	fmt.Printf("│ Driver           │ %s │ Password123!   │\n", driver1User.Phone)
	fmt.Printf("│ Driver           │ %s │ Password123!   │\n", driver2User.Phone)
	fmt.Printf("│ Driver (Car)     │ %s │ Password123!   │\n", driver3User.Phone)
	fmt.Println("└──────────────────┴──────────────────────┴────────────────┘")
	fmt.Println("")
	fmt.Printf("📊 Seeded: 6 users, 3 drivers, 5 completed rides\n")
}

func seedUser(ctx context.Context, db *gorm.DB, fullName, phone, email string, role model.UserRole) *model.User {
	var existing model.User
	if err := db.WithContext(ctx).Where("phone = ? OR email = ?", phone, email).First(&existing).Error; err == nil {
		// Update role if needed
		if existing.Role != role {
			db.WithContext(ctx).Model(&existing).Update("role", role)
		}
		fmt.Printf("  ⏭️  User exists: %s (%s)\n", fullName, phone)
		return &existing
	}

	hash, _ := bcrypt.GenerateFromPassword([]byte("Password123!"), bcrypt.DefaultCost)
	user := &model.User{
		ID:           uuid.New(),
		FullName:     fullName,
		Phone:        phone,
		Email:        email,
		PasswordHash: string(hash),
		Role:         role,
		IsActive:     true,
	}

	if err := db.WithContext(ctx).Create(user).Error; err != nil {
		log.Fatalf("failed to seed user %s: %v", fullName, err)
	}
	fmt.Printf("  ✅ User created: %s (%s) [%s]\n", fullName, phone, role)
	return user
}

func seedDriverProfile(ctx context.Context, db *gorm.DB, userID uuid.UUID, vehicleType model.VehicleType, brand, vehicleModel string, year int, color, plate string, verified bool) *model.DriverProfile {
	var existing model.DriverProfile
	if err := db.WithContext(ctx).Where("user_id = ?", userID).First(&existing).Error; err == nil {
		fmt.Printf("  ⏭️  Driver profile exists: %s\n", plate)
		return &existing
	}

	expiry := time.Now().AddDate(2, 0, 0)
	var verifiedAt *time.Time
	if verified {
		now := time.Now()
		verifiedAt = &now
	}

	// Set initial location around Malang
	lat := -7.9526 + (float64(uuid.New()[0]%100) / 10000)
	lng := 112.6060 + (float64(uuid.New()[0]%100) / 10000)

	profile := &model.DriverProfile{
		ID:             uuid.New(),
		UserID:         userID,
		VehicleType:    vehicleType,
		VehicleBrand:   brand,
		VehicleModel:   vehicleModel,
		VehicleYear:    year,
		VehicleColor:   color,
		LicensePlate:   plate,
		LicenseNumber:  fmt.Sprintf("SIM-%d", time.Now().UnixNano()%1000000),
		LicenseExpiry:  expiry,
		Latitude:       &lat,
		Longitude:      &lng,
		IsOnline:       false,
		IsAvailable:    true,
		RatingAvg:      4.85,
		TotalTrips:     0,
		AcceptanceRate: 95.00,
		IsVerified:     verified,
		VerifiedAt:     verifiedAt,
	}

	if err := db.WithContext(ctx).Create(profile).Error; err != nil {
		log.Fatalf("failed to seed driver profile %s: %v", plate, err)
	}
	fmt.Printf("  ✅ Driver profile: %s %s (%s) [verified=%v]\n", brand, vehicleModel, plate, verified)
	return profile
}

func seedCompletedRide(ctx context.Context, db *gorm.DB, passengerID, driverID uuid.UUID, pickupLat, pickupLng float64, pickupAddr string, dropoffLat, dropoffLng float64, dropoffAddr string, fare float64, rating int) {
	now := time.Now()
	requestedAt := now.Add(-time.Duration(30+uuid.New()[0]%60) * time.Minute)
	matchedAt := requestedAt.Add(12 * time.Second)
	arrivedAt := matchedAt.Add(4 * time.Minute)
	pickedUpAt := arrivedAt.Add(1 * time.Minute)
	completedAt := pickedUpAt.Add(time.Duration(10+uuid.New()[0]%15) * time.Minute)

	ride := &model.Ride{
		ID:                 uuid.New(),
		PassengerID:        passengerID,
		DriverID:           &driverID,
		Status:             model.RideStatusCompleted,
		VehicleType:        model.VehicleMotorcycle,
		PickupLat:          pickupLat,
		PickupLng:          pickupLng,
		PickupAddress:      pickupAddr,
		DropoffLat:         dropoffLat,
		DropoffLng:         dropoffLng,
		DropoffAddress:     dropoffAddr,
		EstimatedDistanceM: 3500,
		EstimatedDurationS: 600,
		ActualDistanceM:    3800,
		ActualDurationS:    int(completedAt.Sub(pickedUpAt).Seconds()),
		BaseFare:           fare,
		SurgeMultiplier:    1.0,
		TotalFare:          fare,
		PaymentMethod:      model.PaymentCash,
		RequestedAt:        requestedAt,
		MatchedAt:          &matchedAt,
		DriverArrivedAt:    &arrivedAt,
		PickedUpAt:         &pickedUpAt,
		CompletedAt:        &completedAt,
	}

	if err := db.WithContext(ctx).Create(ride).Error; err != nil {
		log.Fatalf("failed to seed ride: %v", err)
	}

	// Add rating from passenger
	ratingObj := &model.Rating{
		ID:      uuid.New(),
		RideID:  ride.ID,
		RaterID: passengerID,
		RateeID: driverID,
		Score:   rating,
		Comment: "Good ride!",
	}
	db.WithContext(ctx).Create(ratingObj)

	// Update driver total trips
	db.WithContext(ctx).Model(&model.DriverProfile{}).Where("id = ?", driverID).
		UpdateColumn("total_trips", gorm.Expr("total_trips + 1"))

	fmt.Printf("  ✅ Ride: %s → %s (Rp %.0f, ⭐%d)\n", pickupAddr, dropoffAddr, fare, rating)
}
