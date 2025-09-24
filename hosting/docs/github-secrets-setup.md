# GitHub Secrets Configuration

This document describes how to set up GitHub repository secrets for automated deployment of your CGC infrastructure and applications.

## Required Secrets

### Infrastructure Secrets

| Secret Name | Description | Example |
|-------------|-------------|---------|
| `PULUMI_ACCESS_TOKEN` | Pulumi access token for state management | `pul-abc123...` |
| `DO_ACCESS_TOKEN` | Digital Ocean access token | `dop_v1_abc123...` |

### Application Secrets

| Secret Name | Description | Example |
|-------------|-------------|---------|
| `GOOGLE_API_KEY` | Google API key for image generation | `AIza...` |
| `LEONARDO_API_KEY` | Leonardo AI API key | `leonardo_...` |
| `FREEPIK_API_KEY` | Freepik API key | `freepik_...` |


## How to Add Secrets

1. **Navigate to your GitHub repository**
2. **Go to Settings > Secrets and variables > Actions**
3. **Click "New repository secret"**
4. **Add each secret with the exact name from the table above**

## Getting the Required Values

### 1. Pulumi Access Token
```bash
# Login to Pulumi and get token
pulumi login
# Go to https://app.pulumi.com/account/tokens to create a new token
```

### 2. Digital Ocean Access Token
1. Go to [Digital Ocean API Tokens](https://cloud.digitalocean.com/account/api/tokens)
2. Create a new Personal Access Token
3. Give it read/write permissions

### 3. Digital Ocean Services (Automatic)
Your infrastructure will automatically create and configure:
- **Digital Ocean Spaces**: Object storage for generated content
- **Valkey Database**: Managed in-memory database for caching user votes
- **VPC and Networking**: Private network for secure service communication

No additional setup required - these services are provisioned during infrastructure deployment.

### 4. API Keys for Image Services
- **Google API Key**: [Google Cloud Console](https://console.cloud.google.com/apis/credentials)
- **Leonardo AI**: [Leonardo AI Dashboard](https://app.leonardo.ai/api-access)
- **Freepik API**: [Freepik Developer Portal](https://www.freepik.com/api)

## Security Best Practices

### ✅ Do's
- Use separate API keys for different environments (dev/staging/prod)
- Regularly rotate API keys and tokens
- Use the principle of least privilege for API permissions
- Monitor API usage and set up alerts
- Keep SSH keys specific to deployment (don't reuse personal keys)

### ❌ Don'ts
- Never commit secrets to code or logs
- Don't share secrets between repositories unless necessary
- Don't use production secrets in development workflows
- Don't hardcode secrets in workflow files

## Environment-Specific Secrets

For multiple environments, you can create environment-specific secrets:

```
GOOGLE_API_KEY_DEV
GOOGLE_API_KEY_STAGING
GOOGLE_API_KEY_PROD
```

Then modify the workflow to use the appropriate secret based on the environment.

## Troubleshooting

### Common Issues

1. **"Secret not found" error**
   - Verify secret name matches exactly (case-sensitive)
   - Check that secret is set at repository level, not organization level

2. **SSH connection failures**
   - Ensure SSH key is added to Digital Ocean
   - Verify private key format (should start with `-----BEGIN OPENSSH PRIVATE KEY-----`)
   - Check that droplet is accessible and SSH service is running

3. **API authentication errors**
   - Verify API keys are valid and not expired
   - Check API key permissions and quotas
   - Ensure API keys are for the correct service/project

4. **Pulumi authentication issues**
   - Verify Pulumi access token is valid
   - Check that token has permissions for the organization/project
   - Ensure Pulumi backend is properly configured

## Monitoring and Alerts

Consider setting up monitoring for:
- API key usage and quotas
- Deployment success/failure rates
- Infrastructure costs
- Security events (failed SSH attempts, etc.)

## Secret Rotation Schedule

| Secret Type | Rotation Frequency | Notes |
|-------------|-------------------|-------|
| SSH Keys | Every 90 days | Coordinate with team deployments |
| API Keys | Every 6 months | Check with service providers for best practices |
| Pulumi Tokens | Every 12 months | Ensure team access continuity |
| DO Access Tokens | Every 6 months | Monitor for any unauthorized usage |

## Support

If you encounter issues with secret management:
1. Check the GitHub Actions logs for specific error messages
2. Verify all secrets are properly set in the repository settings
3. Test API keys manually before using in workflows
4. Consult the individual service documentation for API key requirements