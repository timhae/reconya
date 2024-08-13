package device

import (
	"bytes"
	"context"
	"log"
	"net"
	"reconya-ai/internal/config"
	"reconya-ai/internal/network"
	"reconya-ai/models"
	"sort"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type DeviceService struct {
	Config         *config.Config
	collection     *mongo.Collection
	networkService *network.NetworkService
}

func NewDeviceService(db *mongo.Client, collName string, networkService *network.NetworkService, cfg *config.Config) *DeviceService {
	collection := db.Database(cfg.DatabaseName).Collection(collName)
	return &DeviceService{
		Config:         cfg,
		collection:     collection,
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
	if err != nil {
		return nil, err
	}

	s.setTimestamps(device, existingDevice, currentTime)

	updateData := s.buildUpdateData(device, network.ID)

	return s.updateDevice(bson.M{"ipv4": device.IPv4}, updateData)
}

func (s *DeviceService) setTimestamps(device, existingDevice *models.Device, currentTime time.Time) {
	if existingDevice == nil || existingDevice.CreatedAt.IsZero() {
		device.CreatedAt = currentTime
	} else {
		device.CreatedAt = existingDevice.CreatedAt
	}
	device.UpdatedAt = currentTime
}

func (s *DeviceService) buildUpdateData(device *models.Device, networkID primitive.ObjectID) bson.M {
	updateData := bson.M{
		"ipv4":                device.IPv4,
		"hostname":            device.Hostname,
		"mac":                 device.MAC,
		"ports":               device.Ports,
		"last_seen_online_at": device.LastSeenOnlineAt,
		"network_id":          networkID,
		"created_at":          device.CreatedAt,
		"updated_at":          device.UpdatedAt,
	}
	if device.Vendor != nil && *device.Vendor != "" {
		updateData["vendor"] = device.Vendor
	}
	return updateData
}

func (s *DeviceService) updateDevice(filter, updateData bson.M) (*models.Device, error) {
	update := bson.M{"$set": updateData}
	opts := options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After)

	var updatedDevice models.Device
	err := s.collection.FindOneAndUpdate(context.Background(), filter, update, opts).Decode(&updatedDevice)
	if err != nil {
		log.Printf("Error saving device with updated information: %v", err)
		return nil, err
	}
	return &updatedDevice, nil
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
	var devices []models.Device
	cursor, err := s.collection.Find(context.Background(), bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	for cursor.Next(context.Background()) {
		var device models.Device
		if err := cursor.Decode(&device); err != nil {
			return nil, err
		}
		devices = append(devices, device)
	}

	// Sort devices by IP address after fetching
	sortDevicesByIP(devices)

	return devices, nil
}

func (s *DeviceService) FindByID(deviceID string) (*models.Device, error) {
	var device models.Device

	objID, err := primitive.ObjectIDFromHex(deviceID)
	if err != nil {
		log.Printf("Error converting deviceID %s to ObjectID: %v", deviceID, err)
		return nil, err
	}

	err = s.collection.FindOne(context.Background(), bson.M{"_id": objID}).Decode(&device)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		log.Printf("Error finding device with ID %s: %v", deviceID, err)
		return nil, err
	}
	return &device, nil
}

func (s *DeviceService) FindByIPv4(ipv4 string) (*models.Device, error) {
	var device models.Device
	err := s.collection.FindOne(context.Background(), bson.M{"ipv4": ipv4}).Decode(&device)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		log.Printf("Error finding device with IPv4 %s: %v", ipv4, err)
		return nil, err
	}
	return &device, nil
}

func (s *DeviceService) FindAllForNetwork(cidr string) ([]models.Device, error) {
	var devices []models.Device

	network, err := s.networkService.FindByCIDR(cidr)
	if err != nil {
		return nil, err
	}

	if network == nil {
		return devices, nil
	}

	filter := bson.M{"network_id": network.ID}

	cursor, err := s.collection.Find(context.Background(), filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	for cursor.Next(context.Background()) {
		var device models.Device
		if err := cursor.Decode(&device); err != nil {
			return nil, err
		}
		devices = append(devices, device)
	}

	sortDevicesByIP(devices)

	return devices, nil
}

func (s *DeviceService) UpdateDeviceStatuses() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cur, err := s.collection.Find(ctx, bson.M{})
	if err != nil {
		return err
	}
	defer cur.Close(ctx)

	for cur.Next(ctx) {
		var device models.Device
		if err := cur.Decode(&device); err != nil {
			return err
		}

		var status models.DeviceStatus

		if device.LastSeenOnlineAt == nil {
			status = models.DeviceStatusUnknown
		} else {
			duration := time.Since(*device.LastSeenOnlineAt)
			switch {
			case duration <= 3*time.Minute:
				status = models.DeviceStatusOnline
			case duration <= 7*time.Minute:
				status = models.DeviceStatusIdle
			default:
				status = models.DeviceStatusOffline
			}
		}

		if status != device.Status {
			update := bson.M{
				"$set": bson.M{
					"status":     status,
					"updated_at": time.Now(),
				},
			}
			_, err := s.collection.UpdateByID(ctx, device.ID, update)
			if err != nil {
				return err
			}
		}
	}

	return cur.Err()
}
