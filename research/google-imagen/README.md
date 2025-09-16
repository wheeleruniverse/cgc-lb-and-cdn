# Google Imagen

This repository contains two Go examples for working with Google's Imagen AI model.

## Prerequisites

You need a Google API key to use these examples:

1. Get your API key from [Google AI Studio](https://aistudio.google.com/app/apikey)
2. Set the environment variable:
   ```bash
   export GOOGLE_API_KEY=xxx
   ```

## Scripts

### Discover Google Imagen Models
Lists all available Imagen models from the Google AI API with detailed JSON output. This script filters specifically for models containing "imagen" in their name.

**Usage:**
```bash
go run discover_imagen_models.go
```

**Example Output:**
```
2025/09/16 14:09:14 Model: {
  "name": "models/imagen-3.0-generate-002",
  "displayName": "Imagen 3.0",
  "description": "Vertex served Imagen 3.0 002 model",
  "version": "002",
  "tunedModelInfo": {},
  "inputTokenLimit": 480,
  "outputTokenLimit": 8192,
  "supportedActions": [
    "predict"
  ]
}
2025/09/16 14:09:14 ---
2025/09/16 14:09:14 Found 1 Imagen models that support image generation
```

### Create Images with Imagen
Generates 4 images using the Imagen model and saves them as PNG files to the `images/` directory.

**Usage:**
```bash
go run generate_imagen.go
```

**Example Output:**
```
2025/09/16 14:13:48 Generated 4 images successfully
2025/09/16 14:13:48 Saved image 1 to: images/google-imagen-4d239706-cf2a-4045-8a50-637b7b99e965.png
2025/09/16 14:13:48 Saved image 2 to: images/google-imagen-d75309bf-449c-457d-8522-c540eb9fd88f.png
2025/09/16 14:13:48 Saved image 3 to: images/google-imagen-10595871-2eac-46c1-88ff-1c1563097cad.png
2025/09/16 14:13:48 Saved image 4 to: images/google-imagen-14f2f045-a001-49a2-a649-b74b55e98fff.png
```

Generated images are saved to the `images/` directory with UUID-based filenames like `google-imagen-[uuid].png`.