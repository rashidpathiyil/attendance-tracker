# Security Information for Attendance Tracker

## Verifying Application Security

Since Attendance Tracker is an open-source application, we provide multiple ways to verify the authenticity and integrity of the downloaded executable.

### 1. SHA-256 Checksums

Every release includes SHA-256 checksums that you can use to verify that the file you downloaded matches the one we published:

#### Windows PowerShell
```powershell
# Calculate SHA-256 for the downloaded file
Get-FileHash -Algorithm SHA256 -Path "Attendance Tracker.exe"

# Compare it with the contents of the .sha256 file
```

#### macOS/Linux Terminal
```bash
# Calculate SHA-256 for the downloaded file
shasum -a 256 "Attendance Tracker.exe"

# Compare it with the contents of the .sha256 file
```

### 2. Verification Script

We provide a verification script (`verify.ps1`) in each release that automates the security checks:

```powershell
# Download both the .exe and .sha256 files from the release
# Then run:
.\verify.ps1 -ExecutablePath "Attendance Tracker.exe" -ChecksumFile "Attendance Tracker.exe.sha256"
```

### 3. Windows SmartScreen Warning

When running the application for the first time on Windows, you may see a SmartScreen warning. This is normal for open-source applications that don't use a paid code signing certificate.

To proceed safely:
1. Right-click the file and select "Properties"
2. Check that the file details match what you expect
3. When running, click "More info" on the SmartScreen dialog, then "Run anyway"

### 4. Building from Source

For maximum security, you can build the application from source:

```bash
git clone https://github.com/USERNAME/attendance-tracker.git
cd attendance-tracker
./build-windows.sh  # Requires Docker
```

## Permissions Required

The Attendance Tracker application requires the following permissions:

- **File system access**: To store configuration files in your user directory
- **Autostart capability**: To run at system startup (can be disabled in settings)
- **Network access**: To send status updates to your tracking server (configurable)

## Reporting Security Issues

If you find a security vulnerability in Attendance Tracker, please report it by emailing [security@example.com](mailto:security@example.com) rather than opening a public issue.

## Code Transparency

All code used to build the application is available in this repository. Our build process is automated through GitHub Actions workflows which you can inspect in the `.github/workflows` directory. 
