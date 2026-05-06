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

var ubLocations = []struct {
	Lat     float64
	Lng     float64
	Address string
	Zone    model.CampusZone
}{
	{-7.9526, 112.6146, "FILKOM UB Gedung F", model.ZoneFILKOM},
	{-7.9536, 112.6130, "FILKOM UB Gedung E", model.ZoneFILKOM},
	{-7.9510, 112.6100, "FTP UB", model.ZoneFTP},
	{-7.9490, 112.6080, "FEB UB", model.ZoneFEB},
	{-7.9470, 112.6120, "FH UB", model.ZoneFH},
	{-7.9550, 112.6050, "FK UB", model.ZoneFK},
	{-7.9500, 112.6170, "FIA UB", model.ZoneFIA},
	{-7.9480, 112.6150, "FISIP UB", model.ZoneFISIP},
	{-7.9460, 112.6060, "FMIPA UB", model.ZoneFMIPA},
	{-7.9540, 112.6090, "FT UB", model.ZoneFT},
	{-7.9515, 112.6110, "Rektorat UB", model.ZoneRektorat},
	{-7.9530, 112.6160, "GOR Pertamina UB", model.ZoneGOR},
}

var malangLocations = []struct {
	Lat     float64
	Lng     float64
	Address string
}{
	{-7.9666, 112.6326, "Alun-Alun Kota Malang"},
	{-7.9780, 112.6340, "Stasiun Malang"},
	{-7.9270, 112.6530, "Mall Dinoyo City"},
	{-7.9420, 112.6150, "RS Saiful Anwar"},
	{-7.9610, 112.6170, "Jl. Ijen Boulevard"},
	{-7.9350, 112.6250, "Pasar Besar Malang"},
	{-7.9560, 112.6450, "Malang Town Square (MATOS)"},
	{-7.9440, 112.6380, "Mie Gacoan Soehat"},
	{-7.9500, 112.6300, "Indomaret Soekarno Hatta"},
	{-7.9580, 112.6200, "Kos-kosan Jl. Veteran"},
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

	fmt.Println("🌱 Seeding UB Mager database...")
	fmt.Println("")

	ctx := context.Background()
	rng := rand.New(rand.NewSource(42))

	adminUser := seedUser(ctx, db, "Admin UB Mager", "+6280000000000", "admin@student.ub.ac.id", model.RoleAdmin)
	passenger1 := seedUser(ctx, db, "Budi Santoso", "+6281234567001", "budi@student.ub.ac.id", model.RolePassenger)
	passenger2 := seedUser(ctx, db, "Siti Rahayu", "+6281234567002", "siti@student.ub.ac.id", model.RolePassenger)
	passenger3 := seedUser(ctx, db, "Rina Wulandari", "+6281234567003", "rina@student.ub.ac.id", model.RolePassenger)
	passenger4 := seedUser(ctx, db, "Hendra Gunawan", "+6281234567004", "hendra@student.ub.ac.id", model.RolePassenger)
	passenger5 := seedUser(ctx, db, "Dewi Lestari", "+6281234567005", "dewi@student.ub.ac.id", model.RolePassenger)
	driver1User := seedUser(ctx, db, "Agus Prasetyo", "+6289876543001", "agus@student.ub.ac.id", model.RoleDriver)
	driver2User := seedUser(ctx, db, "Dedi Kurniawan", "+6289876543002", "dedi@student.ub.ac.id", model.RoleDriver)
	driver3User := seedUser(ctx, db, "Eko Wijaya", "+6289876543003", "eko@student.ub.ac.id", model.RoleDriver)
	driver4User := seedUser(ctx, db, "Fajar Nugroho", "+6289876543004", "fajar@student.ub.ac.id", model.RoleDriver)

	fmt.Println("")
	fmt.Println("🚗 Seeding driver profiles...")

	driver1 := seedDriverProfile(ctx, db, driver1User.ID, model.VehicleMotorcycle, "Honda", "Vario 160", 2023, "Hitam", "N 1234 AB", true, model.ZoneFILKOM)
	driver2 := seedDriverProfile(ctx, db, driver2User.ID, model.VehicleMotorcycle, "Yamaha", "NMAX 155", 2024, "Biru", "N 5678 CD", true, model.ZoneFTP)
	driver3 := seedDriverProfile(ctx, db, driver3User.ID, model.VehicleCar, "Toyota", "Avanza", 2022, "Putih", "N 9012 EF", true, model.ZoneFEB)
	_ = seedDriverProfile(ctx, db, driver4User.ID, model.VehicleCarXL, "Toyota", "Innova Zenix", 2024, "Silver", "N 3456 GH", false, model.ZoneOther)

	fmt.Println("")
	fmt.Println("🚕 Seeding rides...")

	passengers := []*model.User{passenger1, passenger2, passenger3, passenger4, passenger5}
	drivers := []*model.DriverProfile{driver1, driver2, driver3}
	vehicleTypes := []model.VehicleType{model.VehicleMotorcycle, model.VehicleMotorcycle, model.VehicleCar}

	rideCount := 0
	for daysAgo := 13; daysAgo >= 0; daysAgo-- {
		ridesPerDay := 3 + rng.Intn(5)
		for i := 0; i < ridesPerDay; i++ {
			passenger := passengers[rng.Intn(len(passengers))]
			driverIdx := rng.Intn(len(drivers))
			driver := drivers[driverIdx]
			vt := vehicleTypes[driverIdx]

			pickupIdx := rng.Intn(len(ubLocations))
			dropoffIdx := rng.Intn(len(malangLocations))
			pickup := ubLocations[pickupIdx]
			dropoff := malangLocations[dropoffIdx]

			hour := pickHour(rng)
			baseTime := time.Now().AddDate(0, 0, -daysAgo).Truncate(24 * time.Hour).Add(time.Duration(hour)*time.Hour + time.Duration(rng.Intn(60))*time.Minute)

			fare := 12000 + float64(rng.Intn(30000))
			if vt == model.VehicleCar {
				fare += 15000
			}

			surge := 1.0
			if (hour >= 7 && hour <= 9) || (hour >= 17 && hour <= 19) {
				if rng.Float64() < 0.3 {
					surge = 1.0 + float64(rng.Intn(5))*0.1
				}
			}
			fare *= surge

			pm := model.PaymentCash
			if rng.Float64() < 0.3 {
				pm = model.PaymentEwallet
			}

			if rng.Float64() < 0.12 {
				seedCancelledRide(ctx, db, rng, passenger.ID, driver.ID, vt, pm,
					loc{pickup.Lat, pickup.Lng, pickup.Address}, loc{dropoff.Lat, dropoff.Lng, dropoff.Address},
					fare, surge, baseTime, pickup.Zone, model.CampusZone(""))
			} else {
				rating := 3 + rng.Intn(3)
				seedCompletedRide(ctx, db, rng, passenger.ID, driver.ID, vt, pm,
					loc{pickup.Lat, pickup.Lng, pickup.Address}, loc{dropoff.Lat, dropoff.Lng, dropoff.Address},
					fare, surge, baseTime, rating, pickup.Zone, model.CampusZone(""))
			}
			rideCount++
		}
	}

	seedSearchingRide(ctx, db, passenger1.ID, loc{ubLocations[0].Lat, ubLocations[0].Lng, ubLocations[0].Address}, loc{malangLocations[2].Lat, malangLocations[2].Lng, malangLocations[2].Address}, model.VehicleMotorcycle, ubLocations[0].Zone)
	seedSearchingRide(ctx, db, passenger3.ID, loc{ubLocations[4].Lat, ubLocations[4].Lng, ubLocations[4].Address}, loc{malangLocations[1].Lat, malangLocations[1].Lng, malangLocations[1].Address}, model.VehicleCar, ubLocations[4].Zone)
	rideCount += 2

	fmt.Println("")
	fmt.Println("📋 Seeding tasks...")

	taskCount := 0
	taskData := []struct {
		category    model.TaskCategory
		title       string
		description string
		fee         float64
		pickupIdx   int
		deliveryIdx int
	}{
		{model.TaskCategoryJastipMakanan, "Jastip Mie Gacoan Soehat", "Tolong beliin mie gacoan level 3 + es teh, ambil di Soehat", 5000, 7, 0},
		{model.TaskCategoryJastipMakanan, "Jastip Kopi Janji Jiwa", "Americano ice 1, hazelnut latte 1. Ambil di Janji Jiwa Soehat", 7000, 8, 1},
		{model.TaskCategoryJastipMakanan, "Jastip Ayam Geprek Bensu", "Ayam geprek level 5 + nasi, sambelnya banyakin", 5000, 9, 3},
		{model.TaskCategoryJastipBarang, "Titip beli charger di Indomaret", "Charger type-C merk apapun yang murah, budget max 50rb", 3000, 8, 0},
		{model.TaskCategoryJastipBarang, "Titip ambil paket di JNE Soehat", "Paket atas nama Rina, no resi: JNE123456", 5000, 9, 2},
		{model.TaskCategoryTitipPrint, "Print tugas Algoritma 20 lembar", "File udah di-share ke WA, print bolak-balik A4", 3000, 5, 0},
		{model.TaskCategoryTitipPrint, "Print + jilid laporan PKL", "50 halaman, jilid soft cover biru. File di Google Drive", 15000, 5, 6},
		{model.TaskCategoryAntarJemput, "Antar ke Stasiun Malang", "Butuh antar ke stasiun jam 3 sore, bawa koper 1", 20000, 0, 1},
		{model.TaskCategoryOther, "Bantuin pindahan kos", "Butuh bantuan angkat barang dari kos lama ke kos baru, jarak 500m", 25000, 9, 0},
		{model.TaskCategoryJastipMakanan, "Jastip Bakso Pak Kumis", "Bakso urat 2 porsi + es jeruk 2. Warung depan FT", 8000, 9, 4},
	}

	for i, td := range taskData {
		pickup := ubLocations[td.pickupIdx%len(ubLocations)]
		delivery := ubLocations[td.deliveryIdx%len(ubLocations)]
		creator := passengers[i%len(passengers)]

		status := model.TaskStatusOpen
		var helperID *uuid.UUID
		var acceptedAt, completedAt *time.Time

		if i < 3 {
			status = model.TaskStatusCompleted
			hID := driver1User.ID
			helperID = &hID
			now := time.Now().Add(-time.Duration(2+i) * 24 * time.Hour)
			acceptedAt = &now
			completed := now.Add(45 * time.Minute)
			completedAt = &completed
		} else if i < 5 {
			status = model.TaskStatusAccepted
			hID := driver2User.ID
			helperID = &hID
			now := time.Now().Add(-30 * time.Minute)
			acceptedAt = &now
		}

		task := &model.Task{
			ID:              uuid.New(),
			CreatorID:       creator.ID,
			HelperID:        helperID,
			Category:        td.category,
			Status:          status,
			Title:           td.title,
			Description:     td.description,
			Fee:             td.fee,
			PickupLat:       pickup.Lat,
			PickupLng:       pickup.Lng,
			PickupAddress:   pickup.Address,
			PickupZone:      pickup.Zone,
			DeliveryLat:     delivery.Lat,
			DeliveryLng:     delivery.Lng,
			DeliveryAddress: delivery.Address,
			DeliveryZone:    delivery.Zone,
			AcceptedAt:      acceptedAt,
			CompletedAt:     completedAt,
		}

		if err := db.WithContext(ctx).Create(task).Error; err == nil {
			fmt.Printf("  ✅ Task: [%s] %s (Rp %.0f) [%s]\n", td.category, td.title, td.fee, status)
			taskCount++
		}
	}

	fmt.Println("")
	fmt.Println("💬 Seeding chat messages...")

	var firstRide model.Ride
	db.WithContext(ctx).Where("status = ?", model.RideStatusCompleted).Order("created_at DESC").First(&firstRide)
	if firstRide.ID != uuid.Nil {
		chatMessages := []struct {
			senderID uuid.UUID
			content  string
			offset   time.Duration
		}{
			{firstRide.PassengerID, "Mas udah di mana?", 0},
			{*firstRide.DriverID, "Lagi di depan FILKOM mas, 2 menit lagi sampai", 30 * time.Second},
			{firstRide.PassengerID, "Oke saya di depan gedung F ya, pake jaket hitam", time.Minute},
			{*firstRide.DriverID, "Siap mas, saya yang pake helm merah", 90 * time.Second},
		}

		baseTime := time.Now().Add(-2 * time.Hour)
		for _, msg := range chatMessages {
			rideID := firstRide.ID
			chat := &model.ChatMessage{
				ID:        uuid.New(),
				RideID:    &rideID,
				SenderID:  msg.senderID,
				Content:   msg.content,
				CreatedAt: baseTime.Add(msg.offset),
			}
			db.WithContext(ctx).Create(chat)
		}
		fmt.Printf("  ✅ 4 chat messages seeded for ride %s\n", firstRide.ID.String()[:8])
	}

	fmt.Println("")
	fmt.Println("🚨 Seeding reports...")

	report := &model.Report{
		ID:             uuid.New(),
		ReporterID:     passenger2.ID,
		ReportedUserID: driver2User.ID,
		Category:       model.ReportCategoryRude,
		Description:    "Driver ngebut dan tidak sopan saat berkomunikasi",
		Status:         model.ReportStatusPending,
	}
	db.WithContext(ctx).Create(report)

	report2 := &model.Report{
		ID:             uuid.New(),
		ReporterID:     passenger4.ID,
		ReportedUserID: passenger5.ID,
		Category:       model.ReportCategorySpam,
		Description:    "User ini spam task yang sama berulang kali",
		Status:         model.ReportStatusResolved,
		AdminNote:      "User sudah diperingatkan",
	}
	db.WithContext(ctx).Create(report2)
	fmt.Println("  ✅ 2 reports seeded (1 pending, 1 resolved)")

	fmt.Println("")
	fmt.Println("═══════════════════════════════════════════════")
	fmt.Println("✅ Seeding complete!")
	fmt.Println("═══════════════════════════════════════════════")
	fmt.Println("")
	fmt.Println("📋 Test Accounts (password: Password123!)")
	fmt.Println("┌──────────────────┬──────────────────────┬──────────────────────────────┐")
	fmt.Println("│ Role             │ Phone                │ Email                        │")
	fmt.Println("├──────────────────┼──────────────────────┼──────────────────────────────┤")
	fmt.Printf("│ Admin            │ %s │ %s │\n", adminUser.Phone, adminUser.Email)
	fmt.Printf("│ Passenger        │ %s │ %s    │\n", passenger1.Phone, passenger1.Email)
	fmt.Printf("│ Passenger        │ %s │ %s    │\n", passenger2.Phone, passenger2.Email)
	fmt.Printf("│ Passenger        │ %s │ %s    │\n", passenger3.Phone, passenger3.Email)
	fmt.Printf("│ Passenger        │ %s │ %s  │\n", passenger4.Phone, passenger4.Email)
	fmt.Printf("│ Passenger        │ %s │ %s    │\n", passenger5.Phone, passenger5.Email)
	fmt.Printf("│ Driver (Motor)   │ %s │ %s    │\n", driver1User.Phone, driver1User.Email)
	fmt.Printf("│ Driver (Motor)   │ %s │ %s    │\n", driver2User.Phone, driver2User.Email)
	fmt.Printf("│ Driver (Car)     │ %s │ %s     │\n", driver3User.Phone, driver3User.Email)
	fmt.Printf("│ Driver (Pending) │ %s │ %s   │\n", driver4User.Phone, driver4User.Email)
	fmt.Println("└──────────────────┴──────────────────────┴──────────────────────────────┘")
	fmt.Println("")
	fmt.Printf("📊 Summary: 10 users, 4 drivers, %d rides, %d tasks, 4 chats, 2 reports\n", rideCount, taskCount)
	fmt.Println("")
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
		fmt.Printf("  ⏭️  User exists: %s\n", fullName)
		return &existing
	}

	hash, _ := bcrypt.GenerateFromPassword([]byte("Password123!"), bcrypt.DefaultCost)
	user := &model.User{
		ID:                uuid.New(),
		FullName:          fullName,
		Phone:             phone,
		Email:             email,
		PasswordHash:      string(hash),
		Role:              role,
		IsActive:          true,
		IsStudentVerified: true,
	}

	if err := db.WithContext(ctx).Create(user).Error; err != nil {
		log.Fatalf("failed to seed user %s: %v", fullName, err)
	}
	fmt.Printf("  ✅ User: %s [%s]\n", fullName, role)
	return user
}

