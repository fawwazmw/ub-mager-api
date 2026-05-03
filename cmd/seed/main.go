package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/wardayadev/ub-mager-api/internal/config"
	"github.com/wardayadev/ub-mager-api/internal/database"
	"github.com/wardayadev/ub-mager-api/internal/model"
	"gorm.io/gorm"
)

var locations = []struct {
	Lat     float64
	Lng     float64
	Address string
}{
	{-7.9526, 112.6060, "Universitas Brawijaya"},
	{-7.9666, 112.6326, "Alun-Alun Kota Malang"},
	{-7.9780, 112.6340, "Stasiun Malang"},
	{-7.9270, 112.6530, "Mall Dinoyo City"},
	{-7.9420, 112.6150, "RS Saiful Anwar"},
	{-7.9610, 112.6170, "Jl. Ijen Boulevard"},
	{-7.9350, 112.6250, "Pasar Besar Malang"},
	{-7.9700, 112.6100, "Terminal Arjosari"},
	{-7.9480, 112.6380, "Tugu Malang"},
	{-7.9560, 112.6450, "Malang Town Square"},
}

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
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	adminUser := seedUser(ctx, db, "Admin UB-Mager", "+6280000000000", "admin@ubmager.local", "ADMIN")
	passenger1 := seedUser(ctx, db, "Budi Santoso", "+6281234567001", "budi@example.com", "PASSENGER")
	passenger2 := seedUser(ctx, db, "Siti Rahayu", "+6281234567002", "siti@example.com", "PASSENGER")
	passenger3 := seedUser(ctx, db, "Rina Wulandari", "+6281234567003", "rina@example.com", "PASSENGER")
	passenger4 := seedUser(ctx, db, "Hendra Gunawan", "+6281234567004", "hendra@example.com", "PASSENGER")
	driver1User := seedUser(ctx, db, "Agus Prasetyo", "+6289876543001", "agus@example.com", "DRIVER")
	driver2User := seedUser(ctx, db, "Dedi Kurniawan", "+6289876543002", "dedi@example.com", "DRIVER")
	driver3User := seedUser(ctx, db, "Eko Wijaya", "+6289876543003", "eko@example.com", "DRIVER")
	driver4User := seedUser(ctx, db, "Fajar Nugroho", "+6289876543004", "fajar@example.com", "DRIVER")

	driver1 := seedDriverProfile(ctx, db, driver1User.ID, model.VehicleMotorcycle, "Honda", "Vario 160", 2023, "Hitam", "N 1234 AB", true)
	driver2 := seedDriverProfile(ctx, db, driver2User.ID, model.VehicleMotorcycle, "Yamaha", "NMAX 155", 2024, "Biru", "N 5678 CD", true)
	driver3 := seedDriverProfile(ctx, db, driver3User.ID, model.VehicleCar, "Toyota", "Avanza", 2022, "Putih", "N 9012 EF", true)
	driver4 := seedDriverProfile(ctx, db, driver4User.ID, model.VehicleCarXL, "Toyota", "Innova Zenix", 2024, "Silver", "N 3456 GH", false)

	passengers := []*model.User{passenger1, passenger2, passenger3, passenger4}
	drivers := []*model.DriverProfile{driver1, driver2, driver3}
	vehicleTypes := []model.VehicleType{model.VehicleMotorcycle, model.VehicleMotorcycle, model.VehicleCar}
	payments := []model.PaymentMethod{model.PaymentCash, model.PaymentCash, model.PaymentEwallet}

	rideCount := 0
	for daysAgo := 13; daysAgo >= 0; daysAgo-- {
		ridesPerDay := 3 + rng.Intn(5)
		for i := 0; i < ridesPerDay; i++ {
			passenger := passengers[rng.Intn(len(passengers))]
			driverIdx := rng.Intn(len(drivers))
			driver := drivers[driverIdx]
			vt := vehicleTypes[driverIdx]
			pm := payments[rng.Intn(len(payments))]

			pickupIdx := rng.Intn(len(locations))
			dropoffIdx := (pickupIdx + 1 + rng.Intn(len(locations)-1)) % len(locations)
			pickup := locations[pickupIdx]
			dropoff := locations[dropoffIdx]

			hour := pickHour(rng)
			baseTime := time.Now().AddDate(0, 0, -daysAgo).Truncate(24 * time.Hour).Add(time.Duration(hour) * time.Hour).Add(time.Duration(rng.Intn(60)) * time.Minute)

			fare := 12000 + float64(rng.Intn(30000))
			if vt == model.VehicleCar {
				fare += 15000
			}
			if vt == model.VehicleCarXL {
				fare += 25000
			}

			surge := 1.0
			if hour >= 7 && hour <= 9 || hour >= 17 && hour <= 19 {
				if rng.Float64() < 0.3 {
					surge = 1.0 + float64(rng.Intn(5))*0.1
				}
			}
			fare *= surge

			isCancelled := rng.Float64() < 0.12
			if isCancelled {
				seedCancelledRide(ctx, db, passenger.ID, driver.ID, vt, pm, pickup, dropoff, fare, surge, baseTime)
			} else {
				rating := 3 + rng.Intn(3)
				seedCompletedRide(ctx, db, passenger.ID, driver.ID, vt, pm, pickup, dropoff, fare, surge, baseTime, rating)
			}
			rideCount++
		}
	}

	seedSearchingRide(ctx, db, passenger1.ID, locations[0], locations[3], model.VehicleMotorcycle)
	seedSearchingRide(ctx, db, passenger3.ID, locations[4], locations[7], model.VehicleCar)
	rideCount += 2

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
	fmt.Printf("│ Passenger        │ %s │ Password123!   │\n", passenger3.Phone)
	fmt.Printf("│ Passenger        │ %s │ Password123!   │\n", passenger4.Phone)
	fmt.Printf("│ Driver           │ %s │ Password123!   │\n", driver1User.Phone)
	fmt.Printf("│ Driver           │ %s │ Password123!   │\n", driver2User.Phone)
	fmt.Printf("│ Driver (Car)     │ %s │ Password123!   │\n", driver3User.Phone)
	fmt.Printf("│ Driver (Pending) │ %s │ Password123!   │\n", driver4User.Phone)
	fmt.Println("└──────────────────┴──────────────────────┴────────────────┘")
	fmt.Println("")
	fmt.Printf("📊 Seeded: %d users, 4 drivers (1 pending), %d rides\n", 8, rideCount)
	_ = driver4
}

