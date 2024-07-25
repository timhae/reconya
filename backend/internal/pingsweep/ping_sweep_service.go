package pingsweep

import (
	"log"
	"os"
	"os/exec"
	"reconya-ai/internal/device"
	"reconya-ai/internal/eventlog"
	"reconya-ai/internal/network"
	"reconya-ai/internal/portscan"
	"reconya-ai/models"
)

type PingSweepService struct {
	DeviceService   *device.DeviceService
	EventLogService *eventlog.EventLogService
	NetworkService  *network.NetworkService
	PortScanService *portscan.PortScanService
}

func NewPingSweepService(
	deviceService *device.DeviceService,
	eventLogService *eventlog.EventLogService,
	networkService *network.NetworkService,
	portScanService *portscan.PortScanService) *PingSweepService {
	return &PingSweepService{
		DeviceService:   deviceService,
		EventLogService: eventLogService,
		NetworkService:  networkService,
		PortScanService: portScanService,
	}
}

func (s *PingSweepService) Run() {
	log.Println("Starting new ping sweep scan...")
	network := os.Getenv("NETWORK_RANGE")
	if network == "" {
		log.Fatal("No network range specified in NETWORK_RANGE environment variable")
	}

	devices, err := s.ExecuteSweepScanCommand(network)
	if err != nil {
		log.Printf("Error executing sweep scan: %v\n", err)
		return
	}

	for _, device := range devices {
		updatedDevice, err := s.DeviceService.CreateOrUpdate(&device)
		if err != nil {
			log.Printf("Error updating device %s: %v", device.IPv4, err)
			continue
		}

		s.EventLogService.CreateOne(&models.EventLog{
			Type:     models.DeviceOnline,
			DeviceID: &updatedDevice.ID,
		})

		if s.DeviceService.EligibleForPortScan(updatedDevice) {
			go func(updatedDevice models.Device) {
				s.PortScanService.Run(updatedDevice)
			}(*updatedDevice)
		}
	}

	log.Printf("Ping sweep scan completed. Found %d devices.", len(devices))
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
