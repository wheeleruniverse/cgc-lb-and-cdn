# Cloudflare Migration Guide

This guide explains how to migrate the CGC Load Balancing and CDN project from DigitalOcean to Cloudflare's platform to reduce hosting costs from ~$68/month to $0-5/month while maintaining full functionality.

## Table of Contents

- [Why Migrate to Cloudflare?](#why-migrate-to-cloudflare)
- [Cloudflare Free Tier Limits](#cloudflare-free-tier-limits)
- [Cost Comparison](#cost-comparison)
- [Migration Options](#migration-options)
- [Will Free Tier Be Enough?](#will-free-tier-be-enough)
- [Recommended Migration Path](#recommended-migration-path)
- [Migration Effort Estimates](#migration-effort-estimates)

## Why Migrate to Cloudflare?

After DigitalOcean credits expire, maintaining the current infrastructure costs ~$68/month. Cloudflare's platform offers:

- **Generous free tier** that covers typical portfolio project traffic
- **No egress/bandwidth charges** (unlimited data transfer)
- **Global edge network** for better performance
- **Serverless architecture** that scales automatically
- **$0-5/month** total cost vs $68/month current

## Cloudflare Free Tier Limits

### Cloudflare Pages (Frontend Hosting)
- ✅ **500 builds/month** (more than enough for a portfolio project)
- ✅ **Unlimited sites**
- ✅ **Unlimited requests**
- ✅ **Unlimited bandwidth** (no egress fees!)
- ✅ **100 custom domains**
- ✅ **Free SSL certificates**
- ✅ **Preview deployments** (Git integration)

### Cloudflare Workers (Backend API)
- ✅ **100,000 requests/day** (~3M/month)
- ✅ **Unlimited bandwidth** (no data transfer costs)
- ✅ **10ms CPU time per request**
- ✅ **128MB memory per request**
- ⚠️ Resets daily at 00:00 UTC

### Cloudflare R2 (Object Storage)
Replaces DigitalOcean Spaces for image storage:
- ✅ **10GB storage/month**
- ✅ **1M Class A operations/month** (writes, lists)
- ✅ **10M Class B operations/month** (reads)
- ✅ **Unlimited egress bandwidth** (huge savings vs S3/Spaces!)

### Cloudflare KV (Key-Value Store)
Replaces Valkey for vote storage and caching:
- ✅ **1GB storage**
- ✅ **100,000 reads/day**
- ✅ **1,000 writes/day**
- ✅ **1,000 deletes/day**
- ✅ **1,000 lists/day**

### Cloudflare D1 (SQLite Database)
Optional - for more complex querying:
- ✅ **1GB storage**
- ✅ **5M rows read/day**
- ✅ **100,000 rows written/day**

## Cost Comparison

### Current DigitalOcean Infrastructure (~$68/month)
- Load Balancer: $12/month
- 2 Droplets (s-2vcpu-2gb): $36/month
- Spaces Storage + CDN: $5/month
- Valkey Database: $15/month
- **Total: $68/month**

### Cloudflare Infrastructure ($0-5/month)
- Pages (Frontend): **$0** (free tier)
- Workers (Backend): **$0** (free tier)
- R2 (Images): **$0** (free tier covers up to 10GB)
- KV (Votes/Cache): **$0** (free tier)
- **Total: $0/month** (or $5 if keeping DO Spaces temporarily)

### Annual Savings
- **$816/year** if fully migrated
- **$756/year** if keeping DO Spaces

## Migration Options

### Option 1: Static Site Only (~4-6 hours effort)

**Architecture:**
- Deploy static Next.js site to Cloudflare Pages (free)
- Serve images directly from DO Spaces or R2
- Remove backend API entirely
- Client-side voting (localStorage only, no persistence)

**Cost:** $0-5/month

**Trade-offs:**
- No persistent voting data
- No new image generation
- Pure read-only gallery experience
- Minimal migration effort

### Option 2: Full Migration with Workers Backend (~8-12 hours effort)

**Architecture:**
- Cloudflare Pages for frontend (free)
- Cloudflare Workers for backend API (free)
- Cloudflare KV for vote storage (free)
- Keep DO Spaces ($5/month) or migrate to R2 (free)

**Cost:** $0-5/month

**Trade-offs:**
- Maintains full voting functionality
- No new image generation (display existing pairs only)
- Excellent global performance
- Moderate migration effort

### Option 3: Single DigitalOcean Droplet (~$11-17/month)

**Architecture:**
- Single s-1vcpu-1gb or s-1vcpu-2gb droplet
- No load balancer
- No Valkey (use in-memory cache or SQLite)
- Keep DO Spaces for images

**Cost:** $11-17/month

**Trade-offs:**
- No high availability
- Limited traffic capacity
- Stays within DO ecosystem
- Minimal migration effort

## Will Free Tier Be Enough?

### Daily Request Estimates

For a typical portfolio project:
- **Page views:** 100-1,000/day
- **API calls per page:** 3-5 (fetch pairs, submit vote, get stats)
- **Total API requests:** 300-5,000/day

**Workers Free Tier:** 100,000 requests/day
**Verdict:** ✅ **Way under the limit** (20-330x headroom)

### Storage Requirements

Based on current setup:
- Generated images: ~2-5MB per pair
- 100 pairs: ~500MB
- 500 pairs: ~2.5GB
- 1,000 pairs: ~5GB

**R2 Free Tier:** 10GB
**Verdict:** ✅ **Sufficient for current content** (up to ~1,500 pairs)

### KV Storage Requirements

- Vote data: ~100 bytes per vote
- 10,000 votes: ~1MB
- Metadata: ~10-50MB

**KV Free Tier:** 1GB
**Verdict:** ✅ **More than enough**

### KV Operations

If you get **1,000 page views/day** with voting:
- **Reads:** ~3,000/day (3 reads per page)
- **Writes:** ~500/day (50% vote rate)

**KV Free Tier:**
- Reads: 100,000/day
- Writes: 1,000/day

**Verdict:** ✅ **Well within limits**

### When You'd Exceed Free Tier

You'd only exceed limits if you get:
- **100k+ requests/day** (33M+ page views/month)
- **1k+ votes/day** (1,000+ votes daily)
- **10GB+ images** (2,000+ image pairs)

**At that scale** (which would be amazing!), you'd pay:
- **Workers Paid:** $5/month base + $0.50 per million additional requests
- **KV:** $0.50/month per additional 1GB
- **R2:** $0.015/GB/month over 10GB

**Even with 10x traffic, total cost: ~$5-10/month**

## Recommended Migration Path

**Medium Effort Migration** - Full functionality with Workers backend

### Phase 1: Preparation (Before DO credits expire)
1. **Export existing data:**
   - List all image pairs from DO Spaces
   - Export vote data from Valkey
   - Document current API endpoints

2. **Set up Cloudflare account:**
   - Create Cloudflare account (free)
   - Add your domain to Cloudflare
   - Install Wrangler CLI: `npm install -g wrangler`

3. **Create migration branch:**
   ```bash
   git checkout -b cloudflare-migration
   ```

### Phase 2: Backend Migration (4-6 hours)

1. **Create Workers project structure:**
   ```
   workers/
   ├── src/
   │   ├── index.ts          # Main Worker entry point
   │   ├── handlers/
   │   │   ├── generate.ts   # Image generation endpoint (optional)
   │   │   ├── vote.ts       # Voting endpoint
   │   │   ├── pairs.ts      # Fetch pairs endpoint
   │   │   └── stats.ts      # Statistics endpoint
   │   ├── storage/
   │   │   └── kv.ts         # KV storage abstraction
   │   └── types.ts          # TypeScript types
   ├── wrangler.toml         # Worker configuration
   └── package.json
   ```

2. **Port Go handlers to TypeScript:**
   - Convert existing Go API handlers to TypeScript
   - Replace Valkey calls with KV operations
   - Maintain same API contract (backward compatible)

3. **Set up KV namespaces:**
   ```bash
   wrangler kv:namespace create VOTES
   wrangler kv:namespace create PAIRS
   ```

4. **Deploy Workers:**
   ```bash
   wrangler deploy
   ```

### Phase 3: Frontend Migration (2 hours)

1. **Update API endpoints:**
   - Change backend URL to Workers endpoint
   - No other code changes needed (API contract stays the same)

2. **Deploy to Cloudflare Pages:**
   - Connect GitHub repository to Cloudflare Pages
   - Configure build settings:
     - Build command: `npm run build`
     - Build output directory: `.next`
     - Framework preset: Next.js

3. **Test deployment:**
   - Verify all functionality works
   - Check voting persists correctly
   - Ensure images load from storage

### Phase 4: Image Migration (2-3 hours)

**Option A: Keep DO Spaces temporarily ($5/month)**
- Update CORS settings to allow Cloudflare domains
- No migration needed
- Pay $5/month until ready to migrate

**Option B: Migrate to R2 (free)**
1. **Create R2 bucket:**
   ```bash
   wrangler r2 bucket create cgc-images
   ```

2. **Copy images from DO Spaces to R2:**
   ```bash
   # Using rclone
   rclone sync digitalocean:cgc-lb-and-cdn-content/images cloudflare:cgc-images/images
   ```

3. **Update image URLs in code:**
   - Replace DO Spaces URLs with R2 URLs
   - Update frontend environment variables

### Phase 5: Data Migration (1 hour)

1. **Export Valkey data:**
   ```bash
   # Connect to Valkey and export all pairs
   redis-cli -h <host> -p <port> -a <password> --tls --scan --pattern "pair:*" | \
     xargs redis-cli -h <host> -p <port> -a <password> --tls MGET
   ```

2. **Import to Cloudflare KV:**
   ```typescript
   // Script to bulk import pairs to KV
   // Run using wrangler or Workers API
   ```

3. **Verify data integrity:**
   - Check pair count matches
   - Verify vote counts migrated correctly

### Phase 6: DNS Cutover (1 hour)

1. **Update DNS to point to Cloudflare Pages:**
   - Add CNAME record pointing to Pages deployment
   - Wait for DNS propagation (5-30 minutes)

2. **Test production site:**
   - Verify all functionality works
   - Check SSL certificate
   - Test voting and image loading

3. **Monitor for issues:**
   - Check Workers logs
   - Monitor KV operations
   - Verify no errors in browser console

### Phase 7: Cleanup (Optional)

Once confirmed working:
1. Destroy DO infrastructure: `pulumi destroy`
2. Cancel DO Spaces (if migrated to R2)
3. Remove old deployment workflows

## Migration Effort Estimates

### Low Effort - Static Site Only (~4-6 hours)
**Code changes:** Minimal (~50 lines)
- Modify frontend to fetch images directly
- Remove backend API calls
- Add static export configuration
- Deploy to Pages

**Result:** Read-only gallery, no voting persistence

### Medium Effort - Full Migration (~8-12 hours)
**Code changes:** Medium (~200-300 lines)
- **Frontend (2 hours):** Update API endpoints
- **Backend (4-6 hours):** Port Go to TypeScript Workers
- **Storage (2-3 hours):** Migrate images (optional)
- **Deployment (1 hour):** Configure CI/CD

**Result:** Full functionality with persistence, $0-5/month

### High Effort - With Image Generation (~16-24 hours)
**Additional work:**
- Implement on-demand generation via Workers
- Handle async generation (30s timeout on free tier)
- Queue system for long-running tasks
- Rate limiting and quota management

**This is probably overkill** for a "lite" version after credits expire.

## Next Steps

1. **Review this guide** and decide on migration approach
2. **Create a migration timeline** based on when DO credits expire
3. **Set up Cloudflare account** and install Wrangler CLI
4. **Start with Phase 1** (preparation) while DO infrastructure is still running
5. **Test migration in parallel** before cutting over DNS

## Additional Resources

- [Cloudflare Workers Documentation](https://developers.cloudflare.com/workers/)
- [Cloudflare Pages Documentation](https://developers.cloudflare.com/pages/)
- [Cloudflare KV Documentation](https://developers.cloudflare.com/kv/)
- [Cloudflare R2 Documentation](https://developers.cloudflare.com/r2/)
- [Wrangler CLI Documentation](https://developers.cloudflare.com/workers/wrangler/)

## Questions or Issues?

If you encounter issues during migration:
1. Check Cloudflare Developer Docs
2. Review Workers logs in Cloudflare Dashboard
3. Test locally using `wrangler dev`
4. Verify KV bindings in `wrangler.toml`

---

**Estimated Total Cost After Migration:** $0-5/month (vs $68/month current)
**Estimated Migration Time:** 8-12 hours for full functionality
**Recommended Approach:** Medium Effort - Full Migration with Workers Backend
