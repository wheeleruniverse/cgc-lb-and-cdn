'use client'

import { useState, useEffect } from 'react'
import { motion } from 'framer-motion'

interface LeaderboardEntry {
  provider: string
  wins: number
  losses: number
  total_votes: number
  win_rate: number
}

export default function Leaderboard() {
  const [leaderboard, setLeaderboard] = useState<LeaderboardEntry[]>([])
  const [totalVotes, setTotalVotes] = useState(0)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    const fetchLeaderboard = async () => {
      try {
        setLoading(true)
        const response = await fetch('/api/v1/leaderboard')
        if (!response.ok) {
          throw new Error('Failed to fetch leaderboard')
        }
        const data = await response.json()
        setLeaderboard(data.data.leaderboard || [])

        // Also fetch total votes
        const statsResponse = await fetch('/api/v1/statistics')
        if (statsResponse.ok) {
          const statsData = await statsResponse.json()
          setTotalVotes(statsData.data.total_votes || 0)
        }
      } catch (err) {
        setError(err instanceof Error ? err.Message : 'Unknown error')
      } finally {
        setLoading(false)
      }
    }

    fetchLeaderboard()

    // Refresh every 10 seconds
    const interval = setInterval(fetchLeaderboard, 10000)
    return () => clearInterval(interval)
  }, [])

  if (loading) {
    return (
      <div className="flex items-center justify-center p-8">
        <motion.div
          animate={{ rotate: 360 }}
          transition={{ duration: 1, repeat: Infinity, ease: "linear" }}
          className="w-8 h-8 border-4 border-white border-t-transparent rounded-full"
        />
      </div>
    )
  }

  if (error) {
    return (
      <div className="p-4 bg-red-900/30 rounded-lg backdrop-blur-md">
        <p className="text-red-300 text-sm">Leaderboard unavailable</p>
      </div>
    )
  }

  return (
    <motion.div
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      className="bg-white/10 backdrop-blur-md rounded-xl p-6 border border-white/20"
    >
      <div className="flex items-center justify-between mb-4">
        <h2 className="text-2xl font-bold text-white">Leaderboard</h2>
        <div className="text-sm text-white/60">
          {totalVotes} total votes
        </div>
      </div>

      <div className="space-y-3">
        {leaderboard.map((entry, index) => (
          <motion.div
            key={entry.provider}
            initial={{ opacity: 0, x: -20 }}
            animate={{ opacity: 1, x: 0 }}
            transition={{ delay: index * 0.1 }}
            className="bg-white/5 rounded-lg p-4 hover:bg-white/10 transition-all"
          >
            <div className="flex items-center justify-between">
              <div className="flex items-center space-x-4">
                <div className={`text-2xl font-bold ${
                  index === 0 ? 'text-yellow-300' :
                  index === 1 ? 'text-gray-300' :
                  index === 2 ? 'text-orange-400' :
                  'text-white/60'
                }`}>
                  #{index + 1}
                </div>
                <div>
                  <div className="text-white font-semibold capitalize">
                    {entry.provider}
                  </div>
                  <div className="text-sm text-white/60">
                    {entry.total_votes} votes
                  </div>
                </div>
              </div>
              <div className="text-right">
                <div className="text-xl font-bold text-green-400">
                  {entry.win_rate.toFixed(1)}%
                </div>
                <div className="text-xs text-white/60">
                  {entry.wins}W / {entry.losses}L
                </div>
              </div>
            </div>

            {/* Win rate bar */}
            <div className="mt-2 h-2 bg-white/10 rounded-full overflow-hidden">
              <motion.div
                initial={{ width: 0 }}
                animate={{ width: `${entry.win_rate}%` }}
                transition={{ duration: 1, delay: index * 0.1 }}
                className="h-full bg-gradient-to-r from-green-400 to-blue-500"
              />
            </div>
          </motion.div>
        ))}
      </div>

      {leaderboard.length === 0 && (
        <div className="text-center text-white/60 py-8">
          No votes yet. Be the first to vote!
        </div>
      )}
    </motion.div>
  )
}
