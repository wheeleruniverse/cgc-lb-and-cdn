# Documentation Index

This directory contains all deployment and infrastructure documentation for the CGC Load Balancing and CDN project.

## 📚 Documentation Files

### Quick Start
- **[dual-deployment.md](dual-deployment.md)** - Complete dual-deployment setup guide
  - Overview of full vs lite deployments
  - Step-by-step setup instructions
  - Testing and verification steps
  - Cost optimization strategies

### Deployment Workflows
- **[deployment-workflows.md](deployment-workflows.md)** - GitHub Actions workflows reference
  - All 6 workflows explained in detail
  - Common deployment scenarios
  - Cost comparisons
  - Troubleshooting guide

### Infrastructure Details
- **[github-secrets.md](github-secrets.md)** - Required GitHub secrets setup
  - Digital Ocean credentials
  - Pulumi access token
  - API keys configuration

### DO Spaces Integration
- **[spaces-integration.md](spaces-integration.md)** - DO Spaces integration details
  - How lite deployment fetches real images
  - Build-time data generation
  - Metadata structure
  - Testing locally

- **[spaces-preservation.md](spaces-preservation.md)** - How teardown protects images
  - What gets preserved vs destroyed
  - Cost impact ($5/month vs $68/month)
  - Safety features
  - Manual deletion (if needed)

### Legacy/Reference
- **[cloudflare-migration.md](cloudflare-migration.md)** - Cloudflare migration notes (reference only)

## 🎯 Reading Path by Role

### For First-Time Setup
1. Start with [dual-deployment.md](dual-deployment.md)
2. Set up secrets: [github-secrets.md](github-secrets.md)
3. Review workflows: [deployment-workflows.md](deployment-workflows.md)
4. Deploy and test

### For Developers
1. [dual-deployment.md](dual-deployment.md) - Understand architecture
2. [spaces-integration.md](spaces-integration.md) - How data flows
3. [deployment-workflows.md](deployment-workflows.md) - CI/CD process

### For Cost Optimization
1. [spaces-preservation.md](spaces-preservation.md) - What stays, what goes
2. [deployment-workflows.md](deployment-workflows.md) - Scenario 2 & 3
3. [dual-deployment.md](dual-deployment.md) - Cost strategy section

### For Troubleshooting
1. [deployment-workflows.md](deployment-workflows.md) - Troubleshooting section
2. [spaces-integration.md](spaces-integration.md) - Data generation issues
3. Check workflow logs in GitHub Actions

## 🔗 Related Documentation

### In Other Directories
- **[../../README.md](../../README.md)** - Project overview and architecture
- **[../README.md](../README.md)** - Infrastructure (Pulumi) details
- **[../../frontend/README.md](../../frontend/README.md)** - Frontend application
- **[../../backend/README.md](../../backend/README.md)** - Backend API service
- **[../../frontend/public/static-data/README.md](../../frontend/public/static-data/README.md)** - Static data structure

## 📋 Documentation Standards

All documentation in this directory follows these conventions:

### File Naming
- **lowercase-kebab-case** (e.g., `dual-deployment.md`)
- Descriptive names that indicate content
- No abbreviations unless commonly understood

### Structure
- Clear section headings with emoji prefixes
- Code blocks with syntax highlighting
- Links to related documentation
- Examples and scenarios

### Maintenance
- Keep cost figures up to date
- Update when workflows change
- Add troubleshooting notes from real issues
- Cross-link related documents

## 🔄 Document Relationships

```
dual-deployment.md (main guide)
    ↓
    ├── deployment-workflows.md (detailed workflows)
    │   ├── github-secrets.md (setup)
    │   └── spaces-preservation.md (cost details)
    └── spaces-integration.md (technical details)
```

## 📊 Quick Reference

| Task | Documentation |
|------|---------------|
| **First deployment** | [dual-deployment.md](dual-deployment.md) |
| **Run a workflow** | [deployment-workflows.md](deployment-workflows.md) |
| **Save costs** | [spaces-preservation.md](spaces-preservation.md) |
| **Understand images** | [spaces-integration.md](spaces-integration.md) |
| **Setup secrets** | [github-secrets.md](github-secrets.md) |
| **Troubleshoot** | [deployment-workflows.md](deployment-workflows.md) → Troubleshooting |

## 🆘 Getting Help

1. **Search these docs** using your editor's search
2. **Check workflow logs** in GitHub Actions tab
3. **Review error messages** in deployment summaries
4. **Open an issue** if documentation is unclear

---

**All documentation organized and standardized!** 📚✨
