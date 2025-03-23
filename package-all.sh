#!/bin/bash
# Script to create user-friendly packages for all platforms

echo "Creating user-friendly packages for all platforms..."

# Create directories
mkdir -p dist/installers/{windows,macos,linux}

# Clean up any old Docker containers or images that might be stuck
docker rm -f $(docker ps -aq --filter "ancestor=attendance-tracker-win-installer" 2>/dev/null) 2>/dev/null || true
docker rm -f $(docker ps -aq --filter "ancestor=attendance-tracker-appimage" 2>/dev/null) 2>/dev/null || true
docker rm -f $(docker ps -aq --filter "ancestor=attendance-tracker-deb" 2>/dev/null) 2>/dev/null || true

# Package for Windows - Create an installer with NSIS
# (Requires Docker)
echo "Creating Windows installer..."
cat > Dockerfile.win-installer << EOF
FROM ubuntu:20.04

ENV DEBIAN_FRONTEND=noninteractive
ENV TZ=Etc/UTC

RUN apt-get update && apt-get install -y nsis zip unzip

WORKDIR /app
COPY releases/windows /app/input
RUN mkdir -p /app/output

# Create NSIS installer script
RUN echo '!include "MUI2.nsh"' > installer.nsi && \\
    echo 'Name "Attendance Tracker"' >> installer.nsi && \\
    echo 'OutFile "output/AttendanceTrackerSetup.exe"' >> installer.nsi && \\
    echo 'InstallDir "\$PROGRAMFILES\\\\Attendance Tracker"' >> installer.nsi && \\
    echo 'RequestExecutionLevel admin' >> installer.nsi && \\
    echo '!insertmacro MUI_PAGE_WELCOME' >> installer.nsi && \\
    echo '!insertmacro MUI_PAGE_DIRECTORY' >> installer.nsi && \\
    echo '!insertmacro MUI_PAGE_INSTFILES' >> installer.nsi && \\
    echo '!insertmacro MUI_PAGE_FINISH' >> installer.nsi && \\
    echo '!insertmacro MUI_LANGUAGE "English"' >> installer.nsi && \\
    echo 'Section "Install"' >> installer.nsi && \\
    echo '  SetOutPath "\$INSTDIR"' >> installer.nsi && \\
    echo '  File "input/Attendance Tracker.exe"' >> installer.nsi && \\
    echo '  CreateDirectory "\$SMPROGRAMS\\\\Attendance Tracker"' >> installer.nsi && \\
    echo '  CreateShortcut "\$SMPROGRAMS\\\\Attendance Tracker\\\\Attendance Tracker.lnk" "\$INSTDIR\\\\Attendance Tracker.exe"' >> installer.nsi && \\
    echo '  WriteUninstaller "\$INSTDIR\\\\uninstall.exe"' >> installer.nsi && \\
    echo '  CreateShortcut "\$SMPROGRAMS\\\\Attendance Tracker\\\\Uninstall.lnk" "\$INSTDIR\\\\uninstall.exe"' >> installer.nsi && \\
    echo 'SectionEnd' >> installer.nsi && \\
    echo 'Section "Uninstall"' >> installer.nsi && \\
    echo '  Delete "\$INSTDIR\\\\Attendance Tracker.exe"' >> installer.nsi && \\
    echo '  Delete "\$INSTDIR\\\\uninstall.exe"' >> installer.nsi && \\
    echo '  RMDir "\$INSTDIR"' >> installer.nsi && \\
    echo '  Delete "\$SMPROGRAMS\\\\Attendance Tracker\\\\Attendance Tracker.lnk"' >> installer.nsi && \\
    echo '  Delete "\$SMPROGRAMS\\\\Attendance Tracker\\\\Uninstall.lnk"' >> installer.nsi && \\
    echo '  RMDir "\$SMPROGRAMS\\\\Attendance Tracker"' >> installer.nsi && \\
    echo 'SectionEnd' >> installer.nsi

# Build the installer
RUN makensis installer.nsi
EOF

if command -v docker &> /dev/null && docker info >/dev/null 2>&1; then
    echo "Building Windows installer image..."
    docker build -t attendance-tracker-win-installer -f Dockerfile.win-installer .
    echo "Running Windows installer container..."
    docker run --rm -v "$(pwd)/dist/installers/windows:/app/output" attendance-tracker-win-installer
    echo "✓ Windows installer created: dist/installers/windows/AttendanceTrackerSetup.exe"
else
    echo "⚠️ Docker not available. Skipping Windows installer creation."
    echo "   For Windows, users will need to use the executable directly."
fi

# Package for macOS - Create a DMG
# (Requires being on macOS)
echo "Creating macOS package..."
if [[ "$OSTYPE" == "darwin"* ]]; then
    # Copy the existing macOS package
    cp releases/macos/Attendance\ Tracker-1.0.0-macOS.zip dist/installers/macos/
    echo "✓ macOS package copied: dist/installers/macos/Attendance Tracker-1.0.0-macOS.zip"
