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
2. **Push to main branch**: Infrastructure and applications deploy automatically
3. **Monitor deployment**: Check GitHub Actions for progress and outputs

### Option 2: Local Development

1. **Configure Digital Ocean Token**:
   ```bash
   export DIGITALOCEAN_TOKEN="your_digital_ocean_api_token"
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

   # Set your SSH key name (as it appears in Digital Ocean)
   pulumi config set cgc-hosting:ssh_key_name "your-ssh-key-name"

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
   - Spaces bucket information
   - CDN endpoint

## Post-Deployment Steps

### 1. Deploy Backend Application

SSH into the backend droplet:
```bash
ssh root@<backend-droplet-ip>
```

Deploy your Go application:
```bash
# Clone your repository
cd /opt/cgc-backend
git clone <your-repo-url> .

# Build the application
/usr/local/go/bin/go build -o server cmd/server/main.go

# Start the service
systemctl start cgc-backend.service
systemctl status cgc-backend.service
```

### 2. Deploy Frontend Application

SSH into the frontend droplet:
```bash
ssh root@<frontend-droplet-ip>
```

Deploy your Next.js application:
```bash
# Clone your repository
cd /opt/cgc-frontend
git clone <your-repo-url> .

# Update nginx config with backend IP
sed -i 's/BACKEND_IP/<backend-droplet-ip>/g' /etc/nginx/sites-available/cgc-frontend
systemctl reload nginx

# Install dependencies and build
npm install
npm run build

# Start with PM2
pm2 start ecosystem.config.js
pm2 save
pm2 startup
```

### 3. Configure DNS (Optional)

If you have a custom domain:
1. Point your domain's A record to the Load Balancer IP
2. Update the CDN configuration with your domain

## Configuration Options

| Config Key | Description | Default |
|------------|-------------|---------|
| `digitalocean:region` | Digital Ocean region | `nyc3` |
| `cgc-hosting:ssh_key_name` | SSH key name for droplets | `default` |
| `cgc-hosting:domain` | Custom domain for CDN | `""` |

## Resource Details

### Load Balancer
- **Size**: Small (suitable for moderate traffic)
- **Health Check**: HTTP check on `/health` endpoint
- **Ports**: 80 (HTTP) and 443 (HTTPS)
- **SSL**: Ready for SSL certificate (update certificate ID)

### Droplets
- **Size**: s-1vcpu-1gb (1 CPU, 1GB RAM)
- **Image**: Ubuntu 22.04 LTS
- **Network**: Private VPC with public IP
- **Storage**: 25GB SSD

### Spaces Storage
- **Name**: cgc-generated-content
- **Region**: nyc3
- **ACL**: public-read
- **CDN**: Built-in CDN enabled

### Security
- **Firewall**: Restricts access to necessary ports only
- **VPC**: Private communication between droplets
- **SSH**: Key-based authentication only

## Monitoring and Maintenance

1. **Check Service Status**:
   ```bash
   # Backend
   ssh root@<backend-ip> "systemctl status cgc-backend.service"

   # Frontend
   ssh root@<frontend-ip> "pm2 status"
   ```

2. **View Logs**:
   ```bash
   # Backend logs
   ssh root@<backend-ip> "journalctl -u cgc-backend.service -f"

   # Frontend logs
   ssh root@<frontend-ip> "pm2 logs cgc-frontend"
   ```

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
- 2 Droplets (s-1vcpu-1gb): $12/month ($6 each)
- Spaces Storage (250GB): $5/month
- **Total**: ~$29/month

## Cleanup

To destroy all resources:
```bash
pulumi destroy
```

## Troubleshooting

### Common Issues

1. **SSH Key Not Found**: Ensure your SSH key is uploaded to Digital Ocean with the correct name
2. **API Token**: Verify your Digital Ocean token has the correct permissions
3. **Region**: Some features may not be available in all regions

### Getting Help

- Digital Ocean Documentation: [docs.digitalocean.com](https://docs.digitalocean.com)
- Pulumi Documentation: [pulumi.com/docs](https://www.pulumi.com/docs)
- Project Issues: Check the main repository for known issues

## Security Considerations

1. **API Tokens**: Never commit API tokens to version control
2. **SSH Keys**: Use strong SSH keys and consider key rotation
3. **Firewall**: Review and adjust firewall rules as needed
4. **SSL**: Consider adding SSL certificates for production use
5. **Updates**: Regularly update droplet software and security patches