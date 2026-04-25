# Chat-Server Deployment Commands for Server

# === STEP 1: Run migrations (connect to correct database) ===

docker exec -i chat-server-postgres-1 psql -U postgres -d chatserver < migrations/001_init.sql

# === STEP 2: Run more migrations ===
docker exec -i chat-server-postgres-1 psql -U postgres -d chatserver < migrations/002_channels.sql
docker exec -i chat-server-postgres-1 psql -U postgres -d chatserver -c "ALTER TABLE profiles ADD COLUMN IF NOT EXISTS phone VARCHAR(20);"
docker exec -i chat-server-postgres-1 psql -U postgres -d chatserver -c "ALTER TABLE users ADD COLUMN IF NOT EXISTS phone VARCHAR(20);"

# === STEP 3: Start all services ===
docker-compose up -d --build

# === STEP 4: Verify ===
docker-compose ps
curl http://localhost:8001/auth/login -X POST -H "Content-Type: application/json" -d '{"email":"test@test.com","password":"test"}'