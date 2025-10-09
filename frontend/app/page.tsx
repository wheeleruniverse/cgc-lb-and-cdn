'use client'

import { useState, useEffect } from 'react'
import ImageBattle from './components/ImageBattle'
import LoadingScreen from './components/LoadingScreen'

export default function Home() {
  const [isLoading, setIsLoading] = useState(true)

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
        <ImageBattle />
      )}
    </main>
  )
}