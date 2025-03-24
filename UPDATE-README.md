# Attendance Tracker Auto-Update System

Attendance Tracker includes an automatic update system that checks for and installs new versions directly from the GitHub releases.

## How Updates Work

1. **Automatic Background Checks**: The app silently checks for updates at startup and periodically while running
2. **Manual Checking**: You can manually check for updates through the Help menu
3. **Update Notifications**: If an update is available, a non-intrusive notification appears
4. **Detailed Information**: The update dialog shows version number, release date, changelog, and file size
5. **Simple Installation**: One-click download and installation with progress reporting

## For Users

### Checking for Updates

1. Open the **Help menu**
2. Click on **Check for Updates**
3. If an update is available, follow the prompts to download and install

### Update Notification

When an update is available:
- A notification will appear indicating a new version is available
- Click **View Details** to see more information or **Not Now** to dismiss

### Download and Installation Process

1. Click **Download & Install** to start the update process
2. A progress bar shows download progress, speed, and estimated time
3. After download, the application will prepare to install
4. The app will restart automatically when the installation is complete

## For Developers: Publishing Updates

To publish a new update that users can automatically download:

1. **Increment Version Number**: Update the `VERSION` variable in `build-windows.sh` (e.g., from 1.0.0 to 1.0.1)

2. **Tag Release on GitHub**:
   ```bash
   git tag v1.0.1
   git push origin v1.0.1
   ```

3. **Wait for Automated Build**: GitHub Actions will automatically build the release and publish it

Alternatively, create a release manually:

1. Build the executable with version information:
   ```bash
   ./build-windows.sh
   ```

2. Go to GitHub and create a new release:
   - Use the tag format `v1.0.1` (must start with 'v')
   - Upload the executable from `./releases/windows/`
   - Provide release notes using bullet points (• or - format)

## Version Numbering

Attendance Tracker follows Semantic Versioning:
- **Major.Minor.Patch** (e.g., 1.0.1)
- Increment **Major** for incompatible API changes
- Increment **Minor** for new functionality (backward-compatible)
- Increment **Patch** for bug fixes (backward-compatible)

## Testing Updates Locally

You can test the update system locally using the included PowerShell script:

```powershell
.\check-update.ps1
```

This script creates a local update server that simulates GitHub releases for testing without publishing an actual update.

## Troubleshooting

If updates are not working:

1. Check your internet connection
2. Ensure the application has appropriate permissions
3. Look for error messages in the log file (located in your config directory)
4. Try manually downloading the latest version from GitHub 
