# GitHub Actions Workflows Documentation

This project uses **6 GitHub Actions workflows** to manage dual deployments: a full-featured Digital Ocean deployment and a lite GitHub Pages deployment.

## 📋 Workflow Overview

### 🚀 Full Deployment (Digital Ocean) - "Pricy"

#### 1. pricy-deploy.yml
**Purpose**: Deploy complete infrastructure to Digital Ocean

**Triggers**: Manual (workflow_dispatch)

**Inputs**:
- `droplet_count` (optional): Number of droplets (2-10), default: 2
- `recreate_valkey` (optional): Rebuild Valkey from DO Spaces backup, default: false

**What it does**:
1. Builds Pulumi infrastructure program (Go)
2. Deploys infrastructure:
   - Load Balancer
   - N Droplets (configurable)
   - Valkey database cluster (VPC-only)
   - DO Spaces bucket with CDN
   - VPC and firewall rules
3. Configures API keys and secrets
4. Outputs deployment information (IPs, endpoints)

**Cost Impact**: +$68/month (~$18 per droplet + $12 LB + $15 Valkey + $5 Spaces)

**Usage**:
```
Actions → pricy-deploy.yml → Run workflow
- droplet_count: 2
- recreate_valkey: false
```

**Duration**: ~8-12 minutes

---

#### 2. pricy-teardown.yml
**Purpose**: Destroy all Digital Ocean resources

**Triggers**: Manual (workflow_dispatch)

**Inputs**:
- `confirm`: Must type "DESTROY" to confirm

