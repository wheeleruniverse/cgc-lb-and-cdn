# CGC Infrastructure - Digital Ocean Hosting

This directory contains Infrastructure as Code (IaC) using Pulumi and Go to deploy the CGC Load Balancing and CDN project on Digital Ocean.

## Architecture Overview

The infrastructure includes:

- **1 Load Balancer**: Digital Ocean Load Balancer for distributing traffic
- **2 Droplets**:
  - Backend droplet (Go API server on port 8080)
  - Frontend droplet (Next.js application on port 3000)
- **Spaces Storage**: Object storage for generated images and content
- **Spaces CDN**: Content Delivery Network for serving static files
- **Valkey Database**: Managed in-memory database for caching user votes and session data
- **VPC**: Private network for secure communication between resources
- **Firewall**: Security rules for controlling access

## Prerequisites

1. **Digital Ocean Account**: Sign up at [digitalocean.com](https://www.digitalocean.com)
2. **Digital Ocean API Token**: Create a personal access token in your DO account
3. **SSH Key**: Upload your SSH public key to Digital Ocean
4. **Pulumi**: Install Pulumi CLI: `curl -fsSL https://get.pulumi.com | sh`
5. **Go**: Install Go 1.21 or later

## Setup

### Option 1: GitHub Actions Deployment (Recommended)

1. **Set up GitHub Secrets**: Follow the guide in [`docs/github-secrets-setup.md`](docs/github-secrets-setup.md)
2. **Manual deployment**: Trigger deployment via GitHub Actions workflow_dispatch
3. **Monitor deployment**: Check GitHub Actions for progress and outputs

### Option 2: Local Development

1. **Configure Digital Ocean Access Token**:
   ```bash
   export DIGITALOCEAN_TOKEN="your_digital_ocean_access_token"
   ```

2. **Initialize Pulumi Stack**:
   ```bash
   cd hosting
   pulumi stack init dev
   ```

3. **Configure Stack Settings**:
   ```bash
   # Set the Digital Ocean region
   pulumi config set digitalocean:region nyc3

   # Optional: Set a custom domain for CDN
   pulumi config set cgc-hosting:domain "your-domain.com"
   ```

## Deployment

1. **Build the Infrastructure Program** (optional, Pulumi will do this automatically):
   ```bash
   mkdir -p bin
   go build -o bin/cgc-hosting .
   ```

2. **Deploy Infrastructure**:
   ```bash
   pulumi up
   ```

3. **Review the plan** and confirm deployment when prompted.

4. **Note the outputs**:
   - Load Balancer IP address
   - Backend and Frontend droplet IPs
   - DO Spaces bucket information
   - CDN endpoint
   - Valkey database connection details

## Post-Deployment Steps

### Application Deployment

With the simplified architecture, applications are deployed through infrastructure automation. The droplets are configured in private subnets and accessed only through the load balancer.

Consider using:
- **DigitalOcean App Platform** for containerized applications
- **Infrastructure-as-Code** deployment through Pulumi
- **Container registries** for application images

### 3. Configure DNS (Optional)

If you have a custom domain:
1. Point your domain's A record to the Load Balancer IP
2. Update the CDN configuration with your domain

## Configuration Options

| Config Key | Description | Default |
|------------|-------------|---------|
| `digitalocean:region` | Digital Ocean region | `nyc3` |
| `cgc-hosting:domain` | Custom domain for CDN | `""` |

## Resource Details

### Load Balancer
- **Size**: Small (suitable for moderate traffic)
- **Health Check**: HTTP check on `/health` endpoint
- **Ports**: 80 (HTTP) and 443 (HTTPS)
- **SSL**: Ready for SSL certificate (update certificate ID)

### Droplets
- **Size**: s-2vcpu-2gb (2 vCPU, 2GB RAM)
- **Image**: Ubuntu 22.04 LTS
- **Network**: Private VPC with public IP
- **Storage**: 60GB SSD
- **Note**: The 2 vCPU / 2GB configuration is required to handle UserData script execution while simultaneously serving backend, frontend, and nginx. Smaller instances (s-1vcpu-1gb) were insufficient during testing.

### DO Spaces Storage
- **Name**: cgc-lb-and-cdn-content
- **Region**: nyc3
- **ACL**: public-read
- **CDN**: Built-in CDN enabled

### Valkey Database
- **Engine**: Valkey 8 (Redis-compatible)
- **Size**: db-s-1vcpu-1gb (1 vCPU, 1GB RAM)
- **Node Count**: 1 (single node for cost efficiency)
- **Network**: Private VPC access only
- **Use Case**: User vote caching, session storage, fast data access

### Security
- **Firewall**: Restricts access to necessary ports only
- **VPC**: Private communication between droplets
- **Private Subnets**: Droplets only accessible via load balancer

## Monitoring and Maintenance

1. **Check Service Status**:
   Monitor through DigitalOcean dashboard and load balancer health checks

2. **View Logs**:
   Access logs through DigitalOcean monitoring or centralized logging solution

3. **Load Balancer Health**:
   Check the Digital Ocean dashboard for droplet health status.

## Scaling

To scale the infrastructure:

1. **Horizontal Scaling**: Add more droplets to the load balancer
2. **Vertical Scaling**: Resize existing droplets
3. **CDN**: Spaces CDN automatically scales globally

## Cost Estimation

Monthly costs (approximate):
- Load Balancer: $12/month
- 2 Droplets (s-2vcpu-2gb): $36/month ($18 each)
- Spaces Storage (250GB): $5/month
- Valkey Database (db-s-1vcpu-1gb): $15/month
- **Total**: ~$68/month

**Note**: The s-2vcpu-2gb droplet size is required to handle the full-stack deployment (backend, frontend, nginx, and UserData bootstrapping). Initial testing with s-1vcpu-1gb droplets proved insufficient for concurrent workload execution.

## Cleanup

To destroy all resources:
```bash
pulumi destroy
```

## Troubleshooting

### Common Issues

1. **Access Token**: Verify your Digital Ocean access token has the correct permissions
2. **Region**: Some features may not be available in all regions
3. **Spaces**: Ensure DO Spaces service is available in your selected region

### Getting Help

- Digital Ocean Documentation: [docs.digitalocean.com](https://docs.digitalocean.com)
- Pulumi Documentation: [pulumi.com/docs](https://www.pulumi.com/docs)
- Project Issues: Check the main repository for known issues

## Security Considerations

1. **Access Tokens**: Never commit access tokens to version control
2. **Private Subnets**: Droplets are in private subnets, accessible only via load balancer
3. **Firewall**: Review and adjust firewall rules as needed
4. **SSL**: Consider adding SSL certificates for production use
5. **Updates**: Regularly update droplet software and security patches