package portscan

import (
	"encoding/xml"
	"log"
	"os/exec"
	"reconya-ai/internal/device"
	"reconya-ai/internal/eventlog"
	"reconya-ai/models"
)

// PortScanService struct
type PortScanService struct {
	DeviceService   *device.DeviceService
	EventLogService *eventlog.EventLogService
}

// NewPortScanService creates a new PortScanService
func NewPortScanService(deviceService *device.DeviceService, eventLogService *eventlog.EventLogService) *PortScanService {
	return &PortScanService{
		DeviceService:   deviceService,
		EventLogService: eventLogService,
	}
}

// Run executes a port scan for a given IP address
func (s *PortScanService) Run(ipv4 string) {
	log.Printf("Starting port scan for IP [%s]", ipv4)

	// Record the start of the port scan in event logs
	// [Insert code to log the event]

	device, err := s.DeviceService.FindByIPv4(ipv4)
	if err != nil {
		log.Printf("Error finding device: %v", err)
		return
	}

	if device == nil {
		log.Printf("No device found for IP: %s", ipv4)
		return
	}

	// [Insert code to update device's status]

	ports, err := s.ExecutePortScan(ipv4)
	if err != nil {
		log.Printf("Error executing port scan: %v", err)
		return
	}

	// [Insert code to handle the port scan result, such as updating the device in the database]
	log.Printf("Port scan for IP [%s] completed. Found ports: %v", ipv4, ports)
}

// ExecutePortScan performs the port scan using Nmap
func (s *PortScanService) ExecutePortScan(ipv4 string) ([]models.Port, error) {
	cmd := exec.Command("sudo", "/usr/bin/nmap", "-oX", "-", "-O", ipv4)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}

	ports := s.ParseNmapOutput(string(output))
	return ports, nil
}

// ParseNmapOutput parses the XML output of the Nmap command
func (s *PortScanService) ParseNmapOutput(output string) []models.Port {
	var nmapXML models.NmapXML
	err := xml.Unmarshal([]byte(output), &nmapXML)
	if err != nil {
		log.Printf("Error parsing Nmap XML output: %v", err)
		return nil
	}

	var ports []models.Port
	for _, host := range nmapXML.Hosts {
		for _, xmlPort := range host.Ports {
			port := models.Port{
				Number:   xmlPort.PortID,
				Protocol: xmlPort.Protocol,
				State:    xmlPort.State.State,
				Service:  xmlPort.Service.Name,
			}
			log.Printf("Parsed Port: %+v\n", port) // Debugging log
			ports = append(ports, port)
		}
	}
	return ports
}
