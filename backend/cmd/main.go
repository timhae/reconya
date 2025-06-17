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
	systemRepo         db.SystemStatusRepository
	deviceService      *device.DeviceService
	networkService     *network.NetworkService
	eventLogService    *eventlog.EventLogService
	systemStatusService *systemstatus.SystemStatusService
	portScanService    *portscan.PortScanService
	pingSweepService   *pingsweep.PingSweepService
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()

	// Connect to database
	log.Println("Using SQLite database")
	var err error
	sqliteDB, err = db.ConnectToSQLite(cfg.SQLitePath)
	if err != nil {
		log.Fatalf("Failed to connect to SQLite: %v", err)
	}
	
	// Initialize database schema
	if err := db.InitializeSchema(sqliteDB); err != nil {
		log.Fatalf("Failed to initialize database schema: %v", err)
	}

	// Initialize repositories
	repoFactory = db.NewRepositoryFactory(sqliteDB, cfg.DatabaseName)
	deviceRepo = repoFactory.NewDeviceRepository()
	networkRepo = repoFactory.NewNetworkRepository()
	eventLogRepo = repoFactory.NewEventLogRepository()
	systemRepo = repoFactory.NewSystemStatusRepository()

	// Initialize services
	networkService = network.NewNetworkService(networkRepo, cfg)
	deviceService = device.NewDeviceService(deviceRepo, networkService, cfg)
	eventLogService = eventlog.NewEventLogService(eventLogRepo, deviceService)
	systemStatusService = systemstatus.NewSystemStatusService(systemRepo)
	portScanService = portscan.NewPortScanService(deviceService, eventLogService)
	pingSweepService = pingsweep.NewPingSweepService(cfg, deviceService, eventLogService, networkService, portScanService)
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
	deviceService *device
