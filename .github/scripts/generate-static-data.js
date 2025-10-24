#!/usr/bin/env node

/**
 * Generate static image pairs data from Digital Ocean Spaces
 *
 * This script lists all images in DO Spaces and fetches their metadata
 * to build the image-pairs.json file for lite deployment.
 */

const https = require('https');
const fs = require('fs');
const path = require('path');
const { parseStringPromise } = require('xml2js');

const SPACES_ENDPOINT = 'cgc-lb-and-cdn-content.nyc3.digitaloceanspaces.com';
const BUCKET_NAME = 'cgc-lb-and-cdn-content';
const IMAGES_PREFIX = 'images/';

/**
 * Make HTTPS request and return response
 */
function httpsRequest(url, options = {}) {
  return new Promise((resolve, reject) => {
    const req = https.request(url, options, (res) => {
      let data = '';
      res.on('data', (chunk) => { data += chunk; });
      res.on('end', () => {
        resolve({ statusCode: res.statusCode, headers: res.headers, body: data });
      });
    });
    req.on('error', reject);
    req.end();
  });
}

/**
 * List all objects in DO Spaces bucket with prefix
 */
async function listSpacesObjects(prefix = IMAGES_PREFIX) {
  const url = `https://${SPACES_ENDPOINT}/?list-type=2&prefix=${encodeURIComponent(prefix)}`;

  console.log(`📡 Fetching object list from: ${url}`);

  const response = await httpsRequest(url);

  if (response.statusCode !== 200) {
    throw new Error(`Failed to list objects: ${response.statusCode}`);
  }

  // Parse XML response
  const result = await parseStringPromise(response.body);
  const contents = result.ListBucketResult?.Contents || [];

  console.log(`✅ Found ${contents.length} objects in bucket`);

  return contents.map(item => ({
    key: item.Key[0],
    size: parseInt(item.Size[0]),
    lastModified: item.LastModified[0],
  }));
}

/**
 * Fetch metadata for an image using HEAD request
 */
async function fetchImageMetadata(imageKey) {
  const url = `https://${SPACES_ENDPOINT}/${imageKey}`;

  const response = await httpsRequest(url, { method: 'HEAD' });

  if (response.statusCode !== 200) {
    console.warn(`⚠️ Failed to fetch metadata for ${imageKey}: ${response.statusCode}`);
    return null;
  }

  // Extract X-Amz-Meta-* headers
  const metadata = {};
  for (const [key, value] of Object.entries(response.headers)) {
    if (key.startsWith('x-amz-meta-')) {
      const metaKey = key.replace('x-amz-meta-', '');
      metadata[metaKey] = value;
    }
  }

  return metadata;
}

/**
 * Group images by pair ID
 */
function groupImagesByPair(objects) {
  const pairs = new Map();

  for (const obj of objects) {
    // Match pattern: images/{provider}/{pair-id}/{side}.png
    const match = obj.key.match(/^images\/([^\/]+)\/([^\/]+)\/(left|right)\.(png|jpg|jpeg)$/i);

    if (!match) {
      continue; // Skip non-image files
    }

    const [, provider, pairId, side] = match;

    if (!pairs.has(pairId)) {
      pairs.set(pairId, { pairId, provider, left: null, right: null });
    }

    const pair = pairs.get(pairId);
    pair[side] = `https://${SPACES_ENDPOINT}/${obj.key}`;
  }

  // Filter out incomplete pairs (must have both left and right)
  return Array.from(pairs.values()).filter(pair => pair.left && pair.right);
}

/**
 * Fetch metadata for all image pairs
 */
async function enrichPairsWithMetadata(pairs) {
  console.log(`\n🔍 Fetching metadata for ${pairs.length} pairs...`);

  const enrichedPairs = [];

  for (let i = 0; i < pairs.length; i++) {
    const pair = pairs[i];

    // Extract key from URL
    const leftKey = pair.left.replace(`https://${SPACES_ENDPOINT}/`, '');

    // Fetch metadata from left image (contains pair info)
    const metadata = await fetchImageMetadata(leftKey);

    if (!metadata) {
      console.warn(`⚠️ Skipping pair ${pair.pairId} - no metadata found`);
      continue;
    }

    enrichedPairs.push({
      pair_id: metadata['pair-id'] || pair.pairId,
      prompt: metadata.prompt || 'AI-generated image',
      provider: metadata.provider || pair.provider,
      left_url: pair.left,
      right_url: pair.right,
    });

    // Show progress
    if ((i + 1) % 10 === 0 || i === pairs.length - 1) {
      console.log(`  Progress: ${i + 1}/${pairs.length} pairs processed`);
    }

    // Rate limiting - wait 100ms between requests
    await new Promise(resolve => setTimeout(resolve, 100));
  }

  return enrichedPairs;
}

/**
 * Main function
 */
async function main() {
  console.log('🚀 Generating static data from Digital Ocean Spaces\n');

  try {
    // List all objects in the bucket
    const objects = await listSpacesObjects();

    // Group by pair ID
    console.log('\n📊 Grouping images by pair...');
    const pairs = groupImagesByPair(objects);
    console.log(`✅ Found ${pairs.length} complete image pairs`);

    if (pairs.length === 0) {
      console.warn('\n⚠️ No image pairs found in DO Spaces!');
      console.log('Make sure images are uploaded in the format: images/{provider}/{pair-id}/{left|right}.{png|jpg}');
      process.exit(0);
    }

    // Fetch metadata for each pair
    const enrichedPairs = await enrichPairsWithMetadata(pairs);

    // Shuffle pairs for variety
    const shuffledPairs = enrichedPairs.sort(() => Math.random() - 0.5);

    // Create output data structure
    const outputData = {
      pairs: shuffledPairs,
      description: 'Image pairs generated from Digital Ocean Spaces',
      generatedAt: new Date().toISOString(),
      totalPairs: shuffledPairs.length,
    };

    // Write to file
    const outputPath = path.join(process.cwd(), 'frontend/public/static-data/image-pairs.json');
    fs.writeFileSync(outputPath, JSON.stringify(outputData, null, 2));

    console.log(`\n✅ Successfully generated static data!`);
    console.log(`📁 Output: ${outputPath}`);
    console.log(`📊 Total pairs: ${shuffledPairs.length}`);
    console.log(`\nSample pair:`);
    console.log(JSON.stringify(shuffledPairs[0], null, 2));

  } catch (error) {
    console.error('\n❌ Error generating static data:', error.message);
    console.error(error.stack);
    process.exit(1);
  }
}

// Run if executed directly
if (require.main === module) {
  main();
}

module.exports = { main };
