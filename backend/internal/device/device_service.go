package device

import (
	"context"
	"log"
	"reconya-ai/models"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type DeviceService struct {
	collection *mongo.Collection
}

func NewDeviceService(db *mongo.Client, dbName, collName string) *DeviceService {
	collection := db.Database(dbName).Collection(collName)
	return &DeviceService{collection: collection}
}

func (s *DeviceService) CreateOrUpdate(device *models.Device) (*models.Device, error) {
	filter := bson.M{"ipv4": device.IPv4}

	// Set LastSeenOnlineAt to the current time
	now := time.Now()
	device.LastSeenOnlineAt = &now

	// Ensure you're creating a map that includes the ports data and LastSeenOnlineAt
	updateData := bson.M{
		"ipv4":                device.IPv4,
		"hostname":            device.Hostname,
		"mac":                 device.MAC,
		"vendor":              device.Vendor,
		"ports":               device.Ports,            // Include port data
		"last_seen_online_at": device.LastSeenOnlineAt, // Update LastSeenOnlineAt
		// Add other fields as necessary but exclude the _id field
	}

	update := bson.M{"$set": updateData}
	opts := options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After)

	var updatedDevice models.Device
	err := s.collection.FindOneAndUpdate(context.Background(), filter, update, opts).Decode(&updatedDevice)
	if err != nil {
		log.Printf("Error saving device with updated ports: %v", err)
		return nil, err
	}
	return &updatedDevice, nil
}

// ParseFromNmap takes a bufferStream (the Nmap output as a string) and parses it into a slice of Device structs.
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
						// If hostname is not equal to the IP address, update the device hostname.
						device.Hostname = &hostname
					} else {
						// If hostname equals the IP address, consider it as no hostname provided.
						// To explicitly clear out an existing hostname, point to an empty string instead of nil.
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

			// Check for the MAC address in the following lines
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
		// Handle nil device
		log.Println("Warning: Attempted to check port scan eligibility for a nil device")
		return false
	}

	now := time.Now()
	if device.PortScanEndedAt != nil && device.PortScanEndedAt.Add(24*time.Hour).After(now) {
		return false
	}
	return true
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

	return devices, nil
}

// FindByIPv4 finds a device by its IPv4 address
func (s *DeviceService) FindByIPv4(ipv4 string) (*models.Device, error) {
	var device models.Device
	err := s.collection.FindOne(context.Background(), bson.M{"ipv4": ipv4}).Decode(&device)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			// No device found is not an error in this context, return nil device and nil error
			return nil, nil
		}
		// Log the error and return
		log.Printf("Error finding device with IPv4 %s: %v", ipv4, err)
		return nil, err
	}
	return &device, nil
}
