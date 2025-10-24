#!/bin/bash

# Test script for local development
# Can be run from any directory

set -e

# Get the absolute path to the repository root (2 levels up from this script)
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"

echo "🧪 Testing static data generation locally..."
echo "📁 Repository root: ${REPO_ROOT}"
echo ""

# Install dependencies
echo "📦 Installing script dependencies..."
cd "${SCRIPT_DIR}"
npm install

# Run generation script
echo ""
echo "📡 Generating static data from DO Spaces..."
cd "${REPO_ROOT}"
node "${SCRIPT_DIR}/generate-static-data.js"

# Check output
OUTPUT_FILE="${REPO_ROOT}/frontend/public/static-data/image-pairs.json"
if [ -f "${OUTPUT_FILE}" ]; then
    echo ""
    echo "✅ Generation successful!"
    echo ""
    echo "📊 Summary:"

    # Count pairs
    PAIR_COUNT=$(node -e "const data = require('${OUTPUT_FILE}'); console.log(data.pairs.length);")
    echo "  - Total pairs: $PAIR_COUNT"

    # Show generated timestamp
    GEN_TIME=$(node -e "const data = require('${OUTPUT_FILE}'); console.log(data.generatedAt);")
    echo "  - Generated at: $GEN_TIME"

    echo ""
    echo "📁 Output file: ${OUTPUT_FILE}"
    echo ""
    echo "🚀 Ready to test lite build:"
    echo "  cd ${REPO_ROOT}/frontend"
    echo "  npm run build:lite"
else
    echo ""
    echo "❌ Generation failed - output file not created"
    exit 1
fi
