import type { Metadata } from 'next'
import { Inter } from 'next/font/google'
import './globals.css'

const inter = Inter({ subsets: ['latin'] })

export const metadata: Metadata = {
  title: 'AI Image Battle',
  description: 'Compare AI-generated images and vote for your favorite',
  viewport: 'width=device-width, initial-scale=1, maximum-scale=1, user-scalable=0',
}

export default function RootLayout({
  children,
}: {
  children: React.ReactNode
}) {
  return (
    <html lang="en">
      <body className={`${inter.className} overflow-hidden`}>
        {children}
      </body>
    </html>
  )
}