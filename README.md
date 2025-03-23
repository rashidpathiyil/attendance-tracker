# Attendance Tracker

A cross-platform desktop application for tracking user attendance with automatic idle detection.

## Features

- Manual check-in and check-out functionality
- Automatic mode that tracks user activity
- Cross-platform support (Windows, macOS, Linux)
- Idle time detection
- Configurable settings
- Developer mode for advanced configurations
- System tray integration
- Auto-start on system boot option

## Requirements

- Go 1.18 or higher
- Fyne toolkit and dependencies

### Platform-specific requirements

- **Linux**: Requires X Window System and `xprintidle` utility for idle detection
  ```
  # On Debian/Ubuntu
  sudo apt-get install x11-utils
  ```

- **macOS**: Uses built-in `ioreg` command for idle detection (no additional dependencies)

- **Windows**: Uses PowerShell for idle detection (no additional dependencies)

## Building from source

### Install Go

Download and install Go from [golang.org](https://golang.org/doc/install)

### Install Fyne

```bash
go install fyne.io/fyne/v2/cmd/fyne@latest
```

### Clone the repository

```bash
git clone https://github.com/yourusername/attendance-tracker.git
cd attendance-tracker
```

### Build for your platform

```bash
make build
```

### Cross-platform builds

Build for all supported platforms:

```bash
make build-all
```

Or build for a specific platform:

```bash
make windows  # Windows builds
make darwin   # macOS builds
make linux    # Linux builds
```

### Package the application

Create installable packages for distribution:

```bash
make package-all       # All platforms
make package-windows   # Windows installer
make package-darwin    # macOS .app bundle
make package-linux     # Linux packages
```

## Auto-Start Setup

To configure the application to start automatically when your system boots:

### Windows

1. Create a shortcut to the application executable
2. Press `Win+R`, type `shell:startup`, and press Enter
3. Move the shortcut to the Startup folder that opens

**Alternative method:**
1. Press `Win+R`, type `regedit`, and press Enter
2. Navigate to `HKEY_CURRENT_USER\Software\Microsoft\Windows\CurrentVersion\Run`
3. Right-click in the right pane and select New > String Value
4. Name it "AttendanceTracker" and set the value to the full path of the executable with the `--minimized` flag:
   ```
   C:\Path\To\attendance-tracker.exe --minimized
   ```

### macOS

1. Open System Preferences > Users & Groups
2. Select your user account and click "Login Items"
3. Click the "+" button and select the Attendance Tracker application
4. Check "Hide" to start the app minimized

**Alternative method (via terminal):**
1. Create a LaunchAgent plist file:
   ```bash
   mkdir -p ~/Library/LaunchAgents
   nano ~/Library/LaunchAgents/com.attendance.tracker.plist
   ```

2. Add the following content (adjust the path to your executable):
   ```xml
   <?xml version="1.0" encoding="UTF-8"?>
   <!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
   <plist version="1.0">
   <dict>
       <key>Label</key>
       <string>com.attendance.tracker</string>
       <key>ProgramArguments</key>
       <array>
           <string>/Applications/AttendanceTracker.app/Contents/MacOS/AttendanceTracker</string>
           <string>--minimized</string>
       </array>
       <key>RunAtLoad</key>
       <true/>
   </dict>
   </plist>
   ```

3. Load the LaunchAgent:
   ```bash
   launchctl load ~/Library/LaunchAgents/com.attendance.tracker.plist
   ```

### Linux

1. Create a .desktop file:
   ```bash
   mkdir -p ~/.config/autostart
   nano ~/.config/autostart/attendance-tracker.desktop
   ```

2. Add the following content:
   ```
   [Desktop Entry]
   Type=Application
   Name=Attendance Tracker
   Exec=/path/to/attendance-tracker --minimized
   Terminal=false
   Comment=Attendance Tracking Application
   Categories=Utility;
   ```

3. Make it executable:
   ```bash
   chmod +x ~/.config/autostart/attendance-tracker.desktop
   ```

## Versioning

This application follows [Semantic Versioning](https://semver.org/).

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## Configuration

You can configure the following constants in the `main.go` file:

- `serverEndpoint`: The URL to send status updates
- `deviceID`: The ID of the current device
- `idleTimeout`: The duration of inactivity before auto check-out (default: 5 minutes)
- `checkInterval`: How often to check for activity (default: 1 second)

## Usage

- Click the "Check In" / "Check Out" button to manually toggle your status
- Enable "Auto Mode" to automatically check in when active and check out when idle
- The activity log displays all status changes and server communications 
