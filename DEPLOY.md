# Deploy chat-server to production server

## Prerequisites
- Docker installed on server
- PostgreSQL client (optional)

## Quick Deploy

### 1. Copy project to server
```bash
scp -r chat-server/ user@server:/root/
# or via git
git clone <your-repo> && cd chat-server
```

### 2. Update environment (production)
Create `.env` file:
```bash
cp .env.example .env
nano .env  # Edit with production values
```

Contents:
```bash
DB_HOST=postgres
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=YourSecurePassword123!
DB_NAME=chatserver
JWT_SECRET=GenerateRandomSecretKeyHere!
PORT=8001
PROFILE_PORT=8002
RELAY_PORT=8003
```

### 3. Run on server
```bash
cd chat-server

# Build and start all services
docker-compose up -d --build

# Check status
docker-compose ps

# View logs
docker-compose logs -f
```

### 4. Verify services
```bash
# Auth service
curl http://localhost:8001/health

# Profile service  
curl http://localhost:8002/health

# Relay service
curl http://localhost:8003/health
```

---

## Production Considerations

### 1. Change default JWT secret
```bash
openssl rand -hex 32  # Generate secure secret
```

### 2. Use production PostgreSQL
For production, use managed PostgreSQL (AWS RDS, Cloud SQL, etc.):
```yaml
# docker-compose.yaml - comment out postgres service and use external DB
# profiles:
#   DB_HOST: your-prod-db.rds.amazonaws.com
#   DB_PORT: 5432
```

### 3. SSL/TLS
Put behind nginx or cloudflare for HTTPS:
```nginx
# nginx.conf (example)
server {
    listen 443 ssl;
    ssl_certificate /path/to/cert.pem;
    ssl_certificate_key /path/to/key.pem;
    
    location / {
        proxy_pass http://localhost:8001;
    }
}
```

### 4. Firewall
```bash
# Allow ports
sudo ufw allow 8001/tcp
sudo ufw allow 8002/tcp
sudo ufw allow 8003/tcp
```

### 5. Auto-start on boot
```bash
# Create systemd service
sudo nano /etc/systemd/system/chat-server.service

[Unit]
Description=Chat Server
After=docker.service

[Service]
WorkingDirectory=/root/chat-server
ExecStart=/usr/local/bin/docker-compose up
ExecStop=/usr/local/bin/docker-compose down
Restart=always

[Install]
WantedBy=multi-user.target
```

Enable:
```bash
sudo systemctl daemon-reload
sudo systemctl enable chat-server
sudo systemctl start chat-server
```

---

## Common Commands

```bash
# Restart services
docker-compose restart

# Stop services
docker-compose down

# View logs
docker-compose logs -f auth-service
docker-compose logs -f profile-service  
docker-compose logs -f relay-service

# Rebuild after code changes
docker-compose up -d --build

# Scale relay service for more connections
docker-compose up -d --scale relay-service=3
```

---

## iOS Client Configuration

Update API URLs for production:
```swift
struct APIConfig {
    # Development
    static let authBase = "http://localhost:8001"
    static let profileBase = "http://localhost:8002"
    static let relayWebSocket = "ws://localhost:8003/ws/connect"
    
    # Production - change to your server IP/domain
    // static let authBase = "https://api.yourdomain.com"
    // static let profileBase = "https://api.yourdomain.com"
    // static let relayWebSocket = "wss://api.yourdomain.com/ws/connect"
}
```