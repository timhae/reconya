# Reconya AI 

A network reconnaissance and asset discovery tool built with Go and React, designed to help map and monitor network devices.

## Overview

Reconya AI Go helps users discover, identify, and monitor devices on their network. Key features include:

- Network scanning (port scanning and ping sweeping)
- Device identification and classification
- Network topology visualization
- Event logging and monitoring
- Web-based dashboard interface

## Installation

### Prerequisites

- Docker and Docker Compose
- Go 1.16+
- Node.js 14+ and npm

### Backend Setup

1. Clone the repository:
   ```
   git clone https://github.com/yourusername/reconya-ai-go.git
   cd reconya-ai-go
   ```

2. Set up environment variables:
   ```
   cd backend
   cp .env.example .env
   ```
   Edit `.env` with your configuration.

3. Choose your database:

   **Option 1: Using SQLite (Recommended for simplicity)**
   
   Add the following to your `.env` file:
   ```
   DATABASE_TYPE=sqlite
   SQLITE_PATH=data/reconya.db
   ```
   No additional database setup required!

   **Option 2: Using MongoDB**
   
   Add the following to your `.env` file:
   ```
   DATABASE_TYPE=mongodb
   MONGODB_URI=mongodb://localhost:27017
   ```
   
   Then start MongoDB:
   ```
   ../scripts/dev_create_mongo.sh
   ```

4. Build and run the backend:
   ```
   ../scripts/dev_start_backend.sh
   ```

5. Migration from MongoDB to SQLite (if you were using MongoDB):
   ```
   ../scripts/migrate_to_sqlite.sh
   ```

### Frontend Setup

1. Install dependencies:
   ```
   cd frontend
   npm install
   ```

2. Start the development server:
   ```
   ../scripts/dev_start_frontend.sh
   ```

3. Access the web interface at `http://localhost:3000`

## Usage

1. Log in with credentials configured in your `.env` file
2. Configure network range to scan
3. Run discovery to find devices
4. View and manage discovered devices

## Architecture

- **Backend**: Go API server with MongoDB or SQLite for storage
- **Frontend**: React/TypeScript web application
- **Scanning**: Network operations performed through native Go libraries

### Database Options

The application supports two database options:

#### MongoDB (Default)
- Good for distributed deployments
- Allows for horizontal scaling
- Requires a MongoDB instance

#### SQLite (Recommended for Single-User Deployments)
- Self-contained, no separate database service required
- Simpler setup
- Perfect for personal or small deployments
- Lightweight and portable

## Security Notes

- Always use strong passwords in production
- Use an `.env` file for all sensitive configuration
- Never expose the backend API directly to the internet
- Run with least privilege required for network scanning

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the Creative Commons Attribution-NonCommercial 4.0 International License - see the [LICENSE](LICENSE) file for details. Commercial use requires explicit permission from the author.

## Acknowledgments

- [Nmap](https://nmap.org/) for inspiration and scanning techniques
- [React](https://reactjs.org/) for the frontend framework
- [Go](https://golang.org/) for the backend language
