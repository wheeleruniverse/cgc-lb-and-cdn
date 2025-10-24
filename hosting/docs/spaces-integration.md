# DO Spaces Integration - Real Data for Lite Deployment

## ✅ Updated! Using Real Images from DO Spaces

The lite deployment now uses **real images from Digital Ocean Spaces** instead of mock data!

## 🎯 How It Works

### Build-Time Generation

When `cheap-deploy.yml` runs, it automatically:

1. **Lists Objects** in DO Spaces bucket via HTTPS GET request
2. **Parses XML** response to find all image files
3. **Groups Pairs** by matching pattern: `images/{provider}/{pair-id}/{left|right}.png`
4. **Fetches Metadata** from each image's response headers
5. **Generates JSON** file with all discovered pairs

### Metadata Structure

Each image in DO Spaces has metadata stored in response headers:

```http
X-Amz-Meta-Pair-Id: 00360dbb-50d6-4804-82f2-4a3c1f0de815
X-Amz-Meta-Provider: freepik
X-Amz-Meta-Prompt: A tidal pool reflecting an entire miniature ocean ecosystem.
X-Amz-Meta-Side: left
```

This metadata is:
- Set by the backend when uploading images
- Returned in HTTP response headers
- Publicly accessible (no authentication needed)
- Used to build the image pairs JSON

## 📁 Files Added/Modified

### New Files:

1. **`.github/scripts/generate-static-data.js`**
   - Node.js script that fetches images from DO Spaces
   - Lists objects using S3 XML API
   - Fetches metadata from response headers
   - Generates `frontend/public/static-data/image-pairs.json`

2. **`.github/scripts/package.json`**
   - Dependencies for generation script
   - `xml2js` for parsing S3 XML responses

3. **`.github/scripts/test-local.sh`**
   - Test script for local development
   - Run: `./.github/scripts/test-local.sh`

4. **`frontend/app/services/spacesDataService.ts`**
   - Optional: Runtime fetching service (requires CORS)
   - Not used in current build-time approach

### Modified Files:

1. **`.github/workflows/cheap-deploy.yml`**
   - Added steps to generate static data before build
   - Shows pair count in deployment summary
   - Updated notes to reflect real images

2. **`frontend/.gitignore`**
   - Ignores generated `image-pairs.json` (regenerated each build)

3. **`.gitignore`** (root)
   - Ignores `.github/scripts/node_modules/`

4. **Documentation**:
   - `DUAL-DEPLOYMENT-SETUP.md` - Updated static data section
   - `frontend/public/static-data/README.md` - Detailed how-it-works

## 🚀 Workflow Integration

### cheap-deploy.yml Steps:

```yaml
1. Checkout code
2. Setup Node.js 18
3. Install script dependencies (.github/scripts)
4. Generate static data from DO Spaces  ← NEW!
5. Install frontend dependencies
6. Build frontend in lite mode
7. Upload to GitHub Pages
8. Deploy
```

### Build Output:

```
📡 Fetching real image pairs from DO Spaces...
✅ Found 42 objects in bucket
📊 Grouping images by pair...
✅ Found 21 complete image pairs
🔍 Fetching metadata for 21 pairs...
  Progress: 21/21 pairs processed
✅ Successfully generated static data!
📊 Total pairs: 21
```

## 🔑 Key Benefits

1. **Always Fresh**: Lite deployment stays in sync with DO Spaces
2. **No Manual Updates**: Automatically discovers new images
3. **Real Metadata**: Uses actual prompts and provider info
4. **Build-Time**: No runtime API calls, no CORS needed
5. **Fast Loading**: Static JSON is small and cacheable

## 📊 Data Flow

```
┌─────────────────────────────────────────────────────┐
│ Full Deployment (Digital Ocean)                     │
│                                                      │
│  Backend → Generates Images → Uploads to DO Spaces  │
│                                    ↓                 │
│                          Sets metadata headers      │
└─────────────────────────────────────────────────────┘
                           ↓
┌─────────────────────────────────────────────────────┐
│ GitHub Actions (cheap-deploy.yml)                   │
│                                                      │
│  generate-static-data.js → Fetches from DO Spaces   │
│                                    ↓                 │
│                    Reads metadata headers           │
│                                    ↓                 │
│              Generates image-pairs.json             │
└─────────────────────────────────────────────────────┘
                           ↓
┌─────────────────────────────────────────────────────┐
│ Lite Deployment (GitHub Pages)                      │
│                                                      │
│  Static Site → Loads from image-pairs.json          │
│                                    ↓                 │
│              Displays real images                   │
└─────────────────────────────────────────────────────┘
```

