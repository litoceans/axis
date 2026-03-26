#!/bin/bash
# Axis LLM Gateway - Universal Installation Script
# Auto-detects OS and runs appropriate installer

set -e

echo "🚀 Axis LLM Gateway Installer"
echo ""

# Detect OS
OS="$(uname -s)"

case "$OS" in
    Linux)
        echo "🐧 Detected Linux"
        echo ""
        # Download and run Linux installer
        curl -fsSL https://raw.githubusercontent.com/litoceans/axis/main/scripts/install-linux.sh | bash
        ;;
    Darwin)
        echo "🍎 Detected macOS"
        echo ""
        # Download and run macOS installer
        curl -fsSL https://raw.githubusercontent.com/litoceans/axis/main/scripts/install-macos.sh | bash
        ;;
    *)
        echo "❌ Unsupported OS: $OS"
        echo ""
        echo "Supported platforms:"
        echo "  - Linux (x86_64)"
        echo "  - macOS (Intel/Apple Silicon)"
        echo "  - Windows (use install-windows.ps1)"
        exit 1
        ;;
esac