func seedDriverProfile(ctx context.Context, db *gorm.DB, userID uuid.UUID, vehicleType model.VehicleType, brand, vehicleModel string, year int, color, plate string, verified bool, zone model.CampusZone) *model.DriverProfile {
	var existing model.DriverProfile
	if err := db.WithContext(ctx).Where("user_id = ?", userID).First(&existing).Error; err == nil {
		return &existing
	}

	expiry := time.Now().AddDate(2, 0, 0)
	var verifiedAt *time.Time
	if verified {
		now := time.Now().AddDate(0, -1, 0)
		verifiedAt = &now
	}

	lat := -7.9526 + (float64(uuid.New()[0]%50) / 10000)
	lng := 112.6100 + (float64(uuid.New()[0]%80) / 10000)

	profile := &model.DriverProfile{
		ID:             uuid.New(),
		UserID:         userID,
		VehicleType:    vehicleType,
		VehicleBrand:   brand,
		VehicleModel:   vehicleModel,
		VehicleYear:    year,
		VehicleColor:   color,
		LicensePlate:   plate,
		LicenseNumber:  fmt.Sprintf("SIM-A-%d", time.Now().UnixNano()%1000000),
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
		HomeZone:       zone,
	}

	if err := db.WithContext(ctx).Create(profile).Error; err != nil {
		log.Fatalf("failed to seed driver %s: %v", plate, err)
	}
	fmt.Printf("  ✅ Driver: %s %s (%s) [zone=%s, verified=%v]\n", brand, vehicleModel, plate, zone, verified)
	return profile
}

