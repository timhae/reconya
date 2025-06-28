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

## Important Notice: Docker Implementation Status

⚠️ **Docker networking has been moved to experimental status due to fundamental limitations.**

The fundamental limitation is Docker's network architecture. Even with comprehensive MAC discovery methods, privileged mode, and enhanced capabilities, Docker containers cannot reliably access Layer 2 (MAC address) information across different network segments.

**For full functionality, including complete MAC address discovery, please use the local installation method below.**

Docker files have been moved to the `experimental/` directory for those who want to experiment with containerized deployment, but local installation is the recommended approach.

## Local Installation (Recommended)

### One-Command Installation

The easiest way to install RecoNya with all dependencies:

```bash
git clone https://github.com/Dyneteq/reconya-ai-go.git
cd reconya-ai-go
npm run install
```

This will:
- Detect your operating system (macOS, Windows, Debian, or Red Hat-based)
- Install all required dependencies (Go, Node.js, nmap)
- Configure nmap permissions for MAC address detection
- Set up the RecoNya application
- Install all Node.js dependencies

**After installation, use these commands:**
```bash
npm run start    # Start RecoNya
npm run stop     # Stop RecoNya  
npm run status   # Check service status
npm run uninstall # Uninstall RecoNya
```

Then open your browser to: `http://localhost:3000`  
Default login: `admin` / `password`

### Manual Installation

If you prefer to install manually or the script doesn't work on your system:

#### Prerequisites

1. **Install Go** (1.21 or later): https://golang.org/dl/
2. **Install Node.js** (18 or later): https://nodejs.org/
3. **Install nmap**:
   ```bash
   # macOS
   brew install nmap
   
   # Ubuntu/Debian
   sudo apt-get install nmap
   
   # RHEL/CentOS/Fedora
   sudo yum install nmap  # or dnf install nmap
   ```

4. **Grant nmap privileges** (for MAC address detection):
   ```bash
   sudo chown root:admin $(which nmap)
   sudo chmod u+s $(which nmap)
   ```

#### Setup & Run

1. **Clone the repository:**
   ```bash
   git clone https://github.com/Dyneteq/reconya-ai-go.git
   cd reconya-ai-go
   ```

2. **Setup backend:**
   ```bash
   cd backend
   cp .env.example .env
   # Edit .env file to set your network range and credentials
   go mod download
   ```

3. **Setup frontend:**
   ```bash
   cd ../frontend
   npm install
   ```

4. **Start the application:**

   **Terminal 1 - Backend:**
   ```bash
   cd backend
   go run ./cmd
   ```

   **Terminal 2 - Frontend:**
   ```bash
   cd frontend
   npm start
   ```

5. **Access the application:**
   - Open your browser to: `http://localhost:3000`
   - Default login: `admin` / `password` (check your `.env` file for custom credentials)

## How to Use

1. Login with your credentials (default: `admin` / `password`)
2. Devices will automatically appear as they're discovered on your network
3. Click on devices to see details including:
   - MAC addresses and vendor information
   - Open ports and running services
   - Operating system fingerprints
   - Device screenshots (for web services)
4. Use the network map to visualize device locations
5. Monitor the event log for network activity

## Configuration

Edit the `backend/.env` file to customize:

```bash
LOGIN_USERNAME=admin
LOGIN_PASSWORD=your_secure_password
NETWORK_RANGE="192.168.1.0/24"  # Set to your actual network range
DATABASE_NAME="reconya-dev"
JWT_SECRET_KEY="your_jwt_secret"
SQLITE_PATH="data/reconya-dev.db"
```

## Architecture

- **Backend**: Go API with SQLite database (Port 3008)
- **Frontend**: React/TypeScript with Bootstrap (Port 3000)
- **Scanning**: Multi-strategy network discovery with nmap integration
- **Database**: SQLite for device storage and event logging

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

## Troubleshooting

### Common Issues

**Installation problems**
- Run `npm run status` to check what's missing
- Ensure you have Node.js 14+ installed
- Try running `npm run install` again

**No devices found**
- Verify your network range is correct in `backend/.env` file
- Run `npm run status` to check if nmap is installed and configured
- Check that you're on the same network segment as target devices

**Services won't start**
- Run `npm run stop` to kill any stuck processes
- Check `npm run status` for dependency issues
- Ensure ports 3000 and 3008 are available

**Missing MAC addresses**
- Run `npm run status` to verify nmap permissions
- MAC addresses only visible on same network segment
- Some devices may not respond to ARP requests

**Permission denied errors**
- The installer should handle nmap permissions automatically
- If issues persist, manually run: `sudo chmod u+s $(which nmap)`

**Services keep crashing**
- Check if dependencies are properly installed with `npm run status`
- Verify your `.env` configuration is correct
- Try stopping and restarting: `npm run stop && npm run start`

## Uninstalling RecoNya

To completely remove RecoNya and optionally its dependencies:

```bash
npm run uninstall
```

The uninstall process will:
- Stop any running RecoNya processes
- Remove application files and data
- Remove nmap setuid permissions  
- Optionally remove system dependencies (Go, Node.js, nmap)

**Note:** You'll be asked for confirmation before removing system dependencies since they might be used by other applications.

## Experimental Docker Support

Docker files are available in the `experimental/` directory but are not recommended due to network isolation limitations that prevent proper MAC address discovery. Use local installation for full functionality.

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