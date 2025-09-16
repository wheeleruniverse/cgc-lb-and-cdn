# Leonardo AI

This directory contains a Go example for generating images using Leonardo AI's API.

## Prerequisites

You need a Leonardo AI API key to use this example:

1. Get your API key from [Leonardo AI API Access](https://app.leonardo.ai/api-access)
2. Set the environment variable:
   ```bash
   export LEONARDO_API_KEY=xxx
   ```
3. Install dependencies:
   ```bash
   go mod tidy
   ```

## Script

### Generate Images with Leonardo AI
Generates 4 images using Leonardo AI and saves them as PNG files to the `images/` directory. The script uses the Leonardo Creative model and polls for completion since image generation is asynchronous.

**Usage:**
```bash
go run generate_leonardo.go
```

**Example Output:**
```
2025/09/16 14:13:48 Started generation with ID: abc123-def456-ghi789
2025/09/16 14:13:48 Waiting for generation to complete...
2025/09/16 14:13:53 Generation status: PENDING
2025/09/16 14:13:58 Generation status: COMPLETE
2025/09/16 14:13:58 Generated 4 images successfully
2025/09/16 14:13:58 Saved image 1 to: images/leonardo-ai-4d239706-cf2a-4045-8a50-637b7b99e965.png
2025/09/16 14:13:58 Saved image 2 to: images/leonardo-ai-d75309bf-449c-457d-8522-c540eb9fd88f.png
2025/09/16 14:13:58 Saved image 3 to: images/leonardo-ai-10595871-2eac-46c1-88ff-1c1563097cad.png
2025/09/16 14:13:58 Saved image 4 to: images/leonardo-ai-14f2f045-a001-49a2-a649-b74b55e98fff.png
```

Generated images are saved to the `images/` directory with UUID-based filenames like `leonardo-ai-[uuid].png`.

## Notes

- Leonardo AI generates images asynchronously, so the script polls every 5 seconds until completion
- Uses the Leonardo Creative model (`6bef9f1b-29cb-40c7-b9df-32b51c1f67d3`) by default
- Images are 1024x1024 pixels with a guidance scale of 7 (recommended)
- The same prompt "Robot holding a red skateboard" is used to match the Google Imagen example