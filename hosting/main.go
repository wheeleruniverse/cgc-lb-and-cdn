package main

import (
	"fmt"
	"strconv"

	"github.com/pulumi/pulumi-digitalocean/sdk/v4/go/digitalocean"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		// Load configuration
		cfg := config.New(ctx, "")
		domain := cfg.Require("domain")
		sha := cfg.Require("sha")
		dropletCount := cfg.RequireInt("droplet_count")

		// Get API keys from Pulumi config (passed from GitHub Actions)
		googleAPIKey := cfg.Get("google_api_key")
		leonardoAPIKey := cfg.Get("leonardo_api_key")
		freepikAPIKey := cfg.Get("freepik_api_key")
		useDoSpaces := cfg.Get("use_do_spaces")
		if useDoSpaces == "" {
			useDoSpaces = "false" // Default to local storage
		}

		// Get Spaces credentials from Pulumi config
		spacesAccessKey := cfg.Get("do_spaces_access_key")
		spacesSecretKey := cfg.Get("do_spaces_secret_key")

		// Note: In GitHub Actions, these should be passed as:
		// --config do_spaces_access_key=${{ secrets.DO_SPACES_ACCESS_KEY }}
		// --config do_spaces_secret_key=${{ secrets.DO_SPACES_SECRET_KEY }}

		// Create Spaces bucket for storing generated images and logs
		spaceBucketName := "cgc-lb-and-cdn-content"
		spaceBucketEndpoint := "nyc3.digitaloceanspaces.com"
		spaceBucket, err := digitalocean.NewSpacesBucket(ctx, "cgc-lb-and-cdn-spaces", &digitalocean.SpacesBucketArgs{
			Name:   pulumi.String(spaceBucketName),
			Region: pulumi.String("nyc3"),
			Acl:    pulumi.String("public-read"), // Public read for CDN access
		})
		if err != nil {
			return err
		}

		// Enable CDN on the Spaces bucket
		_, err = digitalocean.NewCdn(ctx, "cgc-lb-and-cdn-cdn", &digitalocean.CdnArgs{
			Origin: spaceBucket.BucketDomainName,
		})
		if err != nil {
			return err
		}

		// Create VPC for the project
		vpc, err := digitalocean.NewVpc(ctx, "cgc-lb-and-cdn-vpc", &digitalocean.VpcArgs{
			Name:    pulumi.String("cgc-lb-and-cdn-vpc"),
			Region:  pulumi.String("nyc3"),
			IpRange: pulumi.String("10.20.0.0/24"),
		})
		if err != nil {
			return err
		}

		// Create Valkey managed database cluster for caching user votes (VPC private only)
		valkeyCluster, err := digitalocean.NewDatabaseCluster(ctx, "cgc-lb-and-cdn-valkey", &digitalocean.DatabaseClusterArgs{
			Name:               pulumi.String("cgc-lb-and-cdn-valkey"),
			Engine:             pulumi.String("valkey"),
			Version:            pulumi.String("8"),
			Size:               pulumi.String("db-s-1vcpu-1gb"), // Small cluster for development
			Region:             pulumi.String("nyc3"),
			NodeCount:          pulumi.Int(1), // Single node for cost efficiency
			PrivateNetworkUuid: vpc.ID(),
			// Tags removed due to permission issues - can be added manually in DO console
		})
		if err != nil {
			return err
		}

		// Create droplets dynamically based on droplet_count
		// Each droplet runs both backend and frontend applications
		droplets := make([]*digitalocean.Droplet, dropletCount)
		for i := 0; i < dropletCount; i++ {
			logicalName := fmt.Sprintf("cgc-lb-and-cdn-droplet-%d", i+1)
			physicalName := fmt.Sprintf("cgc-lb-and-cdn-droplet-%s-%d", sha[:7], i+1)
			droplet, err := digitalocean.NewDroplet(ctx, logicalName, &digitalocean.DropletArgs{
				Name:    pulumi.String(physicalName),
				Image:   pulumi.String("ubuntu-22-04-x64"),
				Size:    pulumi.String("s-1vcpu-1gb"),
				Region:  pulumi.String("nyc3"),
				VpcUuid: vpc.ID(),
				UserData: pulumi.All(valkeyCluster.Host, valkeyCluster.Port, valkeyCluster.Password, spaceBucket.Name, spaceBucket.Region).ApplyT(func(args []interface{}) string {
					bucketName := args[3].(string)
					return getFullStackUserData(sha, googleAPIKey, leonardoAPIKey, freepikAPIKey, useDoSpaces, bucketName, bucketName, spaceBucketEndpoint, args[0].(string), fmt.Sprintf("%v", args[1]), args[2].(string), spacesAccessKey, spacesSecretKey)
				}).(pulumi.StringOutput),
				// Tags removed due to permission issues
			})
			if err != nil {
				return err
			}
			droplets[i] = droplet
		}

		// Note: Domain should already exist in DigitalOcean (manually created or via DNS provider)
		// We'll just use the domain name for DNS records and certificate

		// Create Let's Encrypt certificate for the domain
		certificate, err := digitalocean.NewCertificate(ctx, "cgc-lb-and-cdn-cert", &digitalocean.CertificateArgs{
			Name: pulumi.String("cgc-lb-and-cdn-cert"),
			Type: pulumi.String("lets_encrypt"),
			Domains: pulumi.StringArray{
				pulumi.String(domain),
				pulumi.String("www." + domain),
			},
		})
		if err != nil {
			return err
		}

		// Load balancer to distribute traffic between both droplets
		// Build dynamic droplet ID array for Load Balancer
		dropletIDInputs := make([]interface{}, len(droplets))
		for i, droplet := range droplets {
			dropletIDInputs[i] = droplet.ID()
		}

		// Wait for droplets to have IPs assigned before adding to LB
		loadBalancer, err := digitalocean.NewLoadBalancer(ctx, "cgc-lb-and-cdn-lb", &digitalocean.LoadBalancerArgs{
			Name:    pulumi.String("cgc-lb-and-cdn-lb"),
			Region:  pulumi.String("nyc3"),
			Size:    pulumi.String("lb-small"),
			VpcUuid: vpc.ID(),

			// Connect to all droplets dynamically
			DropletIds: pulumi.All(dropletIDInputs...).ApplyT(func(args []interface{}) []int {
				ids := make([]int, len(args))
				for i, arg := range args {
					idStr := string(arg.(pulumi.ID))
					id, _ := strconv.Atoi(idStr)
					ids[i] = id
				}
				return ids
			}).(pulumi.IntArrayOutput),

			// Forward traffic to backend API on port 8080
			ForwardingRules: digitalocean.LoadBalancerForwardingRuleArray{
				// HTTP traffic (will redirect to HTTPS)
				&digitalocean.LoadBalancerForwardingRuleArgs{
					EntryProtocol:  pulumi.String("http"),
					EntryPort:      pulumi.Int(80),
					TargetProtocol: pulumi.String("http"),
					TargetPort:     pulumi.Int(8080),
				},
				// HTTPS traffic to backend
				&digitalocean.LoadBalancerForwardingRuleArgs{
					EntryProtocol:   pulumi.String("https"),
					EntryPort:       pulumi.Int(443),
					TargetProtocol:  pulumi.String("http"),
					TargetPort:      pulumi.Int(8080),
					CertificateName: certificate.Name,
				},
			},

			// Health check configuration
			// More lenient settings to keep droplets in rotation during temporary issues
			Healthcheck: &digitalocean.LoadBalancerHealthcheckArgs{
				Protocol:               pulumi.String("http"),
				Port:                   pulumi.Int(8080),
				Path:                   pulumi.String("/health"),
				CheckIntervalSeconds:   pulumi.Int(10),
				ResponseTimeoutSeconds: pulumi.Int(10), // Increased from 5s to 10s
				HealthyThreshold:       pulumi.Int(2),  // Reduced from 3 to 2 (faster recovery)
				UnhealthyThreshold:     pulumi.Int(5),  // Increased from 3 to 5 (more tolerant)
			},

			// Sticky sessions
			StickySessions: &digitalocean.LoadBalancerStickySessionsArgs{
				Type:             pulumi.String("cookies"),
				CookieName:       pulumi.String("lb"),
				CookieTtlSeconds: pulumi.Int(300),
			},
		}, pulumi.DependsOn(convertDropletsToResources(droplets)))
		if err != nil {
			return err
		}

		// Create firewall rules (depends on droplets being created first)
		firewall, err := digitalocean.NewFirewall(ctx, "cgc-lb-and-cdn-firewall", &digitalocean.FirewallArgs{
			Name: pulumi.String("cgc-lb-and-cdn-firewall"),

			// Inbound rules
			InboundRules: digitalocean.FirewallInboundRuleArray{
				// HTTP access
				&digitalocean.FirewallInboundRuleArgs{
					Protocol:  pulumi.String("tcp"),
					PortRange: pulumi.String("80"),
					SourceAddresses: pulumi.StringArray{
						pulumi.String("0.0.0.0/0"),
						pulumi.String("::/0"),
					},
				},
				// HTTPS access
				&digitalocean.FirewallInboundRuleArgs{
					Protocol:  pulumi.String("tcp"),
					PortRange: pulumi.String("443"),
					SourceAddresses: pulumi.StringArray{
						pulumi.String("0.0.0.0/0"),
						pulumi.String("::/0"),
					},
				},
				// Backend API port
				&digitalocean.FirewallInboundRuleArgs{
					Protocol:  pulumi.String("tcp"),
					PortRange: pulumi.String("8080"),
					SourceAddresses: pulumi.StringArray{
						pulumi.String("10.20.0.0/24"), // VPC only
					},
				},
				// Frontend Next.js port
				&digitalocean.FirewallInboundRuleArgs{
					Protocol:  pulumi.String("tcp"),
					PortRange: pulumi.String("3000"),
					SourceAddresses: pulumi.StringArray{
						pulumi.String("10.20.0.0/24"), // VPC only
					},
				},
				// Valkey database port (internal VPC access only)
				&digitalocean.FirewallInboundRuleArgs{
					Protocol:  pulumi.String("tcp"),
					PortRange: pulumi.String("25061"), // Standard Valkey port in DO
					SourceAddresses: pulumi.StringArray{
						pulumi.String("10.20.0.0/24"), // VPC only
					},
				},
			},

			// Outbound rules (allow all outbound traffic)
			OutboundRules: digitalocean.FirewallOutboundRuleArray{
				&digitalocean.FirewallOutboundRuleArgs{
					Protocol:  pulumi.String("tcp"),
					PortRange: pulumi.String("1-65535"),
					DestinationAddresses: pulumi.StringArray{
						pulumi.String("0.0.0.0/0"),
						pulumi.String("::/0"),
					},
				},
				&digitalocean.FirewallOutboundRuleArgs{
					Protocol:  pulumi.String("udp"),
					PortRange: pulumi.String("1-65535"),
					DestinationAddresses: pulumi.StringArray{
						pulumi.String("0.0.0.0/0"),
						pulumi.String("::/0"),
					},
				},
			},

			// Associate firewall with all droplets dynamically
			DropletIds: func() pulumi.IntArray {
				intArray := make(pulumi.IntArray, len(droplets))
				for i, droplet := range droplets {
					intArray[i] = droplet.ID().ApplyT(func(id string) (int, error) {
						return strconv.Atoi(id)
					}).(pulumi.IntOutput)
				}
				return intArray
			}(),
		}, pulumi.DependsOn(convertDropletsToResources(droplets)))
		if err != nil {
			return err
		}

		// Note: DNS A (IPv4) and AAAA (IPv6) records are automatically created by DigitalOcean when the
		// Let's Encrypt certificate is issued for the Zone Apex route. The certificate creation process requires
		// domain validation, so DigitalOcean auto-creates DNS records pointing to the load balancer to complete the
		// HTTP-01 challenge.

		// Create DNS A record for www subdomain
		_, err = digitalocean.NewDnsRecord(ctx, "cgc-lb-and-cdn-dns-www-v4", &digitalocean.DnsRecordArgs{
			Domain: pulumi.String(domain),
			Type:   pulumi.String("A"),
			Name:   pulumi.String("www"),
			Value:  loadBalancer.Ip,
			Ttl:    pulumi.Int(3600),
		})
		if err != nil {
			return err
		}

		// Create DNS AAAA record for www subdomain
		_, err = digitalocean.NewDnsRecord(ctx, "cgc-lb-and-cdn-dns-www-v6", &digitalocean.DnsRecordArgs{
			Domain: pulumi.String(domain),
			Type:   pulumi.String("AAAA"),
			Name:   pulumi.String("www"),
			Value:  loadBalancer.Ipv6,
			Ttl:    pulumi.Int(3600),
		})
		if err != nil {
			return err
		}

		// Export important information
		ctx.Export("domain", pulumi.String(domain))
		ctx.Export("domainUrl", pulumi.String("https://"+domain))
		ctx.Export("wwwDomainUrl", pulumi.String("https://www."+domain))
		ctx.Export("certificateId", certificate.ID())
		ctx.Export("loadBalancerIp", loadBalancer.Ip)

		// Export droplet IPs dynamically
		for i, droplet := range droplets {
			ctx.Export(fmt.Sprintf("droplet%dIp", i+1), droplet.Ipv4Address)
		}

		ctx.Export("spacesBucketName", pulumi.String(spaceBucketName))
		ctx.Export("spacesBucketEndpoint", pulumi.String(spaceBucketEndpoint))
		ctx.Export("spacesCdnEndpoint", pulumi.String("https://"+spaceBucketName+"."+spaceBucketEndpoint))
		ctx.Export("vpcId", vpc.ID())
		ctx.Export("firewallId", firewall.ID())
		ctx.Export("valkeyClusterHost", valkeyCluster.Host)
		ctx.Export("valkeyClusterPort", valkeyCluster.Port)
		ctx.Export("valkeyClusterUri", valkeyCluster.Uri)
		ctx.Export("valkeyClusterPassword", valkeyCluster.Password)

		return nil
	})
}

