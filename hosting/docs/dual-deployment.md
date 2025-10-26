# Dual-Deployment Setup Complete! 🎉

Your project now supports **two independent deployment architectures** that can run simultaneously:

1. **Full Deployment** (Digital Ocean) - "Pricy" - ~$68/month
2. **Lite Deployment** (GitHub Pages) - "Cheap" - $0/month

## 📦 What Was Changed

### 1. Frontend Feature Flags System

**New Files**:
- `frontend/.env.full` - Environment variables for full deployment
- `frontend/.env.lite` - Environment variables for lite deployment
- `frontend/app/config.ts` - Configuration system with feature flags
- `frontend/app/services/dataService.ts` - Data service layer abstracting API vs static data

**Modified Files**:
- `frontend/package.json` - Added `build:full`, `build:lite`, `dev:full`, `dev:lite` scripts
- `frontend/next.config.js` - Dynamic configuration based on deployment mode
- `frontend/app/components/ImageBattle.tsx` - Uses data service and feature flags
- `frontend/app/components/WinnersGrid.tsx` - Uses data service and feature flags

**Static Data**:
- `frontend/public/static-data/image-pairs.json` - Pre-generated image pairs for lite mode
- `frontend/public/static-data/winners.json` - Winners placeholder for lite mode
- `frontend/public/static-data/README.md` - Documentation for static data

### 2. GitHub Actions Workflows

**Renamed Workflows** (from 3 to 6 total):
- `deploy.yml` → `pricy-deploy.yml` (Full deployment to DO)
- `teardown.yml` → `pricy-teardown.yml` (Teardown DO resources)
- `refresh.yml` → `pulumi-refresh.yml` (Sync Pulumi state)

**New Workflows**:
- `cheap-deploy.yml` - Deploy static site to GitHub Pages
- `cheap-teardown.yml` - Disable GitHub Pages
- `dns-cutover.yml` - Switch DNS between deployments

**Documentation**:
- `.github/workflows/WORKFLOWS.md` - Comprehensive workflow documentation

### 3. Documentation Updates

**Updated**:
- `README.md` - Added dual-deployment architecture section, workflow tables, deployment flows

## 🚀 Quick Start Guide

### Step 1: Enable GitHub Pages

1. Go to **Settings** > **Pages** in your repository
2. Under **Build and deployment**:
   - Source: **GitHub Actions**
3. Click **Save**

### Step 2: Deploy Both Versions

Run these workflows from the **Actions** tab:

```
1. Actions → pricy-deploy.yml → Run workflow
   - droplet_count: 2
   → Full deployment will be live at DO Load Balancer IP in ~10 minutes

2. Actions → cheap-deploy.yml → Run workflow
   - base_path: (leave empty)
   → Lite deployment will be live at GitHub Pages URL in ~5 minutes
```

### Step 3: Test Both Deployments

**Full Deployment**:
- Access via DO Load Balancer IP (check workflow output)
- Should show "Full (Digital Ocean)" mode indicator
- Has AI generation, live voting, cross-session tracking

**Lite Deployment**:
- Access via GitHub Pages URL (check workflow output)
- Should show "Lite Mode" yellow banner
- Has local voting only, pre-generated images

### Step 4: DNS Cutover (Optional)

Switch your domain to either deployment:

```
Actions → dns-cutover.yml → Run workflow
- target: github-pages (or digital-ocean)
- confirm: SWITCH
```

Wait 1-2 hours for DNS propagation, then visit your domain!

## 💡 Feature Flags in Action

The same codebase adapts based on `NEXT_PUBLIC_DEPLOYMENT_MODE`:

### Full Mode Features:
- ✅ API calls to backend
- ✅ AI image generation
- ✅ Cross-session vote tracking
- ✅ Live statistics
- ✅ Valkey database integration

### Lite Mode Features:
- ✅ Static site (no backend)
- ✅ Local browser voting
- ✅ Pre-generated images from DO Spaces
- ✅ Same UX, gracefully degraded
- ❌ No AI generation
- ❌ No cross-session tracking

## 🎯 Recommended Usage Strategy

### For Daily Use (Cost Savings):
1. Keep **lite deployment** active on GitHub Pages
2. Point domain to GitHub Pages
3. Costs: **$5/month** (DO Spaces + CDN only) 🎉

### For Demos/Interviews:
1. Run `pricy-deploy.yml` (~10 min)
2. Run `dns-cutover.yml` to Digital Ocean
3. Wait 5-10 minutes for DNS
4. Demo full AI features! 🎨
5. Run `dns-cutover.yml` back to GitHub Pages
6. Run `pricy-teardown.yml` to save costs

**Cost**: $5/month baseline + ~$2-3 per demo day (prorated hourly billing)

**What Gets Preserved**:
- ✅ DO Spaces bucket (images)
- ✅ CDN configuration
- ✅ All generated images

### For Development:
```bash
# Test full mode locally (requires backend)
cd frontend
npm run dev:full

# Test lite mode locally (no backend needed)
cd frontend
npm run dev:lite
```

