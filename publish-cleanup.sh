#!/bin/bash
# Final repository cleanup and organization before publication

echo "Cleaning and organizing repository for publication..."

# Create a clean releases directory structure
mkdir -p clean-releases/{windows,macos,linux/{debian,fedora,appimage}}

# ======= Clean up macOS specific files =======
echo "Cleaning macOS specific files..."
find . -name ".DS_Store" -type f -delete
find . -name "._*" -type f -delete

# ======= Organize installers and packages =======
echo "Organizing installers and packages..."

# Copy macOS packages
if [ -f "dist/installers/macos/Attendance Tracker-1.0.0-macOS.zip" ]; then
    cp "dist/installers/macos/Attendance Tracker-1.0.0-macOS.zip" clean-releases/macos/
    echo "✓ MacOS package copied"
elif [ -f "releases/macos/Attendance Tracker-1.0.0-macOS.zip" ]; then
    cp "releases/macos/Attendance Tracker-1.0.0-macOS.zip" clean-releases/macos/
    echo "✓ MacOS package copied"
fi

# Copy Windows packages
if [ -f "dist/installers/windows/AttendanceTrackerSetup.exe" ]; then
    cp "dist/installers/windows/AttendanceTrackerSetup.exe" clean-releases/windows/
    echo "✓ Windows installer copied"
elif [ -f "releases/windows/Attendance Tracker.exe" ]; then
    cp "releases/windows/Attendance Tracker.exe" clean-releases/windows/
    echo "✓ Windows executable copied"
fi

# Copy Windows README
if [ -f "releases/windows/README.md" ]; then
    cp "releases/windows/README.md" clean-releases/windows/
    echo "✓ Windows README copied"
fi

# Copy Linux Debian package
if [ -f "dist/installers/linux/attendance-tracker_1.0.0_amd64.deb" ]; then
    cp "dist/installers/linux/attendance-tracker_1.0.0_amd64.deb" clean-releases/linux/debian/
    echo "✓ Linux Debian package copied"
fi

# Copy Linux AppImage
if [ -f "dist/installers/linux/AttendanceTracker-1.0.0-x86_64.AppImage" ]; then
    cp "dist/installers/linux/AttendanceTracker-1.0.0-x86_64.AppImage" clean-releases/linux/appimage/
    echo "✓ Linux AppImage copied"
fi

# Copy Fedora files
if [ -f "releases/linux/attendance-tracker-1.0.0-fedora-src.tar.gz" ]; then
    cp "releases/linux/attendance-tracker-1.0.0-fedora-src.tar.gz" clean-releases/linux/fedora/
    echo "✓ Fedora source package copied"
fi

# Copy build scripts for Fedora
if [ -d "releases/linux/build-scripts" ]; then
    cp -r "releases/linux/build-scripts" clean-releases/linux/fedora/
    echo "✓ Fedora build scripts copied"
fi

# Copy Linux README
if [ -f "releases/linux/README.md" ]; then
    cp "releases/linux/README.md" clean-releases/linux/
    echo "✓ Linux README copied"
fi

# ======= Remove all temporary and build files =======
echo "Removing temporary and build files..."

# Remove Docker related files
rm -f Dockerfile.* 

# Remove temporary directories
rm -rf tmp dist fyne-cross .fyne-cross

# Remove build scripts that are no longer needed
rm -f build-*.sh build-*.bat

# Remove placeholder files
rm -f "releases/windows/Attendance Tracker.exe"
rm -f releases/linux/bin/attendance-tracker_1.0.0_linux_amd64

# ======= Clean up the final structure =======
echo "Finalizing repository structure..."

# Remove the old releases directory
rm -rf releases

# Rename the clean releases directory to releases
mv clean-releases releases

# Create a top-level README for the releases directory
cat > releases/README.md << EOF
# Attendance Tracker Releases

This directory contains ready-to-use installation packages for different platforms.

## Windows
- \`AttendanceTrackerSetup.exe\` - Windows installer package
- See \`windows/README.md\` for installation instructions

## macOS
- \`Attendance Tracker-1.0.0-macOS.zip\` - macOS application package
- Simply extract and run the application

## Linux
Linux packages are organized by distribution/format:

### Debian/Ubuntu
- \`debian/attendance-tracker_1.0.0_amd64.deb\` - Debian package
- Install with: \`sudo dpkg -i attendance-tracker_1.0.0_amd64.deb\`

### Fedora
- \`fedora/attendance-tracker-1.0.0-fedora-src.tar.gz\` - Fedora source package
- See \`fedora/build-scripts/\` for build instructions

### AppImage (Universal Linux)
- \`appimage/AttendanceTracker-1.0.0-x86_64.AppImage\` - AppImage package
- Make executable: \`chmod +x AttendanceTracker-1.0.0-x86_64.AppImage\`
- Run directly: \`./AttendanceTracker-1.0.0-x86_64.AppImage\`
EOF

# Remove any leftover temporary cleanup scripts
rm -f cleanup.sh

echo "Repository cleanup complete!"
echo "Final packages are organized in the releases/ directory."
echo "The repository is now clean and ready for publication."

# Cleanup script for preparing the repository for release
echo "Starting repository cleanup..."

# Remove macOS specific files
echo "Removing .DS_Store files..."
find . -name ".DS_Store" -type f -delete

# Remove redundant build scripts
echo "Removing redundant build scripts..."
rm -f build-fedora-rpm.sh

# Remove placeholder binaries
echo "Removing placeholder binaries..."
if [ -f releases/linux/bin/attendance-tracker_1.0.0_linux_amd64 ]; then
    file_contents=$(cat releases/linux/bin/attendance-tracker_1.0.0_linux_amd64)
    if [[ $file_contents == *"placeholder"* ]]; then
        echo "Detected placeholder binary, removing..."
        rm -f releases/linux/bin/attendance-tracker_1.0.0_linux_amd64
    else
        echo "Real binary detected, keeping it."
    fi
fi

# Keep only the final scripts
echo "Organizing build scripts..."
# Rename the final script to a more standard name
mv create-fedora-rpm.sh build-rpm.sh 2>/dev/null || echo "Script already renamed or does not exist."

# Update the README to reflect the correct script names
if [ -f README.md ]; then
    sed -i.bak 's/create-fedora-rpm.sh/build-rpm.sh/g' README.md
    rm -f README.md.bak
fi

# Check .gitignore to ensure it excludes temporary files
echo "Updating .gitignore..."
if [ -f .gitignore ]; then
    # Check if .DS_Store is already ignored
    if ! grep -q "\.DS_Store" .gitignore; then
        echo "*.DS_Store" >> .gitignore
    fi
    
    # Ensure build artifacts are ignored
    if ! grep -q "releases/linux/bin/placeholder" .gitignore; then
        echo "# Placeholder binaries" >> .gitignore
        echo "releases/linux/bin/*placeholder*" >> .gitignore
    fi
fi

# Clean up package scripts directory if it exists
echo "Finalizing script organization..."
if [ -d scripts ]; then
    echo "Organizing scripts directory..."
    mkdir -p scripts/build 2>/dev/null
    
    # Move any remaining build scripts to the scripts directory
    mv package-all.sh scripts/build/ 2>/dev/null || echo "Script already moved or does not exist."
    
    # Keep the cleanup script in the root for easy access
    # But make a copy in the scripts directory
    cp publish-cleanup.sh scripts/cleanup.sh 2>/dev/null
fi

echo "Repository cleanup complete."
echo "Final release preparation complete. Repository is ready for release." 
