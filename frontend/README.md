# AI Image Battle Frontend

A mobile-optimized Next.js frontend for comparing AI-generated images with swipe functionality and amazing animations.

## Features

- **2-Image Comparison**: Side-by-side image comparison with voting
- **Mobile-First Design**: Optimized for mobile devices with touch gestures
- **Swipe Functionality**: Swipe left/right to vote for your preferred image
- **Amazing Animations**: Framer Motion animations with spring physics
- **Real-time Stats**: Track votes and voting streaks
- **Responsive UI**: Works on desktop and mobile devices

## Tech Stack

- **Next.js 14** - React framework with App Router
- **TypeScript** - Type-safe development
- **Tailwind CSS** - Utility-first CSS framework
- **Framer Motion** - Advanced animations and gestures
- **@use-gesture/react** - Touch gesture handling
- **@react-spring/web** - Spring physics animations

## Getting Started

1. **Install dependencies:**
   ```bash
   npm install
   ```

2. **Start the development server:**
   ```bash
   npm run dev
   ```

3. **Make sure the backend is running:**
   The backend should be running on `http://localhost:8080`

4. **Open your browser:**
   Navigate to `http://localhost:3000`

## API Integration

The frontend connects to the backend API:

- `GET /api/v1/images/pair` - Fetch random image pairs
- `POST /api/v1/images/rate` - Submit comparison votes

## Mobile Experience

The app is specifically designed for mobile use with:

- Touch-optimized interface
- Swipe gestures for voting
- Haptic feedback (when available)
- Responsive design
- Prevented text selection and zoom

## Animation Features

- **Loading animations** with rotating spinners
- **Swipe animations** with physics-based movement
- **Winner effects** with scale and glow animations
- **Background gradients** that shift over time
- **Smooth transitions** between image pairs

## Development

```bash
# Development server
npm run dev

# Build for production
npm run build

# Start production server
npm start

# Lint code
npm run lint
```

## Deployment

This frontend is designed to work with:
- Digital Ocean App Platform
- Vercel
- Netlify
- Any static hosting provider

Make sure to update the API endpoints in production to point to your deployed backend.