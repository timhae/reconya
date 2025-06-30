package nicidentifier

import (
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"reconya-ai/internal/config"
	"reconya-ai/internal/device"
	"reconya-ai/internal/eventlog"
	"reconya-ai/internal/network"
	"reconya-ai/internal/systemstatus"
	"reconya-ai/models"
)

type NicIdentifierService struct {
	NetworkService      *network.NetworkService
	SystemStatusService *systemstatus.SystemStatusService
	EventLogService     *eventlog.EventLogService
	DeviceService       *device.DeviceService
	Config              *config.Config
}

func NewNicIdentifierService(
	networkService *network.NetworkService,
	systemStatusService *systemstatus.SystemStatusService,
	eventLogService *eventlog.EventLogService,
	deviceService *device.DeviceService,
	config *config.Config) *NicIdentifierService {
	return &NicIdentifierService{
		NetworkService:      networkService,
		SystemStatusService: systemStatusService,
		EventLogService:     eventLogService,
		DeviceService:       deviceService,
		Config:              config,
	}
}

func (s *NicIdentifierService) Identify() {
	log.Printf("Attempting network identification")
	nic := s.getLocalNic()
	fmt.Printf("NIC: %v\n", nic)
	
	// Use configured network CIDR instead of detected interface CIDR
	cidr := s.Config.NetworkCIDR
	log.Printf("Using configured network CIDR: %s", cidr)
	
	publicIP, err := s.getPublicIp()
	if err != nil {
		log.Printf("Failed to get public IP: %v", err)
		return
	}
	log.Printf("Public IP Address found: [%v]", publicIP)

	networkEntity, err := s.NetworkService.FindOrCreate(cidr)
	if err != nil {
		log.Printf("Failed to find or create network: %v", err)
		return
	}

	localDevice := models.Device{
		Name:   nic.Name,
		IPv4:   nic.IPv4,
		Status: models.DeviceStatusOnline,
	}

	savedDevice, err := s.DeviceService.CreateOrUpdate(&localDevice)
	if err != nil {
		log.Printf("Failed to save or update local device: %v", err)
		return
	}

	systemStatus := models.SystemStatus{
		LocalDevice: *savedDevice,
		NetworkID:   networkEntity.ID,
		PublicIP:    &publicIP,
	}

	_, err = s.SystemStatusService.CreateOrUpdate(&systemStatus)
	if err != nil {
		log.Printf("Failed to create or update system status: %v", err)
		return
	}

	device := savedDevice.ID
	s.EventLogService.CreateOne(&models.EventLog{
		Type:     models.LocalIPFound,
		DeviceID: &device,
	})

	s.EventLogService.CreateOne(&models.EventLog{
		Type: models.LocalNetworkFound,
	})
}

func (s *NicIdentifierService) getLocalNic() models.NIC {
	interfaces, err := net.Interfaces()
	if err != nil {
		log.Println("Error getting network interfaces:", err)
		return models.NIC{}
	}

	var candidates []models.NIC
	var dockerInterfaces []models.NIC

	for _, iface := range interfaces {
		fmt.Printf("Checking interface: %s\n", iface.Name)
		if iface.Flags&net.FlagUp == 0 {
			fmt.Printf("Skipping %s: interface is down\n", iface.Name)
			continue
		}
		if iface.Flags&net.FlagLoopback != 0 {
			fmt.Printf("Skipping %s: interface is loopback\n", iface.Name)
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			fmt.Printf("Skipping %s: error getting addresses: %v\n", iface.Name, err)
			continue
		}

		for _, addr := range addrs {
			ip, _, err := net.ParseCIDR(addr.String())
			if err != nil || ip.To4() == nil {
				fmt.Printf("Skipping address %s on %s: not a valid IPv4\n", addr.String(), iface.Name)
				continue
			}

			if !ip.IsLoopback() {
				nic := models.NIC{Name: iface.Name, IPv4: ip.String()}
				
				// Check if this is a Docker or container network
				if s.isDockerOrContainerNetwork(ip.String()) {
					fmt.Printf("Found Docker/container interface: %s with IPv4: %s\n", iface.Name, ip.String())
					dockerInterfaces = append(dockerInterfaces, nic)
				} else {
					fmt.Printf("Found potential host interface: %s with IPv4: %s\n", iface.Name, ip.String())
					candidates = append(candidates, nic)
				}
			}
		}
	}

	// Prefer non-Docker interfaces
	if len(candidates) > 0 {
		// Prioritize common home/office networks
		for _, nic := range candidates {
			if s.isCommonPrivateNetwork(nic.IPv4) {
				fmt.Printf("Selected preferred interface: %s with IPv4: %s\n", nic.Name, nic.IPv4)
				return nic
			}
		}
		// If no common private networks, return first candidate
		fmt.Printf("Selected first non-Docker interface: %s with IPv4: %s\n", candidates[0].Name, candidates[0].IPv4)
		return candidates[0]
	}

	// Fallback to Docker interfaces if no others available
	if len(dockerInterfaces) > 0 {
		fmt.Printf("Using Docker interface as fallback: %s with IPv4: %s\n", dockerInterfaces[0].Name, dockerInterfaces[0].IPv4)
		return dockerInterfaces[0]
	}

	return models.NIC{}
}

// isDockerOrContainerNetwork checks if an IP belongs to common container networks
func (s *NicIdentifierService) isDockerOrContainerNetwork(ip string) bool {
	// Common Docker and container network ranges
	dockerRanges := []string{
		"172.17.0.0/16",    // Default Docker bridge
		"172.18.0.0/16",    // Docker custom networks
		"172.19.0.0/16",
		"172.20.0.0/16",
		"172.21.0.0/16",
		"172.22.0.0/16",
		"172.23.0.0/16",
		"172.24.0.0/16",
		"172.25.0.0/16",
		"172.26.0.0/16",
		"172.27.0.0/16",
		"172.28.0.0/16",
		"172.29.0.0/16",
		"172.30.0.0/16",
		"172.31.0.0/16",
	}

	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return false
	}

	for _, cidr := range dockerRanges {
		_, network, err := net.ParseCIDR(cidr)
		if err != nil {
			continue
		}
		if network.Contains(parsedIP) {
			return true
		}
	}
	return false
}

// isCommonPrivateNetwork checks if an IP belongs to common home/office networks
func (s *NicIdentifierService) isCommonPrivateNetwork(ip string) bool {
	// Common home/office network ranges
	commonRanges := []string{
		"192.168.0.0/16",   // Most common home networks
		"10.0.0.0/8",       // Corporate networks
	}

	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return false
	}

	for _, cidr := range commonRanges {
		_, network, err := net.ParseCIDR(cidr)
		if err != nil {
			continue
		}
		if network.Contains(parsedIP) {
			return true
		}
	}
	return false
}

func (s *NicIdentifierService) getPublicIp() (string, error) {
	resp, err := http.Get("https://api.ipify.org")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	ip, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(ip), nil
}
