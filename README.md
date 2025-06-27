# Reconya

Network reconnaissance and asset discovery tool built with Go and React.

![Dashboard Screenshot](screenshots/dashboard.png)

## Overview

Reconya discovers and monitors devices on your network with real-time updates. Suitable for network administrators, security professionals, and home users.

### Features

- Network scanning with nmap integration
- Device identification (MAC addresses, vendor detection, hostnames)
- Real-time monitoring and event logging
- Web-based dashboard
- Device fingerprinting

## Quick Start

1. **Clone and run:**
   ```bash
   git clone https://github.com/Dyneteq/reconya-ai-go.git
   cd reconya-ai-go
   ./setup.sh
   ```

2. **Access at:** `http://localhost:3001`

3. **Login:** `admin` / `password`

## Daily Usage

```bash
./start.sh    # Start
./stop.sh     # Stop  
./logs.sh     # View logs
```

## How to Use

1. Login with `admin` / `password` (or check your `.env` file for custom credentials)
2. Devices will automatically appear as they're discovered  
3. Click on devices to see details and edit information

> **Note:** The setup script creates a `.env` file with default credentials. You can edit this file to change the username, password, and network range.

---

## Advanced

<details>
<summary>Development Setup</summary>

### Backend
```bash
cd backend
cp .env.example .env
go mod download
go run cmd/main.go
```

### Frontend
```bash
cd frontend
npm install
npm start
```
</details>

<details>
<summary>Network Scanning Issues</summary>

For MAC address detection, install nmap:
```bash
# macOS: brew install nmap
# Ubuntu: sudo apt-get install nmap
```

Grant nmap privileges:
```bash
sudo chown root:admin $(which nmap)
sudo chmod u+s $(which nmap)
```
</details>

<details>
<summary>Troubleshooting</summary>

### Common Issues

**Python 3.12 errors**
- Error: `ModuleNotFoundError: No module named 'distutils.spawn'`
- Solution: Use `docker compose` instead of `docker-compose`

**App not accessible**
- Check ports 3001 and 3008 aren't in use
- Verify containers are running: `docker ps`
- Check logs: `./logs.sh`

**Missing MAC addresses**
- Ensure nmap is installed and has proper permissions
- MAC addresses only visible on same network segment

**CORS issues**
- Check CORS config in `backend/middleware/cors.go`
- Verify API routing in nginx.conf

**Docker IP detection issues**
- Error: Reconya detects Docker internal IP instead of host network
- **Solution 1**: Set correct network range in `.env`:
  ```bash
  NETWORK_RANGE=192.168.1.0/24  # Replace with your actual network
  ```
- **Solution 2**: Use host networking for full network access:
  ```bash
  docker compose -f docker-compose.yml -f docker-compose.host.yml up -d
  ```
- **Solution 3**: Enable network capabilities (already enabled by default):
  ```yaml
  cap_add:
    - NET_ADMIN
    - NET_RAW
  ```

## Architecture

- **Backend**: Go API with SQLite database
- **Frontend**: React/TypeScript with Bootstrap
- **Scanning**: Multi-strategy network discovery with nmap integration
- **Deployment**: Docker Compose

## Scanning Algorithm

### Discovery Process

Reconya uses a multi-layered scanning approach that combines nmap integration with native Go implementations:

**1. Network Discovery (Every 30 seconds)**
- Multiple nmap strategies with automatic fallback
- ICMP ping sweeps (privileged mode)
- TCP connect probes to common ports (fallback)
- ARP table lookups for MAC address resolution

**2. Device Identification**
- IEEE OUI database for vendor identification
- Multi-method hostname resolution (DNS, NetBIOS, mDNS)
- Operating system fingerprinting via nmap
- Device type classification based on ports and vendors

**3. Port Scanning (Background workers)**
- Top 100 ports scan for active services
- Service detection and banner grabbing
- Concurrent scanning with worker pool pattern

**4. Web Service Detection**
- Automatic discovery of HTTP/HTTPS services
- Screenshot capture using headless Chrome
- Service metadata extraction (titles, server headers)

### Scanning Strategies

The system attempts these nmap commands in order:

```bash
# Primary: Privileged ICMP scan
sudo nmap -sn --send-ip -T4 -R --system-dns -oX - <network>

# Fallback 1: Unprivileged ICMP
nmap -sn --send-ip -T4 -oX - <network>

# Fallback 2: ARP scan
nmap -sn -PR -T4 -R --system-dns -oX - <network>

# Fallback 3: TCP SYN probe
nmap -sn -PS80,443,22,21,23,25,53,110,111,135,139,143,993,995 -T4 -oX - <network>
```

### Concurrency Model

- **50 concurrent goroutines** for network scanning
- **3 background workers** for port scanning queue
- **Producer-consumer pattern** for efficient resource utilization
- **Database locking** with retry mechanism for data consistency

## Contributing

1. Fork the repository
2. Create feature branch
3. Make changes and test
4. Submit pull request

## License

Creative Commons Attribution-NonCommercial 4.0 International License. Commercial use requires permission.

## 🌟 Please check my other projects!

- **[Tududi](https://tududi.com)** -  Self-hosted task management with hierarchical organization, multi-language support, and Telegram integration
- **[BreachHarbor](https://breachharbor.com)** - Cybersecurity suite for digital asset protection  
- **[Hevetra](https://hevetra.com)** - Digital tracking for child health milestones