#!/bin/bash

# Exit on error
set -e

# Set application name
APP_NAME="Attendance Tracker"

# Check if at least 3 arguments are provided
if [ $# -lt 3 ]; then
    echo "Usage: $0 VERSION BUILD_DATE COMMIT_SHA"
    echo "Example: $0 1.0.0 $(date +%Y-%m-%d) $(git rev-parse --short HEAD)"
    exit 1
fi

VERSION="$1"
BUILD_DATE="$2"
COMMIT_SHA="$3"

echo "Building $APP_NAME for Windows..."
echo "Version: $VERSION"
echo "Build date: $BUILD_DATE"
echo "Commit: $COMMIT_SHA"
echo ""
echo "Note: Building Windows executables on macOS can be complex."
echo "Alternative options if you encounter problems:"
echo "1. Build using the GitHub Actions workflow (recommended)"
echo "2. Build directly on a Windows machine"
echo "3. Use a Docker container with proper cross-compilation setup"
echo ""

# Check operating system
SYS_OS="unknown"
if [ "$(uname)" == "Darwin" ]; then
    SYS_OS="macos"
elif [ "$(expr substr $(uname -s) 1 5)" == "Linux" ]; then
    SYS_OS="linux"
elif [ "$(expr substr $(uname -s) 1 10)" == "MINGW32_NT" ] || [ "$(expr substr $(uname -s) 1 10)" == "MINGW64_NT" ]; then
    SYS_OS="windows"
fi

# Check for MinGW (only needed for cross-compilation)
HAS_MINGW=0
if [ "$SYS_OS" != "windows" ]; then
    if command -v x86_64-w64-mingw32-gcc &> /dev/null; then
        HAS_MINGW=1
    fi
fi

# Check for Fyne CLI
if ! command -v fyne &> /dev/null; then
    echo "Fyne CLI not found. Installing..."
    go install fyne.io/fyne/v2/cmd/fyne@latest
    
    # Check if installation was successful
    if ! command -v fyne &> /dev/null; then
        echo "Failed to install Fyne CLI automatically. Please install manually:"
        echo "go install fyne.io/fyne/v2/cmd/fyne@latest"
        echo ""
        echo "Make sure your GOPATH/bin is in your PATH environment variable."
        echo "For example, add this to your ~/.bashrc or ~/.zshrc:"
        echo "  export PATH=\$PATH:\$HOME/go/bin"
        exit 1
    fi
    echo "Fyne CLI installed successfully."
fi

# Create versions
VERSION_LDFLAGS="-X main.Version=${VERSION} -X main.BuildDate=${BUILD_DATE} -X main.CommitSHA=${COMMIT_SHA}"

# Create output directory
mkdir -p ./releases/windows

# Cross-compile for Windows
echo "Compiling Windows executable..."
if [ "$GITHUB_ACTIONS" = "true" ]; then
    # Use direct go build with appropriate tags for GitHub Actions
    GOOS=windows GOARCH=amd64 CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc \
    go build -tags "no_native_menus" -ldflags "$VERSION_LDFLAGS" -o "./releases/windows/$APP_NAME.exe"
elif [ "$HAS_MINGW" = "1" ]; then
    # Use direct Go build with MinGW cross-compiler
    echo "Building with MinGW cross-compiler..."
    GOOS=windows GOARCH=amd64 CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc \
    go build -tags "no_native_menus" -ldflags "$VERSION_LDFLAGS" -o "./releases/windows/$APP_NAME.exe"
    
    # Optionally use UPX for compression if available
    if command -v upx &> /dev/null; then
        echo "Compressing with UPX..."
        upx --best "./releases/windows/$APP_NAME.exe" || echo "UPX compression failed, continuing without it"
    fi
else
    # Fallback to simple GOOS=windows without CGO for environments without MinGW
    echo "WARNING: Building without CGO (limited functionality). For full features, install MinGW."
    echo ""
    echo "To install the required MinGW cross-compiler:"
    if [ "$SYS_OS" = "macos" ]; then
        echo "  macOS: brew install mingw-w64"
    elif [ "$SYS_OS" = "linux" ]; then
        echo "  Linux: sudo apt-get install gcc-mingw-w64"
    else
        echo "  Windows: Use native compilation instead"
    fi
    echo ""
    echo "Building with limited functionality..."
    GOOS=windows GOARCH=amd64 CGO_ENABLED=0 \
    go build -tags "mobile" -ldflags "$VERSION_LDFLAGS" -o "./releases/windows/$APP_NAME.exe"
fi

# Generate SHA-256 checksums
echo "Generating checksums..."
cd ./releases/windows
sha256sum "$APP_NAME.exe" > "$APP_NAME.exe.sha256"
cd ../..

# Copy verification tools
echo "Copying verification tools..."
cp verify.ps1 ./releases/windows/ 2>/dev/null || echo "Warning: verify.ps1 not found"
cp SECURITY.md ./releases/windows/VERIFICATION.txt 2>/dev/null || echo "Warning: SECURITY.md not found"
cp install.bat ./releases/windows/ 2>/dev/null || echo "Warning: install.bat not found"

# Create a simple NSIS installer script for Windows
cat > ./releases/windows/install.nsi << EOF
; Attendance Tracker Installer Script
Unicode True

!include "MUI2.nsh"
!include "FileFunc.nsh"

; Metadata
Name "$APP_NAME"
OutFile "$APP_NAME Setup.exe"
InstallDir "\$PROGRAMFILES64\\$APP_NAME"
InstallDirRegKey HKCU "Software\\$APP_NAME" ""
RequestExecutionLevel admin

; Variables
Var StartMenuFolder

; Interface settings
!define MUI_ABORTWARNING
!define MUI_ICON "..\\..\\icon.ico"

; Pages
!insertmacro MUI_PAGE_WELCOME
!insertmacro MUI_PAGE_LICENSE "..\\..\\LICENSE"
!insertmacro MUI_PAGE_DIRECTORY
!insertmacro MUI_PAGE_STARTMENU "Application" \$StartMenuFolder
!insertmacro MUI_PAGE_INSTFILES
!define MUI_FINISHPAGE_RUN "\$INSTDIR\\$APP_NAME.exe"
!insertmacro MUI_PAGE_FINISH

; Uninstaller pages
!insertmacro MUI_UNPAGE_CONFIRM
!insertmacro MUI_UNPAGE_INSTFILES

; Languages
!insertmacro MUI_LANGUAGE "English"

; The main section
Section "MainSection" SEC01
  SetOutPath "\$INSTDIR"
  
  ; Install files
  File "$APP_NAME.exe"
  File "..\\..\\LICENSE"
  File "..\\..\\icon.ico"
  
  ; Store installation folder
  WriteRegStr HKCU "Software\\$APP_NAME" "" \$INSTDIR
  
  ; Create uninstaller
  WriteUninstaller "\$INSTDIR\\uninstall.exe"
  
  ; Create start menu entries
  !insertmacro MUI_STARTMENU_WRITE_BEGIN Application
    CreateDirectory "\$SMPROGRAMS\\\$StartMenuFolder"
    CreateShortcut "\$SMPROGRAMS\\\$StartMenuFolder\\$APP_NAME.lnk" "\$INSTDIR\\$APP_NAME.exe"
    CreateShortcut "\$SMPROGRAMS\\\$StartMenuFolder\\Uninstall.lnk" "\$INSTDIR\\uninstall.exe"
  !insertmacro MUI_STARTMENU_WRITE_END
  
  ; Create registry entries for uninstall
  WriteRegStr HKLM "Software\\Microsoft\\Windows\\CurrentVersion\\Uninstall\\$APP_NAME" "DisplayName" "$APP_NAME"
  WriteRegStr HKLM "Software\\Microsoft\\Windows\\CurrentVersion\\Uninstall\\$APP_NAME" "UninstallString" '"\$INSTDIR\\uninstall.exe"'
  WriteRegStr HKLM "Software\\Microsoft\\Windows\\CurrentVersion\\Uninstall\\$APP_NAME" "QuietUninstallString" '"\$INSTDIR\\uninstall.exe" /S'
  WriteRegStr HKLM "Software\\Microsoft\\Windows\\CurrentVersion\\Uninstall\\$APP_NAME" "InstallLocation" "\$INSTDIR"
  WriteRegStr HKLM "Software\\Microsoft\\Windows\\CurrentVersion\\Uninstall\\$APP_NAME" "DisplayIcon" "\$INSTDIR\\icon.ico"
  WriteRegStr HKLM "Software\\Microsoft\\Windows\\CurrentVersion\\Uninstall\\$APP_NAME" "Publisher" "Rashid Pathiyil"
  WriteRegStr HKLM "Software\\Microsoft\\Windows\\CurrentVersion\\Uninstall\\$APP_NAME" "DisplayVersion" "$VERSION"
  WriteRegDWORD HKLM "Software\\Microsoft\\Windows\\CurrentVersion\\Uninstall\\$APP_NAME" "NoModify" 1
  WriteRegDWORD HKLM "Software\\Microsoft\\Windows\\CurrentVersion\\Uninstall\\$APP_NAME" "NoRepair" 1
  
  ; Calculate and store installed size
  \${GetSize} "\$INSTDIR" "/S=0K" \$0 \$1 \$2
  IntFmt \$0 "0x%08X" \$0
  WriteRegDWORD HKLM "Software\\Microsoft\\Windows\\CurrentVersion\\Uninstall\\$APP_NAME" "EstimatedSize" "\$0"
SectionEnd

; Run the application after install if requested
Function .onInstSuccess
  Exec '"\$INSTDIR\\$APP_NAME.exe"'
FunctionEnd

; Uninstaller
Section "Uninstall"
  ; Kill running process if it's running
  ExecCmd::exec 'taskkill /f /im "$APP_NAME.exe"'
  Sleep 1000
  
  ; Remove files
  Delete "\$INSTDIR\\$APP_NAME.exe"
  Delete "\$INSTDIR\\LICENSE"
  Delete "\$INSTDIR\\icon.ico"
  Delete "\$INSTDIR\\uninstall.exe"
  
  ; Remove start menu items
  !insertmacro MUI_STARTMENU_GETFOLDER Application \$StartMenuFolder
  Delete "\$SMPROGRAMS\\\$StartMenuFolder\\$APP_NAME.lnk"
  Delete "\$SMPROGRAMS\\\$StartMenuFolder\\Uninstall.lnk"
  RMDir "\$SMPROGRAMS\\\$StartMenuFolder"
  
  ; Remove installation directory
  RMDir "\$INSTDIR"
  
  ; Remove registry entries
  DeleteRegKey HKLM "Software\\Microsoft\\Windows\\CurrentVersion\\Uninstall\\$APP_NAME"
  DeleteRegKey HKCU "Software\\$APP_NAME"
SectionEnd
EOF

# Create a batch file for installing with upgrade parameter
cat > ./releases/windows/upgrade.bat << EOF
@echo off
echo Installing $APP_NAME with upgrade support...
"$APP_NAME Setup.exe" /UPGRADE=1
EOF

echo "-------------------------------------------"
echo "Build complete!"
echo "-------------------------------------------"
echo "Files available in ./releases/windows:"
echo "- $APP_NAME.exe"
echo "- $APP_NAME.exe.sha256 (Checksum)"
echo "-------------------------------------------"
echo "To build the installer, run makensis on the install.nsi file:"
echo "makensis ./releases/windows/install.nsi"
echo "-------------------------------------------" 
