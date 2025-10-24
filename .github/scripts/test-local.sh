#!/bin/bash

# Test script for local development
# Run from repository root: ./.github/scripts/test-local.sh

set -e

echo "🧪 Testing static data generation locally..."
echo ""

# Install dependencies
echo "📦 Installing script dependencies..."
cd .github/scripts
npm install
cd ../..

# Run generation script
echo ""
echo "📡 Generating static data from DO Spaces..."
node .github/scripts/generate-static-data.js

# Check output
if [ -f "frontend/public/static-data/image-pairs.json" ]; then
    echo ""
    echo "✅ Generation successful!"
    echo ""
    echo "📊 Summary:"

    # Count pairs
    PAIR_COUNT=$(node -e "const data = require('./frontend/public/static-data/image-pairs.json'); console.log(data.pairs.length);")
    echo "  - Total pairs: $PAIR_COUNT"

    # Show generated timestamp
    GEN_TIME=$(node -e "const data = require('./frontend/public/static-data/image-pairs.json'); console.log(data.generatedAt);")
    echo "  - Generated at: $GEN_TIME"

    echo ""
    echo "📁 Output file: frontend/public/static-data/image-pairs.json"
    echo ""
    echo "🚀 Ready to test lite build:"
    echo "  cd frontend"
    echo "  npm run build:lite"
else
    echo ""
    echo "❌ Generation failed - output file not created"
    exit 1
fi
