#!/bin/bash
# Axis LLM Gateway - macOS Installation Script
# One-command install for macOS systems

set -e

echo "🚀 Installing Axis LLM Gateway..."

# Check if Homebrew is installed
if ! command -v brew &> /dev/null; then
    echo "🍺 Homebrew not found. Installing Homebrew..."
    /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
fi

# Try to install via Homebrew first
echo "📥 Installing via Homebrew..."
if brew install axis 2>/dev/null; then
    echo "✅ Installed via Homebrew!"
else
    echo "📥 Homebrew install failed, downloading binary directly..."
    # Download latest binary
    curl -L https://github.com/litoceans/axis/releases/latest/download/axis-darwin-amd64 -o /usr/local/bin/axis
    chmod +x /usr/local/bin/axis
fi

# Create config directory
echo "⚙️  Setting up configuration..."
mkdir -p ~/.axis
axis --init

echo ""
echo "✅ Axis installed successfully!"
echo ""
echo "Configuration: ~/.axis/axis.yaml"
echo ""
echo "Run Axis:"
echo "  axis serve"
echo ""
echo "Edit ~/.axis/axis.yaml and add your API keys before starting."
