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
	"reconya-ai/models"
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

	s.EventLogService.CreateOne(&models.EventLog{
		Type: models.PingSweep,
	})
	devices, err := s.ExecuteSweepScanCommand(s.Config.NetworkCIDR)
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

		deviceIDStr := device.ID
		s.EventLogService.CreateOne(&models.EventLog{
			Type:     models.DeviceOnline,
			DeviceID: &deviceIDStr,
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
	log.Printf("Executing nmap command on network: %s", network)
	cmd := exec.Command("sudo", "/usr/bin/nmap", "-sn", "--send-ip", "-T4", network)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("nmap command failed: %s\n", string(output))
		return nil, fmt.Errorf("error executing nmap: %w", err)
	}

	log.Printf("nmap command succeeded. Output:\n%s", output)

	devices := s.DeviceService.ParseFromNmap(string(output))
	return devices, nil
}