## 📊 Architecture Comparison

| Feature | Full (DO) | Lite (GH Pages) |
|---------|-----------|-----------------|
| **Platform** | Digital Ocean | GitHub Pages |
| **Backend** | Go API + Valkey | None |
| **Frontend** | Next.js SSR | Next.js Static Export |
| **Voting** | Cross-session (API) | Local (localStorage) |
| **Images** | AI-generated + DO Spaces | Pre-generated from DO Spaces |
| **Cost** | ~$68/month | $0/month |
| **Scalability** | Horizontal (add droplets) | GitHub's CDN |
| **AI Generation** | ✅ Yes | ❌ No |

## 🔐 Static Data (Automatic!)

The lite deployment **automatically fetches real images** from DO Spaces at build time:

1. **Build Process**: `cheap-deploy.yml` runs `.github/scripts/generate-static-data.js`
2. **Fetch Images**: Script lists all images in DO Spaces bucket
3. **Read Metadata**: Fetches metadata from response headers (`X-Amz-Meta-*`)
4. **Generate JSON**: Creates `image-pairs.json` with all discovered pairs
5. **No Manual Update**: Always fresh data from DO Spaces!

### How Images Are Stored

Images must follow this structure in DO Spaces:
```
images/
  ├── freepik/
  │   └── {pair-id}/
  │       ├── left.png    (with metadata headers)
  │       └── right.png   (with metadata headers)
  └── google/
      └── {pair-id}/
          ├── left.png
          └── right.png
```

Metadata headers (set by backend):
- `X-Amz-Meta-Pair-Id`: Unique pair identifier
- `X-Amz-Meta-Provider`: AI provider name
- `X-Amz-Meta-Prompt`: Image generation prompt
- `X-Amz-Meta-Side`: left or right

## 🎓 For Recruiters & Interviews

This setup demonstrates several **production-grade practices**:

### 1. Feature Flags
- Build-time configuration
- Single codebase, multiple deployments
- Graceful degradation

### 2. Blue/Green Deployment
- Both versions live simultaneously
- Independent testing before cutover
- Zero-downtime DNS switching

### 3. Cost Optimization
- Free tier for daily use
- Full features on-demand
- Prorated resource usage

### 4. Infrastructure as Code
- Pulumi for DO deployment
- GitHub Actions for automation
- Reproducible deployments

### 5. Service Layer Abstraction
- Clean separation: API vs static data
- Testable, maintainable code
- Easy to swap implementations

## 📚 Documentation

- **[Workflow Guide](.github/workflows/WORKFLOWS.md)** - Detailed workflow documentation
- **[Main README](README.md)** - Project overview
- **[Infrastructure Guide](hosting/README.md)** - DO deployment details
- **[Static Data Guide](frontend/public/static-data/README.md)** - Lite mode data structure

## 🚨 Important Notes

### GitHub Pages Setup
- Must enable in repository Settings > Pages
- Source must be "GitHub Actions"
- Custom domain configuration is optional

### DNS Management
- Both deployments can run simultaneously
- DNS cutover is manual workflow
- Propagation takes 1-48 hours (usually 1-2 hours)

### Cost Awareness
- DO charges hourly (prorated)
- Tear down when not using to save money
- GH Pages is completely free

### Static Data
- Lite mode uses placeholder data by default
- Update `image-pairs.json` with real images
- Images still served from DO Spaces CDN

## 🎉 What This Enables

1. **Portfolio Showcase**: Free hosting for your portfolio
2. **Demo-Ready**: Spin up full features in 10 minutes
3. **Cost-Effective**: $0/month normally, ~$2-3 per demo
4. **Professional**: Shows sophisticated deployment practices
5. **Recruiter-Friendly**: Clear demonstration of DevOps skills

## 🔧 Troubleshooting

### Build Fails
- Check Node.js version (requires 18+)
- Verify all dependencies installed
- Check environment variables

### DNS Not Working
- Wait 1-2 hours for propagation
- Use https://dnschecker.org to verify
- Check A records in DO console

### GitHub Pages 404
- Verify Pages is enabled in Settings
- Check workflow ran successfully
- Verify base_path setting

### Feature Flag Issues
- Check .env.local file exists
- Verify NEXT_PUBLIC_* variables
- Check console logs for config output

## 🎓 Next Steps

1. ✅ Enable GitHub Pages in repository settings
2. ✅ Run both deployment workflows
3. ✅ Test both deployments independently
4. ✅ Update static data with real images
5. ✅ Configure DNS cutover for your domain
6. ✅ Practice demo flow (deploy → cutover → demo → teardown)

## 💬 Questions?

- Check [WORKFLOWS.md](.github/workflows/WORKFLOWS.md) for detailed scenarios
- Review workflow logs in Actions tab
- Test locally with `npm run dev:full` or `npm run dev:lite`

---

**Congratulations! Your dual-deployment architecture is ready to impress recruiters and save you money!** 🚀