type loc struct {
	Lat     float64
	Lng     float64
	Address string
}

func seedCompletedRide(ctx context.Context, db *gorm.DB, rng *rand.Rand, passengerID, driverID uuid.UUID, vt model.VehicleType, pm model.PaymentMethod, pickup, dropoff loc, fare, surge float64, baseTime time.Time, rating int, pickupZone, dropoffZone model.CampusZone) {
	matchedAt := baseTime.Add(time.Duration(8+rng.Intn(20)) * time.Second)
	arrivedAt := matchedAt.Add(time.Duration(3+rng.Intn(5)) * time.Minute)
	pickedUpAt := arrivedAt.Add(time.Duration(30+rng.Intn(90)) * time.Second)
	completedAt := pickedUpAt.Add(time.Duration(8+rng.Intn(20)) * time.Minute)

	distM := 1500 + float64(rng.Intn(5000))

	ride := &model.Ride{
		ID:                 uuid.New(),
		PassengerID:        passengerID,
		DriverID:           &driverID,
		Status:             model.RideStatusCompleted,
		VehicleType:        vt,
		PickupLat:          pickup.Lat,
		PickupLng:          pickup.Lng,
		PickupAddress:      pickup.Address,
		PickupZone:         pickupZone,
		DropoffLat:         dropoff.Lat,
		DropoffLng:         dropoff.Lng,
		DropoffAddress:     dropoff.Address,
		DropoffZone:        dropoffZone,
		EstimatedDistanceM: distM,
		EstimatedDurationS: int(distM / 6.5),
		ActualDistanceM:    distM * (1.0 + float64(rng.Intn(15))/100),
		ActualDurationS:    int(completedAt.Sub(pickedUpAt).Seconds()),
		BaseFare:           fare * 0.7,
		SurgeMultiplier:    surge,
		TotalFare:          fare,
		PaymentMethod:      pm,
		RequestedAt:        baseTime,
		MatchedAt:          &matchedAt,
		DriverArrivedAt:    &arrivedAt,
		PickedUpAt:         &pickedUpAt,
		CompletedAt:        &completedAt,
	}

	db.WithContext(ctx).Create(ride)

	ratingObj := &model.Rating{
		ID:      uuid.New(),
		RideID:  &ride.ID,
		RaterID: passengerID,
		RateeID: driverID,
		Score:   rating,
	}
	db.WithContext(ctx).Create(ratingObj)
	db.WithContext(ctx).Model(&model.DriverProfile{}).Where("id = ?", driverID).UpdateColumn("total_trips", gorm.Expr("total_trips + 1"))
}

