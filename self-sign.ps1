# Self-signing script for Windows
# This is for DEVELOPMENT USE ONLY - not a replacement for a proper code signing certificate

param (
    [Parameter(Mandatory=$true)]
    [string]$ExecutablePath,
    
    [string]$PublisherName = "Rashid Pathiyil"
)

# Check if the executable exists
if (-not (Test-Path $ExecutablePath)) {
    Write-Error "Executable not found at path: $ExecutablePath"
    exit 1
}

# Create a self-signed certificate if it doesn't exist
$certSubject = "CN=$PublisherName"
$cert = Get-ChildItem Cert:\CurrentUser\My -CodeSigningCert | Where-Object { $_.Subject -eq $certSubject } | Select-Object -First 1

if (-not $cert) {
    Write-Host "Creating a new self-signed code signing certificate for $PublisherName..."
    $cert = New-SelfSignedCertificate -Subject $certSubject -Type CodeSigningCert -CertStoreLocation Cert:\CurrentUser\My
    
    if (-not $cert) {
        Write-Error "Failed to create self-signed certificate."
        exit 1
    }
    
    Write-Host "Certificate created with thumbprint: $($cert.Thumbprint)"
} else {
    Write-Host "Using existing certificate with thumbprint: $($cert.Thumbprint)"
}

# Sign the executable
Write-Host "Signing executable: $ExecutablePath"
$timeStampServer = "http://timestamp.digicert.com"

try {
    Set-AuthenticodeSignature -FilePath $ExecutablePath -Certificate $cert -TimestampServer $timeStampServer
    Write-Host "Executable signed successfully!"
    
    # Display verification details
    $signature = Get-AuthenticodeSignature -FilePath $ExecutablePath
    Write-Host "Signature Status: $($signature.Status)"
    Write-Host "Signed By: $($signature.SignerCertificate.Subject)"
} catch {
    Write-Error "Signing failed: $_"
    exit 1
}

# Remind user about the security warning
Write-Host ""
Write-Host "IMPORTANT: This is a self-signed certificate and will still show security warnings in Windows."
Write-Host "For a production application, consider purchasing a code signing certificate from a trusted CA." 
