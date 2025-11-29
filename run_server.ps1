# Run server with full output capture
$ErrorActionPreference = "Continue"

Write-Host "Starting VTP Server..."
Write-Host "Working directory: $(Get-Location)"

# Change to app directory
cd 'c:\Users\Admin\Desktop\VTP'

# Run the executable
& '.\bin\vtp'

Write-Host "Server exited"
