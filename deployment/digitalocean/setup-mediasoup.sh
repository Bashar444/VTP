#!/bin/bash
# ═══════════════════════════════════════════════════════════════════════════════
# VTP Platform - MediaSoup SFU Server Setup
# Run this on a DigitalOcean Droplet (Ubuntu 22.04 LTS)
#
# Recommended Droplet: Basic $12/month (1 vCPU, 2GB RAM)
# Region: Same as App Platform (e.g., NYC1)
#
# Usage:
#   ssh root@your-droplet-ip
#   curl -sSL https://raw.githubusercontent.com/Bashar444/VTP/main/deployment/digitalocean/setup-mediasoup.sh | bash
# ═══════════════════════════════════════════════════════════════════════════════

set -e

echo "═══════════════════════════════════════════════════════════════════════════"
echo "  VTP Platform - MediaSoup SFU Server Setup"
echo "═══════════════════════════════════════════════════════════════════════════"

# Get public IP
PUBLIC_IP=$(curl -s http://169.254.169.254/metadata/v1/interfaces/public/0/ipv4/address 2>/dev/null || curl -s ifconfig.me)
echo "Public IP: $PUBLIC_IP"

# Update system
echo ""
echo "[1/7] Updating system packages..."
apt-get update && apt-get upgrade -y

# Install Node.js 20 LTS
echo ""
echo "[2/7] Installing Node.js 20 LTS..."
curl -fsSL https://deb.nodesource.com/setup_20.x | bash -
apt-get install -y nodejs

# Install build dependencies for MediaSoup
echo ""
echo "[3/7] Installing build dependencies..."
apt-get install -y python3 python3-pip build-essential git

# Create app directory
echo ""
echo "[4/7] Cloning VTP repository..."
mkdir -p /opt/vtp
cd /opt/vtp
git clone https://github.com/Bashar444/VTP.git .

# Install MediaSoup dependencies
echo ""
echo "[5/7] Installing MediaSoup dependencies..."
cd /opt/vtp/mediasoup-sfu
npm install

# Create systemd service
echo ""
echo "[6/7] Creating systemd service..."
cat > /etc/systemd/system/mediasoup.service << EOF
[Unit]
Description=VTP MediaSoup SFU Server
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=/opt/vtp/mediasoup-sfu
ExecStart=/usr/bin/node src/index.js
Restart=always
RestartSec=10
Environment=NODE_ENV=production
Environment=PORT=3000
Environment=MEDIASOUP_LISTEN_IP=0.0.0.0
Environment=MEDIASOUP_ANNOUNCED_IP=$PUBLIC_IP
Environment=RTC_MIN_PORT=40000
Environment=RTC_MAX_PORT=49999
Environment=LOG_LEVEL=info

[Install]
WantedBy=multi-user.target
EOF

# Enable and start service
systemctl daemon-reload
systemctl enable mediasoup
systemctl start mediasoup

# Configure firewall
echo ""
echo "[7/7] Configuring firewall..."
ufw --force enable
ufw allow 22/tcp      # SSH
ufw allow 3000/tcp    # MediaSoup HTTP API
ufw allow 40000:49999/udp  # WebRTC media (UDP)
ufw allow 40000:49999/tcp  # WebRTC media (TCP fallback)

echo ""
echo "═══════════════════════════════════════════════════════════════════════════"
echo "  ✅ MediaSoup SFU Setup Complete!"
echo "═══════════════════════════════════════════════════════════════════════════"
echo ""
echo "  Server Status:  $(systemctl is-active mediasoup)"
echo "  Public IP:      $PUBLIC_IP"
echo "  HTTP API:       http://$PUBLIC_IP:3000"
echo "  WebSocket:      ws://$PUBLIC_IP:3000"
echo ""
echo "  Add to your .env:"
echo "  MEDIASOUP_URL=http://$PUBLIC_IP:3000"
echo ""
echo "  View logs:      journalctl -u mediasoup -f"
echo "  Restart:        systemctl restart mediasoup"
echo ""
echo "═══════════════════════════════════════════════════════════════════════════"
