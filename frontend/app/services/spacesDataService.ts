/**
 * Real-time data service for fetching images from DO Spaces
 *
 * This service fetches image pairs directly from Digital Ocean Spaces CDN,
 * reading metadata from response headers in real-time.
 *
 * Requirements:
 * - DO Spaces must have CORS enabled
 * - CORS must expose X-Amz-Meta-* headers
 */

import { type ImagePair } from './dataService'

const SPACES_ENDPOINT = 'https://cgc-lb-and-cdn-content.nyc3.digitaloceanspaces.com'
const IMAGES_PREFIX = 'images/'

interface SpacesObject {
  key: string
  lastModified: string
  size: number
}

/**
 * Parse S3 XML listing response
 */
async function parseS3ListingXML(xmlText: string): Promise<SpacesObject[]> {
  const parser = new DOMParser()
  const xmlDoc = parser.parseFromString(xmlText, 'text/xml')

  const contents = xmlDoc.getElementsByTagName('Contents')
  const objects: SpacesObject[] = []

  for (let i = 0; i < contents.length; i++) {
    const item = contents[i]
    const key = item.getElementsByTagName('Key')[0]?.textContent || ''
    const lastModified = item.getElementsByTagName('LastModified')[0]?.textContent || ''
    const size = parseInt(item.getElementsByTagName('Size')[0]?.textContent || '0')

    objects.push({ key, lastModified, size })
  }

  return objects
}

/**
 * List all objects in DO Spaces bucket
 */
export async function listSpacesObjects(): Promise<SpacesObject[]> {
  const url = `${SPACES_ENDPOINT}/?list-type=2&prefix=${encodeURIComponent(IMAGES_PREFIX)}`

  try {
    const response = await fetch(url)

    if (!response.ok) {
      throw new Error(`Failed to list objects: ${response.status}`)
    }

    const xmlText = await response.text()
    return parseS3ListingXML(xmlText)

  } catch (error) {
    console.error('[SpacesDataService] Failed to list objects:', error)

    // If CORS is not configured, this will fail
    if (error instanceof TypeError && error.message.includes('Failed to fetch')) {
      console.error('[SpacesDataService] CORS error detected. Make sure DO Spaces has CORS configured.')
      console.error('[SpacesDataService] Required CORS settings:')
      console.error('  - AllowedOrigins: * (or your GitHub Pages domain)')
      console.error('  - AllowedMethods: GET, HEAD')
      console.error('  - AllowedHeaders: *')
      console.error('  - ExposedHeaders: X-Amz-Meta-*')
    }

    throw error
  }
}

/**
 * Fetch metadata for an image
 */
export async function fetchImageMetadata(imageUrl: string): Promise<Record<string, string>> {
  try {
    const response = await fetch(imageUrl, { method: 'HEAD' })

    if (!response.ok) {
      throw new Error(`Failed to fetch metadata: ${response.status}`)
    }

    // Extract X-Amz-Meta-* headers
    const metadata: Record<string, string> = {}

    response.headers.forEach((value, key) => {
      if (key.startsWith('x-amz-meta-')) {
        const metaKey = key.replace('x-amz-meta-', '')
        metadata[metaKey] = value
      }
    })

    return metadata

  } catch (error) {
    console.error('[SpacesDataService] Failed to fetch metadata:', error)
    throw error
  }
}

/**
 * Group images by pair ID
 */
function groupImagesByPair(objects: SpacesObject[]): Array<{
  pairId: string
  provider: string
  left: string
  right: string
}> {
  const pairs = new Map()

  for (const obj of objects) {
    // Match pattern: images/{provider}/{pair-id}/{side}.png
    const match = obj.key.match(/^images\/([^\/]+)\/([^\/]+)\/(left|right)\.(png|jpg|jpeg)$/i)

    if (!match) continue

    const [, provider, pairId, side] = match
    const imageUrl = `${SPACES_ENDPOINT}/${obj.key}`

    if (!pairs.has(pairId)) {
      pairs.set(pairId, { pairId, provider, left: null, right: null })
    }

    const pair = pairs.get(pairId)
    pair[side] = imageUrl
  }

  // Filter out incomplete pairs
  return Array.from(pairs.values()).filter(pair => pair.left && pair.right)
}

/**
 * Load all image pairs from DO Spaces with metadata
 */
export async function loadImagePairsFromSpaces(): Promise<ImagePair[]> {
  console.log('[SpacesDataService] Loading image pairs from DO Spaces...')

  // List all objects
  const objects = await listSpacesObjects()
  console.log(`[SpacesDataService] Found ${objects.length} objects`)

  // Group by pair ID
  const pairs = groupImagesByPair(objects)
  console.log(`[SpacesDataService] Found ${pairs.length} complete pairs`)

  if (pairs.length === 0) {
    console.warn('[SpacesDataService] No image pairs found!')
    return []
  }

  // Fetch metadata for each pair (from left image)
  const imagePairs: ImagePair[] = []

  for (const pair of pairs) {
    try {
      const metadata = await fetchImageMetadata(pair.left)

      imagePairs.push({
        pair_id: metadata['pair-id'] || pair.pairId,
        prompt: metadata.prompt || 'AI-generated image',
        provider: metadata.provider || pair.provider,
        left_url: pair.left,
        right_url: pair.right,
      })

      // Rate limiting - wait 50ms between requests
      await new Promise(resolve => setTimeout(resolve, 50))

    } catch (error) {
      console.warn(`[SpacesDataService] Failed to load pair ${pair.pairId}:`, error)
      // Continue with other pairs
    }
  }

  console.log(`[SpacesDataService] Successfully loaded ${imagePairs.length} pairs`)

  // Shuffle for variety
  return imagePairs.sort(() => Math.random() - 0.5)
}

/**
 * Check if DO Spaces CORS is configured correctly
 */
export async function checkCorsConfiguration(): Promise<boolean> {
  try {
    const testUrl = `${SPACES_ENDPOINT}/?list-type=2&max-keys=1`
    const response = await fetch(testUrl, { method: 'HEAD' })

    // Check if CORS headers are present
    const hasCors = response.headers.has('access-control-allow-origin')

    if (!hasCors) {
      console.error('[SpacesDataService] CORS not configured on DO Spaces')
      return false
    }

    console.log('[SpacesDataService] CORS is configured correctly')
    return true

  } catch (error) {
    console.error('[SpacesDataService] CORS check failed:', error)
    return false
  }
}
