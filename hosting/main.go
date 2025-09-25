package main

import (
	"fmt"

	"github.com/pulumi/pulumi-digitalocean/sdk/v4/go/digitalocean"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		// Load configuration
		cfg := config.New(ctx, "")
		_ = cfg.Get("domain") // Domain configuration for future SSL setup

		// Get API keys from Pulumi config (passed from GitHub Actions)
		googleAPIKey := cfg.Get("google_api_key")
		leonardoAPIKey := cfg.Get("leonardo_api_key")
		freepikAPIKey := cfg.Get("freepik_api_key")
		useDoSpaces := cfg.Get("use_do_spaces")
		if useDoSpaces == "" {
			useDoSpaces = "false" // Default to local storage
		}

		// Note: Spaces bucket creation requires separate Spaces credentials
		// For now, we'll create a placeholder and configure Spaces manually
		spaceBucketName := "cgc-lb-and-cdn-content"
		spaceBucketEndpoint := "nyc3.digitaloceanspaces.com"

		// Create VPC for the project
		vpc, err := digitalocean.NewVpc(ctx, "cgc-lb-and-cdn-vpc", &digitalocean.VpcArgs{
			Name:    pulumi.String("cgc-lb-and-cdn-vpc"),
			Region:  pulumi.String("nyc3"),
			IpRange: pulumi.String("10.10.0.0/16"),
		})
		if err != nil {
			return err
		}

		// Create Valkey managed database cluster for caching user votes
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

		// Backend droplet for Go API server
		backendDroplet, err := digitalocean.NewDroplet(ctx, "cgc-lb-and-cdn-backend", &digitalocean.DropletArgs{
			Name:    pulumi.String("cgc-lb-and-cdn-backend"),
			Image:   pulumi.String("ubuntu-22-04-x64"),
			Size:    pulumi.String("s-1vcpu-1gb"),
			Region:  pulumi.String("nyc3"),
			VpcUuid: vpc.ID(),
			UserData: pulumi.All(valkeyCluster.Host, valkeyCluster.Port, valkeyCluster.Password).ApplyT(func(args []interface{}) string {
				return getBackendUserData(googleAPIKey, leonardoAPIKey, freepikAPIKey, useDoSpaces, spaceBucketName, spaceBucketEndpoint, args[0].(string), fmt.Sprintf("%v", args[1]), args[2].(string))
			}).(pulumi.StringOutput),
			// Tags removed due to permission issues
		})
		if err != nil {
			return err
		}

		// Frontend droplet for Next.js application
		frontendDroplet, err := digitalocean.NewDroplet(ctx, "cgc-lb-and-cdn-frontend", &digitalocean.DropletArgs{
			Name:    pulumi.String("cgc-lb-and-cdn-frontend"),
			Image:   pulumi.String("ubuntu-22-04-x64"),
			Size:    pulumi.String("s-1vcpu-1gb"),
			Region:  pulumi.String("nyc3"),
			VpcUuid: vpc.ID(),
			UserData: backendDroplet.Ipv4AddressPrivate.ApplyT(func(ip string) string {
				return getFrontendUserData(ip)
			}).(pulumi.StringOutput),
			// Tags removed due to permission issues
		})
		if err != nil {
			return err
		}

		// Load balancer to distribute traffic
		loadBalancer, err := digitalocean.NewLoadBalancer(ctx, "cgc-lb-and-cdn-lb", &digitalocean.LoadBalancerArgs{
			Name:   pulumi.String("cgc-lb-and-cdn-lb"),
			Region: pulumi.String("nyc3"),
			Size:   pulumi.String("lb-small"),

			// Forward HTTP traffic to backend
			ForwardingRules: digitalocean.LoadBalancerForwardingRuleArray{
				// API traffic to backend
				&digitalocean.LoadBalancerForwardingRuleArgs{
					EntryProtocol:  pulumi.String("http"),
					EntryPort:      pulumi.Int(80),
					TargetProtocol: pulumi.String("http"),
					TargetPort:     pulumi.Int(8080),
				},
			},

			// Health check configuration
			Healthcheck: &digitalocean.LoadBalancerHealthcheckArgs{
				Protocol:               pulumi.String("http"),
				Port:                   pulumi.Int(8080),
				Path:                   pulumi.String("/health"),
				CheckIntervalSeconds:   pulumi.Int(10),
				ResponseTimeoutSeconds: pulumi.Int(5),
				HealthyThreshold:       pulumi.Int(3),
				UnhealthyThreshold:     pulumi.Int(3),
			},

			// Sticky sessions
			StickySessions: &digitalocean.LoadBalancerStickySessionsArgs{
				Type:             pulumi.String("cookies"),
				CookieName:       pulumi.String("lb"),
				CookieTtlSeconds: pulumi.Int(300),
			},

			// Droplet tag removed - will need to manually assign droplets to load balancer
		})
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
						pulumi.String("10.10.0.0/16"), // VPC only
					},
				},
				// Frontend Next.js port
				&digitalocean.FirewallInboundRuleArgs{
					Protocol:  pulumi.String("tcp"),
					PortRange: pulumi.String("3000"),
					SourceAddresses: pulumi.StringArray{
						pulumi.String("10.10.0.0/16"), // VPC only
					},
				},
				// Valkey database port (internal VPC access only)
				&digitalocean.FirewallInboundRuleArgs{
					Protocol:  pulumi.String("tcp"),
					PortRange: pulumi.String("25061"), // Standard Valkey port in DO
					SourceAddresses: pulumi.StringArray{
						pulumi.String("10.10.0.0/16"), // VPC only
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

			// Firewall rules will apply broadly - specific droplet targeting removed due to tag permissions
		}, pulumi.DependsOn([]pulumi.Resource{backendDroplet, frontendDroplet}))
		if err != nil {
			return err
		}

		// Export important information
		ctx.Export("loadBalancerIp", loadBalancer.Ip)
		ctx.Export("backendDropletIp", backendDroplet.Ipv4Address)
		ctx.Export("frontendDropletIp", frontendDroplet.Ipv4Address)
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

