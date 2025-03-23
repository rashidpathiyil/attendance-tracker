# Building for Windows

This guide explains how to build the Attendance Tracker application for Windows.

## Prerequisites

1. **Install Go for Windows**
   - Download and install Go from [golang.org](https://golang.org/dl/)
   - Make sure `go` is in your PATH

2. **Install Git for Windows**
   - Download and install Git from [git-scm.com](https://git-scm.com/download/win)
   - Make sure `git` is in your PATH

3. **Install GCC/MinGW**
   - Download and install [TDM-GCC](https://jmeubank.github.io/tdm-gcc/)
   - Make sure the compiler is in your PATH

4. **Install Fyne CLI**
   - Open a command prompt and run:
   ```
   go install fyne.io/fyne/v2/cmd/fyne@latest
   ```

## Building

### Option 1: Using the batch script

1. Clone the repository:
   ```
   git clone https://github.com/yourusername/attendance-tracker.git
   cd attendance-tracker
   ```

2. Run the build script:
   ```
   build-windows.bat
   ```

3. The executable will be created in `releases\windows\` directory.

### Option 2: Using Make

1. Install Make for Windows
   - Download and install [Make for Windows](http://gnuwin32.sourceforge.net/packages/make.htm)
   - Make sure `make` is in your PATH

2. Build the Windows executable:
   ```
   make windows
   ```

3. Package the application:
   ```
   make package-windows
   ```

4. Create a release:
   ```
   mkdir -p releases\windows
   copy "Attendance Tracker.exe" releases\windows\
   ```

### Option 3: Manual build

1. Build with Go:
   ```
   go build -ldflags="-H windowsgui" -o "Attendance Tracker.exe"
   ```

2. Package with Fyne:
   ```
   fyne package -os windows -icon icon.png -name "Attendance Tracker" -appID com.example.attendancetracker -release
   ```

## Testing

1. Run the executable:
   ```
   "Attendance Tracker.exe"
   ```

2. Verify that the application starts and functions correctly.

## Distribution

1. The built executable is self-contained and can be distributed directly.

2. For a cleaner distribution, run the cleanup script (if on Linux/macOS) or manually organize the files:
   ```
   ./cleanup.sh  # On Linux/macOS
   ```

3. The final Windows executable will be in the `releases/windows/` directory. 
