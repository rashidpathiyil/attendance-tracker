package main

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
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
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
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
		// Use the direct Windows API implementation
		return getSystemIdleTimeWindows()

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

func main() {
	// Parse command line flags
	startMinimized := flag.Bool("minimized", false, "Start the application minimized to system tray")
	flag.Parse()

	// Create the Fyne app with system tray support
	a := app.NewWithID("com.attendance.tracker")
	// Try the bundled icon
	a.SetIcon(resourceIconPng)
	w := a.NewWindow("Attendance Tracker")
	w.Resize(fyne.NewSize(400, 300))

	// Set up close intercept
	w.SetCloseIntercept(func() {
		// Minimize to system tray instead of closing
		w.Hide()
	})

	// Start minimized if flag is set
	if *startMinimized {
		// Will be shown in systray only at first
		w.Hide()
	}

	// Create main app instance
	var attendanceApp struct {
		mainWindow      fyne.Window
		status          bool
		developerMode   bool
		autoMode        bool
		showActivityLog bool
		showIdleTime    bool
		lastActivity    time.Time
		serverEndpoint  string
		deviceID        string
		userID          string
		idleTimeout     time.Duration
		checkInterval   time.Duration
		runAtStartup    bool
	}

	attendanceApp.mainWindow = w
	attendanceApp.status = false
	attendanceApp.developerMode = false
	attendanceApp.autoMode = true         // Default to auto mode enabled
	attendanceApp.showActivityLog = false // Default to hiding activity log
	attendanceApp.showIdleTime = true     // Default to showing idle time
	attendanceApp.lastActivity = time.Now()
	attendanceApp.serverEndpoint = defaultServerEndpoint
	attendanceApp.deviceID = defaultDeviceID
	attendanceApp.userID = defaultUserID
	attendanceApp.idleTimeout = defaultIdleTimeout
	attendanceApp.checkInterval = defaultCheckInterval
	attendanceApp.runAtStartup = true // Default to run at startup

	// Create app configuration
	config := NewAppConfig()

	// Set up autostart on first run if enabled in config
	if config.RunAtStartup {
		if err := setupAutostart(true); err != nil {
			fmt.Printf("Failed to set up autostart: %v\n", err)
		}
	}

	// Create activity monitor
	activityMonitor := NewSystemActivityMonitor()

	// Status message display
	statusLabel := widget.NewLabel("Current Status: Checked Out")
	statusLabel.Alignment = fyne.TextAlignCenter

	// Create toggle button for check in/check out
	toggleButton := widget.NewButton("Check In", nil)

	// Create a globally accessible idleTimeLabel
	idleTimeLabel := widget.NewLabel("Idle Time: 0s")

	// Create activity log text area
	activityLog := widget.NewMultiLineEntry()
	activityLog.Disable()
	activityLog.SetMinRowsVisible(5)

	// Function to log activity
	logActivity := func(message string) {
		timestamp := time.Now().Format("15:04:05")
		activityLog.Text = fmt.Sprintf("[%s] %s\n%s", timestamp, message, activityLog.Text)
		activityLog.Refresh()
	}

	// Function to send status to server
	sendStatus := func(status string) {
		// Get current time
		currentTime := time.Now()

		// Format time and date strings
		timeStr := currentTime.Format("15:04:05")
		dateStr := currentTime.Format("2006-01-02")

		// Convert status to event_type format
		eventType := "check-in"
		if status == "checked_out" {
			eventType = "check-out"
		}

		// Create the nested payload content
		payloadContent := PayloadContent{
			Time:     timeStr,
			Date:     dateStr,
			DeviceID: config.DeviceID,
		}

		// Add config details if in developer mode
		if config.DeveloperMode {
			configMap := map[string]interface{}{
				"idle_timeout_mins":   int(config.IdleTimeout.Minutes()),
				"check_interval_secs": int(config.CheckInterval.Seconds()),
				"developer_mode":      true,
				"platform":            runtime.GOOS,
				"auto_mode":           config.AutoMode,
				"show_activity_log":   config.ShowActivityLog,
				"show_idle_time":      config.ShowIdleTime,
			}
			payloadContent.Config = configMap
		}

		// Create the base payload, timestamp is omitted as it's optional
		payload := StatusPayload{
			EventType: eventType,
			UserID:    config.UserID,
			Payload:   payloadContent,
		}

		jsonData, err := json.Marshal(payload)
		if err != nil {
			logActivity(fmt.Sprintf("Error creating JSON: %v", err))
			return
		}

		// Send request in a goroutine to avoid blocking UI
		go func() {
			resp, err := http.Post(config.ServerEndpoint, "application/json", bytes.NewBuffer(jsonData))
			if err != nil {
				logActivity(fmt.Sprintf("Error sending status: %v", err))
				return
			}
			defer resp.Body.Close()

			logActivity(fmt.Sprintf("Status '%s' sent to server (Response: %s)", status, resp.Status))
		}()
	}

	// Function to toggle check in/out status
	toggleStatus := func() {
		attendanceApp.status = !attendanceApp.status

		if attendanceApp.status {
			toggleButton.SetText("Check Out")
			statusLabel.SetText("Current Status: Checked In")
			logActivity("Manually checked in")
			sendStatus("checked_in")
		} else {
			toggleButton.SetText("Check In")
			statusLabel.SetText("Current Status: Checked Out")
			logActivity("Manually checked out")
			sendStatus("checked_out")
		}

		// Update activity timestamp
		activityMonitor.UpdateLastActivity()
	}

	// Create a small settings button (define before using showSettings)
	configButton := widget.NewButtonWithIcon("", theme.SettingsIcon(), nil)
	configButton.Importance = widget.LowImportance

	// Create a function to build the UI
	refreshUI := func() {
		// Create containers for different parts of the UI
		var contentItems []fyne.CanvasObject

		// Build status display section
		statusSection := container.NewVBox(
			container.NewPadded(statusLabel),
			container.NewHBox(
				layout.NewSpacer(),
				toggleButton,
				layout.NewSpacer(),
			),
		)

		contentItems = append(contentItems, statusSection)

		// Add idle time display if enabled
		if config.ShowIdleTime {
			contentItems = append(contentItems, container.NewPadded(idleTimeLabel))
		}

		// Add activity log if enabled
		if config.ShowActivityLog {
			contentItems = append(contentItems, container.NewVBox(
				widget.NewLabel("Activity Log:"),
				activityLog,
			))
		}

		// Create main content with settings button in corner
		mainContent := container.NewBorder(
			nil, nil, nil,
			container.NewVBox(
				configButton,
				layout.NewSpacer(),
			),
			container.NewVBox(contentItems...),
		)

		w.SetContent(mainContent)
	}

	// Function to show settings dialog
	showSettings := func() {
		// Create form elements for basic settings
		serverEntry := widget.NewEntry()
		serverEntry.SetText(config.ServerEndpoint)

		deviceIDEntry := widget.NewEntry()
		deviceIDEntry.SetText(config.DeviceID)

		userIDEntry := widget.NewEntry()
		userIDEntry.SetText(config.UserID)

		// Auto mode toggle in settings
		autoModeCheck := widget.NewCheck("Enable Auto Mode", nil)
		autoModeCheck.SetChecked(config.AutoMode)

		// Activity log visibility toggle
		showLogCheck := widget.NewCheck("Show Activity Log", nil)
		showLogCheck.SetChecked(config.ShowActivityLog)

		// Idle time visibility toggle
		showIdleTimeCheck := widget.NewCheck("Show Idle Time", nil)
		showIdleTimeCheck.SetChecked(config.ShowIdleTime)

		// Developer mode toggle
		devModeCheck := widget.NewCheck("Developer Mode", nil)
		devModeCheck.SetChecked(config.DeveloperMode)

		// Add run at startup toggle (only visible in developer mode)
		runAtStartupCheck := widget.NewCheck("Run at Startup", nil)
		runAtStartupCheck.SetChecked(config.RunAtStartup)
		runAtStartupCheck.Disable() // Initially disabled, will be enabled in dev mode

		// Create advanced settings that are only enabled in developer mode
		idleTimeoutEntry := widget.NewEntry()
		idleTimeoutEntry.SetText(fmt.Sprintf("%d", int(config.IdleTimeout.Minutes())))
		idleTimeoutEntry.Disable()

		checkIntervalEntry := widget.NewEntry()
		checkIntervalEntry.SetText(fmt.Sprintf("%d", int(config.CheckInterval.Seconds())))
		checkIntervalEntry.Disable()

		// Update advanced settings availability based on dev mode
		devModeCheck.OnChanged = func(checked bool) {
			if checked {
				idleTimeoutEntry.Enable()
				checkIntervalEntry.Enable()
				runAtStartupCheck.Enable() // Enable startup toggle in dev mode
			} else {
				idleTimeoutEntry.Disable()
				checkIntervalEntry.Disable()
				runAtStartupCheck.Disable() // Disable startup toggle when not in dev mode
			}
		}

		// Create form items
		items := []*widget.FormItem{
			widget.NewFormItem("Server Endpoint", serverEntry),
			widget.NewFormItem("Device ID", deviceIDEntry),
			widget.NewFormItem("User ID", userIDEntry),
			widget.NewFormItem("", autoModeCheck),
			widget.NewFormItem("", showLogCheck),
			widget.NewFormItem("", showIdleTimeCheck),
			widget.NewFormItem("", devModeCheck),
			widget.NewFormItem("Idle Timeout (minutes)", idleTimeoutEntry),
			widget.NewFormItem("Check Interval (seconds)", checkIntervalEntry),
			widget.NewFormItem("", runAtStartupCheck),
		}

		// Create and show the dialog
		dialog.ShowForm("Application Settings", "Save", "Cancel", items,
			func(confirmed bool) {
				if confirmed {
					// Previous autostart setting to check for changes
					prevRunAtStartup := config.RunAtStartup

					// Save settings
					config.ServerEndpoint = serverEntry.Text
					config.DeviceID = deviceIDEntry.Text
					config.UserID = userIDEntry.Text
					config.AutoMode = autoModeCheck.Checked
					config.ShowActivityLog = showLogCheck.Checked
					config.ShowIdleTime = showIdleTimeCheck.Checked
					config.DeveloperMode = devModeCheck.Checked
					config.RunAtStartup = runAtStartupCheck.Checked

					// Parse developer settings if in dev mode
					if config.DeveloperMode {
						if mins, err := strconv.Atoi(idleTimeoutEntry.Text); err == nil && mins > 0 {
							config.IdleTimeout = time.Duration(mins) * time.Minute
						}

						if secs, err := strconv.Atoi(checkIntervalEntry.Text); err == nil && secs > 0 {
							config.CheckInterval = time.Duration(secs) * time.Second
						}

						// Apply autostart changes if the setting changed
						if prevRunAtStartup != config.RunAtStartup {
							if err := setupAutostart(config.RunAtStartup); err != nil {
								logActivity(fmt.Sprintf("Error setting up autostart: %v", err))
							} else {
								logActivity(fmt.Sprintf("Autostart %s", map[bool]string{true: "enabled", false: "disabled"}[config.RunAtStartup]))
							}
						}
					}

					logActivity(fmt.Sprintf("Settings updated - Auto Mode: %v, Activity Log: %v, Idle Time: %v",
						config.AutoMode, config.ShowActivityLog, config.ShowIdleTime))

					if config.DeveloperMode {
						logActivity(fmt.Sprintf("Developer mode enabled - Idle Timeout: %v, Check Interval: %v",
							config.IdleTimeout, config.CheckInterval))
					}

					// Refresh UI based on new settings
					refreshUI()
				}
			}, w)
	}

	// After defining refreshUI and showSettings, update the button action
	configButton.OnTapped = func() {
		showSettings()
	}

	// Create menu with version info
	mainMenu := fyne.NewMainMenu(
		fyne.NewMenu("File",
			fyne.NewMenuItem("Settings", func() {
				showSettings()
			}),
			fyne.NewMenuItem("About", func() {
				showAboutDialog(w)
			}),
			fyne.NewMenuItem("Quit", func() {
				a.Quit()
			}),
		),
	)

	w.SetMainMenu(mainMenu)

	// Status tracking
	isCheckedIn := false

	// Set toggle button action
	toggleButton.OnTapped = toggleStatus

	// Start the activity monitor in a goroutine
	go func() {
		for {
			time.Sleep(config.CheckInterval)

			// Check for system activity
			hasActivity := activityMonitor.Check()

			if hasActivity && config.AutoMode && !isCheckedIn {
				// Auto check in if activity detected and in auto mode
				isCheckedIn = true
				toggleButton.SetText("Check Out")
				statusLabel.SetText("Current Status: Checked In")
				logActivity("Auto checked in due to system activity")
				sendStatus("checked_in")
			}

			// Get and display current idle time
			idleTime := activityMonitor.IdleTime()
			if idleTime < time.Minute {
				idleTimeLabel.SetText(fmt.Sprintf("Idle Time: %ds", int(idleTime.Seconds())))
			} else {
				idleTimeLabel.SetText(fmt.Sprintf("Idle Time: %dm %ds",
					int(idleTime.Minutes()), int(idleTime.Seconds())%60))
			}

			// Handle auto checkout if needed
			if config.AutoMode && idleTime >= config.IdleTimeout && isCheckedIn {
				isCheckedIn = false
				toggleButton.SetText("Check In")
				statusLabel.SetText("Current Status: Checked Out")
				logActivity("Auto checked out due to inactivity")
				sendStatus("checked_out")
			}
		}
	}()

	// Initial setup
	logActivity("Application started")
	logActivity(fmt.Sprintf("Platform: %s", runtime.GOOS))
	refreshUI()

	w.ShowAndRun()
}

// showAboutDialog displays version information
func showAboutDialog(w fyne.Window) {
	content := widget.NewLabel(fmt.Sprintf(
		"Attendance Tracker\nVersion: %s\nBuild Date: %s\nCommit: %s\nPlatform: %s",
		Version, BuildDate, CommitSHA, runtime.GOOS+"/"+runtime.GOARCH,
	))

	dialog.ShowCustom("About", "Close", content, w)
}
