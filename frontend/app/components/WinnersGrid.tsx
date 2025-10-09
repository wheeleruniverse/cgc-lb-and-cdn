'use client'

import { useState, useEffect } from 'react'
import { motion } from 'framer-motion'

interface WinnerImage {
  image_url: string
  prompt: string
  provider: string
  pair_id: string
  timestamp: string
  vote_count: number
}

interface WinnersGridProps {
  side: 'left' | 'right'
  onClose: () => void
}

export default function WinnersGrid({ side, onClose }: WinnersGridProps) {
  const [winners, setWinners] = useState<WinnerImage[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    const fetchWinners = async () => {
      try {
        setLoading(true)
        setError(null)
        const response = await fetch(`/api/v1/images/winners?side=${side}`)

        if (!response.ok) {
          throw new Error('Failed to fetch winners')
        }

        const data = await response.json()
        setWinners(data.data.winners || [])
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Unknown error')
      } finally {
        setLoading(false)
      }
    }

    fetchWinners()
  }, [side])

  const sideColor = side === 'left' ? 'blue' : 'purple'
  const sideColorClass = side === 'left' ? 'from-blue-900 via-blue-800 to-indigo-900' : 'from-purple-900 via-purple-800 to-indigo-900'

  return (
    <motion.div
      initial={{ opacity: 0 }}
      animate={{ opacity: 1 }}
      exit={{ opacity: 0 }}
      className={`min-h-screen bg-gradient-to-br ${sideColorClass} p-6`}
    >
      {/* Header */}
      <div className="max-w-7xl mx-auto">
        <div className="flex items-center justify-between mb-8">
          <motion.div
            initial={{ x: -20, opacity: 0 }}
            animate={{ x: 0, opacity: 1 }}
          >
            <h1 className="text-3xl md:text-4xl font-bold text-white">
              Team {side === 'left' ? 'Left' : 'Right'} Winners 🏆
            </h1>
            <p className="text-white/60 mt-2">
              All images that won from the {side} position
            </p>
          </motion.div>
          <button
            onClick={onClose}
            className="bg-white/10 hover:bg-white/20 text-white px-6 py-3 rounded-lg backdrop-blur-md border border-white/20 transition-all font-semibold"
          >
            ← Back to Voting
          </button>
        </div>

        {/* Loading State */}
        {loading && (
          <div className="flex items-center justify-center py-20">
            <motion.div
              animate={{ rotate: 360 }}
              transition={{ duration: 1, repeat: Infinity, ease: "linear" }}
              className="w-12 h-12 border-4 border-white border-t-transparent rounded-full"
            />
          </div>
        )}

        {/* Error State */}
        {error && (
          <div className="bg-red-900/30 rounded-lg p-6 text-center">
            <p className="text-red-300">{error}</p>
            <button
              onClick={onClose}
              className="mt-4 bg-red-600 hover:bg-red-700 text-white px-6 py-3 rounded-lg transition-colors"
            >
              Go Back
            </button>
          </div>
        )}

        {/* Winners Grid */}
        {!loading && !error && (
          <>
            {winners.length === 0 ? (
              <div className="bg-white/10 backdrop-blur-md rounded-lg p-12 text-center border border-white/20">
                <div className="text-6xl mb-4">🤷</div>
                <h2 className="text-2xl font-bold text-white mb-2">No Winners Yet</h2>
                <p className="text-white/60 mb-6">
                  No images have won from the {side} position yet. Keep voting!
                </p>
                <button
                  onClick={onClose}
                  className="bg-white/20 hover:bg-white/30 text-white px-6 py-3 rounded-lg transition-all"
                >
                  Back to Voting
                </button>
              </div>
            ) : (
              <>
                <div className="mb-6 bg-white/10 backdrop-blur-md rounded-lg px-6 py-3 border border-white/20 inline-block">
                  <p className="text-white font-semibold">
                    {winners.length} {winners.length === 1 ? 'winner' : 'winners'}
                  </p>
                </div>

                <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-6">
                  {winners.map((winner, index) => (
                    <motion.div
                      key={winner.pair_id}
                      initial={{ opacity: 0, y: 20 }}
                      animate={{ opacity: 1, y: 0 }}
                      transition={{ delay: index * 0.05 }}
                      className="bg-white/10 backdrop-blur-md rounded-xl overflow-hidden border border-white/20 hover:border-white/40 transition-all hover:scale-105"
                    >
                      {/* Image */}
                      <div className="relative aspect-square">
                        <img
                          src={winner.image_url}
                          alt={winner.prompt}
                          className="w-full h-full object-cover"
                          onError={(e) => {
                            const target = e.target as HTMLImageElement
                            target.src = 'data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iMjAwIiBoZWlnaHQ9IjIwMCIgeG1sbnM9Imh0dHA6Ly93d3cudzMub3JnLzIwMDAvc3ZnIj48cmVjdCB3aWR0aD0iMTAwJSIgaGVpZ2h0PSIxMDAlIiBmaWxsPSIjZGRkIi8+PHRleHQgeD0iNTAlIiB5PSI1MCUiIGZvbnQtc2l6ZT0iMTQiIHRleHQtYW5jaG9yPSJtaWRkbGUiIGR5PSIuM2VtIj5JbWFnZSBOb3QgRm91bmQ8L3RleHQ+PC9zdmc+'
                          }}
                        />
                        {/* Winner Badge */}
                        <div className="absolute top-2 right-2 bg-yellow-500 text-yellow-900 rounded-full w-10 h-10 flex items-center justify-center font-bold text-xl shadow-lg">
                          🏆
                        </div>
                        {/* Vote Count Badge */}
                        <div className="absolute top-2 left-2 bg-black/70 backdrop-blur-sm text-white rounded-full px-3 py-1 flex items-center gap-1 text-sm font-semibold shadow-lg">
                          <span>❤️</span>
                          <span>{winner.vote_count}</span>
                        </div>
                      </div>

                      {/* Details */}
                      <div className="p-4">
                        <p className="text-white/80 text-xs line-clamp-2 mb-2">
                          {winner.prompt}
                        </p>
                        <div className="flex items-center justify-between">
                          <span className="text-white/60 text-xs capitalize">
                            {winner.provider}
                          </span>
                          <span className={`text-${sideColor}-400 text-xs font-semibold`}>
                            {side === 'left' ? 'Left' : 'Right'}
                          </span>
                        </div>
                      </div>
                    </motion.div>
                  ))}
                </div>
              </>
            )}
          </>
        )}
      </div>
    </motion.div>
  )
}
