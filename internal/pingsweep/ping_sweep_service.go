package pingsweep

import (
	"log"
	"os/exec"
	"reconya-ai/internal/device"
	"reconya-ai/internal/eventlog"
	"reconya-ai/internal/network"
	"reconya-ai/internal/portscan"
	"reconya-ai/models"
	// Import other necessary packages...
)

type PingSweepService struct {
	DeviceService   *device.DeviceService
	EventLogService *eventlog.EventLogService
	NetworkService  *network.NetworkService
	PortScanService *portscan.PortScanService
}

func NewPingSweepService(deviceService *device.DeviceService, eventLogService *eventlog.EventLogService, networkService *network.NetworkService, portScanService *portscan.PortScanService) *PingSweepService {
	return &PingSweepService{
		DeviceService:   deviceService,
		EventLogService: eventLogService,
		NetworkService:  networkService,
		PortScanService: portScanService,
	}
}

func (s *PingSweepService) Run() {
	log.Println("Starting new ping sweep scan...")
	network := "192.168.144.0/24"

	devices, err := s.ExecuteSweepScanCommand(network)
	if err != nil {
		log.Printf("Error executing sweep scan: %v\n", err)
		return
	}

	for _, device := range devices {
		// Update the device status in the database
		updatedDevice, err := s.DeviceService.CreateOrUpdate(&device)
		if err != nil {
			log.Printf("Error updating device %s: %v", device.IPv4, err)
			continue
		}

		// Check if the device is eligible for a port scan and initiate it concurrently
		if s.DeviceService.EligibleForPortScan(updatedDevice) {
			go func(ip string) {
				s.PortScanService.Run(ip)
			}(updatedDevice.IPv4)
		}
	}

	log.Printf("Ping sweep scan completed. Found %d devices.", len(devices))
}

// eligibleForPortScan determines whether a device is eligible for port scanning
func eligibleForPortScan(device *models.Device) bool {
	// Implement the logic to determine if the device should be port scanned
	return true // Placeholder return value
}

func (s *PingSweepService) ExecuteSweepScanCommand(network string) ([]models.Device, error) {
	cmd := exec.Command("sudo", "/usr/bin/nmap", "-sn", "--send-ip", "-T4", network)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}

	devices := s.ParseNmapOutput(string(output))
	return devices, nil
}

func (s *PingSweepService) ParseNmapOutput(output string) []models.Device {
	return s.DeviceService.ParseFromNmap(output)
}

// Implement other methods as needed...
