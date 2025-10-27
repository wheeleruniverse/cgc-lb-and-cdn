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

// Digital Ocean Spaces configuration
const DO_BUCKET_NAME = 'cgc-lb-and-cdn-content';
const DO_REGION = 'nyc3';
const DO_SPACES_ENDPOINT = `${DO_BUCKET_NAME}.${DO_REGION}.digitaloceanspaces.com`;
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
 * Handles pagination for buckets with >1000 objects (S3 API limit)
 */
async function listSpacesObjects(prefix = IMAGES_PREFIX) {
  let allContents = [];
  let continuationToken = null;
  let pageCount = 0;

  console.log(`📡 Fetching object list from DO Spaces...`);

  do {
    pageCount++;
    const params = new URLSearchParams({
      'list-type': '2',
      'prefix': prefix,
    });

    if (continuationToken) {
      params.append('continuation-token', continuationToken);
    }

    const url = `https://${DO_SPACES_ENDPOINT}/?${params.toString()}`;

    console.log(`  Page ${pageCount}: Fetching...`);

    const response = await httpsRequest(url);

    if (response.statusCode !== 200) {
      throw new Error(`Failed to list objects: ${response.statusCode}`);
    }

    // Parse XML response
    const result = await parseStringPromise(response.body);
    const contents = result.ListBucketResult?.Contents || [];
    const isTruncated = result.ListBucketResult?.IsTruncated?.[0] === 'true';

    allContents = allContents.concat(contents);
    console.log(`  Page ${pageCount}: Found ${contents.length} objects (total so far: ${allContents.length})`);

    // Get continuation token for next page
    continuationToken = isTruncated
      ? result.ListBucketResult?.NextContinuationToken?.[0]
      : null;

  } while (continuationToken);

  console.log(`✅ Completed fetching ${allContents.length} objects from bucket (${pageCount} pages)`);

  return allContents.map(item => ({
    key: item.Key[0],
    size: parseInt(item.Size[0]),
    lastModified: item.LastModified[0],
  }));
}

/**
 * Fetch metadata for an image using HEAD request
 */
async function fetchImageMetadata(imageKey) {
  const url = `https://${DO_SPACES_ENDPOINT}/${imageKey}`;

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
    pair[side] = `https://${DO_SPACES_ENDPOINT}/${obj.key}`;
  }

  // Filter out incomplete pairs (must have both left and right)
  return Array.from(pairs.values()).filter(pair => pair.left && pair.right);
}

/**
 * Fetch metadata for all image pairs (parallel with batching for speed)
 */
async function enrichPairsWithMetadata(pairs) {
  console.log(`\n🔍 Fetching metadata for ${pairs.length} pairs...`);

  const BATCH_SIZE = 50; // Process 50 pairs concurrently
  const enrichedPairs = [];
  let processedCount = 0;

  // Process in batches for better performance
  for (let i = 0; i < pairs.length; i += BATCH_SIZE) {
    const batch = pairs.slice(i, i + BATCH_SIZE);

    // Process batch concurrently
    const batchResults = await Promise.all(
      batch.map(async (pair) => {
        // Extract key from URL
        const leftKey = pair.left.replace(`https://${DO_SPACES_ENDPOINT}/`, '');

        // Fetch metadata from left image (contains pair info)
        const metadata = await fetchImageMetadata(leftKey);

        if (!metadata) {
          console.warn(`⚠️ Skipping pair ${pair.pairId} - no metadata found`);
          return null;
        }

        const provider = metadata.provider || pair.provider;

        // Store only essential data - URLs are built via template in frontend
        return {
          id: metadata['pair-id'] || pair.pairId,
          prompt: metadata.prompt || 'unknown prompt',
          provider,
        };
      })
    );

    // Add successful results
    enrichedPairs.push(...batchResults.filter(result => result !== null));

    processedCount += batch.length;
    console.log(`  Progress: ${processedCount}/${pairs.length} pairs processed (batch ${Math.floor(i / BATCH_SIZE) + 1}/${Math.ceil(pairs.length / BATCH_SIZE)})`);
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

    // Output data is just the array of pairs
    // Frontend builds URLs via template: {cdn}/images/{provider}/{id}/left.png
    const outputData = shuffledPairs;

    // Write to file
    const outputPath = path.join(process.cwd(), 'frontend/public/static-data/image-pairs.json');
    const outputDir = path.dirname(outputPath);

    // Create directory if it doesn't exist
    if (!fs.existsSync(outputDir)) {
      fs.mkdirSync(outputDir, { recursive: true });
    }

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
