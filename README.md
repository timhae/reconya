# Reconya AI 

A powerful network reconnaissance and asset discovery tool built with Go and React, designed to help map and monitor network devices with precision and elegance.

<div align="center">
  <img src="screenshots/dashboard.png" alt="Reconya Dashboard" width="80%">
</div>

## 🌟 Overview

Reconya AI Go helps users discover, identify, and monitor devices on their network with real-time updates and an intuitive interface. Our tool is perfect for network administrators, security professionals, and tech enthusiasts.

### ✨ Key Features

- 🔎 **Advanced Network Scanning** - Comprehensive port scanning and ping sweeping
- 🧩 **Device Identification** - Accurate identification and classification of network devices
- 🕸️ **Network Visualization** - Clear and interactive network topology mapping
- 📊 **Event Monitoring** - Real-time logging and monitoring of network events
- 🖥️ **Modern Dashboard** - Sleek, responsive web interface for all devices

## 🚀 Installation

### 📋 Prerequisites

- 🐳 Docker and Docker Compose (for production deployment)
- 🔹 Go 1.16+
- 🟢 Node.js 14+ and npm

### 💻 Development Setup

#### 🔧 Backend Setup

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

#### 🎨 Frontend Setup

1. Install dependencies:
   ```
   cd frontend
   npm install
   ```

2. Configure environment variables:
   ```
   cp .env.example .env
   ```
   Adjust the `.env` file as needed.

3. Start the development server:
   ```
   ../scripts/dev_start_frontend.sh
   ```

4. Access the web interface at `http://localhost:3000`

### 🏭 Production Deployment

For production environments, we recommend using Docker Compose:

1. Configure environment variables:
   ```
   # Backend
   cd backend
   cp .env.example .env
   # Edit .env with production values

   # Frontend (optional)
   cd ../frontend
   cp .env.example .env.production
   # Edit .env.production with production values
   ```

2. Build and start the containers:
   ```
   docker-compose up -d
   ```

3. Access the application at `http://your-server-ip` (port 80)

#### ⚙️ Production Customization

- 🔧 **NGINX Configuration**: Edit `frontend/nginx.conf` to customize the web server settings
- 🔒 **SSL/TLS**: For HTTPS, use a reverse proxy like Traefik or modify the NGINX configuration
- 💾 **Persistence**: Database files are stored in the `backend/data` directory. Consider mounting this to a persistent volume
- 🔄 **Auto-updates**: Set up a CI/CD pipeline for automated deployments

## 📝 Usage

<div align="center">
  <img src="screenshots/event-logs.png" alt="Event Logs" width="80%">
</div>

1. 🔑 Log in with credentials configured in your `.env` file
2. 🌐 Configure network range to scan in the settings
3. 🔍 Run discovery to find devices on your network
4. 📱 View and manage discovered devices in the dashboard
5. 📊 Monitor network activity through event logs

## 🏗️ Architecture

- 🔙 **Backend**: Go API server with MongoDB or SQLite for storage
- 🖌️ **Frontend**: React/TypeScript web application with responsive Bootstrap UI
- 🔍 **Scanning**: Network operations performed through native Go libraries
- 🔄 **Real-time Updates**: Polling system with configurable intervals

### 💾 Database Options

The application supports two database options to fit different deployment needs:

#### 🔷 MongoDB
- 🌐 Good for distributed deployments
- 🚀 Allows for horizontal scaling
- 🔌 Requires a MongoDB instance

#### 🔶 SQLite (Recommended for Single-User Deployments)
- 📦 Self-contained, no separate database service required
- 🧩 Simpler setup with minimal configuration
- 🏠 Perfect for personal or small deployments
- 🪶 Lightweight and portable

## 🔐 Security Notes

- 🔑 Always use strong passwords in production environments
- 🔒 Use an `.env` file for all sensitive configuration
- 🛡️ Never expose the backend API directly to the internet
- 👮 Run with least privilege required for network scanning
- 🔄 Keep dependencies updated to patch security vulnerabilities
- 🧪 Regularly test your deployment for security issues

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. 🍴 Fork the repository
2. 🌿 Create your feature branch (`git checkout -b feature/amazing-feature`)
3. 💾 Commit your changes (`git commit -m 'Add some amazing feature'`)
4. 🚀 Push to the branch (`git push origin feature/amazing-feature`)
5. 🔍 Open a Pull Request with a detailed description

## 📄 License

This project is licensed under the Creative Commons Attribution-NonCommercial 4.0 International License - see the [LICENSE](LICENSE) file for details. Commercial use requires explicit permission from the author.

## ✨ Features Added in Latest Update

- [Nmap](https://nmap.org/) for inspiration and scanning techniques
- [React](https://reactjs.org/) for the frontend framework
- [Go](https://golang.org/) for the backend language
