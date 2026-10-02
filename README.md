# Cloud Portfolio Challenge Load Balancing and CDN - AI Image Service

> ## 🗄️ Archived — infrastructure decommissioned (October 2026)
>
> This was built for the **Pluralsight Cloud Portfolio Challenge** as a time-boxed demonstration. The live deployment was never meant to be permanent, and it has now been fully torn down.
>
> **Nothing here is running.** All Digital Ocean resources have been destroyed — load balancer, droplets, Valkey cluster, Spaces bucket, and CDN. The custom domain has lapsed, the GitHub Pages site is unpublished, and all deployment credentials have been revoked. This repository is archived and read-only.
>
> **Everything below describes the system as it ran.** The architecture, infrastructure code, and provider integrations are intact and are the point of this repo. But the URLs are dead, the cost figures are historical, and the deployment workflows will not run — their secrets are gone and the resources they target no longer exist. Treat the setup instructions as a record of how it worked, not as something to follow.
>
> **The generated artifacts were preserved offline before teardown:** 2,656 AI image pairs (5,310 objects, 5 GiB) across three providers, along with the 491 generation prompts that existed only as Spaces object metadata. Both were checksum-verified against the bucket as it was deleted.

A **Cloud Portfolio Challenge** implementation showcasing modern cloud architecture with load balancing, CDN, and intelligent AI image generation using Google's Agent Development Kit (ADK).

## 🏗️ Architecture Overview

This project showcases **two independent deployment architectures** that can run simultaneously:

### 🚀 Full Deployment (Digital Ocean) - "Pricy"
Production-grade infrastructure with complete AI capabilities:
- **Load Balancer**: Digital Ocean Load Balancer distributing traffic across multiple full-stack droplets
- **Full-Stack Droplets**: Each droplet runs Go API server, Next.js frontend, and Nginx reverse proxy
- **Automated Deployment**: UserData script handles complete application setup with dedicated service user
- **CDN**: Digital Ocean Spaces CDN for fast global image delivery and log storage
- **Multi-Provider Intelligence**: ADK-powered orchestrator managing multiple AI image providers
- **Managed Database**: Valkey (Redis-compatible, VPC-only) for caching and session management
- **Cost**: ~$68/month

### 🪶 Lite Deployment (GitHub Pages) - "Cheap"
Cost-optimized static deployment with feature flags:
- **Platform**: GitHub Pages (completely free)
- **Frontend**: Static Next.js export with pre-generated images
- **Feature Flags**: Demonstrates production-grade deployment practices
- **Voting**: Local browser-only tracking (no backend)
- **Images**: Served from DO Spaces CDN
- **Cost**: $5/month (DO Spaces + CDN, GitHub Pages is free)

### 🔄 Blue/Green Deployment Strategy
Both deployments can run independently for testing:
- Full deployment on DO Load Balancer IP
- Lite deployment on GitHub Pages URL
- DNS cutover workflow for seamless switching
- Zero-downtime migration between deployments

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
- **Load Balancing**: Distributes traffic across multiple full-stack droplet instances
- **CDN Integration**: Global content delivery via Digital Ocean Spaces
- **Automated Deployment**: UserData script handles complete setup (apps, nginx, services, monitoring)
- **Log Management**: Hourly compressed log uploads to Spaces for centralized monitoring
- **Health Monitoring**: Comprehensive health checks and status endpoints
- **Infrastructure as Code**: Complete Pulumi deployment automation with configurable scaling
- **Security Hardening**: Services run as dedicated non-root user with VPC-isolated database

### Modern Frontend Experience
- **Mobile-First**: Touch-optimized image comparison with swipe gestures
- **Real-time Animations**: Framer Motion with spring physics
- **Progressive Web App**: Responsive design across all devices

### Production-Grade Feature Flags
- **Dual Deployment Support**: Single codebase for both full and lite modes
- **Build-Time Configuration**: Environment-based feature toggles
- **Graceful Degradation**: Lite mode maintains UX without backend
- **Service Layer Abstraction**: Clean separation between API and static data
- **Demonstrable Skill**: Shows sophisticated deployment practices for recruiters

## 🔗 Documentation Links

### Core Documentation
- **[🏗️ Infrastructure & Hosting](hosting/README.md)** - Digital Ocean deployment with Pulumi
- **[⚙️ Backend API Service](backend/README.md)** - Go server with ADK orchestrator
- **[🎨 Frontend Application](frontend/README.md)** - Next.js image comparison interface
- **[🔬 AI Provider Research](research/)** - Individual provider guides and examples

