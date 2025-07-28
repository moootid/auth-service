# Auth Service

A high-performance authentication and authorization service built with Go, designed as part of a microservices video processing cluster. This service handles user authentication, JWT token management, and acts as a proxy for video analysis and transcoding operations.

## 🚀 Features

- **User Authentication**: Registration and login with JWT tokens
- **Password Security**: Bcrypt hashing for secure password storage
- **Profile Management**: User profile retrieval and updates
- **Video Operations Proxy**: Secure access to video analysis and transcoding services
- **Database Integration**: PostgreSQL/CockroachDB support with GORM
- **Metrics & Monitoring**: Prometheus metrics endpoint
- **CORS Support**: Cross-origin resource sharing for web applications
- **Health Checks**: Service health monitoring endpoint
- **S3 Integration**: Video file download from Amazon S3

## 🏗️ Architecture

This service acts as an authentication gateway for a video processing cluster, providing:

- **Authentication Layer**: Secure user management and JWT-based authorization
- **Service Proxy**: Routes authenticated requests to video analysis and transcoding services
- **Data Persistence**: User profiles and video processing job tracking
- **Security Middleware**: Request validation and user context injection

## 📋 API Endpoints

### Public Endpoints

- `GET /health` - Service health check
- `GET /metrics` - Prometheus metrics
- `POST /auth/register` - User registration
- `POST /auth/login` - User login

### Protected Endpoints (Require JWT Token)

- `GET /auth/profile` - Get user profile
- `PUT /auth/profile` - Update user profile

### Video Analysis

- `POST /auth/video/analyze` - Submit video for analysis
- `GET /auth/video/analyze` - List user's video analyses
- `GET /auth/video/analyze/{id}` - Get specific analysis details

### Video Transcoding

- `POST /auth/video/transcode` - Submit video for transcoding
- `GET /auth/video/transcode` - List user's transcoding jobs
- `GET /auth/video/transcode/{id}` - Get specific transcoding job details
- `GET /auth/video/transcode/{id}/download` - Download processed video from S3

## 🛠️ Tech Stack

- **Language**: Go 1.24
- **Web Framework**: Gorilla Mux
- **Database**: PostgreSQL/CockroachDB with GORM
- **Authentication**: JWT tokens with golang-jwt/jwt
- **Password Hashing**: bcrypt
- **Cloud Storage**: AWS S3
- **Monitoring**: Prometheus metrics
- **Containerization**: Docker

## 🚀 Quick Start

### Using Docker Hub Image

```bash
# Pull the latest image
docker pull moootid/auth-service:latest

# Run with environment variables
docker run -p 8080:8080 \
  -e DB_HOST=your-db-host \
  -e DB_PORT=26257 \
  -e DB_USER=your-db-user \
  -e DB_PASSWORD=your-db-password \
  -e DB_NAME=microservices \
  -e JWT_SECRET=your-secret-key \
  moootid/auth-service:latest
```

### Using Docker Compose

1. Clone the repository:

```bash
git clone https://github.com/moootid/auth-service.git
cd auth-service
```

2. Create a `.env` file with your configuration:

```env
DB_HOST=localhost
DB_PORT=26257
DB_USER=root
DB_PASSWORD=your-password
DB_NAME=microservices
DB_SSLMODE=disable
JWT_SECRET=your-secret-key-here
```

3. Run with Docker Compose:

```bash
docker-compose up -d
```

### Building from Source

1. Clone the repository:

```bash
git clone https://github.com/moootid/auth-service.git
cd auth-service
```

2. Install dependencies:

```bash
go mod download
```

3. Set environment variables:

```bash
export DB_HOST=localhost
export DB_PORT=26257
export DB_USER=root
export DB_PASSWORD=your-password
export DB_NAME=microservices
export JWT_SECRET=your-secret-key
```

4. Run the service:

```bash
go run main.go
```

The service will start on port 8080 by default.

## 🔧 Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `PORT` | Service port | `8080` |
| `DB_HOST` | Database host | `localhost` |
| `DB_PORT` | Database port | `26257` |
| `DB_USER` | Database user | `root` |
| `DB_PASSWORD` | Database password | `""` |
| `DB_NAME` | Database name | `microservices` |
| `DB_SSLMODE` | Database SSL mode | `disable` |
| `JWT_SECRET` | JWT signing secret | `your-secret-key` |

### Database Setup

The service uses GORM for database operations and supports PostgreSQL/CockroachDB. Ensure your database is running and accessible with the provided credentials.

Required tables will be automatically created by GORM migrations:

- `users` - User profiles and authentication data
- `video_analyses` - Video analysis job tracking
- `transcoding_jobs` - Video transcoding job tracking

## 📝 Usage Examples

### Register a New User

```bash
curl -X POST http://localhost:8080/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "securepassword"
  }'
```

### Login

```bash
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "securepassword"
  }'
```

### Access Protected Endpoint

```bash
# Use the JWT token from login response
curl -X GET http://localhost:8080/auth/profile \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

### Submit Video for Analysis

```bash
curl -X POST http://localhost:8080/auth/video/analyze \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "video_url": "https://example.com/video.mp4",
    "analysis_type": "content_detection"
  }'
```

## 🔍 Monitoring

- **Health Check**: `GET /health` - Returns service status
- **Metrics**: `GET /metrics` - Prometheus metrics endpoint

## 🏛️ Project Structure

```text
auth-service/
├── main.go                 # Application entry point
├── database/
│   └── db.go              # Database connection and configuration
├── handlers/
│   ├── auth.go            # Authentication handlers
│   ├── analyze.go         # Video analysis proxy handlers
│   └── transcode.go       # Video transcoding proxy handlers
├── middleware/
│   ├── auth.go            # JWT authentication middleware
│   └── metrics.go         # Prometheus metrics middleware
├── models/
│   ├── user.go            # User data models
│   ├── video_analyses.go  # Video analysis models
│   └── transcoding_job.go # Transcoding job models
├── Dockerfile             # Container configuration
├── docker-compose.yml     # Multi-container setup
├── go.mod                 # Go module dependencies
└── go.sum                 # Dependency checksums
```

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🔗 Related Services

This auth service is part of a larger video processing cluster:

- **Video Analysis Service**: Handles video content analysis
- **Video Transcoding Service**: Manages video format conversion and processing
- **Storage Service**: Manages file uploads and downloads

---

**Docker Hub**: [`moootid/auth-service:latest`](https://hub.docker.com/r/moootid/auth-service)  
**GitHub**: [`moootid/auth-service`](https://github.com/moootid/auth-service)
