# CGC Image Service Backend

A Go-based image generation service using Google Agent Development Kit (ADK) for intelligent provider selection and automatic fallback between Freepik, Google Imagen, and Leonardo AI.

## Features

- **Google Agent Development Kit (ADK)**: Intelligent orchestrator agent for provider selection
- **Automatic Fallback**: Switches between providers when quotas or rate limits are hit
- **Random Load Balancing**: Treats all providers equally until errors occur
- **Local Image Storage**: Saves generated images locally (ready for DigitalOcean Spaces integration)
- **RESTful API**: Clean `/generate` endpoint with JSON responses
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
POST /generate
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
        "path": "images/freepik-uuid.png",
        "size": 1234567
      }
    ],
    "provider": "freepik",
    "success": true,
    "request_id": "uuid",
    "duration": "2.5s",
    "metadata": {
      "model": "classic-fast",
      "aspect_ratio": "1:1"
    }
  }
}
```

### Provider Status
```bash
GET /status
```

**Response:**
```json
{
  "data": {
    "freepik": {
      "name": "freepik",
      "available": true,
      "last_success": "2025-09-16T15:30:00Z",
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
  "timestamp": "2025-09-16T15:30:00Z"
}
```

## Configuration

Environment variables:

- `PORT`: Server port (default: 8080)
- `HOST`: Server host (default: 0.0.0.0)
- `IMAGES_DIR`: Directory for saved images (default: images)
- `GIN_MODE`: Gin mode (release, debug, test)

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

## Future Enhancements

- [ ] DigitalOcean Spaces integration for cloud storage
- [ ] Advanced ADK features (conversation memory, tool integration)
- [ ] Webhook notifications for generation completion
- [ ] Image metadata extraction and tagging
- [ ] Performance monitoring and analytics

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