### Deployment Guides
- **[🚀 Dual Deployment Setup](hosting/docs/dual-deployment.md)** - Full quick start guide
- **[📋 Deployment Workflows](hosting/docs/deployment-workflows.md)** - GitHub Actions workflows
- **[🪣 Spaces Integration](hosting/docs/spaces-integration.md)** - DO Spaces and real images
- **[🛡️ Spaces Preservation](hosting/docs/spaces-preservation.md)** - How teardown protects images
- **[🔐 GitHub Secrets](hosting/docs/github-secrets.md)** - Required secrets setup

## 💰 Cost Estimation

### Full Deployment (Digital Ocean)
**Monthly Costs (~$68/month with 2 droplets):**
- Load Balancer: $12/month
- 2 Droplets (s-2vcpu-2gb): $36/month ($18 each)
- Spaces Storage + CDN: $5/month (includes images and logs)
- Valkey Database (VPC-only): $15/month

**Scaling**: Add $18/month per additional droplet. Horizontal scaling is configured via `droplet_count` parameter.

**Note**: The s-2vcpu-2gb droplet configuration is required to handle full-stack deployment (backend Go API, frontend Next.js, nginx reverse proxy, and UserData bootstrapping). Smaller instances were insufficient during initial testing.

### Lite Deployment (GitHub Pages)
**Monthly Costs ($5/month):**
- GitHub Pages: $0/month (completely free)
- Spaces Storage + CDN: $5/month (shared with full deployment)

**Note**: DO Spaces is preserved during teardown to maintain images for lite deployment.

### Cost Optimization Strategy
**Recommended**: Keep lite deployment active ($5/month), spin up full deployment only for demos (~$2-3 per demo day with prorated billing).

**Annual Savings**: ~$756/year compared to continuous full deployment!

**AI Generation Costs:**
- Leverages free tiers across multiple providers
- Automatic cost optimization via intelligent provider selection

## 🚀 Deployment Options

This project uses **GitHub Actions workflows** for automated deployment and management of both architectures.

### Full Deployment Workflows (Digital Ocean)

| Workflow | Purpose | Cost Impact |
|----------|---------|-------------|
| **pricy-deploy.yml** | Deploy full infrastructure to Digital Ocean | +$68/month |
| **pricy-teardown.yml** | Destroy all DO resources | -$68/month |
| **pulumi-refresh.yml** | Sync Pulumi state with actual DO resources | $0 |

### Lite Deployment Workflows (GitHub Pages)

| Workflow | Purpose | Cost Impact |
|----------|---------|-------------|
| **cheap-deploy.yml** | Build and deploy static site to GitHub Pages | $0 |
| **cheap-teardown.yml** | Disable GitHub Pages deployment | $0 |

### DNS Management Workflow

| Workflow | Purpose | Description |
|----------|---------|-------------|
| **dns-cutover.yml** | Switch DNS between deployments | Seamlessly move domain between DO and GitHub Pages |

### 🎯 Typical Deployment Flow

**Initial Setup (Testing Phase):**
1. Run `pricy-deploy.yml` → Full deployment live at DO Load Balancer IP
2. Run `cheap-deploy.yml` → Lite deployment live at GitHub Pages URL
3. Test both deployments independently
4. Run `dns-cutover.yml` → Switch domain to preferred deployment

**Production (Cost Savings):**
1. Run `dns-cutover.yml` (target: github-pages) → Switch to free deployment
2. Run `pricy-teardown.yml` → Save $68/month
3. Keep lite deployment active on GitHub Pages

**Feature Showcase (Demos/Interviews):**
1. Run `pricy-deploy.yml` → Bring up full deployment
2. Run `dns-cutover.yml` (target: digital-ocean) → Show full features
3. Demo complete AI generation capabilities
4. Run `dns-cutover.yml` (target: github-pages) → Switch back to free
5. Run `pricy-teardown.yml` → Save costs again

**👉 [Deployment Workflows Guide](hosting/docs/deployment-workflows.md)**
**👉 [Dual Deployment Setup](hosting/docs/dual-deployment.md)**
**👉 [Infrastructure Setup Guide](hosting/README.md)**

### Local Development

For local development and testing:

```bash
# Full mode (requires backend)
cd frontend
npm run dev:full

# Lite mode (static only)
cd frontend
npm run dev:lite
```

---

**Built for the Pluralsight Cloud Portfolio Challenge** | **Powered by Digital Ocean** | **Enhanced with Google ADK**
