# FreePik

This directory contains a Go example for generating images using FreePik's official API.

## Prerequisites

You need a FreePik API key to use this example:

1. Get your API key from [FreePik Developers Dashboard](https://www.freepik.com/developers/dashboard/api-key)
2. Set the environment variable:
   ```bash
   export FREEPIK_API_KEY=xxx
   ```
3. Install dependencies:
   ```bash
   go mod tidy
   ```

## Script

### Generate Images with FreePik
Generates 4 images using FreePik's text-to-image API and saves them as PNG files to the `images/` directory. FreePik uses asynchronous processing with task-based polling.

**Usage:**
```bash
go run generate_images.go
```

**Example Output:**
```
2025/09/16 15:13:48 Starting image generation with FreePik...
2025/09/16 15:13:48 Started generation with task ID: abc123-def456-ghi789
2025/09/16 15:13:48 Waiting for generation to complete...
2025/09/16 15:13:53 Generation status: IN_PROGRESS
2025/09/16 15:13:58 Generation status: COMPLETED
2025/09/16 15:13:58 Generated 4 images successfully
2025/09/16 15:13:58 Saved image 1 to: images/freepik-4d239706-cf2a-4045-8a50-637b7b99e965.png
2025/09/16 15:13:58 Saved image 2 to: images/freepik-d75309bf-449c-457d-8522-c540eb9fd88f.png
2025/09/16 15:13:58 Saved image 3 to: images/freepik-10595871-2eac-46c1-88ff-1c1563097cad.png
2025/09/16 15:13:58 Saved image 4 to: images/freepik-14f2f045-a001-49a2-a649-b74b55e98fff.png
```

Generated images are saved to the `images/` directory with UUID-based filenames like `freepik-[uuid].png`.

## Notes

- **Official API** - Uses FreePik's documented and supported API endpoints
- **Free tier available** - FreePik provides $5 USD credit to start
- **Asynchronous processing** - Uses task-based polling similar to Leonardo AI
- **Multiple models** - Supports Classic Fast, Imagen3, Flux, and other AI models
- **High quality** - Professional-grade image generation
- **Square aspect ratio** - Images generated at 1:1 ratio by default
- **Same prompt** - Uses "Robot holding a red skateboard" to match other examples

## API Features

**Available Models:**
- Classic Fast (`/v1/ai/text-to-image`)
- Google Imagen3 (`/v1/ai/text-to-image/imagen3`)
- Flux Dev (`/v1/ai/text-to-image/flux-dev`)
- Mystic (`/v1/ai/mystic`)
- Gemini 2.5 Flash (`/v1/ai/gemini-2-5-flash-image-preview`)

**Advantages:**
- ✅ Official API with documentation and support
- ✅ Free tier with $5 credit
- ✅ Multiple cutting-edge AI models
- ✅ Professional quality output
- ✅ Reliable service with proper authentication
- ✅ Webhook support for real-time notifications

**Current Implementation:**
- Uses Classic Fast model for reliable generation
- Downloads images via URL (no base64 decoding needed)
- Polls every 5 seconds until completion
- Handles both SUCCESS and COMPLETED status responses