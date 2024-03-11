package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"reconya-ai/db"
	"reconya-ai/internal/device"
	"reconya-ai/internal/eventlog"
	"reconya-ai/internal/network"
	"reconya-ai/internal/nicidentifier"
	"reconya-ai/internal/pingsweep"
	"reconya-ai/internal/portscan"
	"reconya-ai/internal/systemstatus"
	"reconya-ai/middleware"
	"time"
)

func main() {
	mongoURI := os.Getenv("MONGODB_URI")
	if mongoURI == "" {
		mongoURI = "mongodb://localhost:27017/reconya-mongo-dev"
	}
	db.ConnectToMongo(mongoURI)

	networkService := network.NewNetworkService(db.GetMongoClient(), "reconya-dev", "networks")
	deviceService := device.NewDeviceService(db.GetMongoClient(), "reconya-dev", "devices")
	deviceHandlers := device.NewDeviceHandlers(deviceService)
	eventLogService := eventlog.NewEventLogService(db.GetMongoClient(), "reconya-dev", "event_logs")
	eventLogHandlers := eventlog.NewEventLogHandlers(eventLogService)
	systemStatusService := systemstatus.NewSystemStatusService(db.GetMongoClient(), "reconya-dev", "system_status")
	portScanService := portscan.NewPortScanService(deviceService, eventLogService /* other dependencies */)
	pingSweepService := pingsweep.NewPingSweepService(deviceService, eventLogService, networkService, portScanService)

	net, err := networkService.FindOrCreate("192.168.144.0/24")
	if err != nil {
		log.Fatalf("Failed to find or create network: %v", err)
	}
	log.Printf("Network CIDR: %s", net.CIDR)

	nicService := nicidentifier.NewNicIdentifierService(networkService, systemStatusService, eventLogService)
	nicService.Identify()

	go func() {
		log.Println("Starting Ping Sweep Service...")
		log.Println("Running Ping Sweep Service...")
		pingSweepService.Run()
		log.Println("Ping Sweep Service Run Completed.")
	}()

	go func() {
		ticker := time.NewTicker(3 * time.Minute)
		for range ticker.C {
			log.Println("Running Ping Sweep Service...")
			pingSweepService.Run()
			log.Println("Ping Sweep Service Run Completed.")
		}
	}()

	// Create a new ServeMux
	mux := http.NewServeMux()
	corsRouter := middleware.CORS(mux)

	// Define your routes
	mux.HandleFunc("/devices", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			deviceHandlers.CreateDevice(w, r)
		case http.MethodGet:
			deviceHandlers.GetAllDevices(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/event-log", eventLogHandlers.FindLatest)
	mux.HandleFunc("/event-log/", eventLogHandlers.FindAllByDeviceId)

	mux.HandleFunc("/", handler)

	// Wrap the entire router with the logging middleware
	loggedRouter := middleware.LoggingMiddleware(corsRouter)

	fmt.Println("Server is starting on port 3008...")
	if err := http.ListenAndServe(":3008", loggedRouter); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello, World!")
}