// getBackendUserData returns the cloud-init script for the backend droplet
func getBackendUserData(googleAPIKey, leonardoAPIKey, freepikAPIKey, useDoSpaces, spacesBucket, spacesEndpoint, valkeyHost, valkeyPort, valkeyPassword string) string {
	return fmt.Sprintf(`#!/bin/bash
set -e

# Update system
apt-get update -y
apt-get upgrade -y

# Install required packages
apt-get install -y curl wget git build-essential

# Install Go 1.21
cd /tmp
wget https://go.dev/dl/go1.21.0.linux-amd64.tar.gz
tar -C /usr/local -xzf go1.21.0.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> /etc/profile
echo 'export GOPATH=/opt/go' >> /etc/profile
echo 'export PATH=$PATH:/opt/go/bin' >> /etc/profile

# Create app directory
mkdir -p /opt/cgc-backend
cd /opt/cgc-backend

# Create environment file for the backend service
cat > /opt/cgc-backend/.env << 'ENVEOF'
PORT=8080
HOST=0.0.0.0
GOOGLE_API_KEY=%s
LEONARDO_API_KEY=%s
FREEPIK_API_KEY=%s
USE_DO_SPACES=%s
DO_SPACES_BUCKET=%s
DO_SPACES_ENDPOINT=%s
DO_VALKEY_HOST=%s
DO_VALKEY_PORT=%s
DO_VALKEY_PASSWORD=%s
ENVEOF

# Create systemd service file
cat > /etc/systemd/system/cgc-backend.service << 'EOF'
[Unit]
Description=CGC Backend Service
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=/opt/cgc-backend
Environment=PATH=/usr/local/go/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin
EnvironmentFile=/opt/cgc-backend/.env
ExecStart=/opt/cgc-backend/server
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
EOF

# Enable the service
systemctl enable cgc-backend.service

# Note: The actual application deployment will need to be done separately
# This could be done via CI/CD pipeline or manual deployment
echo "Backend droplet setup completed. Deploy your Go application to /opt/cgc-backend/"
echo "Remember to start the service: systemctl start cgc-backend.service"
`, googleAPIKey, leonardoAPIKey, freepikAPIKey, useDoSpaces, spacesBucket, spacesEndpoint, valkeyHost, valkeyPort, valkeyPassword)
}

// getFrontendUserData returns the cloud-init script for the frontend droplet
func getFrontendUserData(backendPrivateIP string) string {
	return fmt.Sprintf(`#!/bin/bash
set -e

# Update system
apt-get update -y
apt-get upgrade -y

# Install required packages
apt-get install -y curl wget git nginx

# Install Node.js 18
curl -fsSL https://deb.nodesource.com/setup_18.x | sudo -E bash -
apt-get install -y nodejs

# Install PM2 for process management
npm install -g pm2

# Create app directory
mkdir -p /opt/cgc-frontend
cd /opt/cgc-frontend

# Configure Nginx as reverse proxy
cat > /etc/nginx/sites-available/cgc-frontend << 'EOF'
server {
    listen 80;
    server_name _;

    # Serve static files directly
    location /_next/static/ {
        alias /opt/cgc-frontend/.next/static/;
        expires 1y;
        access_log off;
    }

    # Proxy API requests to backend
    location /api/ {
        proxy_pass http://%s:8080;
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
EOF

# Enable the site
ln -sf /etc/nginx/sites-available/cgc-frontend /etc/nginx/sites-enabled/
rm -f /etc/nginx/sites-enabled/default

# Create PM2 ecosystem file
cat > /opt/cgc-frontend/ecosystem.config.js << 'EOF'
module.exports = {
  apps: [{
    name: 'cgc-frontend',
    script: 'npm',
    args: 'start',
    cwd: '/opt/cgc-frontend',
    env: {
      NODE_ENV: 'production',
      PORT: 3000
    }
  }]
}
EOF

# Enable services
systemctl enable nginx
systemctl start nginx

# Note: The actual application deployment will need to be done separately
# This could be done via CI/CD pipeline or manual deployment
echo "Frontend droplet setup completed. Deploy your Next.js application to /opt/cgc-frontend/"
echo "Remember to:"
echo "1. Update BACKEND_IP in nginx config"
echo "2. Run: npm install && npm run build"
echo "3. Start PM2: pm2 start ecosystem.config.js && pm2 save && pm2 startup"
`, backendPrivateIP)
}