else
    echo "⚠️ Not on macOS. Skipping macOS package creation."
fi

# Package for Linux - Create AppImage
# (Requires Docker)
echo "Creating Linux AppImage..."
cat > Dockerfile.appimage << EOF
FROM ubuntu:20.04

ENV DEBIAN_FRONTEND=noninteractive
ENV TZ=Etc/UTC

RUN apt-get update && apt-get install -y wget fuse libfuse2 file

WORKDIR /app
COPY releases/linux/bin/attendance-tracker_1.0.0_linux_amd64 /app/attendance-tracker
COPY releases/linux/attendance-tracker.desktop /app/
COPY icon.png /app/

RUN wget -q -O appimagetool https://github.com/AppImage/AppImageKit/releases/download/continuous/appimagetool-x86_64.AppImage && \\
    chmod +x appimagetool

# Create AppDir structure
RUN mkdir -p AppDir/usr/bin && \\
    mkdir -p AppDir/usr/share/applications && \\
    mkdir -p AppDir/usr/share/icons/hicolor/256x256/apps

RUN cp attendance-tracker AppDir/usr/bin/ && \\
    cp attendance-tracker.desktop AppDir/usr/share/applications/ && \\
    cp attendance-tracker.desktop AppDir/ && \\
    cp icon.png AppDir/usr/share/icons/hicolor/256x256/apps/attendance-tracker.png && \\
    cp icon.png AppDir/

# Create AppImage
RUN chmod +x appimagetool && \\
    ARCH=x86_64 ./appimagetool AppDir AttendanceTracker-1.0.0-x86_64.AppImage
EOF

if command -v docker &> /dev/null && docker info >/dev/null 2>&1; then
    echo "Building Linux AppImage..."
    docker build -t attendance-tracker-appimage -f Dockerfile.appimage .
    echo "Extracting AppImage..."
    docker run --rm -v "$(pwd)/dist/installers/linux:/output" attendance-tracker-appimage sh -c "cp *.AppImage /output/"
    echo "✓ Linux AppImage created: dist/installers/linux/AttendanceTracker-1.0.0-x86_64.AppImage"
    
    # Also create a .deb package
    echo "Creating Debian package..."
    cat > Dockerfile.deb << EOF
FROM ubuntu:20.04

ENV DEBIAN_FRONTEND=noninteractive
ENV TZ=Etc/UTC

RUN apt-get update && apt-get install -y dpkg-dev fakeroot lintian

WORKDIR /app
COPY releases/linux/bin/attendance-tracker_1.0.0_linux_amd64 /app/attendance-tracker
COPY releases/linux/attendance-tracker.desktop /app/
COPY icon.png /app/

# Create Debian package structure
RUN mkdir -p debian/DEBIAN && \\
    mkdir -p debian/usr/bin && \\
    mkdir -p debian/usr/share/applications && \\
    mkdir -p debian/usr/share/icons/hicolor/256x256/apps

# Create control file
RUN echo "Package: attendance-tracker" > debian/DEBIAN/control && \\
    echo "Version: 1.0.0" >> debian/DEBIAN/control && \\
    echo "Section: utils" >> debian/DEBIAN/control && \\
    echo "Priority: optional" >> debian/DEBIAN/control && \\
    echo "Architecture: amd64" >> debian/DEBIAN/control && \\
    echo "Maintainer: Your Name <your.email@example.com>" >> debian/DEBIAN/control && \\
    echo "Description: Attendance tracking application" >> debian/DEBIAN/control && \\
    echo " A simple application to track user attendance." >> debian/DEBIAN/control

# Copy files
RUN cp attendance-tracker debian/usr/bin/ && \\
    cp attendance-tracker.desktop debian/usr/share/applications/ && \\
    cp icon.png debian/usr/share/icons/hicolor/256x256/apps/attendance-tracker.png

# Set permissions
RUN chmod 755 debian/usr/bin/attendance-tracker

# Build package
RUN dpkg-deb --build debian && \\
    mv debian.deb attendance-tracker_1.0.0_amd64.deb
EOF

    echo "Building Debian package..."
    docker build -t attendance-tracker-deb -f Dockerfile.deb .
    echo "Extracting Debian package..."
    docker run --rm -v "$(pwd)/dist/installers/linux:/output" attendance-tracker-deb sh -c "cp *.deb /output/"
    echo "✓ Linux .deb package created: dist/installers/linux/attendance-tracker_1.0.0_amd64.deb"
else
    echo "⚠️ Docker not available. Skipping Linux package creation."
    echo "   For Linux, users will need to use the executable directly."
fi

echo "Package creation complete!"
echo "User-friendly packages are in dist/installers/ directory:"
echo " - Windows: dist/installers/windows/AttendanceTrackerSetup.exe"
echo " - macOS:   dist/installers/macos/Attendance Tracker-1.0.0-macOS.zip"
echo " - Linux:   dist/installers/linux/AttendanceTracker-1.0.0-x86_64.AppImage"
echo "            dist/installers/linux/attendance-tracker_1.0.0_amd64.deb" 
