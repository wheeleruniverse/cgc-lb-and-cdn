# Craiyon

This directory contains a Go example for generating images using Craiyon's unofficial API.

## Prerequisites

**No API key required!** Craiyon is completely free and open to use.

1. Install dependencies:
   ```bash
   go mod tidy
   ```

## Script

### Generate Images with Craiyon
Generates 4 images using Craiyon's free API and saves them as PNG files to the `images/` directory. Since Craiyon uses an unofficial API, generation may take 30-60 seconds.

**Usage:**
```bash
go run generate_craiyon.go
```

**Example Output:**
```
2025/09/16 14:13:48 Starting image generation with Craiyon...
2025/09/16 14:13:48 This may take 30-60 seconds...
2025/09/16 14:14:25 Generated 9 images successfully
2025/09/16 14:14:25 Saved image 1 to: images/craiyon-4d239706-cf2a-4045-8a50-637b7b99e965.png
2025/09/16 14:14:25 Saved image 2 to: images/craiyon-d75309bf-449c-457d-8522-c540eb9fd88f.png
2025/09/16 14:14:25 Saved image 3 to: images/craiyon-10595871-2eac-46c1-88ff-1c1563097cad.png
2025/09/16 14:14:25 Saved image 4 to: images/craiyon-14f2f045-a001-49a2-a649-b74b55e98fff.png
```

Generated images are saved to the `images/` directory with UUID-based filenames like `craiyon-[uuid].png`.

## Notes

- **Completely free** - No API key or account required
- **Unofficial API** - Uses Craiyon's backend endpoint (may change without notice)
- **Slower generation** - Takes 30-60 seconds due to free tier processing
- **Multiple images** - Craiyon typically returns 9 images, but we limit to 4 for consistency
- **No rate limits** - Perfect as an unlimited fallback option
- **Base64 encoded** - Images are returned as base64 strings and decoded locally
- **Same prompt** - Uses "Robot holding a red skateboard" to match other examples

## Pros & Cons

**Advantages:**
- ✅ Completely free and unlimited
- ✅ No authentication required
- ✅ Perfect fallback when quota limits are hit
- ✅ Returns multiple image variations

**Limitations:**
- ⚠️ **Currently blocked by Cloudflare** - API returns 403 errors due to bot protection
- ⚠️ Unofficial API - could change without notice
- ⚠️ Slower generation time (30-60 seconds)
- ⚠️ Lower quality compared to premium services
- ⚠️ No guaranteed uptime or support

## Current Status

**⚠️ API Currently Unavailable**: As of September 2025, Craiyon has implemented Cloudflare protection that blocks automated requests. The API returns a 403 error with a "Just a moment..." challenge page.

**Potential Solutions:**
- Use browser automation tools like Selenium or Playwright
- Wait for Craiyon to provide an official API
- Consider alternative free image generation services