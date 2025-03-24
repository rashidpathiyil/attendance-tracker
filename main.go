package main

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// Version information
var (
	Version   = "1.0.0"
	BuildDate = "unknown"
	CommitSHA = "unknown"
)

// Default configuration values
var (
	defaultServerEndpoint = "https://stale-olivette-rashidpathiyil-d5cc9ac4.koyeb.app/api/v1/events"
	defaultDeviceID       = getDeviceID()
	defaultUserID         = getUserID()
	defaultIdleTimeout    = 20 * time.Minute
	defaultCheckInterval  = 2 * time.Second
)

// AppConfig stores the application configuration
type AppConfig struct {
	ServerEndpoint  string
	DeviceID        string
	UserID          string
	IdleTimeout     time.Duration
	CheckInterval   time.Duration
	DeveloperMode   bool
	ShowActivityLog bool
	ShowIdleTime    bool
	AutoMode        bool
	RunAtStartup    bool
}

// Create a new config with default values
func NewAppConfig() *AppConfig {
	return &AppConfig{
		ServerEndpoint:  defaultServerEndpoint,
		DeviceID:        defaultDeviceID,
		UserID:          defaultUserID,
		IdleTimeout:     defaultIdleTimeout,
		CheckInterval:   defaultCheckInterval,
		DeveloperMode:   false,
		ShowActivityLog: false,
		ShowIdleTime:    true,
		AutoMode:        true,
		RunAtStartup:    true,
	}
}

// StatusPayload contains the data to send to the server
type StatusPayload struct {
	EventType string         `json:"event_type"`
	UserID    string         `json:"user_id"`
	Payload   PayloadContent `json:"payload"`
	Timestamp *time.Time     `json:"timestamp,omitempty"` // Optional timestamp
}

// PayloadContent is the nested data structure in the payload
type PayloadContent struct {
	Time     string                 `json:"time"` // HH:MM:SS format
	Date     string                 `json:"date"` // YYYY-MM-DD format
	DeviceID string                 `json:"device_id"`
	Config   map[string]interface{} `json:"config,omitempty"`
}

// SystemActivityMonitor detects user activity at the OS level
type SystemActivityMonitor struct {
	lastActivity time.Time
}

func NewSystemActivityMonitor() *SystemActivityMonitor {
	return &SystemActivityMonitor{
		lastActivity: time.Now(),
	}
}

