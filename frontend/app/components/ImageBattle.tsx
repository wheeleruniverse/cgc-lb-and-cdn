'use client'

import { useState, useEffect, useCallback } from 'react'
import { motion, AnimatePresence } from 'framer-motion'
import { useDrag } from '@use-gesture/react'
import { useSpring, animated } from '@react-spring/web'

interface ImageInfo {
  id: string
  filename: string
  path: string
  url: string
  provider: string
  size: number
}

interface ImagePair {
  pair_id: string
  left: ImageInfo
  right: ImageInfo
}

export default function ImageBattle() {
  const [imagePair, setImagePair] = useState<ImagePair | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [isVoting, setIsVoting] = useState(false)
  const [showWinnerEffect, setShowWinnerEffect] = useState<'left' | 'right' | null>(null)
  const [teamScores, setTeamScores] = useState({ left: 0, right: 0 })
  const [keyboardHighlight, setKeyboardHighlight] = useState<'left' | 'right' | null>(null)

  // Spring animation for the main container
  const [{ x, rotate, scale }, api] = useSpring(() => ({
    x: 0,
    rotate: 0,
    scale: 1,
    config: { tension: 300, friction: 30 }
  }))

  const fetchImagePair = async () => {
    try {
      setLoading(true)
      setError(null)
      const response = await fetch('/api/v1/images/pair')
      if (!response.ok) {
        throw new Error('Failed to fetch image pair')
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
          left_id: imagePair.left.id,
          right_id: imagePair.right.id,
        }),
      })

      if (!response.ok) {
        throw new Error('Failed to submit vote')
      }

      // Update team scores
      setTeamScores(prev => ({
        ...prev,
        [winner]: prev[winner] + 1
      }))

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

  // Drag gesture handler
  const bind = useDrag(
    ({ down, movement: [mx], direction: [xDir], velocity: [vx] }) => {
      const trigger = Math.abs(mx) > 50 || Math.abs(vx) > 0.5
      const side = xDir < 0 ? 'left' : 'right'

      if (!down && trigger && !isVoting) {
        // Trigger vote
        submitVote(side)
        api.start({ x: xDir < 0 ? -200 : 200, rotate: xDir * 15, scale: 0.9 })
        setTimeout(() => {
          api.start({ x: 0, rotate: 0, scale: 1 })
        }, 300)
      } else {
        // Follow drag
        api.start({
          x: down ? mx : 0,
          rotate: down ? mx * 0.1 : 0,
          scale: down ? 1.05 : 1,
          immediate: down
        })
      }
    },
    { axis: 'x', preventScroll: true }
  )

  useEffect(() => {
    fetchImagePair()
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

  if (error) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <motion.div
          initial={{ opacity: 0, scale: 0.9 }}
          animate={{ opacity: 1, scale: 1 }}
          className="text-center p-8 bg-red-900/30 rounded-lg backdrop-blur-md"
        >
          <h2 className="text-2xl font-bold text-red-300 mb-4">Oops!</h2>
          <p className="text-red-200 mb-6">{error}</p>
          <button
            onClick={fetchImagePair}
            className="px-6 py-3 bg-red-600 hover:bg-red-700 text-white rounded-lg transition-colors"
          >
            Try Again
          </button>
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
          <p className="text-purple-200 text-sm">Swipe left or right to vote!</p>
        </div>

        {/* Team Scores with Vertical Divider */}
        <div className="flex items-center justify-center max-w-md mx-auto">
          {/* Team Left */}
          <div className="flex-1 text-center">
            <div className="bg-blue-500/20 backdrop-blur-md rounded-lg px-4 py-3 border border-blue-400/30">
              <div className="text-2xl font-bold text-blue-300">{teamScores.left}</div>
              <div className="text-xs text-blue-200">Team Left</div>
            </div>
          </div>

          {/* Vertical Divider with VS */}
          <div className="flex flex-col items-center mx-4">
            <div className="w-px h-8 bg-gradient-to-b from-transparent via-white/50 to-transparent"></div>
            <div className="bg-white/20 backdrop-blur-md rounded-full w-12 h-12 flex items-center justify-center border border-white/30 my-2">
              <span className="text-white font-bold text-sm">VS</span>
            </div>
            <div className="w-px h-8 bg-gradient-to-b from-white/50 via-transparent to-transparent"></div>
          </div>

          {/* Team Right */}
          <div className="flex-1 text-center">
            <div className="bg-purple-500/20 backdrop-blur-md rounded-lg px-4 py-3 border border-purple-400/30">
              <div className="text-2xl font-bold text-purple-300">{teamScores.right}</div>
              <div className="text-xs text-purple-200">Team Right</div>
            </div>
          </div>
        </div>
      </motion.div>

      {/* Main Battle Area */}
      <div className="relative z-10 px-4 pb-8">
        <motion.div
          layout
          className="max-w-4xl mx-auto"
        >
          <animated.div
            {...bind()}
            style={{ x, rotate, scale }}
            className="touch-none cursor-grab active:cursor-grabbing"
          >
            <div className="grid grid-cols-2 gap-4 h-[60vh] relative">
              {/* Left Image */}
              <motion.div
                className={`relative rounded-xl overflow-hidden shadow-2xl ${
                  showWinnerEffect === 'left' ? 'ring-4 ring-green-400' : ''
                } ${
                  keyboardHighlight === 'left' ? 'ring-4 ring-blue-400 ring-opacity-75' : ''
                }`}
                whileHover={{ scale: 1.02 }}
                transition={{ duration: 0.2 }}
                animate={keyboardHighlight === 'left' ? { scale: 1.05 } : { scale: 1 }}
              >
                <img
                  src={`http://localhost:8080${imagePair.left.url}`}
                  alt="Left choice"
                  className="w-full h-full object-cover"
                  onError={(e) => {
                    const target = e.target as HTMLImageElement
                    target.src = 'data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iMjAwIiBoZWlnaHQ9IjIwMCIgeG1sbnM9Imh0dHA6Ly93d3cudzMub3JnLzIwMDAvc3ZnIj48cmVjdCB3aWR0aD0iMTAwJSIgaGVpZ2h0PSIxMDAlIiBmaWxsPSIjZGRkIi8+PHRleHQgeD0iNTAlIiB5PSI1MCUiIGZvbnQtc2l6ZT0iMTQiIHRleHQtYW5jaG9yPSJtaWRkbGUiIGR5PSIuM2VtIj5JbWFnZSBOb3QgRm91bmQ8L3RleHQ+PC9zdmc+'
                  }}
                />
                <div className="absolute inset-0 bg-gradient-to-t from-black/50 to-transparent" />
                <div className="absolute bottom-4 left-4 text-white">
                  <div className="text-sm font-medium">{imagePair.left.provider}</div>
                  <div className="text-xs opacity-75">Tap or swipe left</div>
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

              {/* Right Image */}
              <motion.div
                className={`relative rounded-xl overflow-hidden shadow-2xl ${
                  showWinnerEffect === 'right' ? 'ring-4 ring-green-400' : ''
                } ${
                  keyboardHighlight === 'right' ? 'ring-4 ring-blue-400 ring-opacity-75' : ''
                }`}
                whileHover={{ scale: 1.02 }}
                transition={{ duration: 0.2 }}
                animate={keyboardHighlight === 'right' ? { scale: 1.05 } : { scale: 1 }}
              >
                <img
                  src={`http://localhost:8080${imagePair.right.url}`}
                  alt="Right choice"
                  className="w-full h-full object-cover"
                  onError={(e) => {
                    const target = e.target as HTMLImageElement
                    target.src = 'data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iMjAwIiBoZWlnaHQ9IjIwMCIgeG1sbnM9Imh0dHA6Ly93d3cudzMub3JnLzIwMDAvc3ZnIj48cmVjdCB3aWR0aD0iMTAwJSIgaGVpZ2h0PSIxMDAlIiBmaWxsPSIjZGRkIi8+PHRleHQgeD0iNTAlIiB5PSI1MCUiIGZvbnQtc2l6ZT0iMTQiIHRleHQtYW5jaG9yPSJtaWRkbGUiIGR5PSIuM2VtIj5JbWFnZSBOb3QgRm91bmQ8L3RleHQ+PC9zdmc+'
                  }}
                />
                <div className="absolute inset-0 bg-gradient-to-t from-black/50 to-transparent" />
                <div className="absolute bottom-4 right-4 text-white text-right">
                  <div className="text-sm font-medium">{imagePair.right.provider}</div>
                  <div className="text-xs opacity-75">Tap or swipe right</div>
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

            </div>
          </animated.div>


          {/* Mobile tap buttons */}
          <div className="grid grid-cols-2 gap-4 mt-6 md:hidden">
            <button
              onClick={() => submitVote('left')}
              disabled={isVoting}
              className="bg-blue-600 hover:bg-blue-700 disabled:bg-gray-600 text-white py-4 rounded-xl font-semibold text-lg transition-all transform hover:scale-105 active:scale-95"
            >
              Choose Left
            </button>
            <button
              onClick={() => submitVote('right')}
              disabled={isVoting}
              className="bg-purple-600 hover:bg-purple-700 disabled:bg-gray-600 text-white py-4 rounded-xl font-semibold text-lg transition-all transform hover:scale-105 active:scale-95"
            >
              Choose Right
            </button>
          </div>
        </motion.div>

        {/* Desktop and mobile instructions */}
        <motion.div
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          transition={{ delay: 1 }}
          className="mt-8 text-center"
        >
          {/* Mobile instructions */}
          <div className="flex justify-between px-8 text-white/60 text-sm md:hidden">
            <div className="flex items-center">
              <span className="mr-2">←</span> Swipe left
            </div>
            <div className="flex items-center">
              Swipe right <span className="ml-2">→</span>
            </div>
          </div>

          {/* Desktop instructions */}
          <div className="hidden md:flex justify-center items-center space-x-6 text-white/60 text-sm">
            <div className="flex items-center bg-white/10 px-3 py-2 rounded-lg">
              <kbd className="bg-white/20 px-2 py-1 rounded text-xs mr-2">←</kbd>
              <span>Left Arrow = Vote Left</span>
            </div>
            <div className="flex items-center bg-white/10 px-3 py-2 rounded-lg">
              <kbd className="bg-white/20 px-2 py-1 rounded text-xs mr-2">→</kbd>
              <span>Right Arrow = Vote Right</span>
            </div>
          </div>
        </motion.div>
      </div>
    </div>
  )
}