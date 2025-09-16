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
Generates 4 images using FreePik's Classic Fast API and saves them as PNG files to the `images/` directory. FreePik returns images synchronously as base64 data.

**Usage:**
```bash
go run generate_freepik.go
```

**Example Output:**
```
2025/09/16 15:18:02 Starting image generation with FreePik...
2025/09/16 15:18:05 Generated 4 images successfully
2025/09/16 15:18:05 Saved image 1 to: images/freepik-4d239706-cf2a-4045-8a50-637b7b99e965.png
2025/09/16 15:18:05 Saved image 2 to: images/freepik-d75309bf-449c-457d-8522-c540eb9fd88f.png
2025/09/16 15:18:05 Saved image 3 to: images/freepik-10595871-2eac-46c1-88ff-1c1563097cad.png
2025/09/16 15:18:05 Saved image 4 to: images/freepik-14f2f045-a001-49a2-a649-b74b55e98fff.png
```

Generated images are saved to the `images/` directory with UUID-based filenames like `freepik-[uuid].png`.

## Notes

- **Official API** - Uses FreePik's documented and supported API endpoints
- **Free tier available** - FreePik provides $5 USD credit to start
- **Synchronous processing** - Returns images immediately as base64 data
- **Classic Fast model** - Uses the most cost-effective model available
- **High quality** - Professional-grade image generation
- **Square aspect ratio** - Images generated at 1:1 ratio by default
- **Base64 decoding** - Images returned as base64 strings and decoded locally
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
- Uses Classic Fast model (`/v1/ai/text-to-image`) for cost-effective generation
- Receives images as base64 data (synchronous response)
- No polling required - immediate results
- Decodes base64 strings to PNG files automatically