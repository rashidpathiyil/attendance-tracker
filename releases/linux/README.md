# Attendance Tracker for Fedora Linux

This package contains the source code and build scripts for creating an Attendance Tracker package for Fedora Linux.

## Build Instructions

### Option 1: Building RPM package on a Fedora system

1. Install this package on a Fedora system
2. Navigate to the extracted directory
3. Make the build script executable:
   ```
   chmod +x build-scripts/build.sh
   ```
4. Run the build script:
   ```
   ./build-scripts/build.sh
   ```
5. The script will:
   - Install any required dependencies
   - Build the application
   - Create an RPM package in the `dist` directory

### Option 2: Manual installation

If you prefer to install the application manually:

1. Copy the source code to your Fedora system
2. Install dependencies:
   ```
   sudo dnf install golang gcc libX11-devel libXrandr-devel libXinerama-devel libXcursor-devel libXi-devel mesa-libGL-devel
   ```
3. Build the application:
   ```
   go build -o attendance-tracker
   ```
4. Create the desktop entry:
   ```
   sudo cp attendance-tracker /usr/local/bin/
   sudo cp build-scripts/attendance-tracker.desktop /usr/share/applications/
   sudo cp icon.png /usr/share/icons/hicolor/64x64/apps/attendance-tracker.png
   ```

## Usage

After installation, you can:

1. Launch the application from your applications menu
2. Run it from the command line with `attendance-tracker`
3. Auto-start on login by following the instructions in the application settings

## System Requirements

- Fedora Linux (recent version recommended)
- X Window System
- For idle detection on Linux, the `xprintidle` utility is required:
  ```
  sudo dnf install xprintidle
  ``` 
