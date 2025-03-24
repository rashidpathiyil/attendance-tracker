# Test update checking and download process for Attendance Tracker
#
# This script simulates a server-side check for updates by:
# 1. Starting a local HTTP server to serve a version manifest
# 2. Setting an environment variable to point the app to the local server
# 3. Running the app with special flags to force an update check

param (
    [string]$AppPath = "Attendance Tracker.exe",
    [string]$TestVersion = "1.1.0-test"
)

# Check if the app exists
if (-not (Test-Path $AppPath)) {
    Write-Error "Application not found at path: $AppPath"
    exit 1
}

# Create temp directory for test files
$tempDir = [System.IO.Path]::GetTempPath() + [System.IO.Path]::GetRandomFileName()
New-Item -ItemType Directory -Path $tempDir -Force | Out-Null

# Create a version manifest JSON file
$manifestPath = "$tempDir\version.json"
$manifest = @{
    version = $TestVersion
    download_url = "http://localhost:8080/download/attendance-tracker-$TestVersion.exe"
    release_date = (Get-Date).ToString("yyyy-MM-dd")
    release_notes = @(
        "Test update mechanism",
        "Verify download progress works",
        "Confirm installation process functions properly"
    )
    size = 25600000  # About 25 MB
}

# Convert to JSON and save to file
$manifest | ConvertTo-Json | Out-File -FilePath $manifestPath -Encoding utf8

# Create a dummy update file 
$dummyUpdatePath = "$tempDir\attendance-tracker-$TestVersion.exe"
# Create a file with random content to simulate the executable
$randomBytes = New-Object byte[] 1024
$rng = New-Object System.Security.Cryptography.RNGCryptoServiceProvider
$rng.GetBytes($randomBytes)
[System.IO.File]::WriteAllBytes($dummyUpdatePath, $randomBytes)

# Start a simple HTTP server in the background
Write-Host "Starting local update server on port 8080..."
$server = Start-Job -ScriptBlock {
    param($dir)
    $listener = New-Object System.Net.HttpListener
    $listener.Prefixes.Add("http://localhost:8080/")
    $listener.Start()
    
    Write-Host "Server started. Press Ctrl+C to stop."
    
    while ($listener.IsListening) {
        $context = $listener.GetContext()
        $request = $context.Request
        $response = $context.Response
        
        $path = $request.Url.LocalPath
        Write-Host "Request: $path"
        
        if ($path -eq "/version") {
            $content = Get-Content -Path "$dir\version.json" -Raw
            $buffer = [System.Text.Encoding]::UTF8.GetBytes($content)
            $response.ContentLength64 = $buffer.Length
            $response.OutputStream.Write($buffer, 0, $buffer.Length)
            $response.ContentType = "application/json"
        }
        elseif ($path -like "/download/*") {
            $fileName = Split-Path -Leaf $path
            $filePath = "$dir\$fileName"
            
            if (Test-Path $filePath) {
                $content = Get-Content -Path $filePath -Raw -Encoding Byte
                $response.ContentLength64 = $content.Length
                $response.OutputStream.Write($content, 0, $content.Length)
                $response.ContentType = "application/octet-stream"
            }
            else {
                $response.StatusCode = 404
            }
        }
        else {
            $response.StatusCode = 404
        }
        
        $response.Close()
    }
    
    $listener.Stop()
} -ArgumentList $tempDir

# Wait for the server to start
Start-Sleep -Seconds 2

# Set environment variable to override update server
$env:ATTENDANCE_UPDATE_SERVER = "http://localhost:8080/version"

# Run the application with a test flag to force update check
Write-Host "Starting Attendance Tracker with test update server..."
Write-Host "Application should check for updates immediately and find version $TestVersion"
Start-Process -FilePath $AppPath -ArgumentList "--test-updates"

# Wait for the user to finish testing
Write-Host ""
Write-Host "The application should now show update available notifications and allow you to test the download process."
Write-Host "Press any key to stop the test server and clean up..." -ForegroundColor Yellow
$null = $host.UI.RawUI.ReadKey("NoEcho,IncludeKeyDown")

# Clean up
Stop-Job -Job $server
Remove-Job -Job $server
Remove-Item -Path $tempDir -Recurse -Force

Write-Host "Test completed and cleaned up." 
