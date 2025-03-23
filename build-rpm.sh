#!/bin/bash
# Create a Fedora RPM package from an existing executable
# Note: This script requires running on Fedora or using podman/docker with Fedora

echo "Attendance Tracker - Fedora RPM Creator"
echo "========================================"
echo ""
echo "This script creates an RPM package for Fedora Linux."
echo "It requires either:"
echo "1. Running on a Fedora system, or"
echo "2. Using docker/podman with a Fedora image"
echo ""

# Check if we're on Fedora
if [ -f /etc/fedora-release ]; then
    echo "Running on Fedora - proceeding with native build"
    NATIVE_BUILD=true
else
    echo "Not running on Fedora - will use Docker"
    NATIVE_BUILD=false
    
    # Check if Docker/Podman is available
    if ! command -v docker &> /dev/null && ! command -v podman &> /dev/null; then
        echo "Error: Neither Docker nor Podman found. Cannot continue."
        exit 1
    fi
    
    # Determine which container engine to use
    if command -v docker &> /dev/null; then
        CONTAINER_CMD="docker"
    else
        CONTAINER_CMD="podman"
    fi
    
    # Check if Docker/Podman is running
    if [ "$CONTAINER_CMD" = "docker" ] && ! docker info > /dev/null 2>&1; then
        echo "Error: Docker is not running"
        exit 1
    fi
fi

# Create temp directory
mkdir -p tmp/fedora-build
mkdir -p releases/linux/fedora/rpm

# Create the spec file
cat > tmp/fedora-build/attendance-tracker.spec << EOF
Name:           attendance-tracker
Version:        1.0.0
Release:        1%{?dist}
Summary:        Attendance Tracking Application

License:        MIT
URL:            https://example.com/attendance-tracker

Requires:       libX11, libXcursor, libXrandr, libXinerama, mesa-libGL, libXi, libXxf86vm

%description
A simple application to track user attendance.

%install
mkdir -p %{buildroot}%{_bindir}
mkdir -p %{buildroot}%{_datadir}/applications
mkdir -p %{buildroot}%{_datadir}/icons/hicolor/64x64/apps

install -m 755 %{_sourcedir}/attendance-tracker %{buildroot}%{_bindir}/
install -m 644 %{_sourcedir}/attendance-tracker.desktop %{buildroot}%{_datadir}/applications/
install -m 644 %{_sourcedir}/icon.png %{buildroot}%{_datadir}/icons/hicolor/64x64/apps/attendance-tracker.png

%files
%{_bindir}/attendance-tracker
%{_datadir}/applications/attendance-tracker.desktop
%{_datadir}/icons/hicolor/64x64/apps/attendance-tracker.png

%changelog
* Thu Mar 23 2024 Builder <builder@example.com> - 1.0.0-1
- Initial package
EOF

# Create desktop file
cat > tmp/fedora-build/attendance-tracker.desktop << EOF
[Desktop Entry]
Type=Application
Name=Attendance Tracker
Comment=Track user attendance
Exec=attendance-tracker
Icon=attendance-tracker
Categories=Office;Utility;
EOF

# Copy icon
cp icon.png tmp/fedora-build/

# Create a placeholder executable or ask for one
if [ ! -f "releases/linux/bin/attendance-tracker_1.0.0_linux_amd64" ]; then
    echo ""
    echo "No executable found at releases/linux/bin/attendance-tracker_1.0.0_linux_amd64"
    echo "You need to provide a Linux executable for packaging."
    echo ""
    echo "Options:"
    echo "1. Build one on a Linux system with 'go build -o attendance-tracker'"
    echo "2. Create a placeholder (non-functional) for testing"
    echo ""
    read -p "Create a placeholder executable? (y/n): " CREATE_PLACEHOLDER
    
    if [[ "$CREATE_PLACEHOLDER" == "y" || "$CREATE_PLACEHOLDER" == "Y" ]]; then
        mkdir -p releases/linux/bin
        echo '#!/bin/sh' > releases/linux/bin/attendance-tracker_1.0.0_linux_amd64
        echo 'echo "This is a placeholder. Replace with the real executable."' >> releases/linux/bin/attendance-tracker_1.0.0_linux_amd64
        chmod +x releases/linux/bin/attendance-tracker_1.0.0_linux_amd64
        echo "Placeholder created. This is NOT functional!"
    else
        echo "Please provide an executable at releases/linux/bin/attendance-tracker_1.0.0_linux_amd64"
        exit 1
    fi
fi

# Copy the executable
cp releases/linux/bin/attendance-tracker_1.0.0_linux_amd64 tmp/fedora-build/attendance-tracker

if [ "$NATIVE_BUILD" = true ]; then
    # Build RPM on Fedora natively
    echo "Building RPM natively..."
    
    # Set up RPM build environment
    mkdir -p ~/rpmbuild/{SPECS,SOURCES,BUILD,BUILDROOT,RPMS,SRPMS}
    
    # Copy files
    cp tmp/fedora-build/attendance-tracker.spec ~/rpmbuild/SPECS/
    cp tmp/fedora-build/attendance-tracker ~/rpmbuild/SOURCES/
    cp tmp/fedora-build/attendance-tracker.desktop ~/rpmbuild/SOURCES/
    cp tmp/fedora-build/icon.png ~/rpmbuild/SOURCES/
    
    # Build RPM
    cd ~/rpmbuild/SPECS
    rpmbuild -ba attendance-tracker.spec
    
    # Copy RPM to releases directory
    find ~/rpmbuild/RPMS -name "*.rpm" -exec cp {} ../../releases/linux/fedora/rpm/ \;
else
    # Build using Docker/Podman
    echo "Building RPM using $CONTAINER_CMD..."
    
    # Create Dockerfile
    cat > tmp/fedora-build/Dockerfile << EOF
FROM fedora:latest

# Install required packages
RUN dnf install -y rpm-build rpmdevtools

# Set working directory
WORKDIR /build

# Copy files
COPY attendance-tracker.spec /build/
COPY attendance-tracker /build/
COPY attendance-tracker.desktop /build/
COPY icon.png /build/

# Set up RPM build environment
RUN rpmdev-setuptree

# Copy files to RPM build locations
RUN cp attendance-tracker.spec /root/rpmbuild/SPECS/ && \
    cp attendance-tracker /root/rpmbuild/SOURCES/ && \
    cp attendance-tracker.desktop /root/rpmbuild/SOURCES/ && \
    cp icon.png /root/rpmbuild/SOURCES/

# Build RPM
RUN cd /root/rpmbuild/SPECS && \
    rpmbuild -ba attendance-tracker.spec

# Copy RPM to output
RUN mkdir -p /output && \
    find /root/rpmbuild/RPMS -name "*.rpm" -exec cp {} /output/ \;
EOF

    # Build Docker image
    $CONTAINER_CMD build -t attendance-tracker-rpm-builder tmp/fedora-build
    
    # Run container to create RPM
    $CONTAINER_CMD run --rm -v "$(pwd)/releases/linux/fedora/rpm:/output" attendance-tracker-rpm-builder bash -c "cp /root/rpmbuild/RPMS/*/*.rpm /output/"
fi

# Clean up
rm -rf tmp

# Check if RPM was created
if ls releases/linux/fedora/rpm/*.rpm > /dev/null 2>&1; then
    echo "✓ Fedora RPM package created successfully!"
    echo "The RPM is in releases/linux/fedora/rpm/"
    echo ""
    echo "To install on Fedora:"
    echo "sudo dnf install releases/linux/fedora/rpm/attendance-tracker-1.0.0-1.*.rpm"
else
    echo "⚠️ RPM package creation failed."
fi 
