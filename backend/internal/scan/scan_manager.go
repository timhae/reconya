package scan

import (
	"sync"
	"time"
	"reconya-ai/models"
	"reconya-ai/internal/pingsweep"
	"reconya-ai/internal/network"
	"log"
)

// ScanState represents the current state of the scanning system
type ScanState struct {
	IsRunning       bool              `json:"is_running"`
	IsStopping      bool              `json:"is_stopping"`
	CurrentNetwork  *models.Network   `json:"current_network"`
	StartTime       *time.Time        `json:"start_time"`
	LastScanTime    *time.Time        `json:"last_scan_time"`
	ScanCount       int               `json:"scan_count"`
}

// ScanManager manages the network scanning state and operations
type ScanManager struct {
	state           ScanState
	mutex           sync.RWMutex
	pingSweepService *pingsweep.PingSweepService
	networkService  *network.NetworkService
	stopChannel     chan bool
	done            chan bool
}

// NewScanManager creates a new scan manager
func NewScanManager(pingSweepService *pingsweep.PingSweepService, networkService *network.NetworkService) *ScanManager {
	return &ScanManager{
		state: ScanState{
			IsRunning: false,
		},
		pingSweepService: pingSweepService,
		networkService:  networkService,
	}
}

// GetState returns the current scan state
func (sm *ScanManager) GetState() ScanState {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()
	return sm.state
}

// IsRunning returns whether a scan is currently running
func (sm *ScanManager) IsRunning() bool {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()
	return sm.state.IsRunning
}

// GetCurrentNetwork returns the currently selected network for scanning
func (sm *ScanManager) GetCurrentNetwork() *models.Network {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()
	return sm.state.CurrentNetwork
}

// StartScan starts scanning the specified network
func (sm *ScanManager) StartScan(networkID string) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	if sm.state.IsRunning {
		return &ScanError{Type: AlreadyRunning, Message: "A scan is already running"}
	}

	// Get the network to scan
	network, err := sm.networkService.FindByID(networkID)
	if err != nil {
		return &ScanError{Type: NetworkNotFound, Message: "Network not found"}
	}
	if network == nil {
		return &ScanError{Type: NetworkNotFound, Message: "Network not found"}
	}

	// Update state
	now := time.Now()
	sm.state.IsRunning = true
	sm.state.CurrentNetwork = network
	sm.state.StartTime = &now
	sm.state.ScanCount = 0

	// Create channels for communication
	sm.stopChannel = make(chan bool)
	sm.done = make(chan bool)

	// Start the scanning goroutine
	go sm.runScanLoop()

	log.Printf("Started scanning network: %s (%s)", network.Name, network.CIDR)
	return nil
}

// StopScan stops the current scan
func (sm *ScanManager) StopScan() error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	if !sm.state.IsRunning {
		return &ScanError{Type: NotRunning, Message: "No scan is currently running"}
	}

	if sm.state.IsStopping {
		return &ScanError{Type: NotRunning, Message: "Scan is already stopping"}
	}

	// Set stopping state
	sm.state.IsStopping = true
	
	// Signal the scan loop to stop
	close(sm.stopChannel)
	
	// Wait for the scan loop to finish
	go func() {
		<-sm.done
		sm.mutex.Lock()
		defer sm.mutex.Unlock()
		sm.state.IsRunning = false
		sm.state.IsStopping = false
		sm.state.CurrentNetwork = nil
		sm.state.StartTime = nil
		log.Println("Scan stopped successfully")
	}()

	return nil
}

// runScanLoop runs the continuous scanning loop
func (sm *ScanManager) runScanLoop() {
	defer close(sm.done)

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	log.Printf("Starting scan loop for network: %s", sm.state.CurrentNetwork.CIDR)

	// Run first scan immediately
	sm.runSingleScan()

	for {
		select {
		case <-sm.stopChannel:
			log.Println("Scan loop received stop signal")
			return
		case <-ticker.C:
			sm.runSingleScan()
		}
	}
}

// runSingleScan executes a single scan iteration
func (sm *ScanManager) runSingleScan() {
	sm.mutex.Lock()
	network := sm.state.CurrentNetwork
	sm.mutex.Unlock()

	if network == nil {
		return
	}

	log.Printf("Running scan on network: %s", network.CIDR)
	
	// Execute the ping sweep with the current network
	devices, err := sm.pingSweepService.ExecuteSweepScanCommand(network.CIDR)
	if err != nil {
		log.Printf("Error during ping sweep: %v", err)
		return
	}

	log.Printf("Ping sweep found %d devices from scan", len(devices))

	// Process the devices (similar to the original Run method)
	for i, device := range devices {
		log.Printf("Processing device %d/%d: %s", i+1, len(devices), device.IPv4)
		
		// Set the network ID for the device
		device.NetworkID = network.ID
		
		// Update device in database
		updatedDevice, err := sm.pingSweepService.DeviceService.CreateOrUpdate(&device)
		if err != nil {
			log.Printf("Error updating device %s: %v", device.IPv4, err)
			continue
		}
		log.Printf("Successfully saved device: %s", device.IPv4)

		// Create event log
		deviceIDStr := device.ID
		err = sm.pingSweepService.EventLogService.CreateOne(&models.EventLog{
			Type:     models.DeviceOnline,
			DeviceID: &deviceIDStr,
		})
		if err != nil {
			log.Printf("Error creating device online event log: %v", err)
		}

		// Add to port scan queue if eligible
		if sm.pingSweepService.DeviceService.EligibleForPortScan(updatedDevice) {
			// Note: We'll need to expose the port scan queue from ping sweep service
			// For now, let's trigger port scan directly
			go sm.pingSweepService.PortScanService.Run(*updatedDevice)
		}
	}

	// Update scan state
	sm.mutex.Lock()
	now := time.Now()
	sm.state.LastScanTime = &now
	sm.state.ScanCount++
	sm.mutex.Unlock()

	duration := time.Since(*sm.state.StartTime)
	log.Printf("Completed scan iteration %d for network %s. Found %d devices.", sm.state.ScanCount, network.CIDR, len(devices))

	// Create event log for ping sweep completion
	durationInSeconds := float64(duration.Seconds())
	err = sm.pingSweepService.EventLogService.CreateOne(&models.EventLog{
		Type: models.PingSweep,
		DurationSeconds: &durationInSeconds,
	})
	if err != nil {
		log.Printf("Error creating ping sweep completion event log: %v", err)
	}
}

// ScanErrorType represents different types of scan errors
type ScanErrorType string

const (
	AlreadyRunning   ScanErrorType = "already_running"
	NotRunning       ScanErrorType = "not_running"
	NetworkNotFound  ScanErrorType = "network_not_found"
	NoNetworks       ScanErrorType = "no_networks"
)

// ScanError represents a scan-related error
type ScanError struct {
	Type    ScanErrorType `json:"type"`
	Message string        `json:"message"`
}

func (e *ScanError) Error() string {
	return e.Message
}