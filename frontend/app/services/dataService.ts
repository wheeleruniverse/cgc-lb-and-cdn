/**
 * Data Service Layer
 *
 * Abstracts data fetching to support both full (API-based) and lite (static) deployments.
 * This demonstrates production-grade abstraction patterns with feature flags.
 */

import { config } from '../config'

// Type definitions
export interface ImagePair {
  pair_id: string
  prompt: string
  provider: string
  left_url: string
  right_url: string
}

// Optimized format from static data (just the essentials)
interface OptimizedImagePair {
  id: string
  prompt: string
  provider: string
}

export interface WinnerImage {
  image_url: string
  prompt: string
  provider: string
  pair_id: string
  timestamp: string
  vote_count: number
}

export interface Statistics {
  side_wins: {
    left: number
    right: number
  }
}

// LocalStorage keys
const STORAGE_KEYS = {
  VOTED_PAIRS: 'votedPairIds',
  SESSION_ID: 'sessionId',
  LOCAL_VOTES: 'localVotes',
  LOCAL_STATS: 'localStats',
} as const

// Static data cache
let staticPairsCache: ImagePair[] | null = null

/**
 * Generate a random session ID for anonymous user tracking
 */
function generateSessionId(): string {
  const timestamp = Date.now().toString(36)
  const randomPart = Math.random().toString(36).substring(2, 15)
  return `sess_${timestamp}_${randomPart}`
}

/**
 * Get or create session ID
 */
export function getSessionId(): string {
  if (typeof window === 'undefined') return generateSessionId()

  const stored = localStorage.getItem(STORAGE_KEYS.SESSION_ID)
  if (stored) return stored

  const newSessionId = generateSessionId()
  localStorage.setItem(STORAGE_KEYS.SESSION_ID, newSessionId)
  return newSessionId
}

/**
 * Get voted pair IDs from localStorage
 */
export function getVotedPairIds(): string[] {
  if (typeof window === 'undefined') return []
  const stored = localStorage.getItem(STORAGE_KEYS.VOTED_PAIRS)
  return stored ? JSON.parse(stored) : []
}

/**
 * Save voted pair IDs to localStorage
 */
export function saveVotedPairIds(pairIds: string[]): void {
  if (typeof window === 'undefined') return
  localStorage.setItem(STORAGE_KEYS.VOTED_PAIRS, JSON.stringify(pairIds))
}

/**
 * Convert optimized pair format to full format
 * Builds URLs via template: {cdn}/images/{provider}/{id}/left.png
 */
function decodePair(optimized: OptimizedImagePair): ImagePair {
  const cdn = config.cdn.spacesUrl
  const { id, prompt, provider } = optimized

  return {
    pair_id: id,
    prompt: prompt,
    provider: provider,
    left_url: `${cdn}/images/${provider}/${id}/left.png`,
    right_url: `${cdn}/images/${provider}/${id}/right.png`,
  }
}

/**
 * Load static image pairs from JSON file
 */
async function loadStaticPairs(): Promise<ImagePair[]> {
  if (staticPairsCache) return staticPairsCache

  try {
    const basePath = config.basePath || ''
    const response = await fetch(`${basePath}/static-data/image-pairs.json`)
    if (!response.ok) throw new Error('Failed to load static image pairs')

    const data = await response.json()

    // Data is just an array of {id, prompt, provider}
    const optimizedPairs = data as OptimizedImagePair[]
    const pairs = optimizedPairs.map(p => decodePair(p))
    staticPairsCache = pairs

    console.log(`[DataService] Loaded ${pairs.length} pairs (optimized format)`)
    return pairs
  } catch (err) {
    console.error('[DataService] Failed to load static pairs:', err)
    return []
  }
}

/**
 * Fetch image pair
 * - Full mode: Fetch from API
 * - Lite mode: Load from static JSON, filter out voted pairs
 */
export async function fetchImagePair(excludePairIds: string[] = []): Promise<ImagePair> {
  if (config.isLiteMode) {
    // Lite mode: Load from static data
    const allPairs = await loadStaticPairs()
    const availablePairs = allPairs.filter(pair => !excludePairIds.includes(pair.pair_id))

    if (availablePairs.length === 0) {
      throw new Error("You've voted on all available pairs! In lite mode, pairs are limited to pre-generated images.")
    }

    // Return a random available pair
    const randomIndex = Math.floor(Math.random() * availablePairs.length)
    return availablePairs[randomIndex]
  } else {
    // Full mode: Fetch from API
    const sessionId = getSessionId()
    let url = `${config.api.baseUrl}/images/pair`
    const params = new URLSearchParams()

    params.append('session_id', sessionId)
    if (excludePairIds.length > 0) {
      params.append('exclude', excludePairIds.join(','))
    }

    url += `?${params.toString()}`

    const response = await fetch(url)
    if (!response.ok) {
      const errorData = await response.json()
      throw new Error(errorData.error || 'Failed to fetch image pair')
    }

    const data = await response.json()
    return data.data
  }
}