## 🧪 Testing Locally

Test the generation script before deploying:

```bash
# Run test script
./.github/scripts/test-local.sh

# Or manually:
cd .github/scripts
npm install
node generate-static-data.js
```

Output:
```
🚀 Generating static data from Digital Ocean Spaces

📡 Fetching object list from: https://cgc-lb-and-cdn-content.nyc3...
✅ Found 42 objects in bucket

📊 Grouping images by pair...
✅ Found 21 complete image pairs

🔍 Fetching metadata for 21 pairs...
  Progress: 21/21 pairs processed

✅ Successfully generated static data!
📁 Output: frontend/public/static-data/image-pairs.json
📊 Total pairs: 21

Sample pair:
{
  "pair_id": "00360dbb-50d6-4804-82f2-4a3c1f0de815",
  "prompt": "A tidal pool reflecting an entire miniature ocean ecosystem.",
  "provider": "freepik",
  "left_url": "https://cgc-lb-and-cdn-content.nyc3.digitaloceanspaces.com/images/freepik/00360dbb-50d6-4804-82f2-4a3c1f0de815/left.png",
  "right_url": "https://cgc-lb-and-cdn-content.nyc3.digitaloceanspaces.com/images/freepik/00360dbb-50d6-4804-82f2-4a3c1f0de815/right.png"
}
```

## 📋 Requirements

### DO Spaces Structure:

Images must be organized as:
```
cgc-lb-and-cdn-content/
└── images/
    ├── freepik/
    │   └── {pair-id}/
    │       ├── left.png
    │       └── right.png
    ├── google/
    │   └── {pair-id}/
    │       ├── left.png
    │       └── right.png
    └── leonardo/
        └── {pair-id}/
            ├── left.png
            └── right.png
```

### Metadata Headers:

Backend must set these headers when uploading:
```javascript
{
  'x-amz-meta-pair-id': pairId,
  'x-amz-meta-provider': 'freepik',
  'x-amz-meta-prompt': 'Your image prompt here',
  'x-amz-meta-side': 'left' // or 'right'
}
```

### Public Access:

- Bucket must have public read access (already configured)
- CDN must be enabled (already configured)
- No authentication required

## 🎉 Result

### Before:
- ❌ Mock/placeholder data
- ❌ Manual JSON updates required
- ❌ Out of sync with DO Spaces

### After:
- ✅ Real images from DO Spaces
- ✅ Automatic updates on each build
- ✅ Always in sync with full deployment
- ✅ Real prompts and metadata
- ✅ No manual intervention needed

## 🔄 Adding New Images

To add new image pairs:

1. **Generate Images**: Use full DO deployment to create new images
2. **Automatic Upload**: Backend uploads to DO Spaces with metadata
3. **Rebuild Lite**: Run `cheap-deploy.yml` workflow
4. **New Images Available**: Script automatically discovers and includes them!

No manual JSON editing required! 🎉

## 🐛 Troubleshooting

### Script Fails to Fetch

**Error**: `Failed to list objects: 403`

**Solution**: Verify DO Spaces bucket has public read access

---

**Error**: `ENOTFOUND cgc-lb-and-cdn-content.nyc3.digitaloceanspaces.com`

**Solution**: Check network connection and DNS resolution

---

**Error**: `No metadata found`

**Solution**: Verify backend is setting `X-Amz-Meta-*` headers when uploading

### No Pairs Found

**Error**: `Found 0 complete image pairs`

**Solution**:
1. Check images are in correct structure: `images/{provider}/{pair-id}/{left|right}.png`
2. Verify both left.png and right.png exist for each pair
3. Check file extensions are lowercase (.png, not .PNG)

### Build Fails

**Error**: `Cannot find module 'xml2js'`

**Solution**: Make sure `npm install` runs in `.github/scripts` before the generation script

## 📚 Related Documentation

- **[DUAL-DEPLOYMENT-SETUP.md](DUAL-DEPLOYMENT-SETUP.md)** - Overall setup guide
- **[.github/workflows/WORKFLOWS.md](.github/workflows/WORKFLOWS.md)** - Workflow documentation
- **[frontend/public/static-data/README.md](frontend/public/static-data/README.md)** - Static data details

---

**Integration Complete! 🎉 Lite deployment now uses real images from DO Spaces!**
