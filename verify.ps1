# Attendance Tracker - Download Verification Script
# Run this script to verify the authenticity of your downloaded executable

param (
    [Parameter(Mandatory=$true)]
    [string]$ExecutablePath,
    
    [string]$ChecksumFile = ""
)

# Function to calculate SHA-256 hash
function Get-FileHash256($filePath) {
    $hasher = [System.Security.Cryptography.SHA256]::Create()
    $stream = [System.IO.File]::OpenRead($filePath)
    $hash = $hasher.ComputeHash($stream)
    $stream.Close()
    $hasher.Dispose()
    return [BitConverter]::ToString($hash).Replace("-", "").ToLower()
}

# Check if the executable exists
if (-not (Test-Path $ExecutablePath)) {
    Write-Error "Executable not found at path: $ExecutablePath"
    exit 1
}

# Verify the file's digital signature
Write-Host "Checking digital signature..." -ForegroundColor Yellow
$sig = Get-AuthenticodeSignature -FilePath $ExecutablePath
$status = $sig.Status

if ($status -eq "Valid") {
    Write-Host "√ Digital signature is valid." -ForegroundColor Green
    Write-Host "  Signed by: $($sig.SignerCertificate.Subject)" -ForegroundColor Green
} elseif ($status -eq "UnknownError") {
    Write-Host "! No digital signature found." -ForegroundColor Yellow
    Write-Host "  This is expected for open-source releases." -ForegroundColor Yellow
} else {
    Write-Host "✗ Digital signature issue: $status" -ForegroundColor Red
}

# Calculate and verify SHA-256 checksum
Write-Host ""
Write-Host "Calculating SHA-256 checksum..." -ForegroundColor Yellow
$calculatedHash = Get-FileHash256 $ExecutablePath
Write-Host "Calculated SHA-256: $calculatedHash" -ForegroundColor Cyan

# Check against provided checksum file if available
if ($ChecksumFile -and (Test-Path $ChecksumFile)) {
    $expectedHash = Get-Content $ChecksumFile -Raw
    $expectedHash = $expectedHash.Substring(0, 64).Trim().ToLower()
    
    Write-Host "Expected SHA-256:  $expectedHash" -ForegroundColor Cyan
    
    if ($calculatedHash -eq $expectedHash) {
        Write-Host "√ SHA-256 checksum MATCHES!" -ForegroundColor Green
    } else {
        Write-Host "✗ SHA-256 checksum does NOT match!" -ForegroundColor Red
        Write-Host "  The downloaded file may be corrupted or tampered with." -ForegroundColor Red
    }
} else {
    Write-Host ""
    Write-Host "To verify against the official checksum, download the .sha256 file from the GitHub release" -ForegroundColor Yellow
    Write-Host "and run this script with the -ChecksumFile parameter." -ForegroundColor Yellow
}

# Verify file properties
Write-Host ""
Write-Host "File Properties:" -ForegroundColor Yellow
$fileInfo = Get-Item $ExecutablePath
Write-Host "  Name: $($fileInfo.Name)"
Write-Host "  Size: $([Math]::Round($fileInfo.Length / 1KB, 2)) KB"
Write-Host "  Created: $($fileInfo.CreationTime)"
Write-Host "  Modified: $($fileInfo.LastWriteTime)"

# Display guidance for SmartScreen warning
Write-Host ""
Write-Host "Windows SmartScreen Information:" -ForegroundColor Yellow
Write-Host "When running this application for the first time, you may see a Windows SmartScreen warning."
Write-Host "This is normal for open-source applications without a paid code signing certificate."
Write-Host ""
Write-Host "To run the application:" -ForegroundColor Cyan
Write-Host "1. Click 'More info' when the SmartScreen dialog appears"
Write-Host "2. Click 'Run anyway'"
Write-Host ""
Write-Host "If you're concerned about security, you can verify the source code at:"
Write-Host "https://github.com/USERNAME/attendance-tracker" -ForegroundColor Cyan 
