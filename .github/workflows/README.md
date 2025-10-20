# GitHub Actions Workflows

This directory contains automated workflows for managing the CGC infrastructure on Digital Ocean using Pulumi.

## Available Workflows

### 1. Deploy Infrastructure (`deploy.yml`)

**Purpose**: Deploys or updates the complete infrastructure stack to Digital Ocean.

**Trigger**: Manual workflow dispatch (Actions tab → Deploy Infrastructure → Run workflow)

**Parameters**:
- `droplet_count` (optional, default: 2): Number of droplets to deploy (2-10)
  - Maintains high availability architecture
  - Each droplet runs full stack (backend + frontend + nginx)
  - Cost scales at $18/month per droplet
- `recreate_valkey` (optional, default: false): Rebuild Valkey indexes from DO Spaces
  - Set to `true` to rebuild database from single source of truth
  - Useful after data corruption or state drift

**What It Does**:
1. Checks out the code and sets up Go 1.21
2. Installs Pulumi CLI
3. Builds the infrastructure program from `hosting/`
4. Configures Pulumi stack with:
   - Current commit SHA (triggers droplet replacement on new commits)
   - API keys (Google, Leonardo, Freepik) as encrypted secrets
   - DO Spaces credentials
   - Droplet count and Valkey recreation settings
5. Deploys infrastructure with `pulumi up`
6. Outputs deployment results:
   - Load Balancer IP
   - Individual droplet IPs (full-stack instances)
   - Valkey database connection details

**Required Secrets**:
- `PULUMI_ACCESS_TOKEN`: Pulumi Cloud access token
- `DO_ACCESS_TOKEN`: Digital Ocean API token
- `DO_SPACES_ACCESS_KEY`: Spaces access key
- `DO_SPACES_SECRET_KEY`: Spaces secret key
- `GOOGLE_API_KEY`: Google Imagen API key
- `LEONARDO_API_KEY`: Leonardo AI API key
- `FREEPIK_API_KEY`: Freepik API key

**Important Notes**:
- Droplets are automatically replaced when the commit SHA changes (new deployments)
- Applications are deployed via UserData script during droplet provisioning
- No manual deployment steps required after workflow completes
- Estimated runtime: 5-10 minutes

**Cost Impact**: ~$68/month base (2 droplets) + $18/month per additional droplet

---

### 2. Teardown Infrastructure (`teardown.yml`)

**Purpose**: Safely destroys all infrastructure resources to stop incurring costs.

**Trigger**: Manual workflow dispatch (Actions tab → Teardown Infrastructure → Run workflow)

**Parameters**:
- `confirm` (required): Type exactly "DESTROY" to confirm teardown
  - Safety measure to prevent accidental deletion

**What It Does**:
1. Validates confirmation input (must be "DESTROY")
2. Installs doctl (DigitalOcean CLI) for DNS cleanup
3. Cleans up auto-created DNS records (Let's Encrypt validation records)
4. Destroys all Pulumi-managed resources:
   - Load Balancer
   - Droplets (and all applications)
   - Firewall rules
   - Valkey database
   - SSL certificates
5. Handles resources that can't be deleted:
   - VPC (cannot be deleted via API - removed from state)
   - DO Spaces bucket (if not empty - removed from state)
6. Provides manual cleanup checklist

**Required Secrets**: Same as Deploy workflow

**Important Notes**:
- **DESTRUCTIVE OPERATION** - cannot be undone
- VPC will remain (cannot be deleted - this is normal and free)
- DO Spaces bucket may need manual deletion if not empty
- DNS NS records remain (nameserver records for your domain)
- Estimated runtime: 3-5 minutes

**Cost Savings**: ~$68/month (based on 2-droplet deployment)

**Safety Features**:
- Requires exact confirmation text
- Separate job shows error if confirmation doesn't match
- Cleans up auto-created DNS records to prevent orphaned records

---

### 3. Refresh Infrastructure State (`refresh.yml`)

**Purpose**: Synchronizes Pulumi state with actual Digital Ocean resources.

**Trigger**: Manual workflow dispatch (Actions tab → Refresh Infrastructure State → Run workflow)

**Parameters**: None

**What It Does**:
1. Captures current Pulumi state snapshot
2. Queries Digital Ocean API for actual resource state
3. Updates Pulumi state to match reality
4. Removes manually deleted resources from state
5. Shows before/after state comparison

**When To Use**:
- After manually modifying resources in DO console
- After manually deleting resources outside of Pulumi
- Before running Deploy or Teardown to ensure clean state
- When state drift is suspected
- As a diagnostic tool for troubleshooting

**Required Secrets**: Same as Deploy workflow

**Important Notes**:
- Read-only operation (doesn't modify infrastructure)
- Safe to run at any time
- Useful for resolving state drift issues
- Should be run before Deploy if you've manually changed resources
- Estimated runtime: 1-2 minutes

**Example Use Cases**:
```
Scenario 1: Manually deleted a droplet in DO console
→ Run Refresh to update state
→ Run Deploy to recreate the droplet

Scenario 2: Teardown failed due to locked resources
→ Manually unlock/delete resources in DO console
→ Run Refresh to sync state
→ Run Teardown again

Scenario 3: Not sure if infrastructure matches state
→ Run Refresh to check and sync
→ Review output for any drift
```

---

## Workflow Execution Order

### Standard Deployment
1. **Deploy**: Initial infrastructure setup
2. (Use infrastructure)
3. **Teardown**: Clean up when done

### Updating Infrastructure
1. **Deploy**: Push new code commit
2. Workflow automatically replaces droplets with new SHA
3. Applications auto-deploy with new code

### Recovering from Manual Changes
1. **Refresh**: Sync state with reality
2. **Deploy**: Apply desired configuration

### Complete Reset
1. **Teardown**: Destroy everything
2. (Wait for completion)
3. **Deploy**: Fresh deployment

---

## Monitoring Workflow Execution

All workflows provide detailed output in the GitHub Actions summary:

- **Deploy**: Shows IPs, database details, and deployment status
- **Teardown**: Shows cleanup progress and manual steps
- **Refresh**: Shows state changes and drift detection

Access summaries: Actions tab → Select workflow run → Summary

---

## Common Issues

### Deploy Fails with "Resource already exists"
**Solution**: Run Refresh workflow, then retry Deploy

### Teardown Fails with "Resource locked"
**Solution**:
1. Manually unlock resource in DO console
2. Run Refresh workflow
3. Retry Teardown

### Droplets not getting new code on deploy
**Solution**: Commit changes to trigger new SHA (droplets replaced on SHA change)

### API keys not working
**Solution**: Verify secrets are set correctly in repository settings

---

## Security Considerations

1. **Secrets Management**: All secrets encrypted in GitHub and Pulumi
2. **Confirmation Required**: Teardown requires explicit "DESTROY" confirmation
3. **Audit Trail**: All workflow runs logged in GitHub Actions
4. **Least Privilege**: Service accounts should have minimal required permissions

---

## Cost Monitoring

| Workflow | Cost Impact | Duration |
|----------|-------------|----------|
| Deploy | ~$68/month (base 2 droplets) | 5-10 min |
| Teardown | Saves ~$68/month | 3-5 min |
| Refresh | $0 (read-only) | 1-2 min |

**Note**: Costs accumulate while infrastructure is deployed. Run Teardown when not in use to save costs.

---

## Further Reading

- [Infrastructure Setup Guide](../../hosting/README.md)
- [Pulumi Documentation](https://www.pulumi.com/docs)
- [Digital Ocean API Documentation](https://docs.digitalocean.com/reference/api)
