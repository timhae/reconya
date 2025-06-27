# Reconya

Network reconnaissance and asset discovery tool built with Go and React.

## Overview

Reconya discovers and monitors devices on your network with real-time updates. Suitable for network administrators, security professionals, and home users.

### Features

- Network scanning with nmap integration
- Device identification (MAC addresses, vendor detection, hostnames)
- Real-time monitoring and event logging
- Web-based dashboard
- Device fingerprinting

## Installation

### Prerequisites

- Docker and Docker Compose
- Python 3.9+ (for docker-compose compatibility)

### Python 3.12 Users

If using Python 3.12, you may encounter docker-compose issues. Solutions:

**Option 1: Use Docker Compose V2 (recommended)**
```bash
docker compose version  # verify V2 installation
docker compose up -d    # note: no hyphen
```

**Option 2: Legacy docker-compose**
```bash
pip install docker-compose  # requires Python 3.11 or earlier
# or
pipx install docker-compose
```

### Quick Start

1. Clone and setup:
   ```bash
   git clone https://github.com/Dyneteq/reconya-ai-go.git
   cd reconya-ai-go
   ./setup.sh
   ```

2. Access at `http://localhost:3001`

### Manual Setup

1. Configure environment:
   ```bash
   cp .env.example .env
   # edit .env with your settings
   ```

2. Start containers:
   ```bash
   docker compose up -d
   ```

## Usage

1. Login with credentials from your `.env` file
2. Configure network range in settings
3. Run discovery to scan your network
4. Monitor devices in the dashboard

## Development

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

### Network Scanning Setup

For MAC address detection, install nmap:
```bash
# macOS
brew install nmap

# Ubuntu/Debian
sudo apt-get install nmap
```

Grant nmap privileges:
```bash
sudo chown root:admin $(which nmap)
sudo chmod u+s $(which nmap)
```

## Troubleshooting

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

## Architecture

- **Backend**: Go API with SQLite database
- **Frontend**: React/TypeScript with Bootstrap
- **Scanning**: Native Go libraries with nmap integration
- **Deployment**: Docker Compose

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