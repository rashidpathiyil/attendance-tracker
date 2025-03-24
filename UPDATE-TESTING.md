# Testing the Update System in Attendance Tracker

This guide explains how to test the update checking, download, and installation features of the Attendance Tracker application.

## Overview of the Update System

The Attendance Tracker application includes a comprehensive update system with the following features:

1. **Automatic Update Checking**: Checks for updates in the background when the app starts
2. **Manual Update Checking**: Users can check for updates via the Help menu
3. **Update Notifications**: Shows a non-intrusive notification when updates are available
4. **Detailed Update Information**: Displays version number, release date, changelog, and file size
5. **Progress Reporting**: Shows download progress with speed and percentage
6. **Installation**: Handles the installation process, preserving user settings
7. **Update Caching**: Avoids re-downloading the same update multiple times

## Testing the Update System

### 1. Using the PowerShell Test Script

The easiest way to test the update system is to use the included `check-update.ps1` PowerShell script. This script:

- Starts a local HTTP server that simulates an update server
- Overrides the update server URL in the application
- Creates a dummy update file for the application to download
- Runs the application with test flags to force an update check

#### To run the test script:

1. Make sure the Attendance Tracker application is built and available
2. Open a PowerShell window
3. Run the script:

```powershell
.\check-update.ps1
```

The script accepts optional parameters:
- `-AppPath`: Path to the application executable (default: "Attendance Tracker.exe")
- `-TestVersion`: The version to simulate for testing (default: "1.1.0-test")

```powershell
.\check-update.ps1 -AppPath "C:\Path\To\Attendance Tracker.exe" -TestVersion "1.2.0-beta"
```

### 2. Manual Testing Environment Variables

You can also test the update system by setting environment variables:

- `ATTENDANCE_UPDATE_SERVER`: The URL of the update server to check
- `ATTENDANCE_UPDATE_TEST`: Set to "1" to force a simulated update

Example (Windows):
```cmd
set ATTENDANCE_UPDATE_SERVER=http://localhost:8080/version
set ATTENDANCE_UPDATE_TEST=1
"Attendance Tracker.exe" --test-updates
```

Example (macOS/Linux):
```bash
ATTENDANCE_UPDATE_SERVER=http://localhost:8080/version ATTENDANCE_UPDATE_TEST=1 ./attendance-tracker --test-updates
```

### 3. The Update Process Step-by-Step

When testing, you should observe the following process:

1. **Update Check**:
   - The application should display "Checking for updates..." with a progress indicator
   - After a moment, it should show update details if available

2. **Update Dialog**:
   - Should display version, release date, size, and changelog
   - Should have "Download & Install" and "Later" buttons

3. **Download Process**:
   - Shows a progress bar with percentage, downloaded/total size, and speed
   - Can be canceled by clicking "Cancel"

4. **Installation**:
   - After download completes, shows installation progress
   - Preserves current settings
   - Prompts for restart when complete

5. **Update Caching**:
   - If you check for updates again, it should detect the previously downloaded update
   - Offers to install without re-downloading

## Known Limitations in Test Mode

- The downloaded "update" is just a placeholder file, not an actual executable
- The installation process is simulated and doesn't actually replace the executable
- In a real deployment, you would need to implement platform-specific update mechanisms

## Real-World Implementation Notes

In a production environment, you would:

1. Set up a proper update server (GitHub Releases works well)
2. Sign your updates with a code signing certificate
3. Implement platform-specific installers for each supported OS
4. Use proper versioning (Semantic Versioning recommended)

## Troubleshooting

- Check the application logs for update-related messages
- On Windows, make sure the application has write access to its cache directory
- If updates aren't being detected, check that the version in your app is older than the test version 
