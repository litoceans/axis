# Axis LLM Gateway - Windows Installation Script
# One-command install for Windows systems

Write-Host "🚀 Installing Axis LLM Gateway..." -ForegroundColor Green

# Create installation directory
$installDir = "$env:LOCALAPPDATA\axis"
New-Item -ItemType Directory -Force -Path $installDir | Out-Null

# Download latest binary
Write-Host "📥 Downloading Axis binary..." -ForegroundColor Cyan
$downloadUrl = "https://github.com/litoceans/axis/releases/latest/download/axis-windows-amd64.exe"
$outputPath = "$installDir\axis.exe"

try {
    Invoke-WebRequest -Uri $downloadUrl -OutFile $outputPath
    Write-Host "✅ Download complete!" -ForegroundColor Green
} catch {
    Write-Host "❌ Download failed: $_" -ForegroundColor Red
    exit 1
}

# Add to PATH if not already present
$currentPath = [Environment]::GetEnvironmentVariable("Path", "User")
if ($currentPath -notlike "*$installDir*") {
    Write-Host "📝 Adding to PATH..." -ForegroundColor Cyan
    [Environment]::SetEnvironmentVariable("Path", "$currentPath;$installDir", "User")
    $env:Path = [Environment]::GetEnvironmentVariable("Path", "User") + ";" + [Environment]::GetEnvironmentVariable("Path", "Machine")
}

# Create config directory
Write-Host "⚙️  Setting up configuration..." -ForegroundColor Cyan
$configDir = "$env:USERPROFILE\.axis"
New-Item -ItemType Directory -Force -Path $configDir | Out-Null

# Initialize config
Write-Host "🔧 Initializing configuration..." -ForegroundColor Cyan
& "$outputPath" --init

Write-Host ""
Write-Host "✅ Axis installed successfully!" -ForegroundColor Green
Write-Host ""
Write-Host "Configuration: $configDir\axis.yaml"
Write-Host ""
Write-Host "Run Axis:"
Write-Host "  axis serve"
Write-Host ""
Write-Host "Edit $configDir\axis.yaml and add your API keys before starting." -ForegroundColor Yellow
