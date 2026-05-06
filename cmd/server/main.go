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

var version = "dev"

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
	if err := db.AutoMigrate(&model.User{}, &model.DriverProfile{}, &model.Ride{}, &model.Rating{}, &model.Report{}, &model.ChatMessage{}, &model.Task{}); err != nil {
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
	accessTTL, err := strconv.Atoi(cfg.JWTAccessTokenTTL)
	if err != nil || accessTTL <= 0 {
		accessTTL = 15 // default 15 minutes
		log.Warn().Str("raw", cfg.JWTAccessTokenTTL).Int("default", accessTTL).Msg("invalid JWT_ACCESS_TOKEN_TTL, using default")
	}
	refreshTTL, err := strconv.Atoi(cfg.JWTRefreshTokenTTL)
	if err != nil || refreshTTL <= 0 {
		refreshTTL = 10080 // default 7 days in minutes
		log.Warn().Str("raw", cfg.JWTRefreshTokenTTL).Int("default", refreshTTL).Msg("invalid JWT_REFRESH_TOKEN_TTL, using default")
	}
	jwtService := jwtpkg.NewJWTService(cfg.JWTSecret, accessTTL, refreshTTL)

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	driverRepo := repository.NewDriverRepository(db)
	rideRepo := repository.NewRideRepository(db)
	analyticsRepo := repository.NewAnalyticsRepository(db)
	reportRepo := repository.NewReportRepository(db)
	chatRepo := repository.NewChatRepository(db)
	taskRepo := repository.NewTaskRepository(db)

	// Initialize caches
	driverGeoCache := cache.NewDriverGeoCache(rdb)

	// Initialize WebSocket hub
	wsHub := ws.NewHub()
	go wsHub.Run()

	// Initialize services
	authService := service.NewAuthService(userRepo, jwtService)
	driverService := service.NewDriverService(driverRepo, userRepo, driverGeoCache)
	rideService := service.NewRideService(rideRepo, driverRepo, userRepo)
	taskService := service.NewTaskService(taskRepo)

	// Initialize handlers
	authHandler := handler.NewAuthHandler(authService, cfg.ServerEnv == "production", jwtService.GetRefreshTTL())
	userHandler := handler.NewUserHandler(authService)
	driverHandler := handler.NewDriverHandler(driverService)
	rideHandler := handler.NewRideHandler(rideService)
	adminHandler := handler.NewAdminHandler(analyticsRepo)
	reportHandler := handler.NewReportHandler(reportRepo)
	chatHandler := handler.NewChatHandler(chatRepo, wsHub)
	taskHandler := handler.NewTaskHandler(taskService, taskRepo)
	userAdminHandler := handler.NewUserAdminHandler(userRepo)

	// Setup Gin
	if cfg.ServerEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(middleware.RequestID())
	r.Use(middleware.Logger())
	r.Use(middleware.CORS(cfg.CORSOrigins))
	r.Use(middleware.BodyLimit(1 << 20)) // 1MB max request body
	r.Use(middleware.Timeout(30 * time.Second))
	r.Use(gin.Recovery())

	type readyResponse struct {
		Ready bool `json:"ready"`
	}
	type serviceCheck struct {
		Status string `json:"status"`
		Error  string `json:"error,omitempty"`
	}
	type healthResponse struct {
		Status    string                  `json:"status"`
		Service   string                  `json:"service"`
		Version   string                  `json:"version"`
		Time      string                  `json:"time"`
		WsClients int                     `json:"ws_clients"`
		Checks    map[string]serviceCheck `json:"checks"`
	}
	type versionResponse struct {
		Version string `json:"version"`
		Service string `json:"service"`
	}

	r.GET("/ready", func(c *gin.Context) {
		sqlDB, err := db.DB()
		if err != nil || sqlDB.Ping() != nil {
			c.JSON(http.StatusServiceUnavailable, readyResponse{Ready: false})
			return
		}
		if rdb.Ping(c.Request.Context()).Err() != nil {
			c.JSON(http.StatusServiceUnavailable, readyResponse{Ready: false})
			return
		}
		c.JSON(http.StatusOK, readyResponse{Ready: true})
	})

	r.GET("/health", func(c *gin.Context) {
		checks := map[string]serviceCheck{}
		healthy := true

		sqlDB, err := db.DB()
		if err != nil {
			checks["database"] = serviceCheck{Status: "unhealthy", Error: err.Error()}
			healthy = false
		} else if err := sqlDB.Ping(); err != nil {
			checks["database"] = serviceCheck{Status: "unhealthy", Error: err.Error()}
			healthy = false
		} else {
			checks["database"] = serviceCheck{Status: "healthy"}
		}

		if err := rdb.Ping(c.Request.Context()).Err(); err != nil {
			checks["redis"] = serviceCheck{Status: "unhealthy", Error: err.Error()}
			healthy = false
		} else {
			checks["redis"] = serviceCheck{Status: "healthy"}
		}

		status := "healthy"
		code := http.StatusOK
		if !healthy {
			status = "degraded"
			code = http.StatusServiceUnavailable
		}

		c.JSON(code, healthResponse{
			Status:    status,
			Service:   "ub-mager-api",
			Version:   version,
			Time:      time.Now().Format(time.RFC3339),
			WsClients: wsHub.GetOnlineCount(),
			Checks:    checks,
		})
	})

	r.GET("/version", func(c *gin.Context) {
		c.JSON(http.StatusOK, versionResponse{Version: version, Service: "ub-mager-api"})
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

			// Reports
			reports := protected.Group("/reports")
			{
				reports.POST("", reportHandler.Submit)
			}

			// Tasks
			tasks := protected.Group("/tasks")
			{
				tasks.POST("", taskHandler.Create)
				tasks.GET("", taskHandler.Feed)
				tasks.GET("/mine", taskHandler.MyTasks)
				tasks.GET("/helping", taskHandler.MyHelperTasks)
				tasks.GET("/:id", taskHandler.GetByID)
				tasks.PUT("/:id/accept", taskHandler.Accept)
				tasks.PUT("/:id/status", taskHandler.UpdateStatus)
				tasks.PUT("/:id/cancel", taskHandler.Cancel)
			}

			// Ride routes — Passenger
			rides := protected.Group("/rides")
			{
				rides.POST("/estimate", rideHandler.Estimate)
				rides.POST("", middleware.RateLimitByUser(3, time.Minute), rideHandler.RequestRide)
				rides.GET("/active", rideHandler.GetActiveRide)
				rides.GET("/history", rideHandler.GetHistory)
				rides.GET("/:id", rideHandler.GetRide)
				rides.PUT("/:id/cancel", rideHandler.CancelRide)
				rides.POST("/:id/rate", rideHandler.RateRide)
				rides.POST("/:id/messages", chatHandler.SendMessage)
				rides.GET("/:id/messages", chatHandler.GetMessages)
				// Driver actions on rides
				rides.PUT("/:id/accept", rideHandler.AcceptRide)
				rides.PUT("/:id/status", rideHandler.DriverUpdateStatus)
			}

			// Admin routes
			admin := protected.Group("/admin")
			admin.Use(middleware.RoleRequired(string(model.RoleAdmin)))
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
				admin.GET("/rides/:id/messages", chatHandler.AdminGetMessages)

				// Reports management
				admin.GET("/reports", reportHandler.List)
				admin.GET("/reports/pending-count", reportHandler.CountPending)
				admin.PUT("/reports/:id/resolve", reportHandler.Resolve)

				admin.GET("/tasks", taskHandler.AdminList)

				admin.GET("/users", userAdminHandler.ListUsers)
				admin.PUT("/users/:id/suspend", userAdminHandler.SuspendUser)
				admin.PUT("/users/:id/unsuspend", userAdminHandler.UnsuspendUser)
			}

			// Analytics routes (accessible by admin)
			analytics := protected.Group("/analytics")
			analytics.Use(middleware.RoleRequired(string(model.RoleAdmin)))
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
	r.GET("/ws", ws.HandleWebSocket(wsHub, jwtService, cfg.CORSOrigins))

	// Start server
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.ServerPort),
		Handler:      r,
		ReadTimeout:  time.Duration(cfg.ReadTimeoutSec) * time.Second,
		WriteTimeout: time.Duration(cfg.WriteTimeoutSec) * time.Second,
		IdleTimeout:  time.Duration(cfg.IdleTimeoutSec) * time.Second,
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

	if err := rdb.Close(); err != nil {
		log.Warn().Err(err).Msg("redis close error")
	}

	sqlDB, err := db.DB()
	if err == nil {
		if err := sqlDB.Close(); err != nil {
			log.Warn().Err(err).Msg("postgres close error")
		}
	}

	log.Info().Msg("server stopped")
}
