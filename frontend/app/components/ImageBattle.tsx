'use client'

import { useState, useEffect, useCallback } from 'react'
import { motion, AnimatePresence } from 'framer-motion'
import { useDrag } from '@use-gesture/react'
import { useSpring, animated } from '@react-spring/web'
import WinnersGrid from './WinnersGrid'

interface ImagePair {
  pair_id: string
  prompt: string
  provider: string
  left_url: string
  right_url: string
}

// Generate a random session ID for anonymous user tracking
function generateSessionId(): string {
  // Create a unique session ID using timestamp + random values
  const timestamp = Date.now().toString(36)
  const randomPart = Math.random().toString(36).substring(2, 15)
  return `sess_${timestamp}_${randomPart}`
}

export default function ImageBattle() {
  const [imagePair, setImagePair] = useState<ImagePair | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [isVoting, setIsVoting] = useState(false)
  const [showWinnerEffect, setShowWinnerEffect] = useState<'left' | 'right' | null>(null)
  const [teamScores, setTeamScores] = useState({ left: 0, right: 0 })
  const [keyboardHighlight, setKeyboardHighlight] = useState<'left' | 'right' | null>(null)
  const [votedPairIds, setVotedPairIds] = useState<string[]>(() => {
    // Load voted pairs from localStorage on mount
    if (typeof window !== 'undefined') {
      const stored = localStorage.getItem('votedPairIds')
      return stored ? JSON.parse(stored) : []
    }
    return []
  })
  const [showWinnersGrid, setShowWinnersGrid] = useState<'left' | 'right' | null>(null)

  // Session ID for tracking anonymous users
  const [sessionId] = useState<string>(() => {
    // Load or create session ID from localStorage
    if (typeof window !== 'undefined') {
      const stored = localStorage.getItem('sessionId')
      if (stored) {
        return stored
      }
      // Generate new session ID
      const newSessionId = generateSessionId()
      localStorage.setItem('sessionId', newSessionId)
      return newSessionId
    }
    return generateSessionId()
  })

  // Fetch team scores from backend
  const fetchTeamScores = async () => {
    try {
      const response = await fetch('/api/v1/statistics')
      if (response.ok) {
        const data = await response.json()
        if (data.data && data.data.side_wins) {
          setTeamScores({
            left: data.data.side_wins.left || 0,
            right: data.data.side_wins.right || 0,
          })
        }
      }
    } catch (err) {
      console.error('Failed to fetch team scores:', err)
    }
  }

  // Spring animations for individual images
  const [{ x: xLeft, scale: scaleLeft }, apiLeft] = useSpring(() => ({
    x: 0,
    scale: 1,
    config: { tension: 300, friction: 30 }
  }))

  const [{ x: xRight, scale: scaleRight }, apiRight] = useSpring(() => ({
    x: 0,
    scale: 1,
    config: { tension: 300, friction: 30 }
  }))

  const fetchImagePair = async () => {
    try {
      setLoading(true)
      setError(null)

      // Build URL with session ID and excluded pair IDs
      let url = '/api/v1/images/pair'
      const params = new URLSearchParams()

      // Add session ID for session-based tracking
      params.append('session_id', sessionId)

      // Add excluded pair IDs (for backwards compatibility and extra filtering)
      if (votedPairIds.length > 0) {
        params.append('exclude', votedPairIds.join(','))
      }

      url += `?${params.toString()}`

      const response = await fetch(url)

      if (!response.ok) {
        // Parse error response from backend
        const errorData = await response.json()
        throw new Error(errorData.error || 'Failed to fetch image pair')
      }

      const data = await response.json()
      setImagePair(data.data)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Unknown error')
    } finally {
      setLoading(false)
    }
  }

  const submitVote = async (winner: 'left' | 'right') => {
    if (!imagePair || isVoting) return

    setIsVoting(true)
    setShowWinnerEffect(winner)

    try {
      const response = await fetch('/api/v1/images/rate', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          pair_id: imagePair.pair_id,
          winner,
        }),
      })

      if (!response.ok) {
        throw new Error('Failed to submit vote')
      }

      // Add this pair to voted pairs and save to localStorage
      const newVotedPairIds = [...votedPairIds, imagePair.pair_id]
      setVotedPairIds(newVotedPairIds)
      if (typeof window !== 'undefined') {
        localStorage.setItem('votedPairIds', JSON.stringify(newVotedPairIds))
      }

      // Fetch updated team scores from backend
      fetchTeamScores()

      // Wait for animation to complete
      setTimeout(() => {
        setShowWinnerEffect(null)
        setIsVoting(false)
        fetchImagePair()
      }, 1500)

    } catch (err) {
      console.error('Vote submission failed:', err)
      setShowWinnerEffect(null)
      setIsVoting(false)
    }
  }

  // Drag gesture handler for LEFT image
  const bindLeft = useDrag(
    ({ down, movement: [mx], velocity: [vx], tap }) => {
      if (isVoting) return

      // If it's a tap/click, vote immediately
      if (tap) {
        submitVote('left')
        return
      }

      const trigger = Math.abs(mx) > 80 || Math.abs(vx) > 0.5

      if (!down && trigger) {
        // Vote for left when dragging left image
        submitVote('left')
        apiLeft.start({ x: -200, scale: 1.1 })
        setTimeout(() => {
          apiLeft.start({ x: 0, scale: 1 })
        }, 300)
      } else {
        // Follow drag
        apiLeft.start({
          x: down ? mx : 0,
          scale: down ? 1.05 : 1,
          immediate: down
        })
      }
    },
    { axis: 'x', preventScroll: true, filterTaps: true }
  )

  // Drag gesture handler for RIGHT image
  const bindRight = useDrag(
    ({ down, movement: [mx], velocity: [vx], tap }) => {
      if (isVoting) return

      // If it's a tap/click, vote immediately
      if (tap) {
        submitVote('right')
        return
      }

      const trigger = Math.abs(mx) > 80 || Math.abs(vx) > 0.5

      if (!down && trigger) {
        // Vote for right when dragging right image
        submitVote('right')
        apiRight.start({ x: 200, scale: 1.1 })
        setTimeout(() => {
          apiRight.start({ x: 0, scale: 1 })
        }, 300)
      } else {
        // Follow drag
        apiRight.start({
          x: down ? mx : 0,
          scale: down ? 1.05 : 1,
          immediate: down
        })
      }
    },
    { axis: 'x', preventScroll: true, filterTaps: true }
  )

  useEffect(() => {
    fetchImagePair()
    fetchTeamScores()

    // Refresh team scores every 30 seconds
    const interval = setInterval(fetchTeamScores, 30000)
    return () => clearInterval(interval)
  }, [])

  // Keyboard controls for desktop
  useEffect(() => {
    const handleKeyPress = (event: KeyboardEvent) => {
      if (isVoting || !imagePair) return

      if (event.key === 'ArrowLeft') {
        event.preventDefault()
        setKeyboardHighlight('left')
        setTimeout(() => setKeyboardHighlight(null), 200)
        submitVote('left')
      } else if (event.key === 'ArrowRight') {
        event.preventDefault()
        setKeyboardHighlight('right')
        setTimeout(() => setKeyboardHighlight(null), 200)
        submitVote('right')
      }
    }

    window.addEventListener('keydown', handleKeyPress)
    return () => window.removeEventListener('keydown', handleKeyPress)
  }, [isVoting, imagePair])

  // Show winners grid if requested
  if (showWinnersGrid) {
    return (
      <WinnersGrid
        side={showWinnersGrid}
        onClose={() => setShowWinnersGrid(null)}
      />
    )
  }

  if (error) {
    // Check if it's the "no pairs yet" error (friendlier UI)
    const isNoPairsYet = error.includes('No image pairs available yet')
    const isAllVoted = error.includes("You've voted on all available pairs")

    return (
      <div className="flex items-center justify-center min-h-screen bg-gradient-to-br from-purple-900 via-blue-900 to-indigo-900">
        <motion.div
          initial={{ opacity: 0, scale: 0.9 }}
          animate={{ opacity: 1, scale: 1 }}
          className={`text-center p-8 rounded-lg backdrop-blur-md max-w-md mx-4 ${
            isNoPairsYet ? 'bg-blue-900/40 border-2 border-blue-400/30' :
            isAllVoted ? 'bg-green-900/40 border-2 border-green-400/30' :
            'bg-red-900/30'
          }`}
        >
          {isNoPairsYet ? (
            <>
              <motion.div
                animate={{ rotate: 360 }}
                transition={{ duration: 2, repeat: Infinity, ease: "linear" }}
                className="w-16 h-16 mx-auto mb-4"
              >
                🎨
              </motion.div>
              <h2 className="text-2xl font-bold text-blue-200 mb-4">Getting Ready...</h2>
              <p className="text-blue-100 mb-6 leading-relaxed">{error}</p>
              <button
                onClick={fetchImagePair}
                className="px-6 py-3 bg-blue-600 hover:bg-blue-700 text-white rounded-lg transition-colors font-semibold"
              >
                Check Again
              </button>
            </>
          ) : isAllVoted ? (
            <>
              <div className="text-6xl mb-4">🎉</div>
              <h2 className="text-2xl font-bold text-green-200 mb-4">All Done!</h2>
              <p className="text-green-100 mb-6 leading-relaxed">{error}</p>
              <button
                onClick={() => {
                  // Clear voted pairs and reset
                  setVotedPairIds([])
                  if (typeof window !== 'undefined') {
                    localStorage.removeItem('votedPairIds')
                  }
                  fetchImagePair()
                }}
                className="px-6 py-3 bg-green-600 hover:bg-green-700 text-white rounded-lg transition-colors font-semibold"
              >
                Start Over
              </button>
            </>
          ) : (
            <>
              <h2 className="text-2xl font-bold text-red-300 mb-4">Oops!</h2>
              <p className="text-red-200 mb-6">{error}</p>
              <button
                onClick={fetchImagePair}
                className="px-6 py-3 bg-red-600 hover:bg-red-700 text-white rounded-lg transition-colors"
              >
                Try Again
              </button>
            </>
          )}
        </motion.div>
      </div>
    )
  }

  if (loading || !imagePair) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <motion.div
          animate={{ rotate: 360 }}
          transition={{ duration: 1, repeat: Infinity, ease: "linear" }}
          className="w-12 h-12 border-4 border-white border-t-transparent rounded-full"
        />
      </div>
    )
  }

  return (
    <div className="min-h-screen relative overflow-hidden bg-gradient-to-br from-purple-900 via-blue-900 to-indigo-900">
      {/* Background Effects */}
      <div className="absolute inset-0 overflow-hidden">
        <motion.div
          animate={{
            background: [
              "radial-gradient(circle at 20% 80%, rgba(120, 119, 198, 0.3) 0%, transparent 50%)",
              "radial-gradient(circle at 80% 20%, rgba(255, 119, 198, 0.3) 0%, transparent 50%)",
              "radial-gradient(circle at 40% 40%, rgba(119, 255, 198, 0.3) 0%, transparent 50%)",
            ],
          }}
          transition={{ duration: 8, repeat: Infinity, ease: "easeInOut" }}
          className="absolute inset-0"
        />
      </div>

      {/* Header with Team Scores */}
      <motion.div
        initial={{ y: -50, opacity: 0 }}
        animate={{ y: 0, opacity: 1 }}
        className="relative z-10 pt-6 pb-4 px-6"
      >
        <div className="text-center mb-6">
          <h1 className="text-3xl font-bold text-white mb-2">AI Image Battle</h1>
          <p className="text-purple-200 text-sm">Click your favorite image!</p>
        </div>

        {/* Team Scores with Vertical Divider */}
        <div className="flex items-center justify-center max-w-md mx-auto">
          {/* Team Left */}
          <button
            onClick={() => setShowWinnersGrid('left')}
            className="flex-1 text-center group transition-transform hover:scale-105 active:scale-95"
          >
            <div className="bg-blue-500/20 backdrop-blur-md rounded-lg px-4 py-3 border border-blue-400/30 group-hover:border-blue-400/60 group-hover:bg-blue-500/30 transition-all cursor-pointer">
              <div className="text-2xl font-bold text-blue-300">{teamScores.left}</div>
              <div className="text-xs text-blue-200 group-hover:text-blue-100">Team Left</div>
              <div className="text-[10px] text-blue-300/60 group-hover:text-blue-200/80 mt-1">Click to view winners</div>
            </div>
          </button>

          {/* Vertical Divider with VS */}
          <div className="flex flex-col items-center mx-4">
            <div className="w-px h-8 bg-gradient-to-b from-transparent via-white/50 to-transparent"></div>
            <div className="bg-white/20 backdrop-blur-md rounded-full w-12 h-12 flex items-center justify-center border border-white/30 my-2">
              <span className="text-white font-bold text-sm">VS</span>
            </div>
            <div className="w-px h-8 bg-gradient-to-b from-white/50 via-transparent to-transparent"></div>
          </div>

          {/* Team Right */}
          <button
            onClick={() => setShowWinnersGrid('right')}
            className="flex-1 text-center group transition-transform hover:scale-105 active:scale-95"
          >
            <div className="bg-purple-500/20 backdrop-blur-md rounded-lg px-4 py-3 border border-purple-400/30 group-hover:border-purple-400/60 group-hover:bg-purple-500/30 transition-all cursor-pointer">
              <div className="text-2xl font-bold text-purple-300">{teamScores.right}</div>
              <div className="text-xs text-purple-200 group-hover:text-purple-100">Team Right</div>
              <div className="text-[10px] text-purple-300/60 group-hover:text-purple-200/80 mt-1">Click to view winners</div>
            </div>
          </button>
        </div>
      </motion.div>

      {/* Main Battle Area */}
      <div className="relative z-10 px-4 pb-8">
        <motion.div
          layout
          className="max-w-4xl mx-auto"
        >
          {/* Prompt Display */}
          <motion.div
            initial={{ opacity: 0, y: -10 }}
            animate={{ opacity: 1, y: 0 }}
            className="mb-4 text-center"
          >
            <div className="bg-white/10 backdrop-blur-md rounded-lg px-6 py-3 border border-white/20">
              <p className="text-white/60 text-xs uppercase tracking-wider mb-1">Prompt</p>
              <p className="text-white text-sm md:text-base font-medium">{imagePair.prompt}</p>
            </div>
          </motion.div>

          <div className="grid grid-cols-2 gap-4 h-[60vh] relative">
            {/* Left Image */}
            <animated.div
              {...bindLeft()}
              style={{ x: xLeft, scale: scaleLeft }}
              className="touch-none cursor-pointer hover:cursor-grab active:cursor-grabbing"
            >
              <motion.div
                className={`relative rounded-xl overflow-hidden shadow-2xl h-full ${
                  showWinnerEffect === 'left' ? 'ring-4 ring-green-400' : ''
                } ${
                  keyboardHighlight === 'left' ? 'ring-4 ring-blue-400 ring-opacity-75' : ''
                }`}
                whileHover={{ scale: 1.02 }}
                transition={{ duration: 0.2 }}
                animate={keyboardHighlight === 'left' ? { scale: 1.05 } : { scale: 1 }}
              >
                <img
                  src={imagePair.left_url}
                  alt="Left choice"
                  className="w-full h-full object-cover"
                  onError={(e) => {
                    const target = e.target as HTMLImageElement
                    target.src = 'data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iMjAwIiBoZWlnaHQ9IjIwMCIgeG1sbnM9Imh0dHA6Ly93d3cudzMub3JnLzIwMDAvc3ZnIj48cmVjdCB3aWR0aD0iMTAwJSIgaGVpZ2h0PSIxMDAlIiBmaWxsPSIjZGRkIi8+PHRleHQgeD0iNTAlIiB5PSI1MCUiIGZvbnQtc2l6ZT0iMTQiIHRleHQtYW5jaG9yPSJtaWRkbGUiIGR5PSIuM2VtIj5JbWFnZSBOb3QgRm91bmQ8L3RleHQ+PC9zdmc+'
                  }}
                />
                <div className="absolute inset-0 bg-gradient-to-t from-black/50 to-transparent" />
                <div className="absolute bottom-4 left-4 text-white">
                  <div className="text-sm font-medium">{imagePair.provider}</div>
                  <div className="text-xs opacity-75">Click to vote</div>
                </div>

                {showWinnerEffect === 'left' && (
                  <motion.div
                    initial={{ opacity: 0, scale: 0 }}
                    animate={{ opacity: 1, scale: 1 }}
                    className="absolute inset-0 flex items-center justify-center"
                  >
                    <div className="bg-green-500 text-white rounded-full p-6 text-4xl font-bold shadow-2xl">
                      ✓
                    </div>
                  </motion.div>
                )}
              </motion.div>
            </animated.div>

            {/* Right Image */}
            <animated.div
              {...bindRight()}
              style={{ x: xRight, scale: scaleRight }}
              className="touch-none cursor-pointer hover:cursor-grab active:cursor-grabbing"
            >
              <motion.div
                className={`relative rounded-xl overflow-hidden shadow-2xl h-full ${
                  showWinnerEffect === 'right' ? 'ring-4 ring-green-400' : ''
                } ${
                  keyboardHighlight === 'right' ? 'ring-4 ring-blue-400 ring-opacity-75' : ''
                }`}
                whileHover={{ scale: 1.02 }}
                transition={{ duration: 0.2 }}
                animate={keyboardHighlight === 'right' ? { scale: 1.05 } : { scale: 1 }}
              >
                <img
                  src={imagePair.right_url}
                  alt="Right choice"
                  className="w-full h-full object-cover"
                  onError={(e) => {
                    const target = e.target as HTMLImageElement
                    target.src = 'data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iMjAwIiBoZWlnaHQ9IjIwMCIgeG1sbnM9Imh0dHA6Ly93d3cudzMub3JnLzIwMDAvc3ZnIj48cmVjdCB3aWR0aD0iMTAwJSIgaGVpZ2h0PSIxMDAlIiBmaWxsPSIjZGRkIi8+PHRleHQgeD0iNTAlIiB5PSI1MCUiIGZvbnQtc2l6ZT0iMTQiIHRleHQtYW5jaG9yPSJtaWRkbGUiIGR5PSIuM2VtIj5JbWFnZSBOb3QgRm91bmQ8L3RleHQ+PC9zdmc+'
                  }}
                />
                <div className="absolute inset-0 bg-gradient-to-t from-black/50 to-transparent" />
                <div className="absolute bottom-4 right-4 text-white text-right">
                  <div className="text-sm font-medium">{imagePair.provider}</div>
                  <div className="text-xs opacity-75">Click to vote</div>
                </div>

                {showWinnerEffect === 'right' && (
                  <motion.div
                    initial={{ opacity: 0, scale: 0 }}
                    animate={{ opacity: 1, scale: 1 }}
                    className="absolute inset-0 flex items-center justify-center"
                  >
                    <div className="bg-green-500 text-white rounded-full p-6 text-4xl font-bold shadow-2xl">
                      ✓
                    </div>
                  </motion.div>
                )}
              </motion.div>
            </animated.div>
          </div>


        </motion.div>

        {/* Instructions */}
        <motion.div
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          transition={{ delay: 1 }}
          className="mt-8 text-center"
        >
          <div className="flex flex-wrap justify-center items-center gap-4 text-white/60 text-sm">
            <div className="flex items-center bg-white/10 px-4 py-2 rounded-lg">
              <span>👆 Click or drag an image to vote for it</span>
            </div>
            <div className="hidden md:flex items-center bg-white/10 px-3 py-2 rounded-lg">
              <kbd className="bg-white/20 px-2 py-1 rounded text-xs mr-2">←</kbd>
              <span>Vote Left</span>
            </div>
            <div className="hidden md:flex items-center bg-white/10 px-3 py-2 rounded-lg">
              <kbd className="bg-white/20 px-2 py-1 rounded text-xs mr-2">→</kbd>
              <span>Vote Right</span>
            </div>
          </div>
        </motion.div>
      </div>
    </div>
  )
}