// MongoDB seed script for Reconya Go
// Run with: mongo mongodb://localhost:27017/reconya-dev seed_database.js

// Clear existing collections
db.networks.drop();
db.devices.drop();
db.event_logs.drop();
db.system_status.drop();

// Create network
const networkId = ObjectId();
db.networks.insertOne({
  _id: networkId,
  cidr: "192.168.1.0/24"
});

// Function to generate random MAC address
function generateMAC() {
  const hexDigits = "0123456789ABCDEF";
  let mac = "";
  for (let i = 0; i < 6; i++) {
    let segment = "";
    for (let j = 0; j < 2; j++) {
      segment += hexDigits.charAt(Math.floor(Math.random() * 16));
    }
    mac += (i > 0 ? ":" : "") + segment;
  }
  return mac;
}

// Vendors for MAC addresses
const vendors = [
  "Apple, Inc.", 
  "Cisco Systems", 
  "Dell Inc.", 
  "Intel Corporate", 
  "Samsung Electronics",
  "Sony Corporation", 
  "Netgear Inc.", 
  "TP-LINK Technologies", 
  "ASUSTek Computer", 
  "Microsoft Corporation"
];

// Services for ports
const services = [
  { port: "22", protocol: "tcp", service: "ssh" },
  { port: "80", protocol: "tcp", service: "http" },
  { port: "443", protocol: "tcp", service: "https" },
  { port: "21", protocol: "tcp", service: "ftp" },
  { port: "25", protocol: "tcp", service: "smtp" },
  { port: "53", protocol: "udp", service: "dns" },
  { port: "3306", protocol: "tcp", service: "mysql" },
  { port: "8080", protocol: "tcp", service: "http-alt" },
  { port: "1433", protocol: "tcp", service: "ms-sql" },
  { port: "5432", protocol: "tcp", service: "postgresql" }
];

// Device types and names
const deviceTypes = [
  { type: "Router", names: ["Main Router", "Guest Router", "Backup Router"] },
  { type: "Switch", names: ["Core Switch", "Access Switch", "Distribution Switch"] },
  { type: "Server", names: ["Web Server", "Database Server", "File Server", "Mail Server", "Application Server"] },
  { type: "Desktop", names: ["Workstation", "Developer PC", "Office PC", "Reception PC"] },
  { type: "Mobile", names: ["iPhone", "Android Phone", "iPad", "Samsung Tablet"] },
  { type: "IoT", names: ["Smart TV", "IP Camera", "Smart Thermostat", "Voice Assistant"] },
  { type: "Printer", names: ["Office Printer", "Color Printer", "Label Printer"] }
];

// Generate devices
const devices = [];
const deviceIds = [];
const now = new Date();
const timeOffset = 1000 * 60 * 60 * 24; // 1 day in milliseconds