func pickHour(rng *rand.Rand) int {
	r := rng.Float64()
	switch {
	case r < 0.25:
		return 7 + rng.Intn(3)
	case r < 0.50:
		return 11 + rng.Intn(3)
	case r < 0.75:
		return 17 + rng.Intn(3)
	default:
		return rng.Intn(24)
	}
}

func seedUser(ctx context.Context, db *gorm.DB, fullName, phone, email string, role model.UserRole) *model.User {
	var existing model.User
	if err := db.WithContext(ctx).Where("phone = ? OR email = ?", phone, email).First(&existing).Error; err == nil {
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
		IsOnline:       verified,
		IsAvailable:    true,
		RatingAvg:      4.50 + float64(uuid.New()[0]%50)/100,
		TotalTrips:     0,
		AcceptanceRate: 90.00 + float64(uuid.New()[0]%10),
		IsVerified:     verified,
		VerifiedAt:     verifiedAt,
	}

	if err := db.WithContext(ctx).Create(profile).Error; err != nil {
		log.Fatalf("failed to seed driver profile %s: %v", plate, err)
	}
	fmt.Printf("  ✅ Driver profile: %s %s (%s) [verified=%v]\n", brand, vehicleModel, plate, verified)
	return profile
}

type loc struct {
	Lat     float64
	Lng     float64
	Address string
}

