#!/bin/bash
# Axis LLM Gateway - Linux Installation Script
# One-command install for Linux systems

set -e

echo "🚀 Installing Axis LLM Gateway..."

# Download latest binary
echo "📥 Downloading Axis binary..."
curl -L https://github.com/litoceans/axis/releases/latest/download/axis-linux-amd64 -o /usr/local/bin/axis
chmod +x /usr/local/bin/axis

# Create config directory
echo "⚙️  Setting up configuration..."
mkdir -p ~/.axis
axis --init

# Create systemd service
echo "📝 Creating systemd service..."
cat > /etc/systemd/system/axis.service << 'EOF'
[Unit]
Description=Axis LLM Gateway
After=network.target

[Service]
Type=simple
User=root
ExecStart=/usr/local/bin/axis serve
Restart=on-failure
RestartSec=10
Environment=PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin

[Install]
WantedBy=multi-user.target
EOF

# Enable and start service
echo "🔧 Enabling and starting Axis service..."
systemctl daemon-reload
systemctl enable axis
systemctl start axis

echo ""
echo "✅ Axis installed successfully!"
echo ""
echo "Configuration: ~/.axis/axis.yaml"
echo "Service status: systemctl status axis"
echo "Logs: journalctl -u axis -f"
echo ""
echo "Edit ~/.axis/axis.yaml and add your API keys, then restart:"
echo "  systemctl restart axis"
