# Chat-Server Microservices - Complete API Scheme

## Services

### 1. Auth Service (port 8001)

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | /auth/register | Register new user (username, email, phone, password) |
| POST | /auth/login | Login with email or phone |
| POST | /auth/refresh | Refresh JWT token |
| GET | /auth/verify | Verify token & get user info |

### 2. Profile Service (port 8002)

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | /profiles/:userId | Get user profile |
| PUT | /profiles/:userId | Update profile |
| GET | /profiles/search?q=query | Search users by username or phone |
| POST | /channels | Create channel |
| GET | /channels | Get user's channels |
| GET | /channels/search?q=query | Search channels |
| GET | /channels/:channelId | Get channel info |
| POST | /channels/:channelId/join | Join channel |
| POST | /channels/:channelId/leave | Leave channel |
| GET | /channels/:channelId/members | Get channel members |

### 3. Relay Service (port 8003)

| Method | Endpoint | Description |
|--------|----------|-------------|
| WS | /ws/connect | WebSocket for real-time messaging |
| GET | /health | Health check |

---

## Data Models

### User (Auth)
```json
{
  "id": "uuid",
  "username": "string",
  "email": "string",
  "phone": "+1234567890",
  "created_at": "timestamp",
  "updated_at": "timestamp"
}
```

### Profile
```json
{
  "user_id": "uuid",
  "username": "string", 
  "display_name": "string",
  "avatar_url": "url",
  "status": "string",
  "phone": "+1234567890"
}
```

### Channel
```json
{
  "id": "uuid",
  "name": "string",
  "created_by": "uuid"
}
```

### Channel Member
```json
{
  "user_id": "uuid",
  "username": "string",
  "role": "admin|member"
}
```

---

## Request/Response Examples

### Register
```http
POST /auth/register
Content-Type: application/json

{
  "username": "johndoe",
  "email": "john@example.com",
  "phone": "+1234567890",
  "password": "securepass123"
}

Response (201):
{
  "token": "eyJhbGci...",
  "user": {
    "id": "uuid",
    "username": "johndoe",
    "email": "john@example.com",
    "phone": "+1234567890",
    "created_at": "2026-04-25T20:00:00Z"
  }
}
```

### Login
```http
POST /auth/login
Content-Type: application/json

{
  "email": "john@example.com",
  "password": "securepass123"
}
# OR
{
  "phone": "+1234567890",
  "password": "securepass123"
}
```

### Search Users
```http
GET /profiles/search?q=john
GET /profiles/search?q=+1234567890
# Returns: [Profile, Profile, ...]
```

### Update Profile
```http
PUT /profiles/{userId}
Content-Type: application/json

{
  "display_name": "John Doe",
  "avatar_url": "https://example.com/avatar.jpg",
  "status": "Online",
  "phone": "+1234567890"
}
```

### Create Channel
```http
POST /channels
Content-Type: application/json

{
  "name": "general-chat"
}

Response:
{
  "id": "uuid",
  "name": "general-chat", 
  "created_by": "uuid"
}
```

### Channel Membership
```http
POST /channels/{channelId}/join
POST /channels/{channelId}/leave
GET /channels/{channelId}/members
```

---

## WebSocket Messages

### Connect
```javascript
ws://localhost:8003/ws/connect?token=JWT_TOKEN
```

### Message Format
```json
{
  "type": "message",
  "from": "user-id",
  "to": "user-id",
  "channel_id": "channel-id",
  "content": "encrypted-message",
  "timestamp": "ISO8601"
}
```

### Message Types
- `message` - chat message
- `typing` - user is typing
- `read` - message read receipt
- `online` - user online status

---

## iOS Swift Integration

### Configuration
```swift
struct APIConfig {
    static let authBase = "http://localhost:8001"
    static let profileBase = "http://localhost:8002"
    static let relayWebSocket = "ws://localhost:8003/ws/connect"
}
```

### API Client
```swift
class ChatAPIClient {
    static let shared = ChatAPIClient()
    var token: String?
    
    // MARK: - Auth
    func register(username: String, email: String, phone: String, password: String) async throws -> AuthResponse
    func login(identifier: String, password: String) async throws -> AuthResponse  // identifier = email OR phone
    func verify() async throws -> User
    
    // MARK: - Profile  
    func getProfile(userId: String) async throws -> Profile
    func updateProfile(userId: String, update: ProfileUpdate) async throws
    func searchUsers(query: String) async throws -> [Profile]
    
    // MARK: - Channels
    func createChannel(name: String) async throws -> Channel
    func getChannels() async throws -> [Channel]
    func searchChannels(query: String) async throws -> [Channel]
    func joinChannel(channelId: String) async throws
    func leaveChannel(channelId: String) async throws
    func getChannelMembers(channelId: String) async throws -> [Member]
}
```

### WebSocket Client (using Starscream)
```swift
import Starscream

class ChatWebSocket: WebSocketDelegate {
    var socket: WebSocket?
    
    func connect(token: String) {
        var request = URLRequest(url: URL(string: "ws://localhost:8003/ws/connect?token=\(token)")!)
        socket = WebSocket(request: request)
        socket?.delegate = self
        socket?.connect()
    }
    
    func sendMessage(to: String, content: String, channelId: String? = nil) {
        let msg = WSMessage(type: "message", from: currentUserId, to: to, channelId: channelId, content: content, timestamp: Date())
        socket?.write(string: msg.toJSON())
    }
}
```

---

## Database Schema

### users
- id (UUID, PK)
- username (VARCHAR 50, UNIQUE)
- email (VARCHAR 255, UNIQUE)
- phone (VARCHAR 20)
- password_hash (VARCHAR 255)
- created_at, updated_at (TIMESTAMP)

### profiles
- user_id (UUID, FK → users.id)
- display_name (VARCHAR 100)
- avatar_url (TEXT)
- status (TEXT)
- phone (VARCHAR 20)
- updated_at (TIMESTAMP)

### channels
- id (UUID, PK)
- name (VARCHAR 100)
- created_by (UUID → users.id)
- created_at, updated_at (TIMESTAMP)

### channel_members
- channel_id (UUID → channels.id)
- user_id (UUID → users.id)
- role (VARCHAR 20) - admin/member
- joined_at (TIMESTAMP)
- PRIMARY KEY (channel_id, user_id)