/**
 * Submit vote for image pair
 * - Full mode: Submit to API
 * - Lite mode: Store in localStorage only
 */
export async function submitVote(pairId: string, winner: 'left' | 'right'): Promise<void> {
  if (config.isLiteMode) {
    // Lite mode: Store vote locally
    if (typeof window === 'undefined') return

    // Load current local votes
    const localVotesStr = localStorage.getItem(STORAGE_KEYS.LOCAL_VOTES)
    const localVotes = localVotesStr ? JSON.parse(localVotesStr) : []

    // Add new vote
    localVotes.push({
      pair_id: pairId,
      winner,
      timestamp: new Date().toISOString(),
    })

    localStorage.setItem(STORAGE_KEYS.LOCAL_VOTES, JSON.stringify(localVotes))

    // Update local stats
    const statsStr = localStorage.getItem(STORAGE_KEYS.LOCAL_STATS)
    const stats = statsStr ? JSON.parse(statsStr) : { left: 0, right: 0 }
    stats[winner] = (stats[winner] || 0) + 1
    localStorage.setItem(STORAGE_KEYS.LOCAL_STATS, JSON.stringify(stats))

    console.log('[DataService] Vote stored locally:', { pairId, winner, stats })
  } else {
    // Full mode: Submit to API
    const response = await fetch(`${config.api.baseUrl}/images/rate`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        pair_id: pairId,
        winner,
      }),
    })

    if (!response.ok) {
      throw new Error('Failed to submit vote')
    }
  }
}

/**
 * Fetch statistics (team scores)
 * - Full mode: Fetch from API
 * - Lite mode: Load from localStorage
 */
export async function fetchStatistics(): Promise<Statistics> {
  if (config.isLiteMode) {
    // Lite mode: Load from localStorage
    if (typeof window === 'undefined') {
      return { side_wins: { left: 0, right: 0 } }
    }

    const statsStr = localStorage.getItem(STORAGE_KEYS.LOCAL_STATS)
    const stats = statsStr ? JSON.parse(statsStr) : { left: 0, right: 0 }

    return {
      side_wins: {
        left: stats.left || 0,
        right: stats.right || 0,
      },
    }
  } else {
    // Full mode: Fetch from API
    const response = await fetch(`${config.api.baseUrl}/statistics`)
    if (!response.ok) {
      throw new Error('Failed to fetch statistics')
    }

    const data = await response.json()
    return data.data
  }
}

/**
 * Fetch winners for a side
 * - Full mode: Fetch from API
 * - Lite mode: Load from localStorage votes
 */
export async function fetchWinners(side: 'left' | 'right'): Promise<WinnerImage[]> {
  if (config.isLiteMode) {
    // Lite mode: Build winners from local votes
    if (typeof window === 'undefined') return []

    const localVotesStr = localStorage.getItem(STORAGE_KEYS.LOCAL_VOTES)
    const localVotes = localVotesStr ? JSON.parse(localVotesStr) : []

    // Filter votes for the requested side
    const sideVotes = localVotes.filter((vote: any) => vote.winner === side)

    // Load static pairs to get image data
    const allPairs = await loadStaticPairs()

    // Group votes by pair_id and count
    const voteCounts: Record<string, number> = {}
    sideVotes.forEach((vote: any) => {
      voteCounts[vote.pair_id] = (voteCounts[vote.pair_id] || 0) + 1
    })

    // Build winners array
    const winners: WinnerImage[] = Object.entries(voteCounts).map(([pairId, count]) => {
      const pair = allPairs.find(p => p.pair_id === pairId)
      const imageUrl = side === 'left' ? pair?.left_url : pair?.right_url

      return {
        image_url: imageUrl || '',
        prompt: pair?.prompt || 'Unknown',
        provider: pair?.provider || 'Unknown',
        pair_id: pairId,
        timestamp: new Date().toISOString(),
        vote_count: count as number,
      }
    })

    // Sort by vote count descending
    return winners.sort((a, b) => b.vote_count - a.vote_count)
  } else {
    // Full mode: Fetch from API
    const response = await fetch(`${config.api.baseUrl}/images/winners?side=${side}`)
    if (!response.ok) {
      throw new Error('Failed to fetch winners')
    }

    const data = await response.json()
    return data.data.winners || []
  }
}

/**
 * Reset all local data (for lite mode)
 */
export function resetLocalData(): void {
  if (typeof window === 'undefined') return

  localStorage.removeItem(STORAGE_KEYS.VOTED_PAIRS)
  localStorage.removeItem(STORAGE_KEYS.LOCAL_VOTES)
  localStorage.removeItem(STORAGE_KEYS.LOCAL_STATS)

  console.log('[DataService] Local data reset')
}
