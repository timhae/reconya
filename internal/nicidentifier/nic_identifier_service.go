package nicidentifier

import (
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"reconya-ai/internal/eventlog"
	"reconya-ai/internal/network"
	"reconya-ai/internal/systemstatus"
	"reconya-ai/models"
	"strings"
)

// NicIdentifierService struct
type NicIdentifierService struct {
	NetworkService      *network.NetworkService
	SystemStatusService *systemstatus.SystemStatusService
	EventLogService     *eventlog.EventLogService
}

// NewNicIdentifierService creates a new instance of NicIdentifierService
func NewNicIdentifierService(networkSvc *network.NetworkService, systemStatusSvc *systemstatus.SystemStatusService, eventLogSvc *eventlog.EventLogService) *NicIdentifierService {
	return &NicIdentifierService{
		NetworkService:      networkSvc,
		SystemStatusService: systemStatusSvc,
		EventLogService:     eventLogSvc,
	}
}

// Identify performs the NIC identification process
func (s *NicIdentifierService) Identify() {
	nic := s.getLocalNic()
	cidr := extractCIDR(nic.IPv4)
	publicIP, err := s.getPublicIp()
	if err != nil {
		log.Printf("Failed to get public IP: %v", err)
		return
	}

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

	systemStatus := models.SystemStatus{
		LocalDevice: localDevice,
		Network:     networkEntity,
		PublicIP:    &publicIP,
	}

	_, err = s.SystemStatusService.CreateOrUpdate(&systemStatus)
	if err != nil {
		log.Printf("Failed to create or update system status: %v", err)
		return
	}

	s.EventLogService.CreateOne(&models.EventLog{
		Type:        models.LocalIPFound,
		Description: "Local IPv4 Address found",
	})

	s.EventLogService.CreateOne(&models.EventLog{
		Type:        models.LocalNetworkFound,
		Description: "Local Network found",
	})
}

func (s *NicIdentifierService) getLocalNic() models.NIC {
	interfaces, err := net.Interfaces()
	if err != nil {
		log.Println("Error getting network interfaces:", err)
		return models.NIC{}
	}

	for _, iface := range interfaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			ip, _, err := net.ParseCIDR(addr.String())
			if err != nil || ip.To4() == nil {
				continue
			}

			if !ip.IsLoopback() && strings.HasPrefix(iface.Name, "eth") {
				return models.NIC{Name: iface.Name, IPv4: ip.String()}
			}
		}
	}

	return models.NIC{}
}

func (s *NicIdentifierService) getPublicIp() (string, error) {
	resp, err := http.Get("https://api.ipify.org")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	ip, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(ip), nil
}