// convertDropletsToResources converts a slice of droplets to a slice of pulumi.Resource
func convertDropletsToResources(droplets []*digitalocean.Droplet) []pulumi.Resource {
	resources := make([]pulumi.Resource, len(droplets))
	for i, droplet := range droplets {
		resources[i] = droplet
	}
	return resources
}

// getFullStackUserData returns cloud-init script to deploy both backend and frontend on each droplet
func getFullStackUserData(sha, googleAPIKey, leonardoAPIKey, freepikAPIKey, useDoSpaces, leftBucket, rightBucket, spacesEndpoint, valkeyHost, valkeyPort, valkeyPassword, spacesAccessKey, spacesSecretKey string) string {
	return fmt.Sprintf(`#!/bin/bash
set -e

# Deployment configuration (passed from Pulumi)
DEPLOYMENT_SHA="%s"
GOOGLE_API_KEY="%s"
LEONARDO_API_KEY="%s"
FREEPIK_API_KEY="%s"
USE_DO_SPACES="%s"
DO_SPACES_LEFT_BUCKET="%s"
DO_SPACES_RIGHT_BUCKET="%s"
DO_SPACES_ENDPOINT="%s"
DO_SPACES_ACCESS_KEY="%s"
DO_SPACES_SECRET_KEY="%s"
DO_VALKEY_HOST="%s"
DO_VALKEY_PORT="%s"
DO_VALKEY_PASSWORD="%s"

# Setup logging
LOGFILE="/var/log/cgc-lb-and-cdn-deployment.log"
exec > >(tee -a "$LOGFILE") 2>&1

echo "================================"
echo "Cloud Portfolio Challenge LB and CDN Deployment Started: $(date)"
echo "Deployment SHA: ${DEPLOYMENT_SHA}"
echo "================================"

# Set non-interactive mode for all apt operations
export DEBIAN_FRONTEND=noninteractive

# Set log upload interval (in minutes) - default to 5 for testing, can be changed via env
LOG_UPLOAD_INTERVAL_MINUTES="${LOG_UPLOAD_INTERVAL_MINUTES:-5}"
echo "[$(date)] Log upload interval set to: ${LOG_UPLOAD_INTERVAL_MINUTES} minutes"

# Update system
echo "[$(date)] Updating system packages..."
apt-get update -y
apt-get upgrade -y -o Dpkg::Options::="--force-confdef" -o Dpkg::Options::="--force-confold"

# Install required packages
echo "[$(date)] Installing required packages..."
apt-get install -y -o Dpkg::Options::="--force-confdef" -o Dpkg::Options::="--force-confold" curl wget git build-essential nginx s3cmd

# Configure s3cmd for DigitalOcean Spaces
echo "[$(date)] Configuring S3 access for Spaces..."
cat > /root/.s3cfg << S3CFG
[default]
host_base = ${DO_SPACES_ENDPOINT}
host_bucket = %%(bucket)s.${DO_SPACES_ENDPOINT}
access_key = ${DO_SPACES_ACCESS_KEY}
secret_key = ${DO_SPACES_SECRET_KEY}
use_https = True
S3CFG

# Install Go 1.23
cd /tmp
wget https://go.dev/dl/go1.23.2.linux-amd64.tar.gz
tar -C /usr/local -xzf go1.23.2.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> /etc/profile
echo 'export GOPATH=/opt/go' >> /etc/profile
echo 'export PATH=$PATH:/opt/go/bin' >> /etc/profile

# Install Node.js 18
curl -fsSL https://deb.nodesource.com/setup_18.x | sudo -E bash -
apt-get install -y -o Dpkg::Options::="--force-confdef" -o Dpkg::Options::="--force-confold" nodejs

# Install PM2 for process management
npm install -g pm2

# ===================
# BACKEND SETUP
# ===================

# Create backend app directory
mkdir -p /opt/cgc-lb-and-cdn-backend
cd /opt/cgc-lb-and-cdn-backend

# Create environment file for the backend service
cat > /opt/cgc-lb-and-cdn-backend/.env << ENVEOF
PORT=8080
HOST=0.0.0.0
GOOGLE_API_KEY=${GOOGLE_API_KEY}
LEONARDO_API_KEY=${LEONARDO_API_KEY}
FREEPIK_API_KEY=${FREEPIK_API_KEY}
USE_DO_SPACES=${USE_DO_SPACES}
DO_SPACES_LEFT_BUCKET=${DO_SPACES_LEFT_BUCKET}
DO_SPACES_RIGHT_BUCKET=${DO_SPACES_RIGHT_BUCKET}
DO_SPACES_ENDPOINT=${DO_SPACES_ENDPOINT}
DO_SPACES_ACCESS_KEY=${DO_SPACES_ACCESS_KEY}
DO_SPACES_SECRET_KEY=${DO_SPACES_SECRET_KEY}
DO_VALKEY_HOST=${DO_VALKEY_HOST}
DO_VALKEY_PORT=${DO_VALKEY_PORT}
DO_VALKEY_PASSWORD=${DO_VALKEY_PASSWORD}
ENVEOF

# Create systemd service file for backend
cat > /etc/systemd/system/cgc-lb-and-cdn-backend.service << 'EOF'
[Unit]
Description=Cloud Portfolio Challenge Load Balancer and CDN Backend Service
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=/opt/cgc-lb-and-cdn-backend
Environment=PATH=/usr/local/go/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin
EnvironmentFile=/opt/cgc-lb-and-cdn-backend/.env
ExecStart=/opt/cgc-lb-and-cdn-backend/server
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
EOF

# Enable backend service
systemctl enable cgc-lb-and-cdn-backend.service

# Clone repository and deploy backend
echo "[$(date)] Cloning repository and deploying backend..."
cd /opt/cgc-lb-and-cdn-backend
git clone https://github.com/wheeleruniverse/cgc-lb-and-cdn.git repo
cp -r repo/backend/* .
rm -rf repo

# Verify environment variables are set
echo "[$(date)] Verifying environment variables..."
echo "GOOGLE_API_KEY: ${GOOGLE_API_KEY:0:10}..."
echo "LEONARDO_API_KEY: ${LEONARDO_API_KEY:0:10}..."
echo "FREEPIK_API_KEY: ${FREEPIK_API_KEY:0:10}..."
echo "VALKEY_HOST: $DO_VALKEY_HOST"
echo "VALKEY_PORT: $DO_VALKEY_PORT"

# Build and start backend
echo "[$(date)] Building backend application..."
export PATH=$PATH:/usr/local/go/bin
export GOPATH=/opt/go
export HOME=/root
export GOCACHE=/opt/go/cache
mkdir -p "$GOCACHE"
cd /opt/cgc-lb-and-cdn-backend
go mod download
go build -o server ./cmd/server

echo "[$(date)] Starting backend service..."
systemctl start cgc-lb-and-cdn-backend.service
sleep 5

# Check backend service status
echo "[$(date)] Checking backend service status..."
systemctl status cgc-lb-and-cdn-backend.service --no-pager || true

# Test health endpoint
echo "[$(date)] Testing health endpoint..."
curl -v http://localhost:8080/health || echo "Health check failed!"

# ===================
# FRONTEND SETUP
# ===================

# Create frontend app directory
mkdir -p /opt/cgc-lb-and-cdn-frontend
cd /opt/cgc-lb-and-cdn-frontend

# Configure Nginx as reverse proxy
cat > /etc/nginx/sites-available/cgc-lb-and-cdn-frontend << 'NGINXEOF'
server {
    listen 80;
    server_name _;

    # Serve static files directly
    location /_next/static/ {
        alias /opt/cgc-lb-and-cdn-frontend/.next/static/;
        expires 1y;
        access_log off;
    }

    # Proxy API requests to local backend
    location /api/ {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    # Proxy everything else to Next.js
    location / {
        proxy_pass http://localhost:3000;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
NGINXEOF

# Enable the site
ln -sf /etc/nginx/sites-available/cgc-lb-and-cdn-frontend /etc/nginx/sites-enabled/
rm -f /etc/nginx/sites-enabled/default

# Create PM2 ecosystem file for frontend
cat > /opt/cgc-lb-and-cdn-frontend/ecosystem.config.js << 'PM2EOF'
module.exports = {
  apps: [{
    name: 'cgc-lb-and-cdn-frontend',
    script: 'npm',
    args: 'start',
    cwd: '/opt/cgc-lb-and-cdn-frontend',
    env: {
      NODE_ENV: 'production',
      PORT: 3000
    }
  }]
}
PM2EOF

# Enable and start nginx
systemctl enable nginx
systemctl start nginx

# Clone repository and deploy frontend
echo "[$(date)] Cloning repository and deploying frontend..."
cd /opt/cgc-lb-and-cdn-frontend
git clone https://github.com/wheeleruniverse/cgc-lb-and-cdn.git repo
cp -r repo/frontend/* .
rm -rf repo

# Build and start frontend
echo "[$(date)] Building frontend application..."
npm install
npm run build

echo "[$(date)] Starting frontend with PM2..."
pm2 start ecosystem.config.js
pm2 save
pm2 startup systemd -u root --hp /root
env PATH=$PATH:/usr/bin pm2 startup systemd -u root --hp /root | tail -n 1 | bash

# Check PM2 status
echo "[$(date)] Checking PM2 status..."
pm2 list

# ===================
# SETUP LOG UPLOAD
# ===================

# Create script to upload logs to Spaces (if credentials are available)
cat > /usr/local/bin/upload-logs.sh << 'UPLOADEOF'
#!/bin/bash
set -e

HOSTNAME=$(hostname)
TIMESTAMP=$(date +%%Y%%m%%d-%%H%%M%%S)
LOGFILE="/var/log/cgc-lb-and-cdn-deployment.log"
LAST_UPLOAD_FILE="/var/log/cgc-lb-and-cdn-last-upload.txt"

# Track last upload time
CURRENT_TIME=$(date +%%s)
if [ -f "$LAST_UPLOAD_FILE" ]; then
  LAST_UPLOAD_TIME=$(cat "$LAST_UPLOAD_FILE")
  TIME_SINCE_UPLOAD=$(( CURRENT_TIME - LAST_UPLOAD_TIME ))
else
  LAST_UPLOAD_TIME=0
  TIME_SINCE_UPLOAD=0
fi

# Create a consolidated log file
CONSOLIDATED="/tmp/cgc-lb-and-cdn-logs-${TIMESTAMP}.log"
{
  echo "================================"
  echo "Cloud Portfolio Challenge LB and CDN Log Upload"
  echo "Hostname: $HOSTNAME"
  echo "Timestamp: $(date)"
  echo "Time since last upload: ${TIME_SINCE_UPLOAD} seconds"
  echo "================================"
  echo ""

  # Check if there are new logs
  LOG_SIZE=$(wc -l < "$LOGFILE" 2>/dev/null || echo "0")

  if [ "$LOG_SIZE" -eq 0 ] && [ "$TIME_SINCE_UPLOAD" -lt 300 ]; then
    echo "ℹ️  No new deployment logs since last upload."
    echo "   This upload confirms the log monitoring system is still active."
  fi

  echo "=== Deployment Log (${LOG_SIZE} lines) ==="
  cat "$LOGFILE" 2>/dev/null || echo "No deployment log available"
  echo ""

  echo "=== Backend Service Status ==="
  systemctl status cgc-lb-and-cdn-backend.service --no-pager -l 2>/dev/null || echo "Backend service not running"
  echo ""

  echo "=== Backend Service Log (Last 100 lines) ==="
  journalctl -u cgc-lb-and-cdn-backend.service --no-pager -n 100 2>/dev/null || echo "No backend logs available"
  echo ""

  echo "=== PM2 Status ==="
  pm2 list 2>/dev/null || echo "PM2 not running"
  echo ""

  echo "=== PM2 Logs (Last 100 lines) ==="
  pm2 logs --nostream --lines 100 2>/dev/null || echo "No PM2 logs available"
  echo ""

  echo "=== Nginx Error Log (Last 50 lines) ==="
  tail -n 50 /var/log/nginx/error.log 2>/dev/null || echo "No nginx errors"
  echo ""

  echo "=== Nginx Access Log (Last 20 lines) ==="
  tail -n 20 /var/log/nginx/access.log 2>/dev/null || echo "No nginx access logs"
  echo ""

  echo "=== Health Check Test ==="
  curl -s http://localhost:8080/health 2>&1 || echo "Health check failed or backend not responding"
  echo ""

  echo "=== System Info ==="
  echo "Uptime: $(uptime)"
  echo "Load Average: $(cat /proc/loadavg)"
  echo "Memory: $(free -h | grep Mem)"
  echo "Disk: $(df -h / | tail -n 1)"
  echo "Active Connections: $(netstat -an | grep ESTABLISHED | wc -l)"
  echo ""

  echo "================================"
  echo "Log upload completed: $(date)"
  echo "================================"
} > "$CONSOLIDATED"

# Upload to Spaces if s3cmd is configured
if [ -f /root/.s3cfg ] && [ -n "${DO_SPACES_ACCESS_KEY}" ]; then
  s3cmd put "$CONSOLIDATED" "s3://${DO_SPACES_LEFT_BUCKET}/logs/${HOSTNAME}/${TIMESTAMP}.log" 2>&1 && \
    echo "✅ Logs uploaded to Spaces: s3://${DO_SPACES_LEFT_BUCKET}/logs/${HOSTNAME}/${TIMESTAMP}.log" || \
    echo "❌ Failed to upload logs to Spaces"

  # Update last upload timestamp
  echo "$CURRENT_TIME" > "$LAST_UPLOAD_FILE"
else
  echo "⚠️  Spaces credentials not configured, logs saved locally only at: $CONSOLIDATED"
fi

# Cleanup old local log files (keep last 20)
find /tmp -name "cgc-lb-and-cdn-logs-*.log" -mtime +1 -delete 2>/dev/null || true

echo "Log collection and upload complete."
UPLOADEOF

chmod +x /usr/local/bin/upload-logs.sh

# Set up cron job to upload logs based on LOG_UPLOAD_INTERVAL_MINUTES
echo "[$(date)] Setting up cron job with ${LOG_UPLOAD_INTERVAL_MINUTES} minute interval..."
echo "*/${LOG_UPLOAD_INTERVAL_MINUTES} * * * * LOG_UPLOAD_INTERVAL_MINUTES=${LOG_UPLOAD_INTERVAL_MINUTES} DO_SPACES_ACCESS_KEY=${DO_SPACES_ACCESS_KEY} DO_SPACES_SECRET_KEY=${DO_SPACES_SECRET_KEY} /usr/local/bin/upload-logs.sh >> /var/log/cgc-lb-and-cdn-log-upload.log 2>&1" | crontab -

# Verify cron job was set
echo "[$(date)] Cron job configured:"
crontab -l

# Upload initial logs
echo "[$(date)] Uploading initial logs..."
/usr/local/bin/upload-logs.sh

# ===================
# DEPLOYMENT COMPLETE
# ===================

echo "================================"
echo "[$(date)] Full-stack droplet deployment completed!"
echo "Backend API running on localhost:8080"
echo "Frontend running on localhost:3000"
echo "Nginx proxy running on port 80"
echo "Deployment logs: /var/log/cgc-lb-and-cdn-deployment.log"
echo "Upload logs: /var/log/cgc-lb-and-cdn-log-upload.log"
echo "Log upload interval: ${LOG_UPLOAD_INTERVAL_MINUTES} minutes"
echo "================================"
`,
		// All variables passed once at the beginning (13 placeholders total)
		sha,
		googleAPIKey,
		leonardoAPIKey,
		freepikAPIKey,
		useDoSpaces,
		leftBucket,
		rightBucket,
		spacesEndpoint,
		valkeyHost,
		valkeyPort,
		valkeyPassword,
		spacesAccessKey,
		spacesSecretKey,
	)
}
