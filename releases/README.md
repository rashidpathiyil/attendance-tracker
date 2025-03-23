# Attendance Tracker Releases

This directory contains ready-to-use installation packages for different platforms.

## Windows
- `AttendanceTrackerSetup.exe` - Windows installer package
- See `windows/README.md` for installation instructions

## macOS
- `Attendance Tracker-1.0.0-macOS.zip` - macOS application package
- Simply extract and run the application

## Linux
Linux packages are organized by distribution/format:

### Debian/Ubuntu
- `debian/attendance-tracker_1.0.0_amd64.deb` - Debian package
- Install with: `sudo dpkg -i attendance-tracker_1.0.0_amd64.deb`

### Fedora
- `fedora/attendance-tracker-1.0.0-fedora-src.tar.gz` - Fedora source package
- See `fedora/build-scripts/` for build instructions

### AppImage (Universal Linux)
- `appimage/AttendanceTracker-1.0.0-x86_64.AppImage` - AppImage package
- Make executable: `chmod +x AttendanceTracker-1.0.0-x86_64.AppImage`
- Run directly: `./AttendanceTracker-1.0.0-x86_64.AppImage`