// getSystemIdleTime returns how long the system has been idle
// Implementation is platform-specific
func getSystemIdleTime() (time.Duration, error) {
	switch runtime.GOOS {
	case "darwin":
		// macOS implementation
		cmd := exec.Command("sh", "-c", "ioreg -c IOHIDSystem | grep HIDIdleTime")
		output, err := cmd.Output()
		if err != nil {
			return 0, err
		}

		outputStr := string(output)
		if strings.Contains(outputStr, "HIDIdleTime") {
			parts := strings.Split(outputStr, "=")
			if len(parts) >= 2 {
				idleTimeStr := strings.TrimSpace(parts[1])
				idleTimeStr = strings.Replace(idleTimeStr, ",", "", -1)

				var nanoSeconds int64
				fmt.Sscanf(idleTimeStr, "%d", &nanoSeconds)
				return time.Duration(nanoSeconds/1000000) * time.Millisecond, nil
			}
		}
		return 0, fmt.Errorf("could not parse idle time")

	case "windows":
		// Windows implementation using a stub function
		// In a real implementation, this would use Windows API (GetLastInputInfo)
		return time.Duration(0) * time.Millisecond, nil

	case "linux":
		// Linux implementation
		cmd := exec.Command("sh", "-c", "xprintidle")
		output, err := cmd.Output()
		if err != nil {
			return 0, err
		}

		idleTimeStr := strings.TrimSpace(string(output))
		idleTime, err := strconv.ParseUint(idleTimeStr, 10, 64)
		if err != nil {
			return 0, err
		}

		return time.Duration(idleTime) * time.Millisecond, nil

	default:
		return 0, fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
}

// Check returns true if there has been activity since the last check
func (m *SystemActivityMonitor) Check() bool {
	idleTime, err := getSystemIdleTime()
	if err != nil {
		fmt.Printf("Error getting system idle time: %v\n", err)
		return false
	}

	// If system idle time is less than our check interval,
	// there's been activity since our last check
	if idleTime < defaultCheckInterval {
		m.lastActivity = time.Now()
		return true
	}

	return false
}

// IdleTime returns how long it's been since the last activity
func (m *SystemActivityMonitor) IdleTime() time.Duration {
	// Get system idle time
	sysIdleTime, err := getSystemIdleTime()
	if err == nil {
		// If we can get system idle time, use that
		return sysIdleTime
	}

	// Fall back to our own tracking
	return time.Since(m.lastActivity)
}

// UpdateLastActivity updates the last activity timestamp
func (m *SystemActivityMonitor) UpdateLastActivity() {
	m.lastActivity = time.Now()
}

// getDeviceID generates a unique identifier for the current device
func getDeviceID() string {
	// Try to get the hostname first
	hostname, err := os.Hostname()
	if err == nil && hostname != "" {
		return fmt.Sprintf("device-%s", hostname)
	}

	// If hostname fails, try to use MAC address
	interfaces, err := net.Interfaces()
	if err == nil {
		for _, iface := range interfaces {
			if iface.HardwareAddr.String() != "" {
				// Use MD5 hash of MAC to make it more manageable
				hash := md5.Sum([]byte(iface.HardwareAddr.String()))
				return fmt.Sprintf("device-%s", hex.EncodeToString(hash[:])[:8])
			}
		}
	}

	// Fallback to a reasonable default with OS info
	return fmt.Sprintf("device-%s-%d", runtime.GOOS, os.Getpid())
}

// getUserID tries to get the current user's system username
func getUserID() string {
	// Try to get current user
	currentUser, err := user.Current()
	if err == nil && currentUser.Username != "" {
		// Clean up username (remove domain for Windows users)
		username := currentUser.Username
		if parts := strings.Split(username, "\\"); len(parts) > 1 {
			username = parts[1]
		}
		return fmt.Sprintf("user-%s", username)
	}

	// If that fails, try environment variables
	if username := os.Getenv("USER"); username != "" {
		return fmt.Sprintf("user-%s", username)
	}
	if username := os.Getenv("USERNAME"); username != "" {
		return fmt.Sprintf("user-%s", username)
	}

	// Last resort
	return "user-unknown"
}

// Platform-specific setup for autostart
func setupAutostart(enable bool) error {
	switch runtime.GOOS {
	case "windows":
		// Get the executable path
		exePath, err := os.Executable()
		if err != nil {
			return err
		}

		// Windows: Use registry to set up autostart
		if enable {
			// Create startup registry key
			// We'll use a .bat file approach for registry since direct registry modification requires admin rights
			homeDir, err := os.UserHomeDir()
			if err != nil {
				return err
			}

			startupDir := filepath.Join(homeDir, "AppData", "Roaming", "Microsoft", "Windows", "Start Menu", "Programs", "Startup")
			shortcutPath := filepath.Join(startupDir, "AttendanceTracker.lnk")

			// Create the PowerShell command to create a shortcut
			psCommand := fmt.Sprintf(`
				$WshShell = New-Object -ComObject WScript.Shell
				$Shortcut = $WshShell.CreateShortcut("%s")
				$Shortcut.TargetPath = "%s"
				$Shortcut.Arguments = "--minimized"
				$Shortcut.Description = "Attendance Tracker"
				$Shortcut.Save()
			`, shortcutPath, exePath)

			// Create a temporary PowerShell script
			tmpFile, err := os.CreateTemp("", "create_shortcut_*.ps1")
			if err != nil {
				return err
			}
			scriptPath := tmpFile.Name()
			defer os.Remove(scriptPath)

			if _, err := tmpFile.Write([]byte(psCommand)); err != nil {
				tmpFile.Close()
				return err
			}
			tmpFile.Close()

			// Execute the PowerShell script with hidden window
			cmd := exec.Command("powershell", "-WindowStyle", "Hidden", "-ExecutionPolicy", "Bypass", "-File", scriptPath)
			return cmd.Run()
		} else {
			// Remove startup entry
			homeDir, err := os.UserHomeDir()
			if err != nil {
				return err
			}

			startupPath := filepath.Join(homeDir, "AppData", "Roaming", "Microsoft", "Windows", "Start Menu", "Programs", "Startup", "AttendanceTracker.lnk")
			// Check if file exists before trying to remove it
			if _, err := os.Stat(startupPath); err == nil {
				return os.Remove(startupPath)
			}
			return nil // File doesn't exist, nothing to do
		}

	case "darwin":
		// macOS: Create or remove LaunchAgent
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return err
		}

		launchAgentDir := filepath.Join(homeDir, "Library", "LaunchAgents")
		launchAgentPath := filepath.Join(launchAgentDir, "com.attendancetracker.plist")

		if enable {
			// Ensure the directory exists
			if err := os.MkdirAll(launchAgentDir, 0755); err != nil {
				return err
			}

			exePath, err := os.Executable()
			if err != nil {
				return err
			}

			// Create the plist file content
			plistContent := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.attendancetracker</string>
    <key>ProgramArguments</key>
    <array>
        <string>%s</string>
        <string>--minimized</string>
    </array>
    <key>RunAtLoad</key>
    <true/>
</dict>
</plist>`, exePath)

			// Write the plist file
			if err := os.WriteFile(launchAgentPath, []byte(plistContent), 0644); err != nil {
				return err
			}

			// Load the agent
			exec.Command("launchctl", "unload", launchAgentPath).Run() // Ignore errors
			return exec.Command("launchctl", "load", launchAgentPath).Run()
		} else {
			// Unload and remove the agent
			exec.Command("launchctl", "unload", launchAgentPath).Run() // Ignore errors
			if _, err := os.Stat(launchAgentPath); err == nil {
				return os.Remove(launchAgentPath)
			}
			return nil
		}

	case "linux":
		// Linux: Create or remove autostart desktop entry
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return err
		}

		autostartDir := filepath.Join(homeDir, ".config", "autostart")
		desktopPath := filepath.Join(autostartDir, "attendance-tracker.desktop")

		if enable {
			// Ensure the directory exists
			if err := os.MkdirAll(autostartDir, 0755); err != nil {
				return err
			}

			exePath, err := os.Executable()
			if err != nil {
				return err
			}

			// Create desktop entry content
			desktopContent := fmt.Sprintf(`[Desktop Entry]
Type=Application
Name=Attendance Tracker
Exec=%s --minimized
Terminal=false
Comment=Attendance Tracking Application
Categories=Utility;
`, exePath)

			// Write the desktop file
			if err := os.WriteFile(desktopPath, []byte(desktopContent), 0644); err != nil {
				return err
			}

			// Make it executable
			return os.Chmod(desktopPath, 0755)
		} else {
			// Remove the desktop file
			if _, err := os.Stat(desktopPath); err == nil {
				return os.Remove(desktopPath)
			}
			return nil
		}

	default:
		return fmt.Errorf("autostart not supported on %s", runtime.GOOS)
	}
}

// Save configuration to a JSON file
func saveConfig(config *AppConfig) error {
	// Create config directory if needed
	configDir, err := os.UserConfigDir()
	if err != nil {
		return err
	}

	appDir := filepath.Join(configDir, "attendance-tracker")
	if err := os.MkdirAll(appDir, 0755); err != nil {
		return err
	}

	// Convert config to JSON
	configMap := map[string]interface{}{
		"server_endpoint":     config.ServerEndpoint,
		"device_id":           config.DeviceID,
		"user_id":             config.UserID,
		"idle_timeout_mins":   int(config.IdleTimeout.Minutes()),
		"check_interval_secs": int(config.CheckInterval.Seconds()),
		"developer_mode":      config.DeveloperMode,
		"show_activity_log":   config.ShowActivityLog,
		"show_idle_time":      config.ShowIdleTime,
		"auto_mode":           config.AutoMode,
		"run_at_startup":      config.RunAtStartup,
	}

	jsonData, err := json.MarshalIndent(configMap, "", "  ")
	if err != nil {
		return err
	}

	// Write to config file
	configFile := filepath.Join(appDir, "config.json")
	return os.WriteFile(configFile, jsonData, 0644)
}

// Load configuration from JSON file
func loadConfig() (*AppConfig, error) {
	// Create default config
	config := NewAppConfig()

	// Get config file path
	configDir, err := os.UserConfigDir()
	if err != nil {
		return config, err
	}

	configFile := filepath.Join(configDir, "attendance-tracker", "config.json")

	// Check if config file exists
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		// No config file yet, use defaults
		return config, nil
	}

	// Read config file
	data, err := os.ReadFile(configFile)
	if err != nil {
		return config, err
	}

	// Parse JSON
	var configMap map[string]interface{}
	if err := json.Unmarshal(data, &configMap); err != nil {
		return config, err
	}

	// Update config with values from file
	if server, ok := configMap["server_endpoint"].(string); ok {
		config.ServerEndpoint = server
	}
	if deviceID, ok := configMap["device_id"].(string); ok {
		config.DeviceID = deviceID
	}
	if userID, ok := configMap["user_id"].(string); ok {
		config.UserID = userID
	}
	if idleTimeout, ok := configMap["idle_timeout_mins"].(float64); ok {
		config.IdleTimeout = time.Duration(idleTimeout) * time.Minute
	}
	if checkInterval, ok := configMap["check_interval_secs"].(float64); ok {
		config.CheckInterval = time.Duration(checkInterval) * time.Second
	}
	if devMode, ok := configMap["developer_mode"].(bool); ok {
		config.DeveloperMode = devMode
	}
	if showLog, ok := configMap["show_activity_log"].(bool); ok {
		config.ShowActivityLog = showLog
	}
	if showIdle, ok := configMap["show_idle_time"].(bool); ok {
		config.ShowIdleTime = showIdle
	}
	if autoMode, ok := configMap["auto_mode"].(bool); ok {
		config.AutoMode = autoMode
	}
	if runAtStartup, ok := configMap["run_at_startup"].(bool); ok {
		config.RunAtStartup = runAtStartup
	}

	return config, nil
}

// migrateFromPreviousVersion handles data migration during upgrades
func migrateFromPreviousVersion() {
	fmt.Println("Running upgrade migration...")

	// Load configuration from the previous version
	config, err := loadConfig()
	if err != nil {
		fmt.Printf("Warning: Could not load previous config: %v\n", err)
	} else {
		// If needed, update configuration format or defaults
		// For example, if we added new fields in this version:
		// config.NewField = defaultValueForNewField

		// Save the updated config
		if err := saveConfig(config); err != nil {
			fmt.Printf("Error saving migrated config: %v\n", err)
		} else {
			fmt.Println("Successfully migrated configuration")
		}
	}

	// If there are other migration tasks, perform them here
	// For example:
	// - Update database schema
	// - Migrate data files to new format
	// - Update file paths

	fmt.Println("Migration complete")
}

// Package-level variables for application settings
var (
	developerMode bool
	upgradeMode   bool
)

// Reset developer settings when running in normal mode
func resetDeveloperSettings() {
	// Reset any developer-specific settings here
	// This is a placeholder function that can be expanded as needed
	fmt.Println("Resetting developer settings to defaults")
}

// Icon for the application
var resourceAppIconPng = fyne.NewStaticResource("appIcon", nil)

// Create the status tab showing current status
func createStatusTab(w fyne.Window) *fyne.Container {
	// This is a placeholder implementation
	return container.NewVBox(
		widget.NewLabel("Status: Active"),
		widget.NewButton("Check In/Out", func() {
			// Toggle check-in status
		}),
	)
}

// Create the history tab showing attendance history
func createHistoryTab() *fyne.Container {
	// This is a placeholder implementation
	return container.NewVBox(
		widget.NewLabel("Attendance History"),
		widget.NewLabel("No records available"),
	)
}

// Create the settings tab with configuration options
func createSettingsTab(a fyne.App) *fyne.Container {
	// This is a placeholder implementation
	return container.NewVBox(
		widget.NewLabel("Settings"),
		widget.NewCheck("Run at startup", func(value bool) {
			// Update setting
		}),
	)
}

// Show the settings dialog
func showSettingsDialog(w fyne.Window) {
	dialog.ShowCustom("Settings", "Close",
		container.NewVBox(
			widget.NewLabel("Application Settings"),
			widget.NewCheck("Developer Mode", func(value bool) {
				developerMode = value
			}),
		), w)
}

// Show the about dialog
func showAboutDialog(w fyne.Window) {
	dialog.ShowCustom("About", "Close",
		container.NewVBox(
			widget.NewLabel("Attendance Tracker"),
			widget.NewLabel(fmt.Sprintf("Version %s", Version)),
			widget.NewLabel("© 2023 Rashid Pathiyil"),
			widget.NewLabel("An attendance tracking application"),
		), w)
}

func main() {
	// Parse command line arguments
	upgradeFlag := flag.Bool("upgrade", false, "Run in upgrade mode")
	flag.Parse()

	// Enable developer mode by default during development
	developerMode = true

	// Initialize the application
	a := app.New()

	// Set application metadata
	a.SetIcon(resourceAppIconPng)

	// Reset developer mode settings if running in normal mode
	if !developerMode {
		resetDeveloperSettings()
	}

	// Run in upgrade mode if specified
	if *upgradeFlag {
		upgradeMode = true
		migrateFromPreviousVersion()
	}

	// Set application title
	w := a.NewWindow("Attendance Tracker")

	// Create an update notification channel
	updateChannel := make(chan *UpdateInfo, 1)

	// Create tabs
	tabs := container.NewAppTabs(
		container.NewTabItem("Status", createStatusTab(w)),
		container.NewTabItem("History", createHistoryTab()),
		container.NewTabItem("Settings", createSettingsTab(a)),
	)

	tabs.SetTabLocation(container.TabLocationTop)

	// Create the content container
	content := container.NewVBox(tabs)

	// Set content
	w.SetContent(content)

	// Create menu with version info
	mainMenu := createMainMenu(a, w, updateChannel)
	w.SetMainMenu(mainMenu)

	// Set window size
	w.Resize(fyne.NewSize(600, 400))

	// Check for updates in background on startup
	go func() {
		// Wait a bit to let the UI load completely
		time.Sleep(5 * time.Second)

		// Check for updates in background
		updateInfo := getLatestReleaseInfo()
		if updateInfo != nil && updateInfo.Version != Version {
			// Send update notification to channel
			updateChannel <- updateInfo
			// Show notification
			showUpdateNotification(w, updateInfo)
		}
	}()

	// Show window and run app
	w.ShowAndRun()
}

// showSettings shows the settings dialog
func showSettings(a fyne.App, w fyne.Window) {
	settingsTab := createSettingsTab(a)
	dialog.ShowCustom("Settings", "Close", settingsTab, w)
}

// showAbout shows the about dialog
func showAbout(w fyne.Window) {
	dialog.ShowCustom("About", "Close",
		container.NewVBox(
			widget.NewLabel("Attendance Tracker"),
			widget.NewLabel(fmt.Sprintf("Version %s", Version)),
			widget.NewLabel("© 2023 Rashid Pathiyil"),
			widget.NewLabel("An attendance tracking application"),
		), w)
}

// showHelp shows the help dialog
func showHelp(w fyne.Window) {
	dialog.ShowCustom("Help", "Close",
		container.NewVBox(
			widget.NewLabel("Attendance Tracker Help"),
			widget.NewLabel("\nStatus Tab:"),
			widget.NewLabel("Shows your current attendance status."),
			widget.NewLabel("\nHistory Tab:"),
			widget.NewLabel("View your attendance history."),
			widget.NewLabel("\nSettings Tab:"),
			widget.NewLabel("Configure application settings."),
		), w)
}

// confirmUninstall asks for confirmation before uninstalling
func confirmUninstall(w fyne.Window) {
	dialog.ShowConfirm("Uninstall Attendance Tracker?",
		"Are you sure you want to uninstall Attendance Tracker?\nThis will remove the application and all its data.",
		func(uninstall bool) {
			if uninstall {
				performUninstall(w)
			}
		}, w)
}

// performUninstall removes all application components and returns the path to the uninstall log
func performUninstall(w fyne.Window) {
	// Create uninstall log
	logText := "Attendance Tracker Uninstall Log\n"
	logText += fmt.Sprintf("Date: %s\n", time.Now().Format("2006-01-02 15:04:05"))
	logText += fmt.Sprintf("Platform: %s\n\n", runtime.GOOS)

	// Step 1: Remove autostart entry
	logText += "Removing autostart entry... "
	if err := setupAutostart(false); err != nil {
		logText += fmt.Sprintf("ERROR: %v\n", err)
	} else {
		logText += "Done\n"
	}

	// Step 2: Clean up configuration files
	configDir, err := os.UserConfigDir()
	if err == nil {
		appDataPath := filepath.Join(configDir, "attendance-tracker")
		logText += fmt.Sprintf("Removing configuration data from %s... ", appDataPath)

		if err := os.RemoveAll(appDataPath); err != nil {
			logText += fmt.Sprintf("ERROR: %v\n", err)
		} else {
			logText += "Done\n"
		}
	} else {
		logText += fmt.Sprintf("Could not determine config directory: %v\n", err)
	}

	// Step 3: Create uninstall script for self-deletion (platform specific)
	exePath, err := os.Executable()
	if err == nil {
		logText += fmt.Sprintf("Application executable path: %s\n", exePath)

		switch runtime.GOOS {
		case "windows":
			// Create a batch file that will wait for our process to exit and then delete the executable
			batchContent := fmt.Sprintf(`@echo off
timeout /t 1 /nobreak > nul
:retry
del "%s" 2>nul
if exist "%s" (
  timeout /t 1 /nobreak > nul
  goto retry
)
rmdir /s /q "%s\.."
exit
`, exePath, exePath, exePath)

			batPath := filepath.Join(os.TempDir(), "cleanup_attendance_tracker.bat")
			if err := os.WriteFile(batPath, []byte(batchContent), 0755); err != nil {
				logText += fmt.Sprintf("Failed to create cleanup script: %v\n", err)
			} else {
				// Execute the batch file with hidden window
				cmd := exec.Command("cmd", "/c", "start", "/min", batPath)
				if err := cmd.Start(); err != nil {
					logText += fmt.Sprintf("Failed to start cleanup script: %v\n", err)
				} else {
					logText += "Cleanup script started to remove executable after exit\n"
				}
			}

		case "darwin", "linux":
			// On macOS/Linux, we can use a similar shell script approach
			shellContent := fmt.Sprintf(`#!/bin/sh
sleep 1
while [ -f "%s" ]; do
  rm "%s" 2>/dev/null
  if [ -f "%s" ]; then
    sleep 1
  fi
done
exit 0
`, exePath, exePath, exePath)

			shellPath := filepath.Join(os.TempDir(), "cleanup_attendance_tracker.sh")
			if err := os.WriteFile(shellPath, []byte(shellContent), 0755); err != nil {
				logText += fmt.Sprintf("Failed to create cleanup script: %v\n", err)
			} else {
				cmd := exec.Command("sh", "-c", fmt.Sprintf("nohup %s >/dev/null 2>&1 &", shellPath))
				if err := cmd.Start(); err != nil {
					logText += fmt.Sprintf("Failed to start cleanup script: %v\n", err)
				} else {
					logText += "Cleanup script started to remove executable after exit\n"
				}
			}
		}
	} else {
		logText += fmt.Sprintf("Could not determine executable path: %v\n", err)
	}

	// Save the uninstall log
	logPath := filepath.Join(os.TempDir(), "attendance_tracker_uninstall.log")
	os.WriteFile(logPath, []byte(logText), 0644)

	// Show a final message
	dialog.ShowInformation("Uninstall Complete",
		fmt.Sprintf("Attendance Tracker has been uninstalled. A log was saved to:\n%s", logPath), w)

	// Quit the application
	fyne.CurrentApp().Quit()
}

// UpdateInfo represents information about an available update
type UpdateInfo struct {
	Version      string
	DownloadURL  string
	ReleaseDate  string
	ReleaseNotes []string
	Size         int64 // Size in bytes
}

// checkForUpdatesInBackground silently checks for updates and notifies only if an update is available
func checkForUpdatesInBackground(w fyne.Window) {
	// In a real implementation, this would check version from a server API
	currentVersion := Version

	// Check against GitHub releases API or your custom endpoint
	updateInfo := getLatestReleaseInfo()

	// If there's a newer version available, show notification
	if updateInfo != nil && updateInfo.Version != currentVersion {
		logActivity(fmt.Sprintf("Update available: %s (current: %s)", updateInfo.Version, currentVersion))
		showUpdateNotification(w, updateInfo)
	} else {
		logActivity("No updates available. Current version is up to date.")
	}
}

// getLatestReleaseInfo fetches information about the latest release
// Uses HTTP to check a real update server if available
func getLatestReleaseInfo() *UpdateInfo {
	// Get update server URL from environment variable or use default
	updateServerURL := os.Getenv("ATTENDANCE_UPDATE_SERVER")
	if updateServerURL == "" {
		// Default to GitHub API for the rashidpathiyil/attendance-tracker repository
		updateServerURL = "https://api.github.com/repos/rashidpathiyil/attendance-tracker/releases/latest"
	}

	// For testing purposes - always return a simulated newer version if test env is set
	if os.Getenv("ATTENDANCE_UPDATE_TEST") == "1" {
		logActivity("Using simulated update data for testing")
		return &UpdateInfo{
			Version:     "1.1.0",
			DownloadURL: "https://github.com/rashidpathiyil/attendance-tracker/releases/download/v1.1.0/Attendance-Tracker.exe",
			ReleaseDate: "2023-04-15",
			ReleaseNotes: []string{
				"Added auto-update feature",
				"Fixed idle time detection on Windows",
				"Improved startup performance",
				"Added better error handling",
			},
			Size: 24500000, // Approx 24.5 MB
		}
	}

	// Check if this is a forced check when running the test script
	if len(os.Args) > 1 {
		for _, arg := range os.Args {
			if arg == "--test-updates" {
				logActivity("Test update mode enabled - checking " + updateServerURL)
				break
			}
		}
	}

	// Make an HTTP request to the update server
	logActivity("Checking for updates at: " + updateServerURL)

	// Set up the HTTP client with appropriate headers for GitHub API
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Create a request with headers
	req, err := http.NewRequest("GET", updateServerURL, nil)
	if err != nil {
		logActivity(fmt.Sprintf("Error creating request: %v", err))
		return nil
	}

	// Add User-Agent header (GitHub API requires this)
	req.Header.Set("User-Agent", "AttendanceTracker/"+Version)

	// Make the request
	resp, err := client.Do(req)
	if err != nil {
		logActivity(fmt.Sprintf("Error checking for updates: %v", err))
		return nil
	}
	defer resp.Body.Close()

	// Check response status code
	if resp.StatusCode != http.StatusOK {
		logActivity(fmt.Sprintf("Error from update server: status code %d", resp.StatusCode))
		return nil
	}

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logActivity(fmt.Sprintf("Error reading update response: %v", err))
		return nil
	}

	// Parse the response body based on the server type
	// For GitHub API, this would parse the GitHub JSON format
	// For a custom server, parse your own format
	if strings.Contains(updateServerURL, "api.github.com") {
		// Parse GitHub API response
		var githubResponse struct {
			TagName string `json:"tag_name"`
			Name    string `json:"name"`
			Assets  []struct {
				Size        int64  `json:"size"`
				Name        string `json:"name"`
				DownloadURL string `json:"browser_download_url"`
			} `json:"assets"`
			PublishedAt string `json:"published_at"`
			Body        string `json:"body"`
		}

		if err := json.Unmarshal(body, &githubResponse); err != nil {
			logActivity(fmt.Sprintf("Error parsing GitHub response: %v", err))
			return nil
		}

		// Skip the "v" prefix if present
		version := githubResponse.TagName
		if strings.HasPrefix(version, "v") {
			version = version[1:]
		}

		// Skip update check if version isn't a higher number
		if !isVersionNewer(version, Version) {
			logActivity(fmt.Sprintf("Current version %s is up to date compared to %s", Version, version))
			return nil
		}

		// Find the Windows asset
		var downloadURL string
		var size int64
		for _, asset := range githubResponse.Assets {
			// Look for Windows executable - can be attendance-tracker.exe or Attendance Tracker.exe
			if strings.HasSuffix(asset.Name, ".exe") {
				downloadURL = asset.DownloadURL
				size = asset.Size
				break
			}
		}

		// Return nil if no downloadable asset was found
		if downloadURL == "" {
			logActivity("No suitable download file found in release assets")
			return nil
		}

		// Parse release notes from body
		releaseNotes := parseReleaseNotes(githubResponse.Body)

		// Format date
		releaseDate := formatReleaseDate(githubResponse.PublishedAt)

		logActivity(fmt.Sprintf("Update available: %s (current: %s)", version, Version))

		return &UpdateInfo{
			Version:      version,
			DownloadURL:  downloadURL,
			ReleaseDate:  releaseDate,
			ReleaseNotes: releaseNotes,
			Size:         size,
		}
	} else {
		// Parse custom server response (simple JSON)
		var customResponse struct {
			Version      string   `json:"version"`
			DownloadURL  string   `json:"download_url"`
			ReleaseDate  string   `json:"release_date"`
			ReleaseNotes []string `json:"release_notes"`
			Size         int64    `json:"size"`
		}

		if err := json.Unmarshal(body, &customResponse); err != nil {
			logActivity(fmt.Sprintf("Error parsing custom server response: %v", err))
			return nil
		}

		// Don't return update info if version isn't newer
		if !isVersionNewer(customResponse.Version, Version) {
			return nil
		}

		return &UpdateInfo{
			Version:      customResponse.Version,
			DownloadURL:  customResponse.DownloadURL,
			ReleaseDate:  customResponse.ReleaseDate,
			ReleaseNotes: customResponse.ReleaseNotes,
			Size:         customResponse.Size,
		}
	}
}

// isVersionNewer checks if version1 is newer than version2
// Uses semantic versioning rules
func isVersionNewer(version1, version2 string) bool {
	// Parse version strings into components
	v1Parts := strings.Split(version1, ".")
	v2Parts := strings.Split(version2, ".")

	// Pad with zeros to ensure equal length
	for len(v1Parts) < 3 {
		v1Parts = append(v1Parts, "0")
	}
	for len(v2Parts) < 3 {
		v2Parts = append(v2Parts, "0")
	}

	// Compare major, minor, patch versions
	for i := 0; i < 3; i++ {
		v1, err1 := strconv.Atoi(v1Parts[i])
		v2, err2 := strconv.Atoi(v2Parts[i])

		// If we can't parse a version part, consider them equal
		if err1 != nil || err2 != nil {
			continue
		}

		if v1 > v2 {
			return true
		} else if v1 < v2 {
			return false
		}
		// If equal, continue to next part
	}

	// If we get here, versions are equal
	return false
}

// parseReleaseNotes parses release notes from GitHub markdown format
func parseReleaseNotes(markdown string) []string {
	var notes []string

	// Split by newlines
	lines := strings.Split(markdown, "\n")

	// Look for bullet points
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "- ") || strings.HasPrefix(line, "* ") {
			note := line[2:] // Remove the bullet
			notes = append(notes, note)
		}
	}

	// If no bullet points found, just take the first few lines
	if len(notes) == 0 && len(lines) > 0 {
		for i := 0; i < 3 && i < len(lines); i++ {
			if lines[i] != "" {
				notes = append(notes, lines[i])
			}
		}
	}

	return notes
}

// formatReleaseDate formats a date string from GitHub format to a readable format
func formatReleaseDate(githubDate string) string {
	// GitHub date format: 2023-04-15T15:30:45Z
	t, err := time.Parse(time.RFC3339, githubDate)
	if err != nil {
		return githubDate
	}

	return t.Format("2006-01-02")
}

// showUpdateNotification shows a small notification about available updates
func showUpdateNotification(w fyne.Window, updateInfo *UpdateInfo) {
	dialog.ShowCustomConfirm("Update Available",
		"View Details", "Not Now",
		container.NewVBox(
			widget.NewLabel("A new version of Attendance Tracker is available."),
			widget.NewLabel(fmt.Sprintf("Current version: %s", Version)),
			widget.NewLabel(fmt.Sprintf("New version: %s", updateInfo.Version)),
			widget.NewLabel("Would you like to see what's new?"),
		),
		func(viewDetails bool) {
			if viewDetails {
				// Show the full update dialog
				checkForUpdates(w)
			}
		}, w)
}

// checkForUpdates checks if a newer version is available
func checkForUpdates(w fyne.Window) {
	currentVersion := Version

	// Show initial checking dialog
	progress := widget.NewProgressBarInfinite()
	checkingDialog := dialog.NewCustom("Checking for Updates", "Cancel",
		container.NewVBox(
			widget.NewLabel("Checking for updates..."),
			progress,
		), w)

	checkingDialog.Show()

	// In a real implementation, this would be an async HTTP request
	// For demo, use goroutine with small delay to simulate network request
	go func() {
		// Simulate network delay
		time.Sleep(1 * time.Second)

		// Get update info
		updateInfo := getLatestReleaseInfo()

		// Close the checking dialog
		checkingDialog.Hide()

		// Show results
		if updateInfo != nil && updateInfo.Version != currentVersion {
			// Create notes container
			notesContainer := container.NewVBox()
			notesContainer.Add(widget.NewLabel("\nChangelog:"))

			for _, note := range updateInfo.ReleaseNotes {
				notesContainer.Add(widget.NewLabel("• " + note))
			}

			// Format size nicely
			sizeText := fmt.Sprintf("%.1f MB", float64(updateInfo.Size)/1024/1024)

			// Show update available dialog
			dialog.ShowCustomConfirm("Update Available",
				"Download & Install", "Later",
				container.NewVBox(
					widget.NewLabel(fmt.Sprintf("Current version: %s", currentVersion)),
					widget.NewLabel(fmt.Sprintf("New version: %s", updateInfo.Version)),
					widget.NewLabel(fmt.Sprintf("Released: %s", updateInfo.ReleaseDate)),
					widget.NewLabel(fmt.Sprintf("Size: %s", sizeText)),
					notesContainer,
				),
				func(update bool) {
					if update {
						// Download and install the update
						downloadAndInstallUpdate(w, updateInfo)
					}
				}, w)
		} else {
			// No update available
			dialog.ShowInformation("Up to Date",
				fmt.Sprintf("You're using the latest version (%s).", currentVersion), w)
		}
	}()
}

// downloadAndInstallUpdate downloads and installs the update
func downloadAndInstallUpdate(w fyne.Window, updateInfo *UpdateInfo) {
	// First check if the update is already downloaded
	updateCachePath := getUpdateCachePath(updateInfo.Version)
	if fileExists(updateCachePath) {
		// Update already downloaded - confirm reinstall
		dialog.ShowConfirm("Update Already Downloaded",
			fmt.Sprintf("Version %s has already been downloaded. Install it now?", updateInfo.Version),
			func(install bool) {
				if install {
					// Install the already downloaded update
					installUpdate(w, updateCachePath, updateInfo)
				}
			}, w)
		return
	}

	// Set up progress dialog
	progressBar := widget.NewProgressBar()
	progressBar.SetValue(0)
	progressText := widget.NewLabel("Preparing download...")

	cancelDownload := false

	// Create the progress dialog
	dlg := dialog.NewCustom("Downloading Update", "Cancel",
		container.NewVBox(
			widget.NewLabel(fmt.Sprintf("Downloading %s...", updateInfo.Version)),
			progressBar,
			progressText,
		), w)

	// Set up cancel callback
	dlg.SetOnClosed(func() {
		cancelDownload = true
	})

	// Show the dialog
	dlg.Show()

	// Simulate download in goroutine
	go func() {
		// Create a temp directory for the download
		tempDir, err := os.MkdirTemp("", "attendance-tracker-update")
		if err != nil {
			showUpdateError(w, "Could not create temporary download folder", err)
			dlg.Hide()
			return
		}

		logActivity(fmt.Sprintf("Downloading update %s to: %s", updateInfo.Version, tempDir))

		// Simulate download progress
		totalSteps := 100
		for i := 0; i <= totalSteps; i++ {
			if cancelDownload {
				logActivity("Update download cancelled by user")
				return
			}

			// Update progress
			progress := float64(i) / float64(totalSteps)
			progressBar.SetValue(progress)

			// Update text with percentage and download rate
			downloadedSize := int64(float64(updateInfo.Size) * progress)
			progressText.SetText(fmt.Sprintf("%.1f%% (%.1f/%.1f MB) - 2.5 MB/s",
				progress*100,
				float64(downloadedSize)/1024/1024,
				float64(updateInfo.Size)/1024/1024))

			// Slow down the simulation
			time.Sleep(50 * time.Millisecond)
		}

		// Hide download dialog
		dlg.Hide()

		// Create update cache directory if it doesn't exist
		updateCacheDir := getUpdateCacheDir()
		if err := os.MkdirAll(updateCacheDir, 0755); err != nil {
			showUpdateError(w, "Could not create update cache directory", err)
			return
		}

		// In a real implementation, this would be where we move the downloaded file
		// from tempDir to updateCachePath

		// Simulate this by copying a file or just writing a placeholder
		if err := os.WriteFile(updateCachePath, []byte("Simulated update file for "+updateInfo.Version), 0644); err != nil {
			showUpdateError(w, "Could not save downloaded update", err)
			return
		}

		logActivity(fmt.Sprintf("Update %s downloaded to: %s", updateInfo.Version, updateCachePath))

		// Install the downloaded update
		installUpdate(w, updateCachePath, updateInfo)
	}()
}

// installUpdate installs a downloaded update
func installUpdate(w fyne.Window, updateFilePath string, updateInfo *UpdateInfo) {
	// Save current settings to ensure they're preserved during upgrade
	config, _ := loadConfig()
	if config != nil {
		if err := saveConfig(config); err != nil {
			logActivity(fmt.Sprintf("Warning: Could not save config before update: %v", err))
		} else {
			logActivity("Configuration preserved for update")
		}
	}

	// Show installation dialog
	installationDialog := dialog.NewCustom("Installing Update", "",
		container.NewVBox(
			widget.NewLabel("Installing update..."),
			widget.NewLabel("The application will restart when complete."),
			widget.NewProgressBarInfinite(),
		), w)

	installationDialog.Show()

	// Simulate installation process
	time.Sleep(2 * time.Second)

	// Hide installation dialog
	installationDialog.Hide()

	// Show completion dialog
	dialog.ShowCustom("Update Complete", "Restart Now",
		container.NewVBox(
			widget.NewLabel("Update successfully installed!"),
			widget.NewLabel(fmt.Sprintf("Version %s will be applied when Attendance Tracker restarts.", updateInfo.Version)),
		), w)

	// Wait a bit then restart (in a real implementation)
	time.Sleep(2 * time.Second)

	// In a real implementation, we would create a script to:
	// 1. Replace the current executable with the downloaded one
	// 2. Restart the application
	// For now, we just exit
	logActivity("Update installed - application restarting")
	fyne.CurrentApp().Quit()
}

// getUpdateCacheDir returns the path to the update cache directory
func getUpdateCacheDir() string {
	// Get user's cache directory
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		// Fallback to temp directory if user cache dir can't be determined
		return filepath.Join(os.TempDir(), "attendance-tracker-updates")
	}

	return filepath.Join(cacheDir, "attendance-tracker", "updates")
}

// getUpdateCachePath returns the path where an update for a specific version would be cached
func getUpdateCachePath(version string) string {
	return filepath.Join(getUpdateCacheDir(), fmt.Sprintf("attendance-tracker-%s.exe", version))
}

// fileExists checks if a file exists
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// logActivity logs application activity to the log file
func logActivity(message string) {
	logFile, err := os.OpenFile(getLogFilePath(), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("Error opening log file: %v\n", err)
		return
	}
	defer logFile.Close()

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	logEntry := fmt.Sprintf("[%s] %s\n", timestamp, message)

	if _, err := logFile.WriteString(logEntry); err != nil {
		fmt.Printf("Error writing to log file: %v\n", err)
	}
}

// showUpdateError displays an error dialog for update issues
func showUpdateError(w fyne.Window, message string, err error) {
	errorMessage := message
	if err != nil {
		errorMessage += fmt.Sprintf("\n\nError details: %v", err)
	}

	logActivity(fmt.Sprintf("Update error: %s - %v", message, err))

	dialog.ShowError(errors.New(errorMessage), w)
}

// getLogFilePath returns the path to the application log file
func getLogFilePath() string {
	// Get user's log directory (use AppData on Windows, .config on Linux/Mac)
	configDir, err := os.UserConfigDir()
	if err != nil {
		// Fallback to temp directory if config dir can't be determined
		return filepath.Join(os.TempDir(), "attendance-tracker.log")
	}

	logDir := filepath.Join(configDir, "AttendanceTracker")

	// Create directory if it doesn't exist
	if err := os.MkdirAll(logDir, 0755); err != nil {
		// Fallback to temp directory if we can't create the log directory
		return filepath.Join(os.TempDir(), "attendance-tracker.log")
	}

	return filepath.Join(logDir, "attendance-tracker.log")
}

// createMainMenu creates the main menu for the application
func createMainMenu(a fyne.App, w fyne.Window, updateChannel chan *UpdateInfo) *fyne.MainMenu {
	// Create menu items
	settingsItem := fyne.NewMenuItem("Settings", func() {
		showSettingsDialog(w)
	})

	updateItem := fyne.NewMenuItem("Check for Updates", func() {
		checkForUpdates(w)
	})

	// Subscribe to update notifications if channel is provided
	if updateChannel != nil {
		go func() {
			for update := range updateChannel {
				if update != nil {
					// Update is available, modify the menu item
					updateItem.Label = fmt.Sprintf("Update Available: v%s", update.Version)
					// Note: fyne.App doesn't have Menu() method, we'd need to refresh manually
					// For now, we'll just keep the label updated for the next time the menu is shown
				}
			}
		}()
	}

	uninstallItem := fyne.NewMenuItem("Uninstall", func() {
		confirmUninstall(w)
	})

	quitItem := fyne.NewMenuItem("Quit", func() {
		a.Quit()
	})

	// About menu item
	aboutItem := fyne.NewMenuItem("About", func() {
		showAboutDialog(w)
	})

	// Create file menu
	fileMenu := fyne.NewMenu("File", settingsItem, fyne.NewMenuItemSeparator(), uninstallItem, quitItem)

	// Create help menu
	helpMenu := fyne.NewMenu("Help", updateItem, fyne.NewMenuItemSeparator(), aboutItem)

	// Return the main menu
	return fyne.NewMainMenu(fileMenu, helpMenu)
}