func seedCancelledRide(ctx context.Context, db *gorm.DB, rng *rand.Rand, passengerID, driverID uuid.UUID, vt model.VehicleType, pm model.PaymentMethod, pickup, dropoff loc, fare, surge float64, baseTime time.Time, pickupZone, dropoffZone model.CampusZone) {
	cancelledAt := baseTime.Add(time.Duration(2+rng.Intn(10)) * time.Minute)
	reasons := []string{"Berubah pikiran", "Driver terlalu lama", "Salah lokasi pickup", "Ada keperluan mendadak"}

	ride := &model.Ride{
		ID:                 uuid.New(),
		PassengerID:        passengerID,
		Status:             model.RideStatusCancelled,
		VehicleType:        vt,
		PickupLat:          pickup.Lat,
		PickupLng:          pickup.Lng,
		PickupAddress:      pickup.Address,
		PickupZone:         pickupZone,
		DropoffLat:         dropoff.Lat,
		DropoffLng:         dropoff.Lng,
		DropoffAddress:     dropoff.Address,
		DropoffZone:        dropoffZone,
		EstimatedDistanceM: 2000 + float64(rng.Intn(4000)),
		EstimatedDurationS: 300 + rng.Intn(600),
		BaseFare:           fare * 0.7,
		SurgeMultiplier:    surge,
		TotalFare:          fare,
		PaymentMethod:      pm,
		RequestedAt:        baseTime,
		CancelledAt:        &cancelledAt,
		CancellationReason: reasons[rng.Intn(len(reasons))],
	}

	db.WithContext(ctx).Create(ride)
}

func seedSearchingRide(ctx context.Context, db *gorm.DB, passengerID uuid.UUID, pickup, dropoff loc, vt model.VehicleType, pickupZone model.CampusZone) {
	ride := &model.Ride{
		ID:                 uuid.New(),
		PassengerID:        passengerID,
		Status:             model.RideStatusSearching,
		VehicleType:        vt,
		PickupLat:          pickup.Lat,
		PickupLng:          pickup.Lng,
		PickupAddress:      pickup.Address,
		PickupZone:         pickupZone,
		DropoffLat:         dropoff.Lat,
		DropoffLng:         dropoff.Lng,
		DropoffAddress:     dropoff.Address,
		EstimatedDistanceM: 3000,
		EstimatedDurationS: 450,
		BaseFare:           15000,
		SurgeMultiplier:    1.0,
		TotalFare:          15000,
		PaymentMethod:      model.PaymentCash,
		RequestedAt:        time.Now().Add(-time.Duration(2+rand.Intn(5)) * time.Minute),
	}

	db.WithContext(ctx).Create(ride)
	fmt.Printf("  🔍 Searching: %s → %s\n", pickup.Address, dropoff.Address)
}
