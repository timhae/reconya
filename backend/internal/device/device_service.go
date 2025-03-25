package device

import (
	"bytes"
	"context"
	"log"
	"net"
	"reconya-ai/db"
	"reconya-ai/internal/config"
	"reconya-ai/internal/network"
	"reconya-ai/models"
	"sort"
	"strings"
	"time"
)

type DeviceService struct {
	Config         *config.Config
	repository     db.DeviceRepository
	networkService *network.NetworkService
}

func NewDeviceService(deviceRepo db.DeviceRepository, networkService *network.NetworkService, cfg *config.Config) *DeviceService {
	return &DeviceService{
		Config:         cfg,
		repository:     deviceRepo,
		networkService: networkService,
	}
}

func (s *DeviceService) CreateOrUpdate(device *models.Device) (*models.Device, error) {
	currentTime := time.Now()
	device.LastSeenOnlineAt = &currentTime

	network, err := s.networkService.FindOrCreate(s.Config.NetworkCIDR)
	if err != nil {
		return nil, err
	}

	existingDevice, err := s.FindByIPv4(device.IPv4)
	if err != nil && err != db.ErrNotFound {
		return nil, err
	}

	s.setTimestamps(device, existingDevice, currentTime)

	// Set network ID
	device.NetworkID = network.ID

	// Set status if not already set
	if device.Status == "" {
		device.Status = models.DeviceStatusOnline
	}

	// If the device doesn't have a name, use the IP address
	if device.Name == "" {
		device.Name = device.IPv4
	}

	// Save the device
	return s.repository.CreateOrUpdate(context.Background(), device)
}

func (s *DeviceService) setTimestamps(device, existingDevice *models.Device, currentTime time.Time) {
	if existingDevice == nil || existingDevice.CreatedAt.IsZero() {
		device.CreatedAt = currentTime
	} else {
		device.CreatedAt = existingDevice.CreatedAt
	}
	device.UpdatedAt = currentTime
}

func (s *DeviceService) ParseFromNmap(bufferStream string) []models.Device {
	log.Println("Starting Nmap parse")
	var devices []models.Device
	lines := strings.Split(bufferStream, "\n")

	for i := 0; i < len(lines); i++ {
		line := lines[i]
		if strings.HasPrefix(line, "Nmap scan report for") {
			device := models.Device{}
			parts := strings.Fields(line)
			if len(parts) >= 5 {
				ipAddress := strings.Trim(parts[len(parts)-1], "()")
				device.IPv4 = ipAddress

				if len(parts) == 6 {
					hostname := strings.Trim(parts[4], "()")
					if hostname != ipAddress {
						device.Hostname = &hostname
					} else {
						empty := ""
						device.Hostname = &empty
					}
				}

				if device.Hostname != nil {
					log.Printf("Found device - IP: %s, Hostname: %s", device.IPv4, *device.Hostname)
				} else {
					log.Printf("Found device - IP: %s, Hostname: <nil>", device.IPv4)
				}

			}

			if i+2 < len(lines) && strings.HasPrefix(lines[i+2], "MAC Address: ") {
				macParts := strings.Fields(lines[i+2])
				if len(macParts) >= 3 {
					mac := macParts[2]
					device.MAC = &mac
					log.Printf("Device MAC Address: %s", *device.MAC)
				}
			}

			devices = append(devices, device)
		}
	}

	log.Printf("Finished parsing Nmap output. Total devices found: %d", len(devices))
	return devices
}

func (s *DeviceService) EligibleForPortScan(device *models.Device) bool {
	if device == nil {
		log.Println("Warning: Attempted to check port scan eligibility for a nil device")
		return false
	}

	now := time.Now()
	if device.PortScanEndedAt != nil && device.PortScanEndedAt.Add(24*time.Hour).After(now) {
		return false
	}
	return true
}

func sortDevicesByIP(devices []models.Device) {
	sort.Slice(devices, func(i, j int) bool {
		ip1 := net.ParseIP(devices[i].IPv4)
		ip2 := net.ParseIP(devices[j].IPv4)
		return bytes.Compare(ip1, ip2) < 0
	})
}

func (s *DeviceService) FindAll() ([]models.Device, error) {
	ctx := context.Background()
	devices, err := s.repository.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	// Convert pointers to values
	deviceValues := make([]models.Device, len(devices))
	for i, d := range devices {
		deviceValues[i] = *d
	}

	// Sort devices by IP address
	sortDevicesByIP(deviceValues)

	return deviceValues, nil
}

func (s *DeviceService) FindByID(deviceID string) (*models.Device, error) {
	ctx := context.Background()
	device, err := s.repository.FindByID(ctx, deviceID)
	if err == db.ErrNotFound {
		return nil, nil
	}
	if err != nil {
		log.Printf("Error finding device with ID %s: %v", deviceID, err)
		return nil, err
	}
	return device, nil
}

func (s *DeviceService) FindByIPv4(ipv4 string) (*models.Device, error) {
	ctx := context.Background()
	device, err := s.repository.FindByIP(ctx, ipv4)
	if err == db.ErrNotFound {
		return nil, nil
	}
	if err != nil {
		log.Printf("Error finding device with IPv4 %s: %v", ipv4, err)
		return nil, err
	}
	return device, nil
}

func (s *DeviceService) FindAllForNetwork(cidr string) ([]models.Device, error) {
	var deviceValues []models.Device

	network, err := s.networkService.FindByCIDR(cidr)
	if err != nil {
		return nil, err
	}

	if network == nil {
		return deviceValues, nil
	}
	// Get all devices first
	ctx := context.Background()
	allDevices, err := s.repository.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	// Filter devices by network ID
	for _, d := range allDevices {
		// Make sure we're comparing non-empty values
		if d.NetworkID != "" && network.ID != "" && d.NetworkID == network.ID {
			deviceValues = append(deviceValues, *d)
		} else if d.NetworkID == "" && network.ID != "" {
			// If device has no network ID but belongs to the current network
			// The device might be in this network but the ID wasn't saved
			// This is a workaround for existing data
			d.NetworkID = network.ID
			_, err := s.repository.CreateOrUpdate(context.Background(), d)
			if err != nil {
				log.Printf("Error updating device network ID: %v", err)
			} else {
				deviceValues = append(deviceValues, *d)
			}
		} else {
			log.Printf("Skipping device %s (network ID mismatch)", d.IPv4)
		}
	}

	sortDevicesByIP(deviceValues)
	return deviceValues, nil
}

func (s *DeviceService) UpdateDeviceStatuses() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Let the repository handle the device status update logic
	return s.repository.UpdateDeviceStatuses(ctx, 7*time.Minute)
}
