'use client'

import { useState, useEffect } from 'react'
import ImageBattle from './components/ImageBattle'
import LoadingScreen from './components/LoadingScreen'
import Leaderboard from './components/Leaderboard'

export default function Home() {
  const [isLoading, setIsLoading] = useState(true)
  const [showLeaderboard, setShowLeaderboard] = useState(false)

  useEffect(() => {
    // Simulate initial loading
    const timer = setTimeout(() => {
      setIsLoading(false)
    }, 2000)

    return () => clearTimeout(timer)
  }, [])

  return (
    <main className="min-h-screen w-full relative overflow-hidden">
      {isLoading ? (
        <LoadingScreen />
      ) : (
        <div className="flex flex-col lg:flex-row min-h-screen">
          {/* Main Battle Area */}
          <div className="flex-1">
            <ImageBattle />
          </div>

          {/* Leaderboard Sidebar - Desktop */}
          <div className="hidden lg:block w-96 bg-gradient-to-br from-purple-950 via-indigo-950 to-blue-950 border-l border-white/10 overflow-y-auto">
            <div className="p-6">
              <Leaderboard />
            </div>
          </div>

          {/* Leaderboard Toggle Button - Mobile */}
          <button
            onClick={() => setShowLeaderboard(!showLeaderboard)}
            className="lg:hidden fixed bottom-4 right-4 bg-purple-600 hover:bg-purple-700 text-white px-6 py-3 rounded-full shadow-lg font-semibold z-50 transition-all"
          >
            {showLeaderboard ? 'Hide' : 'Show'} Leaderboard
          </button>

          {/* Leaderboard Modal - Mobile */}
          {showLeaderboard && (
            <div className="lg:hidden fixed inset-0 bg-black/80 backdrop-blur-sm z-40 flex items-center justify-center p-4">
              <div className="bg-gradient-to-br from-purple-900 to-indigo-900 rounded-xl max-w-md w-full max-h-[80vh] overflow-y-auto p-6">
                <Leaderboard />
              </div>
            </div>
          )}
        </div>
      )}
    </main>
  )
}