package pingsweep

import (
	"fmt"
	"log"
	"os/exec"
	"reconya-ai/internal/config"
	"reconya-ai/internal/device"
	"reconya-ai/internal/eventlog"
	"reconya-ai/internal/network"
	"reconya-ai/internal/portscan"
	"reconya-ai/internal/util"
	"reconya-ai/models"
	"time"
)

type PingSweepService struct {
	Config          *config.Config
	DeviceService   *device.DeviceService
	EventLogService *eventlog.EventLogService
	NetworkService  *network.NetworkService
	PortScanService *portscan.PortScanService
}

func NewPingSweepService(
	cfg *config.Config,
	deviceService *device.DeviceService,
	eventLogService *eventlog.EventLogService,
	networkService *network.NetworkService,
	portScanService *portscan.PortScanService) *PingSweepService {
	return &PingSweepService{
		Config:          cfg,
		DeviceService:   deviceService,
		EventLogService: eventLogService,
		NetworkService:  networkService,
		PortScanService: portScanService,
	}
}
func (s *PingSweepService) Run() {
	log.Println("Starting new ping sweep scan...")

	// Use retry logic for creating initial ping sweep event log
	err := util.RetryOnLock(func() error {
		return s.EventLogService.CreateOne(&models.EventLog{
			Type: models.PingSweep,
		})
	})
	
	if err != nil {
		log.Printf("Error creating ping sweep event log: %v", err)
	}
	devices, err := s.ExecuteSweepScanCommand(s.Config.NetworkCIDR)
	if err != nil {
		log.Printf("Error executing sweep scan: %v\n", err)
		return
	}

	// Create a list of devices that need port scanning
	var devicesToPingScan []models.Device

	// First update all devices - ONE AT A TIME
	for _, device := range devices {
		// Use retry logic for updating device
		updatedDevice, err := util.RetryOnLockWithResult(func() (*models.Device, error) {
			return s.DeviceService.CreateOrUpdate(&device)
		})
		
		if err != nil {
			log.Printf("Error updating device %s after retries: %v", device.IPv4, err)
			continue
		}

		deviceIDStr := device.ID
		// Use retry logic for creating event log
		eventLogErr := util.RetryOnLock(func() error {
			return s.EventLogService.CreateOne(&models.EventLog{
				Type:     models.DeviceOnline,
				DeviceID: &deviceIDStr,
			})
		})
		
		if eventLogErr != nil {
			log.Printf("Error creating device online event log: %v", eventLogErr)
		}

		// Add to the list if eligible for port scan - but limit to max 3 devices per scan
		if s.DeviceService.EligibleForPortScan(updatedDevice) && len(devicesToPingScan) < 3 {
			devicesToPingScan = append(devicesToPingScan, *updatedDevice)
		}
	}

	// Run port scans ONE AT A TIME, not concurrently
	if len(devicesToPingScan) > 0 {
		log.Printf("Running port scans for %d devices, one at a time", len(devicesToPingScan))
		for _, deviceToScan := range devicesToPingScan {
			// Run the port scan synchronously
			s.PortScanService.Run(deviceToScan)
			
			// Add a small delay between port scans to reduce database contention
			time.Sleep(500 * time.Millisecond)
		}
	}

	log.Printf("Ping sweep scan completed. Found %d devices.", len(devices))
	log.Printf("Ping sweep scan completed. Found %d devices.", len(devices))
}

func (s *PingSweepService) ExecuteSweepScanCommand(network string) ([]models.Device, error) {
	log.Printf("Executing nmap command on network: %s", network)
	// In Docker, we don't need sudo and nmap might be in a different location
	cmd := exec.Command("nmap", "-sn", "--send-ip", "-T4", network)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("nmap command failed: %s\n", string(output))
		return nil, fmt.Errorf("error executing nmap: %w", err)
	}

	log.Printf("nmap command succeeded. Output:\n%s", output)

	devices := s.DeviceService.ParseFromNmap(string(output))
	return devices, nil
}
