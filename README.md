# Cloud Portfolio Challenge Load Balancing and CDN - AI Image Service

A **Cloud Portfolio Challenge** implementation showcasing modern cloud architecture with load balancing, CDN, and intelligent AI image generation using Google's Agent Development Kit (ADK).

## 🏗️ Architecture Overview

This project implements a distributed AI image generation service with the following components:

- **Load Balancer**: Digital Ocean Load Balancer distributing traffic across multiple VMs
- **Backend Services**: Go-based API servers running on separate droplets
- **Frontend Application**: Next.js web application with mobile-optimized image comparison
- **CDN**: Digital Ocean Spaces CDN for fast global image delivery
- **Multi-Provider Intelligence**: ADK-powered orchestrator managing multiple AI image providers
- **Managed Database**: Valkey (Redis-compatible) for caching and session management

## 📁 Project Structure

```
├── backend/          # Go API server with ADK orchestrator
├── frontend/         # Next.js web application
├── hosting/          # Pulumi infrastructure as code
└── research/         # AI provider research and examples
```

## 🚀 Quick Start

### 1. Infrastructure Deployment
Deploy the complete infrastructure to Digital Ocean:

```bash
cd hosting
pulumi up
```

**👉 [Infrastructure Setup Guide](hosting/README.md)**

### 2. Backend Service
Start the AI image generation API:

```bash
cd backend
export FREEPIK_API_KEY=your_key
export GOOGLE_API_KEY=your_key
export LEONARDO_API_KEY=your_key
go run cmd/server/main.go
```

**👉 [Backend Documentation](backend/README.md)**

### 3. Frontend Application
Launch the image comparison interface:

```bash
cd frontend
npm install
npm run dev
```

**👉 [Frontend Documentation](frontend/README.md)**

## 🤖 AI Provider Research

Explore individual AI image generation providers and their capabilities:

| Provider | Status | Features | Documentation |
|----------|--------|----------|---------------|
| **Freepik** | ✅ Integrated | Official API, $5 free credit, sync generation | [📖 Guide](research/freepik/README.md) |
| **Google Imagen** | ✅ Integrated | High-quality Imagen 3.0, Vertex AI | [📖 Guide](research/google-imagen/README.md) |
| **Leonardo AI** | ✅ Integrated | Creative models, async generation, free tier | [📖 Guide](research/leonardo-ai/README.md) |
| **Craiyon** | ❌ Broken | Cloudflare protection blocks API access | [📖 Guide](research/craiyon/README.md) |

## 🎯 Key Features

### Intelligent Provider Management
- **Google ADK Integration**: Orchestrator agent with automatic provider selection
- **Reactive Fallback**: Seamlessly switches providers when quotas are hit
- **Cost Optimization**: Prioritizes free tiers and manages API costs intelligently

### Production-Ready Architecture
- **Load Balancing**: Distributes traffic across multiple backend instances
- **CDN Integration**: Global content delivery via Digital Ocean Spaces
- **Health Monitoring**: Comprehensive health checks and status endpoints
- **Infrastructure as Code**: Complete Pulumi deployment automation

### Modern Frontend Experience
- **Mobile-First**: Touch-optimized image comparison with swipe gestures
- **Real-time Animations**: Framer Motion with spring physics
- **Progressive Web App**: Responsive design across all devices

## 🔗 Documentation Links

- **[🏗️ Infrastructure & Hosting](hosting/README.md)** - Digital Ocean deployment with Pulumi
- **[⚙️ Backend API Service](backend/README.md)** - Go server with ADK orchestrator
- **[🎨 Frontend Application](frontend/README.md)** - Next.js image comparison interface
- **[🔬 AI Provider Research](research/)** - Individual provider guides and examples

## 💰 Cost Estimation

**Digital Ocean Monthly Costs (~$44/month):**
- Load Balancer: $12/month
- 2 Droplets (1vCPU/1GB): $12/month
- Spaces Storage + CDN: $5/month
- Valkey Database: $15/month

**AI Generation Costs:**
- Leverages free tiers across multiple providers
- Automatic cost optimization via intelligent provider selection

## 🚀 Deployment Options

### GitHub Actions (Recommended)
Automated deployment via GitHub Actions with infrastructure provisioning.

### Local Development
Manual deployment for development and testing environments.

**👉 [Complete Deployment Guide](hosting/README.md)**

---

**Built for the Pluralsight Cloud Portfolio Challenge** | **Powered by Digital Ocean** | **Enhanced with Google ADK**
