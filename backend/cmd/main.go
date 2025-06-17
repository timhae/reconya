package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"reconya-ai/db"
	"reconya-ai/internal/auth"
	"reconya-ai/internal/config"
	"reconya-ai/internal/device"
	"reconya-ai/internal/eventlog"
	"reconya-ai/internal/network"
	"reconya-ai/internal/nicidentifier"
	"reconya-ai/internal/pingsweep"
	"reconya-ai/internal/portscan"
	"reconya-ai/internal/systemstatus"
	"reconya-ai/middleware"
)

func runDeviceUpdater(service *device.DeviceService) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			err := service.UpdateDeviceStatuses()
			if err != nil {
				log.Printf("Failed to update device statuses: %v", err)
				// Add a delay after an error to allow other operations to complete
				time.Sleep(1 * time.Second)
			}
		}
	}
}

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Create repositories factory
	var repoFactory *db.RepositoryFactory
	var sqliteDB *sql.DB

	log.Println("Using SQLite database")
	sqliteDB, err = db.ConnectToSQLite(cfg.SQLitePath)
	if err != nil {
		log.Fatalf("Failed to connect to SQLite: %v", err)
	}
	
	// Initialize database schema
	if err := db.InitializeSchema(sqliteDB); err != nil {
		log.Fatalf("Failed to initialize database schema: %v", err)
	}

	repoFactory = db.NewRepositoryFactory(sqliteDB, cfg.DatabaseName)

	// Create repositories
	networkRepo := repoFactory.NewNetworkRepository()
	deviceRepo := repoFactory.NewDeviceRepository()
	eventLogRepo := repoFactory.NewEventLogRepository()
	systemStatusRepo := repoFactory.NewSystemStatusRepository()

	// Create database manager for concurrent access control
	dbManager := db.NewDBManager()

	// Initialize services with repositories
	networkService := network.NewNetworkService(networkRepo, cfg, dbManager)
	deviceService := device.NewDeviceService(deviceRepo, networkService, cfg, dbManager)
	eventLogService := eventlog.NewEventLogService(eventLogRepo, deviceService, dbManager)
	systemStatusService := systemstatus.NewSystemStatusService(systemStatusRepo)
	portScanService := portscan.NewPortScanService(deviceService, eventLogService)
	pingSweepService := pingsweep.NewPingSweepService(cfg, deviceService, eventLogService, networkService, portScanService)
	nicService := nicidentifier.NewNicIdentifierService(networkService, systemStatusService, eventLogService, deviceService)
	
	authHandlers := auth.NewAuthHandlers(cfg)
	middlewareHandlers := middleware.NewMiddleware(cfg)

	nicService.Identify()
	go runPingSweepService(pingSweepService)
	go runDeviceUpdater(deviceService)

	mux := setupRouter(deviceService, eventLogService, systemStatusService, networkService, authHandlers, middlewareHandlers, cfg)
	loggedRouter := middleware.LoggingMiddleware(mux)

	server := &http.Server{
		Addr:    ":3008",
		Handler: loggedRouter,
	}

	go func() {
		log.Println("Server is starting on port 3008...")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("ListenAndServe: %v", err)
		}
	}()

	waitForShutdown(server)
}

func setupRouter(
	deviceService *device.DeviceService,
	eventLogService *eventlog.EventLogService,
	systemStatusService *systemstatus.SystemStatusService,
	networkService *network.NetworkService,
	authHandlers *auth.AuthHandlers,
	middlewareHandlers *middleware.Middleware,
	cfg *config.Config) http.Handler {
	deviceHandlers := device.NewDeviceHandlers(deviceService, cfg)
	eventLogHandlers := eventlog.NewEventLogHandlers(eventLogService)
	systemStatusHandlers := systemstatus.NewSystemStatusHandlers(systemStatusService)
	networkHandlers := network.NewNetworkHandlers(networkService)

	mux := http.NewServeMux()
	corsRouter := middleware.SetupCORS()(mux)

	mux.HandleFunc("/login", authHandlers.LoginHandler)
	mux.HandleFunc("/check-auth", authHandlers.CheckAuthHandler)
	
	// In development Docker environment, make these endpoints accessible without auth
	// In production, uncomment the middlewareHandlers.AuthMiddleware wrapper
	mux.HandleFunc("/devices", deviceHandlers.GetAllDevices)
	mux.HandleFunc("/system-status/latest", systemStatusHandlers.GetLatestSystemStatus)
	mux.HandleFunc("/event-log", eventLogHandlers.FindLatest)
	mux.HandleFunc("/event-log/", eventLogHandlers.FindAllByDeviceId)
	mux.HandleFunc("/network", networkHandlers.GetNetwork)

	return corsRouter
}

func runPingSweepService(service *pingsweep.PingSweepService) {
	log.Println("Starting initial ping sweep service run...")
	service.Run()

	// Use 30 seconds for development to see updates more quickly
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	log.Printf("Ping sweep service scheduled to run every 30 seconds")
	for range ticker.C {
		log.Println("Running scheduled ping sweep...")
		service.Run()
	}
}

func waitForShutdown(server *http.Server) {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	log.Println("Shutting down the server...")
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server Shutdown Failed:%+v", err)
	}
	log.Println("Server gracefully stopped")
}
