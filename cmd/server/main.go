package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/wardayadev/ub-mager-api/internal/cache"
	"github.com/wardayadev/ub-mager-api/internal/config"
	"github.com/wardayadev/ub-mager-api/internal/database"
	"github.com/wardayadev/ub-mager-api/internal/handler"
	"github.com/wardayadev/ub-mager-api/internal/middleware"
	"github.com/wardayadev/ub-mager-api/internal/model"
	jwtpkg "github.com/wardayadev/ub-mager-api/internal/pkg/jwt"
	"github.com/wardayadev/ub-mager-api/internal/repository"
	"github.com/wardayadev/ub-mager-api/internal/service"
	"github.com/wardayadev/ub-mager-api/internal/ws"
)



func main() {
	// Setup zerolog
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	if os.Getenv("SERVER_ENV") != "production" {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	// Load config
	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load config")
	}

	// Connect to PostgreSQL
	db, err := database.NewPostgres(cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to postgres")
	}
	log.Info().Msg("connected to PostgreSQL")

	// Auto-migrate
	if err := db.AutoMigrate(&model.User{}, &model.DriverProfile{}, &model.Ride{}, &model.Rating{}); err != nil {
		log.Fatal().Err(err).Msg("failed to run auto-migration")
	}
	log.Info().Msg("database migrated")

	// Connect to Redis
	rdb, err := database.NewRedis(cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to redis")
	}
	log.Info().Msg("connected to Redis")

	// Initialize JWT service
	accessTTL, _ := strconv.Atoi(cfg.JWTAccessTokenTTL)
	refreshTTL, _ := strconv.Atoi(cfg.JWTRefreshTokenTTL)
	jwtService := jwtpkg.NewJWTService(cfg.JWTSecret, accessTTL, refreshTTL)

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	driverRepo := repository.NewDriverRepository(db)
	rideRepo := repository.NewRideRepository(db)
	analyticsRepo := repository.NewAnalyticsRepository(db)

	// Initialize caches
	driverGeoCache := cache.NewDriverGeoCache(rdb)

	// Initialize WebSocket hub
	wsHub := ws.NewHub()
	go wsHub.Run()

	// Initialize services
	authService := service.NewAuthService(userRepo, jwtService)
	driverService := service.NewDriverService(driverRepo, userRepo, driverGeoCache)
	rideService := service.NewRideService(rideRepo, driverRepo, userRepo)

	// Initialize handlers
	authHandler := handler.NewAuthHandler(authService)
	userHandler := handler.NewUserHandler(authService)
	driverHandler := handler.NewDriverHandler(driverService)
	rideHandler := handler.NewRideHandler(rideService)
	adminHandler := handler.NewAdminHandler(analyticsRepo)

	// Setup Gin
	if cfg.ServerEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(middleware.RequestID())
	r.Use(middleware.Logger())
	r.Use(middleware.CORS())
	r.Use(gin.Recovery())

	r.GET("/ready", func(c *gin.Context) {
		sqlDB, err := db.DB()
		if err != nil || sqlDB.Ping() != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"ready": false})
			return
		}
		if rdb.Ping(c.Request.Context()).Err() != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"ready": false})
			return
		}
		c.JSON(http.StatusOK, gin.H{"ready": true})
	})

	r.GET("/health", func(c *gin.Context) {
		checks := gin.H{}
		healthy := true

		sqlDB, err := db.DB()
		if err != nil {
			checks["database"] = gin.H{"status": "unhealthy", "error": err.Error()}
			healthy = false
		} else if err := sqlDB.Ping(); err != nil {
			checks["database"] = gin.H{"status": "unhealthy", "error": err.Error()}
			healthy = false
		} else {
			checks["database"] = gin.H{"status": "healthy"}
		}

		if err := rdb.Ping(c.Request.Context()).Err(); err != nil {
			checks["redis"] = gin.H{"status": "unhealthy", "error": err.Error()}
			healthy = false
		} else {
			checks["redis"] = gin.H{"status": "healthy"}
		}

		status := "healthy"
		code := http.StatusOK
		if !healthy {
			status = "degraded"
			code = http.StatusServiceUnavailable
		}

		c.JSON(code, gin.H{
			"status":      status,
			"service":     "ub-mager-api",
			"version":     "1.0.0",
			"time":        time.Now().Format(time.RFC3339),
			"ws_clients":  wsHub.GetOnlineCount(),
			"checks":      checks,
		})
	})

	// API v1 routes
	v1 := r.Group("/api/v1")
	{
		// Auth (public, rate limited)
		auth := v1.Group("/auth")
		{
			auth.POST("/register", middleware.RateLimit(5, time.Minute), authHandler.Register)
			auth.POST("/login", middleware.RateLimit(10, time.Minute), authHandler.Login)
			auth.POST("/refresh", middleware.RateLimit(30, time.Minute), authHandler.Refresh)
			auth.POST("/logout", authHandler.Logout)
		}

		// Protected routes
		protected := v1.Group("")
		protected.Use(middleware.AuthRequired(jwtService))
		{
			// User profile
			users := protected.Group("/users")
			{
				users.GET("/me", userHandler.GetProfile)
				users.PUT("/me", userHandler.UpdateProfile)
				users.PUT("/me/password", userHandler.ChangePassword)
			}

			// Driver routes
			drivers := protected.Group("/drivers")
			{
				drivers.POST("/register", driverHandler.Register)
				drivers.GET("/me", driverHandler.GetProfile)
				drivers.PUT("/me/status", driverHandler.ToggleStatus)
				drivers.PUT("/me/location", driverHandler.UpdateLocation)
			}

			// Ride routes — Passenger
			rides := protected.Group("/rides")
			{
				rides.POST("/estimate", rideHandler.Estimate)
				rides.POST("", rideHandler.RequestRide)
				rides.GET("/active", rideHandler.GetActiveRide)
				rides.GET("/history", rideHandler.GetHistory)
				rides.GET("/:id", rideHandler.GetRide)
				rides.PUT("/:id/cancel", rideHandler.CancelRide)
				rides.POST("/:id/rate", rideHandler.RateRide)
				// Driver actions on rides
				rides.PUT("/:id/accept", rideHandler.AcceptRide)
				rides.PUT("/:id/status", rideHandler.DriverUpdateStatus)
			}

			// Admin routes
			admin := protected.Group("/admin")
			admin.Use(middleware.RoleRequired("ADMIN"))
			{
				admin.GET("/dashboard", adminHandler.GetDashboardStats)
				admin.GET("/drivers", adminHandler.ListDrivers)
				admin.GET("/drivers/nearby", driverHandler.GetNearbyDrivers)
			admin.GET("/drivers/:id", adminHandler.GetDriverDetail)
			admin.GET("/drivers/:id/rides", adminHandler.GetDriverRides)
			admin.PUT("/drivers/:id/verify", adminHandler.VerifyDriver)
			admin.PUT("/drivers/:id/status", adminHandler.ToggleDriverOnline)
			admin.GET("/activity", adminHandler.GetRecentActivity)
			admin.PUT("/rides/bulk-cancel", adminHandler.BulkCancelStuckRides)
			admin.GET("/rides", adminHandler.ListRides)
			admin.GET("/rides/counts", adminHandler.GetRideCountsByStatus)
				admin.GET("/rides/:id", adminHandler.GetRideDetail)
				admin.PUT("/rides/:id/cancel", adminHandler.CancelRide)
			}

			// Analytics routes (accessible by admin)
			analytics := protected.Group("/analytics")
			analytics.Use(middleware.RoleRequired("ADMIN"))
			{
				analytics.GET("/revenue", adminHandler.GetRevenueStats)
				analytics.GET("/revenue/daily", adminHandler.GetDailyRevenue)
				analytics.GET("/rides", adminHandler.GetRideStats)
				analytics.GET("/drivers/leaderboard", adminHandler.GetDriverLeaderboard)
			analytics.GET("/peak-hours", adminHandler.GetPeakHours)
			}
		}
	}

	// WebSocket endpoint
	r.GET("/ws", ws.HandleWebSocket(wsHub, jwtService))

	// Start server
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.ServerPort),
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Info().Str("port", cfg.ServerPort).Msg("UB-Mager API started")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("server failed")
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal().Err(err).Msg("server forced to shutdown")
	}

	log.Info().Msg("server stopped")
}