**What it does**:
1. Cleans up auto-created DNS records (Let's Encrypt validation)
2. Runs `pulumi destroy` to remove all resources
3. Handles resources that can't be deleted via API (VPC, Spaces)
4. Provides manual cleanup instructions if needed

**Cost Impact**: -$68/month (saves ~$68/month)

**Safety**: Requires typing "DESTROY" to prevent accidental deletion

**Usage**:
```
Actions → pricy-teardown.yml → Run workflow
- confirm: DESTROY
```

**Duration**: ~5-8 minutes

---

#### 3. pulumi-refresh.yml
**Purpose**: Sync Pulumi state with actual Digital Ocean resources

**Triggers**: Manual (workflow_dispatch)

**What it does**:
1. Runs `pulumi refresh` to check actual resource state
2. Updates Pulumi state to match reality
3. Removes manually deleted resources from state
4. Useful for fixing state drift

**Cost Impact**: $0 (read-only operation)

**When to use**:
- After manually modifying resources in DO console
- Before running deploy/teardown after manual changes
- When Pulumi state seems out of sync

**Usage**:
```
Actions → pulumi-refresh.yml → Run workflow
```

**Duration**: ~2-3 minutes

---

### 🪶 Lite Deployment (GitHub Pages) - "Cheap"

#### 4. cheap-deploy.yml
**Purpose**: Build and deploy static site to GitHub Pages

**Triggers**: Manual (workflow_dispatch)

**Inputs**:
- `base_path` (optional): Base path for GitHub Pages
  - Leave empty for custom domain (wheeleraiduel.online)
  - Use `/repo-name` for username.github.io/repo-name

**What it does**:
1. Installs Node.js dependencies
2. Builds frontend in lite mode (`npm run build:lite`)
3. Generates static HTML export (Next.js `output: 'export'`)
4. Uploads artifact to GitHub Pages
5. Deploys to GitHub Pages environment

**Cost Impact**: $0 (completely free)

**Features in Lite Mode**:
- Local-only voting (no backend)
- Pre-generated images from DO Spaces
- No AI generation
- No cross-session tracking
- Static site with feature flags

**Usage**:
```
Actions → cheap-deploy.yml → Run workflow
- base_path: (leave empty for custom domain)
```

**Duration**: ~3-5 minutes

**Output**: GitHub Pages URL (e.g., https://username.github.io/repo-name)

---

#### 5. cheap-teardown.yml
**Purpose**: Disable GitHub Pages deployment

**Triggers**: Manual (workflow_dispatch)

**Inputs**:
- `confirm`: Must type "DISABLE" to confirm

**What it does**:
1. Provides instructions for manual teardown
2. GitHub Pages cannot be fully disabled via API
3. User must manually disable in Settings > Pages

**Cost Impact**: $0

**Manual Steps Required**:
1. Go to **Settings** > **Pages**
2. Under **Source**, select **None**
3. Click **Save**

**Safety**: Requires typing "DISABLE" to prevent accidental removal

**Usage**:
```
Actions → cheap-teardown.yml → Run workflow
- confirm: DISABLE
```

---

### 🔄 DNS Management

#### 6. dns-cutover.yml
**Purpose**: Switch DNS between Digital Ocean and GitHub Pages

**Triggers**: Manual (workflow_dispatch)

**Inputs**:
- `target`: Choose deployment
  - `github-pages`: Switch to GitHub Pages (lite)
  - `digital-ocean`: Switch to Digital Ocean (full)
- `confirm`: Must type "SWITCH" to confirm

**What it does**:
1. Determines target IP addresses:
   - **GitHub Pages**: Uses GitHub's IPs (185.199.108-111.153)
   - **Digital Ocean**: Fetches Load Balancer IP from Pulumi
2. Removes existing A records for apex (@) and www
3. Creates new A records pointing to target
4. Verifies new DNS configuration
5. Provides propagation timeline

**DNS Records Updated**:
- `wheeleraiduel.online` (apex)
- `www.wheeleraiduel.online`

**Cost Impact**: $0 (DNS changes only)

**Safety**: Requires typing "SWITCH" to prevent accidental changes

**Propagation Time**: 1-48 hours (usually 1-2 hours)

**Usage**:
```
Actions → dns-cutover.yml → Run workflow
- target: github-pages (or digital-ocean)
- confirm: SWITCH
```

**Duration**: ~2-3 minutes

---

## 🎯 Common Deployment Scenarios

### Scenario 1: Initial Setup (Both Deployments)

**Goal**: Set up both deployments for testing

```
1. Actions → pricy-deploy.yml (droplet_count: 2)
   → Full deployment live at DO Load Balancer IP

2. Actions → cheap-deploy.yml (base_path: empty)
   → Lite deployment live at GitHub Pages URL

3. Test both independently:
   - Full: http://[DO_LB_IP]
   - Lite: https://[username].github.io/[repo]

4. Actions → dns-cutover.yml (target: digital-ocean)
   → Domain points to full deployment
```

**Cost**: $68/month (DO), $5/month (DO Spaces for both)

---

### Scenario 2: Switch to Free (Cost Savings)

**Goal**: Move to GitHub Pages to save $68/month

```
1. Ensure cheap-deploy.yml has been run

2. Actions → dns-cutover.yml (target: github-pages, confirm: SWITCH)
   → Domain now points to GitHub Pages

3. Wait 1-2 hours for DNS propagation

4. Verify domain is serving lite deployment

5. Actions → pricy-teardown.yml (confirm: DESTROY)
   → Destroy DO resources, save $68/month
```

**Cost Before**: $68/month
**Cost After**: $5/month (DO Spaces + CDN preserved)
**Savings**: $63/month 🎉

**Note**: DO Spaces bucket is automatically preserved to maintain images for lite deployment.

---

### Scenario 3: Demo Day (Bring Up Full Features)

**Goal**: Show off full capabilities for interview/demo

```
1. Actions → pricy-deploy.yml (droplet_count: 2)
   → Deploy full infrastructure (~10 minutes)

2. Wait for deployment to complete

3. Actions → dns-cutover.yml (target: digital-ocean, confirm: SWITCH)
   → Switch domain to full deployment

4. Wait 5-10 minutes for DNS propagation

5. Demo full AI generation features! 🎨

6. After demo:
   Actions → dns-cutover.yml (target: github-pages, confirm: SWITCH)
   → Switch back to free deployment

7. Actions → pricy-teardown.yml (confirm: DESTROY)
   → Save costs until next demo
```

**Cost**: ~$2-3 for a day of demos (prorated)

---

### Scenario 4: Fixing State Drift

**Goal**: Sync Pulumi state after manual changes

```
1. Manual changes made in DO console (e.g., resized droplet)

2. Actions → pulumi-refresh.yml
   → Sync Pulumi state with actual DO resources

3. State is now accurate for next deploy/teardown
```

---

### Scenario 5: Scaling Up (Traffic Spike)

**Goal**: Add more droplets for increased capacity

```
1. Actions → pricy-deploy.yml (droplet_count: 4)
   → Scale from 2 to 4 droplets

2. Load balancer automatically distributes traffic
```

**Cost Impact**: +$36/month (2 additional droplets at $18 each)

---

## 🔐 Required GitHub Secrets

Configure these secrets in **Settings** > **Secrets and variables** > **Actions**:

### Digital Ocean
- `DO_ACCESS_TOKEN`: Digital Ocean API token
- `DO_SPACES_ACCESS_KEY`: DO Spaces access key
- `DO_SPACES_SECRET_KEY`: DO Spaces secret key

### Pulumi
- `PULUMI_ACCESS_TOKEN`: Pulumi Cloud access token

### API Keys (for full deployment)
- `GOOGLE_API_KEY`: Google Vertex AI API key
- `LEONARDO_API_KEY`: Leonardo AI API key
- `FREEPIK_API_KEY`: Freepik API key

---

## 🚨 Important Notes

### GitHub Pages Limitations
- **No backend**: Lite mode has no API server
- **Local voting only**: Votes stored in browser localStorage
- **Pre-generated images**: Cannot generate new images
- **No cross-session data**: Each user's data is isolated to their browser

### Digital Ocean Costs
- **Always running**: Full deployment costs $68/month 24/7
- **Prorated**: DO charges are prorated (hourly billing)
- **Save costs**: Tear down when not actively using

### DNS Propagation
- **TTL**: 1 hour (3600 seconds)
- **Propagation**: 1-48 hours (usually 1-2 hours)
- **Testing**: Use DNS checker: https://dnschecker.org

### Blue/Green Testing
- Both deployments can run simultaneously
- Test independently before DNS cutover
- Zero-downtime switching

### Feature Flags
- **Single codebase**: Same code for both deployments
- **Build-time**: Flags set during build process
- **Runtime detection**: App knows which mode it's in
- **Graceful degradation**: UI adapts to available features

---

## 📊 Cost Comparison

| Deployment | Monthly Cost | Features | Notes |
|------------|--------------|----------|-------|
| **Full (DO)** | $68 | AI generation, live voting, cross-session tracking, multi-provider support | Includes Spaces |
| **Lite (GH Pages)** | $5 | Static site, local voting, pre-generated images | DO Spaces only |

**Cost Breakdown**:
- **DO Spaces + CDN**: $5/month (always preserved, needed for both deployments)
- **DO Infrastructure**: $63/month (Load Balancer + Droplets + Valkey)
- **GitHub Pages**: $0/month (completely free)

**Recommended Strategy**:
- Keep lite deployment active ($5/month)
- Bring up full deployment for demos/interviews (adds $63/month)
- Tear down full deployment after each use (saves $63/month)
- **DO Spaces is automatically preserved** during teardown

**Estimated Demo Costs**: $5/month baseline + $2-3 per demo day (prorated)

---

## 🔗 Related Documentation

- **[../../README.md](../../README.md)** - Project overview and quick start
- **[dual-deployment.md](dual-deployment.md)** - Complete dual-deployment setup guide
- **[spaces-integration.md](spaces-integration.md)** - DO Spaces integration details
- **[spaces-preservation.md](spaces-preservation.md)** - How teardown protects images
- **[../README.md](../README.md)** - Infrastructure documentation
- **[../../frontend/README.md](../../frontend/README.md)** - Frontend application documentation

---

**Questions?** Check the [Issues](https://github.com/your-repo/issues) page or review workflow logs in the Actions tab.
