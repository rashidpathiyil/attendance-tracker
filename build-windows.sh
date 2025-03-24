#!/bin/bash

# Exit on error
set -e

# Configuration
APP_NAME="Attendance Tracker"
VERSION="1.0.0"  # Update this for each release
BUILD_DATE=$(date +"%Y-%m-%d")
COMMIT_SHA=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")

echo "Building $APP_NAME for Windows..."
echo "Version: $VERSION"
echo "Build date: $BUILD_DATE"
echo "Commit: $COMMIT_SHA"

# Generate Windows resource file
cat > resource_windows.rc << EOF
1 VERSIONINFO
FILEVERSION     1,0,0,0
PRODUCTVERSION  1,0,0,0
BEGIN
  BLOCK "StringFileInfo"
  BEGIN
    BLOCK "080904E4"
    BEGIN
      VALUE "CompanyName", "Rashid Pathiyil"
      VALUE "FileDescription", "Attendance Tracker Application"
      VALUE "FileVersion", "$VERSION"
      VALUE "InternalName", "attendance_tracker"
      VALUE "LegalCopyright", "Copyright © 2023 Rashid Pathiyil"
      VALUE "OriginalFilename", "Attendance Tracker.exe"
      VALUE "ProductName", "Attendance Tracker"
      VALUE "ProductVersion", "$VERSION"
    END
  END
  BLOCK "VarFileInfo"
  BEGIN
    VALUE "Translation", 0x809, 1252
  END
END

1 ICON "icon.ico"
EOF

# Convert icon.png to icon.ico if needed
if [ ! -f icon.ico ]; then
  if command -v convert &> /dev/null; then
    echo "Converting icon.png to icon.ico..."
    convert icon.png -define icon:auto-resize=64,48,32,16 icon.ico
  else
    echo "Warning: ImageMagick not found, can't convert icon. Using default icon."
    # Copy a default icon or create a minimal one
  fi
fi

# Create version information
# Fix the ldflags parameter format
VERSION_LDFLAGS="-X main.Version=$VERSION -X main.BuildDate=$BUILD_DATE -X main.CommitSHA=$COMMIT_SHA"

# Create directory for Windows release
mkdir -p ./releases/windows

# Compile the Windows version with version info
echo "Compiling Windows executable..."
GOOS=windows GOARCH=amd64 go build -ldflags "$VERSION_LDFLAGS" -o "./releases/windows/$APP_NAME.exe"

# Generate SHA-256 checksum for security verification
echo "Generating SHA-256 checksum..."
if command -v sha256sum &> /dev/null; then
  (cd ./releases/windows && sha256sum "$APP_NAME.exe" > "$APP_NAME.exe.sha256")
elif command -v shasum &> /dev/null; then
  (cd ./releases/windows && shasum -a 256 "$APP_NAME.exe" > "$APP_NAME.exe.sha256")
else
  echo "Warning: Could not generate SHA-256 checksum, neither sha256sum nor shasum found."
fi

# Copy verification scripts if they exist
if [ -f "verify.ps1" ]; then
  cp verify.ps1 ./releases/windows/
fi
if [ -f "SECURITY.md" ]; then
  cp SECURITY.md ./releases/windows/VERIFICATION.txt
fi

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
