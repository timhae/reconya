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

// Run executes a port scan for a given IP address and updates device info.
func (s *PortScanService) Run(ipv4 string) {
	log.Printf("Starting port scan for IP [%s]", ipv4)

	device, err := s.DeviceService.FindByIPv4(ipv4)
	if err != nil {
		log.Printf("Error finding device: %v", err)
		return
	}

	if device == nil {
		log.Printf("No device found for IP: %s", ipv4)
		return
	}

	ports, vendor, err := s.ExecutePortScan(ipv4) // Adjusted to also return vendor
	if err != nil {
		log.Printf("Error executing port scan: %v", err)
		return
	}

	device.Ports = ports
	if vendor != "" {
		device.Vendor = &vendor // Update device with vendor info if available
	}

	_, err = s.DeviceService.CreateOrUpdate(device) // Now expects device to include vendor info
	if err != nil {
		log.Printf("Error saving device with updated ports: %v", err)
		return
	}

	log.Printf("Port scan for IP [%s] completed. Found ports: %+v, Vendor: %s", ipv4, ports, vendor)
}

// ExecutePortScan performs the port scan using Nmap and returns ports and vendor.
func (s *PortScanService) ExecutePortScan(ipv4 string) ([]models.Port, string, error) {
	cmd := exec.Command("sudo", "/usr/bin/nmap", "-oX", "-", "-O", ipv4)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, "", err
	}

	ports, vendor := s.ParseNmapOutput(string(output))
	return ports, vendor, nil
}

// ParseNmapOutput parses the XML output of the Nmap command to extract ports and the vendor.
func (s *PortScanService) ParseNmapOutput(output string) ([]models.Port, string) {
	var nmapXML models.NmapXML
	err := xml.Unmarshal([]byte(output), &nmapXML)
	if err != nil {
		log.Printf("Error parsing Nmap XML output: %v", err)
		return nil, ""
	}

	var ports []models.Port
	var vendor string
	for _, host := range nmapXML.Hosts {
		// Iterate through addresses to find the vendor for "mac" type addresses.
		for _, address := range host.Addresses {
			if address.AddrType == "mac" && address.Vendor != "" {
				vendor = address.Vendor // Capture the vendor information.
				break                   // Assuming you're scanning one device at a time or only need the first MAC vendor.
			}
		}

		// Parse port information from each host.
		for _, xmlPort := range host.Ports {
			port := models.Port{
				Number:   xmlPort.PortID,
				Protocol: xmlPort.Protocol,
				State:    xmlPort.State.State,
				Service:  xmlPort.Service.Name,
			}
			ports = append(ports, port)
		}
	}

	return ports, vendor
}
