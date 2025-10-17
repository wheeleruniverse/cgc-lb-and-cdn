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

		// Get Valkey recreate option from Pulumi config (optional)
		recreateValkey := cfg.Get("recreate_valkey")
		if recreateValkey == "" {
			recreateValkey = "false" // Default to not recreating
		}

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
		// By attaching to a VPC via PrivateNetworkUuid, the cluster is automatically VPC-only
		// with no public endpoint - this is the most secure configuration
		valkeyCluster, err := digitalocean.NewDatabaseCluster(ctx, "cgc-lb-and-cdn-valkey", &digitalocean.DatabaseClusterArgs{
			Name:               pulumi.String("cgc-lb-and-cdn-valkey"),
			Engine:             pulumi.String("valkey"),
			Version:            pulumi.String("8"),
			Size:               pulumi.String("db-s-1vcpu-1gb"), // Small cluster for development
			Region:             pulumi.String("nyc3"),
			NodeCount:          pulumi.Int(1), // Single node for cost efficiency
			PrivateNetworkUuid: vpc.ID(),      // VPC attachment = automatic private-only access
			// Tags removed due to permission issues - can be added manually in DO console
		})
		if err != nil {
			return err
		}

		// Note: No DatabaseFirewall needed - VPC attachment already restricts access to VPC-only
		// The cluster has no public endpoint and is only accessible from droplets in the VPC

		// Create droplets dynamically based on droplet_count
		// Each droplet runs both backend and frontend applications
		droplets := make([]*digitalocean.Droplet, dropletCount)
		for i := 0; i < dropletCount; i++ {
			logicalName := fmt.Sprintf("cgc-lb-and-cdn-droplet-%d", i+1)
			physicalName := fmt.Sprintf("cgc-lb-and-cdn-droplet-%s-%d", sha[:7], i+1)
			droplet, err := digitalocean.NewDroplet(ctx, logicalName, &digitalocean.DropletArgs{
				Name:    pulumi.String(physicalName),
				Image:   pulumi.String("ubuntu-22-04-x64"),
				Size:    pulumi.String("s-2vcpu-2gb"),
				Region:  pulumi.String("nyc3"),
				VpcUuid: vpc.ID(),
				UserData: pulumi.All(valkeyCluster.Host, valkeyCluster.Port, valkeyCluster.Password, spaceBucket.Name, spaceBucket.Region).ApplyT(func(args []interface{}) string {
					return getFullStackUserData(UserDataConfig{
						DeploymentSHA:   sha,
						GoogleAPIKey:    googleAPIKey,
						LeonardoAPIKey:  leonardoAPIKey,
						FreepikAPIKey:   freepikAPIKey,
						UseDoSpaces:     useDoSpaces,
						SpacesBucket:    args[3].(string),
						SpacesEndpoint:  spaceBucketEndpoint,
						SpacesAccessKey: spacesAccessKey,
						SpacesSecretKey: spacesSecretKey,
						ValkeyHost:      args[0].(string),
						ValkeyPort:      fmt.Sprintf("%v", args[1]),
						ValkeyPassword:  args[2].(string),
						RecreateValkey:  recreateValkey,
					})
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

			// Forward traffic to Nginx on port 80 (Nginx routes /api/ to backend:8080, everything else to frontend:3000)
			ForwardingRules: digitalocean.LoadBalancerForwardingRuleArray{
				// HTTP traffic
				&digitalocean.LoadBalancerForwardingRuleArgs{
					EntryProtocol:  pulumi.String("http"),
					EntryPort:      pulumi.Int(80),
					TargetProtocol: pulumi.String("http"),
					TargetPort:     pulumi.Int(80),
				},
				// HTTPS traffic
				&digitalocean.LoadBalancerForwardingRuleArgs{
					EntryProtocol:   pulumi.String("https"),
					EntryPort:       pulumi.Int(443),
					TargetProtocol:  pulumi.String("http"),
					TargetPort:      pulumi.Int(80),
					CertificateName: certificate.Name,
				},
			},

			// Health check configuration
			// Check Nginx on port 80 which will proxy /health to backend:8080
			// More tolerant settings to allow time for initial deployment
			Healthcheck: &digitalocean.LoadBalancerHealthcheckArgs{
				Protocol:               pulumi.String("http"),
				Port:                   pulumi.Int(80),
				Path:                   pulumi.String("/health"),
				CheckIntervalSeconds:   pulumi.Int(10),
				ResponseTimeoutSeconds: pulumi.Int(15), // Increased to 15s for slower responses
				HealthyThreshold:       pulumi.Int(2),  // Needs 2 successful checks to mark UP
				UnhealthyThreshold:     pulumi.Int(8),  // Increased to 8 (80 seconds of failures before marking DOWN)
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

// UserDataConfig holds all configuration needed to generate the cloud-init user data script
type UserDataConfig struct {
	DeploymentSHA   string
	GoogleAPIKey    string
	LeonardoAPIKey  string
	FreepikAPIKey   string
	UseDoSpaces     string
	SpacesBucket    string
	SpacesEndpoint  string
	SpacesAccessKey string
	SpacesSecretKey string
	ValkeyHost      string
	ValkeyPort      string
	ValkeyPassword  string
	RecreateValkey  string
}

// getFullStackUserData returns cloud-init script to deploy both backend and frontend on each droplet
func getFullStackUserData(config UserDataConfig) string {
	return fmt.Sprintf(`#!/bin/bash
set -e

# Deployment configuration (passed from Pulumi)
DEPLOYMENT_SHA="%s"
GOOGLE_API_KEY="%s"
LEONARDO_API_KEY="%s"
FREEPIK_API_KEY="%s"
USE_DO_SPACES="%s"
DO_SPACES_BUCKET="%s"
DO_SPACES_ENDPOINT="%s"
DO_SPACES_ACCESS_KEY="%s"
DO_SPACES_SECRET_KEY="%s"
DO_VALKEY_HOST="%s"
DO_VALKEY_PORT="%s"
DO_VALKEY_PASSWORD="%s"
RECREATE_VALKEY="%s"

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
apt-get install -y -o Dpkg::Options::="--force-confdef" -o Dpkg::Options::="--force-confold" curl wget git build-essential nginx s3cmd redis-tools

# Install DigitalOcean Metrics Agent for monitoring
echo "[$(date)] Installing DigitalOcean Metrics Agent..."
curl -sSL https://repos.insights.digitalocean.com/install.sh | bash
echo "[$(date)] ✅ Metrics agent installed - dashboard metrics now available"

# Create dedicated service user for running all services (don't use root!)
echo "[$(date)] Creating dedicated service user..."
useradd -r -s /bin/bash -d /var/lib/cgc-lb-and-cdn-service -m cgc-lb-and-cdn-service
echo "[$(date)] ✅ Service user 'cgc-lb-and-cdn-service' created"

# Configure s3cmd for DigitalOcean Spaces (for the service user)
echo "[$(date)] Configuring S3 access for Spaces..."
cat > /var/lib/cgc-lb-and-cdn-service/.s3cfg << S3CFG
[default]
host_base = ${DO_SPACES_ENDPOINT}
host_bucket = %%(bucket)s.${DO_SPACES_ENDPOINT}
access_key = ${DO_SPACES_ACCESS_KEY}
secret_key = ${DO_SPACES_SECRET_KEY}
use_https = True
S3CFG

# Configure lifecycle policy to automatically delete old logs after 7 days
echo "[$(date)] Configuring lifecycle policy for automatic log cleanup..."
cat > /tmp/lifecycle-policy.xml << 'LIFECYCLE'
<?xml version="1.0" encoding="UTF-8"?>
<LifecycleConfiguration>
    <Rule>
        <ID>delete-old-logs</ID>
        <Status>Enabled</Status>
        <Filter>
            <Prefix>logs/</Prefix>
        </Filter>
        <Expiration>
            <Days>7</Days>
        </Expiration>
    </Rule>
</LifecycleConfiguration>
LIFECYCLE

# Apply lifecycle policy (only run on first droplet to avoid conflicts)
HOSTNAME=$(hostname)
if [[ "$HOSTNAME" =~ -1$ ]]; then
  s3cmd -c /var/lib/cgc-lb-and-cdn-service/.s3cfg setlifecycle /tmp/lifecycle-policy.xml s3://${DO_SPACES_BUCKET} 2>&1 && \
    echo "[$(date)] ✅ Lifecycle policy applied - logs will auto-delete after 7 days" || \
    echo "[$(date)] ⚠️  Failed to apply lifecycle policy (may already exist)"
fi
rm -f /tmp/lifecycle-policy.xml

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
DO_SPACES_BUCKET=${DO_SPACES_BUCKET}
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
User=cgc-lb-and-cdn-service
Group=cgc-lb-and-cdn-service
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

# Set ownership of backend directory to service user
chown -R cgc-lb-and-cdn-service:cgc-lb-and-cdn-service /opt/cgc-lb-and-cdn-backend

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

# Recreate Valkey indexes from DO Spaces if requested (rebuilds from single source of truth)
if [ "${RECREATE_VALKEY}" = "true" ]; then
  echo "[$(date)] Recreating Valkey indexes from DO Spaces..."
  echo "[$(date)] This will rebuild the image pair indexes from the data stored in DO Spaces"

  # First, flush the existing Valkey data
  echo "[$(date)] Flushing existing Valkey data..."
  redis-cli -h ${DO_VALKEY_HOST} -p ${DO_VALKEY_PORT} -a ${DO_VALKEY_PASSWORD} --tls FLUSHALL && \
    echo "[$(date)] ✅ Valkey database flushed successfully" || \
    echo "[$(date)] ⚠️  Failed to flush Valkey database"

  # List all objects in the images/ prefix from DO Spaces
  echo "[$(date)] Reading image pairs from DO Spaces (bucket: ${DO_SPACES_BUCKET})..."

  # Structure: images/<provider>/<pair_id>/<side>.png
  # We need to find all unique pair_id values
  TEMP_LISTING="/tmp/spaces-listing.txt"
  s3cmd ls --recursive "s3://${DO_SPACES_BUCKET}/images/" > "$TEMP_LISTING" 2>&1 || {
    echo "[$(date)] ⚠️  Failed to list DO Spaces objects"
    echo "[$(date)] Continuing with empty Valkey - bootstrap will generate new pairs"
  }

  if [ -f "$TEMP_LISTING" ] && [ -s "$TEMP_LISTING" ]; then
    # Extract unique pair IDs from the listing
    # Expected format: 2024-01-01 12:00  12345  s3://bucket/images/provider/pair-id/side.png
    PAIR_COUNT=0

    # Process each unique pair_id
    for PROVIDER_DIR in $(s3cmd ls "s3://${DO_SPACES_BUCKET}/images/" | grep DIR | awk '{print $2}'); do
      PROVIDER_NAME=$(basename "$PROVIDER_DIR" | sed 's:/$::')
      echo "[$(date)] Processing provider: ${PROVIDER_NAME}"

      for PAIR_DIR in $(s3cmd ls "${PROVIDER_DIR}" | grep DIR | awk '{print $2}'); do
        PAIR_ID=$(basename "$PAIR_DIR" | sed 's:/$::')

        # Check if both left.png and right.png exist
        LEFT_URL="https://${DO_SPACES_BUCKET}.${DO_SPACES_ENDPOINT}/images/${PROVIDER_NAME}/${PAIR_ID}/left.png"
        RIGHT_URL="https://${DO_SPACES_BUCKET}.${DO_SPACES_ENDPOINT}/images/${PROVIDER_NAME}/${PAIR_ID}/right.png"

        # Get metadata from left image (contains prompt and other info)
        METADATA_FILE="/tmp/metadata-${PAIR_ID}.txt"
        s3cmd info "s3://${DO_SPACES_BUCKET}/images/${PROVIDER_NAME}/${PAIR_ID}/left.png" > "$METADATA_FILE" 2>&1

        # Extract prompt from metadata (stored as x-amz-meta-prompt header)
        # Use sed to trim whitespace instead of xargs to avoid quote interpretation issues
        PROMPT=$(grep -i "x-amz-meta-prompt" "$METADATA_FILE" | cut -d: -f2- | sed -e 's/^[[:space:]]*//' -e 's/[[:space:]]*$//')
        if [ -z "$PROMPT" ]; then
          PROMPT="Unknown prompt"
        fi

        # Escape double quotes in prompt for JSON
        PROMPT=$(echo "$PROMPT" | sed 's/"/\\"/g')

        # Store the pair in Valkey using the same format as the backend
        # We'll use redis-cli to store the JSON directly
        TIMESTAMP=$(date -Iseconds)
        PAIR_JSON="{\"pair_id\":\"${PAIR_ID}\",\"prompt\":\"${PROMPT}\",\"provider\":\"${PROVIDER_NAME}\",\"left_url\":\"${LEFT_URL}\",\"right_url\":\"${RIGHT_URL}\",\"timestamp\":\"${TIMESTAMP}\"}"

        # Store pair in Valkey
        redis-cli -h ${DO_VALKEY_HOST} -p ${DO_VALKEY_PORT} -a ${DO_VALKEY_PASSWORD} --tls \
          SET "pair:${PAIR_ID}" "$PAIR_JSON" >/dev/null 2>&1 && \
        redis-cli -h ${DO_VALKEY_HOST} -p ${DO_VALKEY_PORT} -a ${DO_VALKEY_PASSWORD} --tls \
          LPUSH "pairs:all" "${PAIR_ID}" >/dev/null 2>&1 && \
          PAIR_COUNT=$((PAIR_COUNT + 1))

        rm -f "$METADATA_FILE"
      done
    done

    echo "[$(date)] ✅ Recreated ${PAIR_COUNT} image pairs in Valkey from DO Spaces"
    rm -f "$TEMP_LISTING"
  else
    echo "[$(date)] ⚠️  No images found in DO Spaces, Valkey will be empty"
    echo "[$(date)] Bootstrap process will generate initial pairs"
  fi
else
  echo "[$(date)] Valkey recreation not requested, preserving existing data"
fi

# Bootstrap: Generate initial image pairs to prevent empty database
echo "[$(date)] Bootstrapping image pairs..."
for i in 1 2; do
  echo "[$(date)] Generating bootstrap image pair $i/2..."
  curl -s -X POST http://localhost:8080/api/v1/generate \
    -H "Content-Type: application/json" \
    -d "{\"prompt\": \"bootstrap-image-$i\"}" || echo "Bootstrap generation $i failed"

  # Sleep between generations to avoid rate limits
  if [ $i -lt 2 ]; then
    sleep 10
  fi
done
echo "[$(date)] Bootstrap complete - initial image pairs generated"

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

    # Proxy health check to backend
    location /health {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
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
    args: 'start -- -H 0.0.0.0 -p 3000',
    cwd: '/opt/cgc-lb-and-cdn-frontend',
    env: {
      NODE_ENV: 'production'
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

# Set ownership of frontend directory to service user
chown -R cgc-lb-and-cdn-service:cgc-lb-and-cdn-service /opt/cgc-lb-and-cdn-frontend

# Build and start frontend as service user
echo "[$(date)] Building frontend application..."
su - cgc-lb-and-cdn-service -c "cd /opt/cgc-lb-and-cdn-frontend && npm install && npm run build"

echo "[$(date)] Starting frontend with PM2..."
su - cgc-lb-and-cdn-service -c "cd /opt/cgc-lb-and-cdn-frontend && pm2 start ecosystem.config.js"

# Configure PM2 to start on boot for service user
su - cgc-lb-and-cdn-service -c "pm2 save"
env PATH=$PATH:/usr/bin /usr/bin/pm2 startup systemd -u cgc-lb-and-cdn-service --hp /var/lib/cgc-lb-and-cdn-service

# Check PM2 status
echo "[$(date)] Checking PM2 status..."
su - cgc-lb-and-cdn-service -c "pm2 list"

# Diagnostic checks
echo "[$(date)] Running diagnostic checks..."
echo "=== Listening ports ==="
netstat -tlnp | grep -E ':(80|443|3000|8080)' || echo "No services listening on expected ports!"

echo ""
echo "=== Backend health check ==="
sleep 3
curl -v http://localhost:8080/health || echo "Backend health check failed!"

echo ""
echo "=== Frontend health check ==="
curl -v http://localhost:3000 || echo "Frontend health check failed!"

echo ""
echo "=== Nginx health check ==="
curl -v http://localhost:80 || echo "Nginx health check failed!"

echo ""
echo "=== Backend logs (last 20 lines) ==="
journalctl -u cgc-lb-and-cdn-backend.service -n 20 --no-pager || true

echo ""
echo "=== Frontend logs (last 20 lines) ==="
su - cgc-lb-and-cdn-service -c "pm2 logs cgc-lb-and-cdn-frontend --lines 20 --nostream" || true

echo ""
echo "=== Nginx error log (last 20 lines) ==="
tail -n 20 /var/log/nginx/error.log || true

# ===================
# SETUP LOG UPLOAD
# ===================

# Create script to upload logs to Spaces (if credentials are available)
cat > /usr/local/bin/upload-logs.sh << 'UPLOADEOF'
#!/bin/bash
set -e

# Source environment variables for cron
if [ -f /var/lib/cgc-lb-and-cdn-service/.env-cron ]; then
  source /var/lib/cgc-lb-and-cdn-service/.env-cron
fi

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

# Set PM2_HOME for the service user
export PM2_HOME="/var/lib/cgc-lb-and-cdn-service/.pm2"

# Create a consolidated log file
CONSOLIDATED="/tmp/cgc-lb-and-cdn-logs-${TIMESTAMP}.log"
{
  echo "================================"
  echo "Cloud Portfolio Challenge LB and CDN Log Upload"
  echo "Hostname: $HOSTNAME"
  echo "Timestamp: $(date)"
  echo "Time since last upload: ${TIME_SINCE_UPLOAD} seconds"
  echo "Running as user: $(whoami)"
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
  echo "Active Connections: $(ss -tan | grep ESTAB | wc -l)"
  echo ""

  echo "================================"
  echo "Log upload completed: $(date)"
  echo "================================"
} > "$CONSOLIDATED"

# Upload to Spaces if s3cmd is configured
S3CFG_PATH="/var/lib/cgc-lb-and-cdn-service/.s3cfg"
if [ -f "$S3CFG_PATH" ]; then
  # Use the config file explicitly to ensure it's found by cron
  s3cmd -c "$S3CFG_PATH" put "$CONSOLIDATED" "s3://${DO_SPACES_BUCKET}/logs/${HOSTNAME}/${TIMESTAMP}.log" 2>&1 && \
    echo "✅ Logs uploaded to Spaces: s3://${DO_SPACES_BUCKET}/logs/${HOSTNAME}/${TIMESTAMP}.log" || \
    echo "❌ Failed to upload logs to Spaces"

  # Update last upload timestamp
  echo "$CURRENT_TIME" > "$LAST_UPLOAD_FILE"
else
  echo "⚠️  Spaces config not found at $S3CFG_PATH, logs saved locally only at: $CONSOLIDATED"
fi

# Cleanup old local log files (keep last 20)
find /tmp -name "cgc-lb-and-cdn-logs-*.log" -mtime +1 -delete 2>/dev/null || true

echo "Log collection and upload complete."
UPLOADEOF

chmod +x /usr/local/bin/upload-logs.sh

# Set up cron jobs
echo "[$(date)] Setting up cron jobs..."

# Create script to generate image pairs automatically with distributed locking
cat > /usr/local/bin/generate-images.sh << 'GENEOF'
#!/bin/bash
# Automatically generate image pairs to pre-populate the database
# Uses distributed locking via Valkey to prevent multiple droplets from generating simultaneously
# This ensures:
#   1. Only one droplet generates images at a time (prevents API quota exhaustion)
#   2. No duplicate pairs are created
#   3. Valkey isn't overwhelmed with concurrent writes

# Source environment variables for cron
if [ -f /var/lib/cgc-lb-and-cdn-service/.env-cron ]; then
  source /var/lib/cgc-lb-and-cdn-service/.env-cron
fi

LOGFILE="/var/log/cgc-lb-and-cdn-image-generation.log"
TIMESTAMP=$(date '+%Y-%m-%d %H:%M:%S')
LOCK_KEY="image-generation-lock"
LOCK_TTL=300  # Lock expires after 5 minutes (in case process dies)
HOSTNAME=$(hostname)

# Function to acquire distributed lock using Valkey
acquire_lock() {
  # Use redis-cli with Valkey credentials to implement SETNX (SET if Not eXists)
  # Returns 1 if lock acquired, 0 if already held by another droplet
  redis-cli -h ${DO_VALKEY_HOST} -p ${DO_VALKEY_PORT} -a ${DO_VALKEY_PASSWORD} --tls \
    SET "$LOCK_KEY" "$HOSTNAME" NX EX $LOCK_TTL 2>/dev/null | grep -q "OK"
  return $?
}

# Function to release distributed lock
release_lock() {
  # Only release if we own the lock (check value matches our hostname)
  LOCK_OWNER=$(redis-cli -h ${DO_VALKEY_HOST} -p ${DO_VALKEY_PORT} -a ${DO_VALKEY_PASSWORD} --tls \
    GET "$LOCK_KEY" 2>/dev/null)

  if [ "$LOCK_OWNER" = "$HOSTNAME" ]; then
    redis-cli -h ${DO_VALKEY_HOST} -p ${DO_VALKEY_PORT} -a ${DO_VALKEY_PASSWORD} --tls \
      DEL "$LOCK_KEY" >/dev/null 2>&1
  fi
}

# Add random jitter (0-30 seconds) to prevent simultaneous lock attempts
JITTER=$((RANDOM % 30))
echo "[$TIMESTAMP] Waiting ${JITTER}s jitter before attempting lock..." >> "$LOGFILE"
sleep $JITTER

# Try to acquire the distributed lock
if acquire_lock; then
  echo "[$TIMESTAMP] 🔒 Lock acquired by $HOSTNAME, starting image generation..." >> "$LOGFILE"

  # Ensure lock is released on exit (even if script fails)
  trap release_lock EXIT

  # Call the backend API to generate a new image pair
  RESPONSE=$(curl -s -w "\n%{http_code}" -X POST http://localhost:8080/api/v1/generate \
    -H "Content-Type: application/json" \
    -d '{"prompt": "auto-generated"}' 2>&1)

  HTTP_CODE=$(echo "$RESPONSE" | tail -n 1)
  BODY=$(echo "$RESPONSE" | head -n -1)

  if [ "$HTTP_CODE" = "200" ]; then
    PAIR_ID=$(echo "$BODY" | grep -o '"pair_id":"[^"]*"' | cut -d'"' -f4)
    PROVIDER=$(echo "$BODY" | grep -o '"provider":"[^"]*"' | cut -d'"' -f4)
    echo "[$TIMESTAMP] ✅ Successfully generated image pair: $PAIR_ID (Provider: $PROVIDER)" >> "$LOGFILE"
  else
    echo "[$TIMESTAMP] ❌ Failed to generate images (HTTP $HTTP_CODE): $BODY" >> "$LOGFILE"
  fi

  # Lock will be released by trap on EXIT
else
  echo "[$TIMESTAMP] ⏭️  Lock held by another droplet, skipping this run" >> "$LOGFILE"
fi
GENEOF

chmod +x /usr/local/bin/generate-images.sh

# Create environment file that scripts can source
cat > /var/lib/cgc-lb-and-cdn-service/.env-cron << ENVEOF
export DO_SPACES_BUCKET="${DO_SPACES_BUCKET}"
export DO_VALKEY_HOST="${DO_VALKEY_HOST}"
export DO_VALKEY_PORT="${DO_VALKEY_PORT}"
export DO_VALKEY_PASSWORD="${DO_VALKEY_PASSWORD}"
ENVEOF
chown cgc-lb-and-cdn-service:cgc-lb-and-cdn-service /var/lib/cgc-lb-and-cdn-service/.env-cron
chmod 600 /var/lib/cgc-lb-and-cdn-service/.env-cron

# Create cron job file
cat > /etc/cron.d/cgc-lb-and-cdn << CRONEOF
# Cron jobs for cgc-lb-and-cdn service
SHELL=/bin/bash
PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin

# Upload logs every ${LOG_UPLOAD_INTERVAL_MINUTES} minutes
*/${LOG_UPLOAD_INTERVAL_MINUTES} * * * * cgc-lb-and-cdn-service /usr/local/bin/upload-logs.sh >> /var/log/cgc-lb-and-cdn-log-upload.log 2>&1

# Generate image pairs every 15 minutes to keep database populated
*/15 * * * * cgc-lb-and-cdn-service /usr/local/bin/generate-images.sh

# Generate extra images during peak hours (9 AM - 9 PM) every 5 minutes
*/5 9-21 * * * cgc-lb-and-cdn-service /usr/local/bin/generate-images.sh
CRONEOF

# Set proper permissions on cron file
chmod 0644 /etc/cron.d/cgc-lb-and-cdn

echo "[$(date)] ✅ Cron jobs configured in /etc/cron.d/cgc-lb-and-cdn"

# Verify cron jobs were set
echo "[$(date)] Cron job file contents:"
cat /etc/cron.d/cgc-lb-and-cdn

# Set proper ownership for log files so service user can access them
chown cgc-lb-and-cdn-service:cgc-lb-and-cdn-service /var/log/cgc-lb-and-cdn-deployment.log
chmod 644 /var/log/cgc-lb-and-cdn-deployment.log

# Upload initial logs as the service user
echo "[$(date)] Uploading initial logs..."
su - cgc-lb-and-cdn-service -c "/usr/local/bin/upload-logs.sh"

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
echo ""
echo "Services configured to auto-start on boot:"
echo "  ✓ cgc-lb-and-cdn-backend.service (systemd)"
echo "  ✓ nginx.service (systemd)"
echo "  ✓ cgc-lb-and-cdn-frontend (PM2)"
echo ""
echo "[$(date)] Rebooting to apply kernel updates and verify auto-start..."
echo "================================"

# Reboot to apply updates and verify services auto-start
# Services will automatically start on boot:
# - Backend: systemctl enabled
# - Nginx: systemctl enabled
# - Frontend: PM2 startup configured
shutdown -r +1 "Rebooting to apply system updates and verify service auto-start"
`,
		// All variables passed from config struct (13 placeholders total)
		config.DeploymentSHA,
		config.GoogleAPIKey,
		config.LeonardoAPIKey,
		config.FreepikAPIKey,
		config.UseDoSpaces,
		config.SpacesBucket,
		config.SpacesEndpoint,
		config.SpacesAccessKey,
		config.SpacesSecretKey,
		config.ValkeyHost,
		config.ValkeyPort,
		config.ValkeyPassword,
		config.RecreateValkey,
	)
}