func seedCompletedRide(ctx context.Context, db *gorm.DB, passengerID, driverID uuid.UUID, vt model.VehicleType, pm model.PaymentMethod, pickup, dropoff loc, fare, surge float64, baseTime time.Time, rating int) {
	requestedAt := baseTime
	matchedAt := requestedAt.Add(time.Duration(8+rand.Intn(20)) * time.Second)
	arrivedAt := matchedAt.Add(time.Duration(3+rand.Intn(5)) * time.Minute)
	pickedUpAt := arrivedAt.Add(time.Duration(30+rand.Intn(90)) * time.Second)
	completedAt := pickedUpAt.Add(time.Duration(8+rand.Intn(20)) * time.Minute)

	distM := 1500 + float64(rand.Intn(5000))
	durS := int(distM / 6.5)

	ride := &model.Ride{
		ID:                 uuid.New(),
		PassengerID:        passengerID,
		DriverID:           &driverID,
		Status:             model.RideStatusCompleted,
		VehicleType:        vt,
		PickupLat:          pickup.Lat,
		PickupLng:          pickup.Lng,
		PickupAddress:      pickup.Address,
		DropoffLat:         dropoff.Lat,
		DropoffLng:         dropoff.Lng,
		DropoffAddress:     dropoff.Address,
		EstimatedDistanceM: distM,
		EstimatedDurationS: durS,
		ActualDistanceM:    distM * (1.0 + float64(rand.Intn(15))/100),
		ActualDurationS:    int(completedAt.Sub(pickedUpAt).Seconds()),
		BaseFare:           fare * 0.7,
		SurgeMultiplier:    surge,
		TotalFare:          fare,
		PaymentMethod:      pm,
		RequestedAt:        requestedAt,
		MatchedAt:          &matchedAt,
		DriverArrivedAt:    &arrivedAt,
		PickedUpAt:         &pickedUpAt,
		CompletedAt:        &completedAt,
	}

	if err := db.WithContext(ctx).Create(ride).Error; err != nil {
		log.Fatalf("failed to seed ride: %v", err)
	}

	ratingObj := &model.Rating{
		ID:      uuid.New(),
		RideID:  ride.ID,
		RaterID: passengerID,
		RateeID: driverID,
		Score:   rating,
	}
	db.WithContext(ctx).Create(ratingObj)

	db.WithContext(ctx).Model(&model.DriverProfile{}).Where("id = ?", driverID).
		UpdateColumn("total_trips", gorm.Expr("total_trips + 1"))

	fmt.Printf("  ✅ Ride: %s → %s (Rp %.0f, ⭐%d, %s)\n", pickup.Address, dropoff.Address, fare, rating, vt)
}

func seedCancelledRide(ctx context.Context, db *gorm.DB, passengerID, driverID uuid.UUID, vt model.VehicleType, pm model.PaymentMethod, pickup, dropoff loc, fare, surge float64, baseTime time.Time) {
	requestedAt := baseTime
	cancelledAt := requestedAt.Add(time.Duration(2+rand.Intn(10)) * time.Minute)

	reasons := []string{"Changed my mind", "Driver took too long", "Found another ride", "Wrong pickup location"}

	ride := &model.Ride{
		ID:                 uuid.New(),
		PassengerID:        passengerID,
		Status:             model.RideStatusCancelled,
		VehicleType:        vt,
		PickupLat:          pickup.Lat,
		PickupLng:          pickup.Lng,
		PickupAddress:      pickup.Address,
		DropoffLat:         dropoff.Lat,
		DropoffLng:         dropoff.Lng,
		DropoffAddress:     dropoff.Address,
		EstimatedDistanceM: 2000 + float64(rand.Intn(4000)),
		EstimatedDurationS: 300 + rand.Intn(600),
		BaseFare:           fare * 0.7,
		SurgeMultiplier:    surge,
		TotalFare:          fare,
		PaymentMethod:      pm,
		RequestedAt:        requestedAt,
		CancelledAt:        &cancelledAt,
		CancellationReason: reasons[rand.Intn(len(reasons))],
	}

	if err := db.WithContext(ctx).Create(ride).Error; err != nil {
		log.Fatalf("failed to seed cancelled ride: %v", err)
	}
	fmt.Printf("  ❌ Cancelled: %s → %s (%s)\n", pickup.Address, dropoff.Address, ride.CancellationReason)
}

func seedSearchingRide(ctx context.Context, db *gorm.DB, passengerID uuid.UUID, pickup, dropoff loc, vt model.VehicleType) {
	requestedAt := time.Now().Add(-time.Duration(2+rand.Intn(5)) * time.Minute)

	ride := &model.Ride{
		ID:                 uuid.New(),
		PassengerID:        passengerID,
		Status:             model.RideStatusSearching,
		VehicleType:        vt,
		PickupLat:          pickup.Lat,
		PickupLng:          pickup.Lng,
		PickupAddress:      pickup.Address,
		DropoffLat:         dropoff.Lat,
		DropoffLng:         dropoff.Lng,
		DropoffAddress:     dropoff.Address,
		EstimatedDistanceM: 3000,
		EstimatedDurationS: 450,
		BaseFare:           15000,
		SurgeMultiplier:    1.0,
		TotalFare:          15000,
		PaymentMethod:      model.PaymentCash,
		RequestedAt:        requestedAt,
	}

	if err := db.WithContext(ctx).Create(ride).Error; err != nil {
		log.Fatalf("failed to seed searching ride: %v", err)
	}
	fmt.Printf("  🔍 Searching: %s → %s\n", pickup.Address, dropoff.Address)
}