for (let i = 1; i < 30; i++) {
  // Skip some IPs to make it realistic
  if ([5, 10, 15, 20, 25].includes(i)) continue;
  
  // Choose random device type
  const deviceType = deviceTypes[Math.floor(Math.random() * deviceTypes.length)];
  const name = deviceType.names[Math.floor(Math.random() * deviceType.names.length)];
  
  // Generate MAC and vendor
  const mac = generateMAC();
  const vendor = vendors[Math.floor(Math.random() * vendors.length)];
  
  // Randomize last seen time
  const lastSeen = new Date(now.getTime() - Math.random() * timeOffset);
  
  // Randomize status based on last seen time
  let status = "offline";
  if (now.getTime() - lastSeen.getTime() < 1000 * 60 * 60) { // Less than 1 hour
    status = Math.random() > 0.3 ? "online" : "idle";
  }
  
  // Generate hostname
  const hostname = `${name.toLowerCase().replace(/\s+/g, '-')}-${i}`;
  
  // Generate random ports
  const ports = [];
  const numPorts = Math.floor(Math.random() * 5); // 0-4 ports
  const usedPortIndices = new Set();
  
  for (let j = 0; j < numPorts; j++) {
    let portIndex;
    do {
      portIndex = Math.floor(Math.random() * services.length);
    } while (usedPortIndices.has(portIndex));
    
    usedPortIndices.add(portIndex);
    const port = services[portIndex];
    ports.push({
      number: port.port,
      protocol: port.protocol,
      state: "open",
      service: port.service
    });
  }
  
  // Create device object
  const deviceId = ObjectId();
  deviceIds.push(deviceId);
  
  const device = {
    _id: deviceId,
    name: `${name} ${i}`,
    ipv4: `192.168.1.${i}`,
    mac: mac,
    vendor: vendor,
    status: status,
    network_id: networkId,
    ports: ports,
    hostname: hostname,
    created_at: new Date(now.getTime() - Math.random() * timeOffset * 7), // Up to 7 days ago
    updated_at: new Date(now.getTime() - Math.random() * timeOffset / 24), // Up to 1 hour ago
    last_seen_online_at: lastSeen
  };
  
  // Only add port scan times for some devices
  if (Math.random() > 0.5) {
    const scanStart = new Date(now.getTime() - Math.random() * timeOffset * 2);
    device.port_scan_started_at = scanStart;
    device.port_scan_ended_at = new Date(scanStart.getTime() + Math.random() * 1000 * 60 * 5); // 0-5 minutes scan duration
  }
  
  devices.push(device);
}

// Insert devices
db.devices.insertMany(devices);

// Event log types
const eventTypes = [
  "DEVICE_DISCOVERED",
  "DEVICE_UPDATED",
  "PORT_SCAN_STARTED",
  "PORT_SCAN_COMPLETED",
  "DEVICE_ONLINE",
  "DEVICE_OFFLINE"
];

// Generate event logs
const eventLogs = [];
for (let i = 0; i < 50; i++) {
  const deviceIndex = Math.floor(Math.random() * deviceIds.length);
  const deviceId = deviceIds[deviceIndex];
  const eventType = eventTypes[Math.floor(Math.random() * eventTypes.length)];
  const device = devices[deviceIndex];
  
  let description = "";
  switch (eventType) {
    case "DEVICE_DISCOVERED":
      description = `Device ${device.name} (${device.ipv4}) discovered on network`;
      break;
    case "DEVICE_UPDATED":
      description = `Device ${device.name} (${device.ipv4}) information updated`;
      break;
    case "PORT_SCAN_STARTED":
      description = `Port scan started for ${device.name} (${device.ipv4})`;
      break;
    case "PORT_SCAN_COMPLETED":
      description = `Port scan completed for ${device.name} (${device.ipv4}), found ${device.ports.length} open ports`;
      break;
    case "DEVICE_ONLINE":
      description = `Device ${device.name} (${device.ipv4}) is now online`;
      break;
    case "DEVICE_OFFLINE":
      description = `Device ${device.name} (${device.ipv4}) is now offline`;
      break;
  }
  
  eventLogs.push({
    type: eventType,
    description: description,
    device_id: deviceId.toString(),
    created_at: new Date(now.getTime() - Math.random() * timeOffset * 14), // Up to 14 days ago
    updated_at: new Date(now.getTime() - Math.random() * timeOffset * 14) // Up to 14 days ago
  });
}

// Insert event logs
db.event_logs.insertMany(eventLogs);

// Insert system status
db.system_status.insertOne({
  version: "1.0.0",
  uptime: Math.floor(Math.random() * 1000000),
  last_scan: new Date(now.getTime() - Math.random() * 1000 * 60 * 60 * 3), // Up to 3 hours ago
  memory_usage: Math.floor(Math.random() * 1000),
  cpu_usage: Math.floor(Math.random() * 100),
  disk_usage: Math.floor(Math.random() * 100),
  created_at: now,
  updated_at: now
});

print("Database seeded successfully!");
print(`Created ${devices.length} devices`);
print(`Created ${eventLogs.length} event logs`);
print("Created 1 network");
print("Created 1 system status entry");