# CGC Image Service Backend

A Go-based image generation service using Google Agent Development Kit (ADK) for intelligent provider selection and automatic fallback between Freepik, Google Imagen, and Leonardo AI.

## Features

- **Google Agent Development Kit (ADK)**: Intelligent orchestrator agent for provider selection
- **Automatic Fallback**: Switches between providers when quotas or rate limits are hit
- **Random Load Balancing**: Treats all providers equally until errors occur
- **DigitalOcean Spaces CDN**: All images served from CDN (no local storage)
- **Valkey Vote Persistence**: Redis-compatible caching for user votes and leaderboards
- **Real-time Leaderboard**: Track provider performance with win rates and statistics
- **RESTful API**: Clean endpoints with JSON responses
- **Health Monitoring**: Provider status tracking and health endpoints

## Architecture

### Agent Development Kit (ADK)
- **Orchestrator Agent**: Manages provider selection and fallback logic
- **Provider Agents**: Wrap each image generation service with agent capabilities
- **Random Selection**: Ensures equal treatment of all providers until errors occur
- **Automatic Fallback**: Seamlessly switches to working providers when others fail

### Supported Providers
- **Freepik**: Official API with $5 free credit
- **Google Imagen**: High-quality image generation with Imagen 3.0
- **Leonardo AI**: Creative AI with free tier and async generation

## Prerequisites

Set up API keys as environment variables:

```bash
export FREEPIK_API_KEY=your_freepik_api_key
export GOOGLE_API_KEY=your_google_api_key
export LEONARDO_API_KEY=your_leonardo_api_key
```

## Installation

1. Install dependencies:
```bash
cd backend
go mod tidy
```

2. Run the server:
```bash
go run cmd/server/main.go
```

The server will start on `http://localhost:8080` by default.

## API Endpoints

### Generate Images
```bash
POST /api/v1/generate
```

**Request:**
```json
{
  "prompt": "Robot holding a red skateboard",
  "count": 4
}
```

**Response:**
```json
{
  "data": {
    "images": [
      {
        "id": "uuid",
        "filename": "freepik-uuid.png",
        "path": "https://cgc-lb-and-cdn-content.nyc3.digitaloceanspaces.com/freepik-uuid.png",
        "url": "https://cgc-lb-and-cdn-content.nyc3.digitaloceanspaces.com/freepik-uuid.png",
        "size": 1234567
      }
    ],
    "provider": "freepik",
    "success": true,
    "request_id": "uuid",
    "duration": "2.5s"
  }
}
```

### Get Random Image Pair
```bash
GET /api/v1/images/pair
```

**Response:**
```json
{
  "data": {
    "pair_id": "uuid",
    "left": {
      "id": "freepik-uuid",
      "filename": "freepik-uuid.png",
      "url": "https://cdn-url/freepik-uuid.png",
      "provider": "freepik"
    },
    "right": {
      "id": "leonardo-ai-uuid",
      "filename": "leonardo-ai-uuid.png",
      "url": "https://cdn-url/leonardo-ai-uuid.png",
      "provider": "leonardo-ai"
    }
  }
}
```

### Submit Vote
```bash
POST /api/v1/images/rate
```

**Request:**
```json
{
  "pair_id": "uuid",
  "winner": "left",
  "left_id": "freepik-uuid",
  "right_id": "leonardo-ai-uuid"
}
```

**Response:**
```json
{
  "data": {
    "success": true,
    "pair_id": "uuid",
    "winner": "left",
    "message": "Rating submitted successfully",
    "timestamp": "2025-10-06T12:00:00Z"
  }
}
```

### Get Leaderboard
```bash
GET /api/v1/leaderboard
```

**Response:**
```json
{
  "data": {
    "leaderboard": [
      {
        "provider": "freepik",
        "wins": 150,
        "losses": 50,
        "total_votes": 200,
        "win_rate": 75.0
      }
    ],
    "timestamp": "2025-10-06T12:00:00Z"
  }
}
```

### Get Statistics
```bash
GET /api/v1/statistics
```

**Response:**
```json
{
  "data": {
    "providers": {
      "freepik": {
        "provider": "freepik",
        "wins": 150,
        "losses": 50,
        "total_votes": 200,
        "win_rate": 75.0
      }
    },
    "total_votes": 500,
    "timestamp": "2025-10-06T12:00:00Z"
  }
}
```

### Provider Status
```bash
GET /api/v1/status
```

**Response:**
```json
{
  "data": {
    "freepik": {
      "name": "freepik",
      "available": true,
      "last_success": "2025-10-06T12:00:00Z",
      "error_count": 0,
      "quota_hit": false,
      "rate_limited": false
    }
  }
}
```

### Health Check
```bash
GET /health
```

**Response:**
```json
{
  "status": "healthy",
  "available_providers": 3,
  "total_providers": 3,
  "timestamp": "2025-10-06T12:00:00Z"
}
```

## Configuration

Environment variables:

**Server:**
- `PORT`: Server port (default: 8080)
- `HOST`: Server host (default: 0.0.0.0)
- `GIN_MODE`: Gin mode (release, debug, test)

**API Keys:**
- `GOOGLE_API_KEY`: Google Imagen API key
- `LEONARDO_API_KEY`: Leonardo AI API key
- `FREEPIK_API_KEY`: Freepik API key

**DigitalOcean Spaces (required):**
- `DO_SPACES_BUCKET`: Spaces bucket name (e.g., cgc-lb-and-cdn-content)
- `DO_SPACES_ENDPOINT`: Spaces endpoint (e.g., nyc3.digitaloceanspaces.com)
- `DO_SPACES_ACCESS_KEY`: Spaces access key
- `DO_SPACES_SECRET_KEY`: Spaces secret key

**Valkey Database (required for leaderboard):**
- `DO_VALKEY_HOST`: Valkey cluster host
- `DO_VALKEY_PORT`: Valkey port (default: 25061)
- `DO_VALKEY_PASSWORD`: Valkey password

## Error Handling

The system automatically detects and handles:

- **Quota Limits**: When providers hit daily/monthly limits
- **Rate Limits**: When providers limit requests per second/minute
- **API Errors**: Authentication, network, or service errors
- **Automatic Fallback**: Seamlessly switches to available providers

## Provider Implementation

Each provider implements the `ImageProvider` interface:

```go
type ImageProvider interface {
    Generate(ctx context.Context, req *ImageRequest) (*ImageResponse, error)
    GetStatus() *ProviderStatus
    GetName() string
    IsAvailable() bool
    HandleError(err error) *ProviderError
}
```

## Recent Updates

- ✅ DigitalOcean Spaces CDN integration (all images served from CDN)
- ✅ Valkey vote persistence and caching
- ✅ Real-time leaderboard with provider statistics
- ✅ Vote tracking and analytics
- ✅ Removed local storage (production-only deployment)

## Future Enhancements

- [ ] Advanced ADK features (conversation memory, tool integration)
- [ ] Webhook notifications for generation completion
- [ ] Image metadata extraction and tagging
- [ ] Performance monitoring and analytics
- [ ] WebSocket support for real-time leaderboard updates

## Development

Project structure:
```
backend/
├── cmd/server/          # Application entry point
├── internal/
│   ├── agents/          # ADK framework and orchestrator
│   ├── providers/       # Image generation providers
│   ├── models/          # Data models and types
│   ├── handlers/        # HTTP handlers
│   └── config/          # Configuration management
├── pkg/
│   └── utils/           # Utility functions
└── images/              # Generated images storage
```

Run tests:
```bash
go test ./...
```

Build for production:
```bash
go build -o bin/server cmd/server/main.go
```