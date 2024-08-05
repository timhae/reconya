package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"reconya-ai/db"
	"reconya-ai/internal/device"
	"reconya-ai/internal/eventlog"
	"reconya-ai/internal/network"
	"reconya-ai/internal/nicidentifier"
	"reconya-ai/internal/pingsweep"
	"reconya-ai/internal/portscan"
	"reconya-ai/internal/systemstatus"
	"reconya-ai/middleware"

	"github.com/dgrijalva/jwt-go"
	"github.com/joho/godotenv"
)

var jwtKey = []byte("my_secret_key")

type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type Claims struct {
	Username string `json:"username"`
	jwt.StandardClaims
}

func setupCORS() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	mongoURI := os.Getenv("MONGODB_URI")
	if mongoURI == "" {
		log.Fatal("MONGODB_URI is not set in .env file")
	}
	db.ConnectToMongo(mongoURI)

	networkService := network.NewNetworkService(db.GetMongoClient(), "reconya-dev", "networks")
	deviceService := device.NewDeviceService(db.GetMongoClient(), "reconya-dev", "devices")
	deviceHandlers := device.NewDeviceHandlers(deviceService)
	eventLogService := eventlog.NewEventLogService(db.GetMongoClient(), "reconya-dev", "event_logs", deviceService)
	eventLogHandlers := eventlog.NewEventLogHandlers(eventLogService)
	systemStatusService := systemstatus.NewSystemStatusService(db.GetMongoClient(), "reconya-dev", "system_status")
	systemStatusHandlers := systemstatus.NewSystemStatusHandlers(systemStatusService)
	portScanService := portscan.NewPortScanService(deviceService, eventLogService)
	pingSweepService := pingsweep.NewPingSweepService(deviceService, eventLogService, networkService, portScanService)

	network := os.Getenv("NETWORK_RANGE")
	if network == "" {
		log.Fatal("NETWORK_RANGE is not set in .env file")
	}
	net, err := networkService.FindOrCreate(network)
	if err != nil {
		log.Fatalf("Failed to find or create network: %v", err)
	}
	log.Printf("Network CIDR: %s", net.CIDR)

	log.Printf("Starting network identification")
	nicService := nicidentifier.NewNicIdentifierService(networkService, systemStatusService, eventLogService, deviceService)
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

	mux := http.NewServeMux()
	corsRouter := setupCORS()(mux)

	mux.HandleFunc("/login", LoginHandler)
	mux.HandleFunc("/check-auth", CheckAuthHandler)
	mux.HandleFunc("/devices", AuthMiddleware(deviceHandlers.GetAllDevices))
	mux.HandleFunc("/system-status/latest", AuthMiddleware(systemStatusHandlers.GetLatestSystemStatus))
	mux.HandleFunc("/event-log", AuthMiddleware(eventLogHandlers.FindLatest))
	mux.HandleFunc("/event-log/", AuthMiddleware(eventLogHandlers.FindAllByDeviceId))

	mux.HandleFunc("/", handler)

	loggedRouter := middleware.LoggingMiddleware(corsRouter)

	fmt.Println("Server is starting on port 3008...")
	if err := http.ListenAndServe(":3008", loggedRouter); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	var creds Credentials
	err := json.NewDecoder(r.Body).Decode(&creds)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	expectedUsername := os.Getenv("LOGIN_USERNAME")
	expectedPassword := os.Getenv("LOGIN_PASSWORD")
	if expectedUsername == "" || expectedPassword == "" {
		log.Fatal("USERNAME or PASSWORD is not set in .env file")
	}

	if creds.Username != expectedUsername || creds.Password != expectedPassword {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	expirationTime := time.Now().Add(5 * time.Minute)
	claims := &Claims{
		Username: creds.Username,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"token": tokenString})
}

func CheckAuthHandler(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	tokenStr := authHeader[len("Bearer "):]
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})

	if err != nil {
		if err == jwt.ErrSignatureInvalid {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if !token.Valid {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		tokenStr := authHeader[len("Bearer "):]
		claims := &Claims{}

		token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
			return jwtKey, nil
		})

		if err != nil {
			if err == jwt.ErrSignatureInvalid {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if !token.Valid {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello, World!")
}
