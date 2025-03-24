@echo off
echo ===================================================
echo Attendance Tracker - Installation Helper
echo ===================================================
echo.
echo This script will:
echo 1. Verify the authenticity of the executable
echo 2. Run the application if verification passes
echo.
echo Press Ctrl+C to cancel or any key to continue...
pause > nul

echo.
echo Verifying executable...
powershell -ExecutionPolicy Bypass -File verify.ps1 "Attendance Tracker.exe"
if %ERRORLEVEL% NEQ 0 (
    echo.
    echo Verification failed! The executable may have been tampered with.
    echo Please download a fresh copy from the official GitHub repository.
    echo.
    echo Press any key to exit...
    pause > nul
    exit /b 1
)

echo.
echo Verification successful! Starting Attendance Tracker...
echo.
start "" "Attendance Tracker.exe"

echo.
echo If Windows SmartScreen appears, click "More Info" then "Run anyway"
echo.
echo Press any key to exit this installer...
pause > nul
exit /b 0 
