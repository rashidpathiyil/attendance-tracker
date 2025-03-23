# Makefile for Attendance Tracker

APPNAME=attendance-tracker
VERSION=1.0.0
BUILDDATE=$(shell date -u +"%Y-%m-%d")
COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
LDFLAGS=-ldflags "-X main.Version=$(VERSION) -X main.BuildDate=$(BUILDDATE) -X main.CommitSHA=$(COMMIT)"

# Platforms
PLATFORMS=windows linux darwin

.PHONY: all clean build build-all windows darwin linux release install-autostart-windows install-autostart-macos install-autostart-linux

all: build

build:
	go build $(LDFLAGS) -o $(APPNAME)

# Build for current platform with embedded resources
bundle:
	fyne bundle -o bundled.go icon.png
	go build $(LDFLAGS) -o $(APPNAME)

# Cross-platform builds
build-all: windows linux darwin

windows:
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(APPNAME)_$(VERSION)_windows_amd64.exe
	GOOS=windows GOARCH=386 go build $(LDFLAGS) -o $(APPNAME)_$(VERSION)_windows_386.exe

linux:
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(APPNAME)_$(VERSION)_linux_amd64
	GOOS=linux GOARCH=386 go build $(LDFLAGS) -o $(APPNAME)_$(VERSION)_linux_386
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o $(APPNAME)_$(VERSION)_linux_arm64

darwin:
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(APPNAME)_$(VERSION)_darwin_amd64
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(APPNAME)_$(VERSION)_darwin_arm64

# Packaging with Fyne
package-all: package-windows package-darwin package-linux

package-windows:
	fyne package -os windows -icon icon.png -name $(APPNAME) -appID com.example.$(APPNAME) -release

package-darwin:
	fyne package -os darwin -icon icon.png -name $(APPNAME) -appID com.example.$(APPNAME) -release

package-linux:
	fyne package -os linux -icon icon.png -name $(APPNAME) -appID com.example.$(APPNAME) -release

# Install auto-start configurations
install-autostart-windows: build
	@echo "Installing Windows autostart entry..."
	@powershell -Command "$$WshShell = New-Object -ComObject WScript.Shell; \
		$$Shortcut = $$WshShell.CreateShortcut(\"$$env:APPDATA\\Microsoft\\Windows\\Start Menu\\Programs\\Startup\\$(APPNAME).lnk\"); \
		$$Shortcut.TargetPath = \"$$(pwd)\\$(APPNAME).exe\"; \
		$$Shortcut.Arguments = \"--minimized\"; \
		$$Shortcut.Description = \"Attendance Tracker\"; \
		$$Shortcut.Save()"
	@echo "Windows autostart shortcut created in Startup folder"

install-autostart-macos: build
	@echo "Installing macOS LaunchAgent..."
	@mkdir -p ~/Library/LaunchAgents
	@echo "<?xml version=\"1.0\" encoding=\"UTF-8\"?>" > ~/Library/LaunchAgents/com.$(APPNAME).plist
	@echo "<!DOCTYPE plist PUBLIC \"-//Apple//DTD PLIST 1.0//EN\" \"http://www.apple.com/DTDs/PropertyList-1.0.dtd\">" >> ~/Library/LaunchAgents/com.$(APPNAME).plist
	@echo "<plist version=\"1.0\">" >> ~/Library/LaunchAgents/com.$(APPNAME).plist
	@echo "<dict>" >> ~/Library/LaunchAgents/com.$(APPNAME).plist
	@echo "    <key>Label</key>" >> ~/Library/LaunchAgents/com.$(APPNAME).plist
	@echo "    <string>com.$(APPNAME)</string>" >> ~/Library/LaunchAgents/com.$(APPNAME).plist
	@echo "    <key>ProgramArguments</key>" >> ~/Library/LaunchAgents/com.$(APPNAME).plist
	@echo "    <array>" >> ~/Library/LaunchAgents/com.$(APPNAME).plist
	@echo "        <string>$$(pwd)/$(APPNAME)</string>" >> ~/Library/LaunchAgents/com.$(APPNAME).plist
	@echo "        <string>--minimized</string>" >> ~/Library/LaunchAgents/com.$(APPNAME).plist
	@echo "    </array>" >> ~/Library/LaunchAgents/com.$(APPNAME).plist
	@echo "    <key>RunAtLoad</key>" >> ~/Library/LaunchAgents/com.$(APPNAME).plist
	@echo "    <true/>" >> ~/Library/LaunchAgents/com.$(APPNAME).plist
	@echo "</dict>" >> ~/Library/LaunchAgents/com.$(APPNAME).plist
	@echo "</plist>" >> ~/Library/LaunchAgents/com.$(APPNAME).plist
	@launchctl unload ~/Library/LaunchAgents/com.$(APPNAME).plist 2>/dev/null || true
	@launchctl load ~/Library/LaunchAgents/com.$(APPNAME).plist
	@echo "macOS LaunchAgent installed and loaded"

install-autostart-linux: build
	@echo "Installing Linux autostart entry..."
	@mkdir -p ~/.config/autostart
	@echo "[Desktop Entry]" > ~/.config/autostart/$(APPNAME).desktop
	@echo "Type=Application" >> ~/.config/autostart/$(APPNAME).desktop
	@echo "Name=Attendance Tracker" >> ~/.config/autostart/$(APPNAME).desktop
	@echo "Exec=$$(pwd)/$(APPNAME) --minimized" >> ~/.config/autostart/$(APPNAME).desktop
	@echo "Terminal=false" >> ~/.config/autostart/$(APPNAME).desktop
	@echo "Comment=Attendance Tracking Application" >> ~/.config/autostart/$(APPNAME).desktop
	@echo "Categories=Utility;" >> ~/.config/autostart/$(APPNAME).desktop
	@chmod +x ~/.config/autostart/$(APPNAME).desktop
	@echo "Linux autostart entry created in ~/.config/autostart"

# Detect OS and install appropriate autostart
install-autostart:
	@case "$$(uname -s)" in \
		Darwin) make install-autostart-macos ;; \
		Linux) make install-autostart-linux ;; \
		CYGWIN*|MINGW*|MSYS*) make install-autostart-windows ;; \
		*) echo "Unsupported operating system" ;; \
	esac

# Clean build artifacts
clean:
	rm -f $(APPNAME)
	rm -f $(APPNAME)_*
	rm -f *.exe
	rm -f bundled.go

# Clean autostart entries
clean-autostart:
	@case "$$(uname -s)" in \
		Darwin) \
			launchctl unload ~/Library/LaunchAgents/com.$(APPNAME).plist 2>/dev/null || true; \
			rm -f ~/Library/LaunchAgents/com.$(APPNAME).plist; \
			echo "Removed macOS LaunchAgent" ;; \
		Linux) \
			rm -f ~/.config/autostart/$(APPNAME).desktop; \
			echo "Removed Linux autostart entry" ;; \
		CYGWIN*|MINGW*|MSYS*) \
			rm -f "$$APPDATA/Microsoft/Windows/Start Menu/Programs/Startup/$(APPNAME).lnk"; \
			echo "Removed Windows autostart shortcut" ;; \
		*) echo "Unsupported operating system" ;; \
	esac

# Release target to create distribution packages
release: build-all
	mkdir -p release
	zip -j release/$(APPNAME)_$(VERSION)_windows_amd64.zip $(APPNAME)_$(VERSION)_windows_amd64.exe
	zip -j release/$(APPNAME)_$(VERSION)_windows_386.zip $(APPNAME)_$(VERSION)_windows_386.exe
	tar -czf release/$(APPNAME)_$(VERSION)_linux_amd64.tar.gz $(APPNAME)_$(VERSION)_linux_amd64
	tar -czf release/$(APPNAME)_$(VERSION)_linux_386.tar.gz $(APPNAME)_$(VERSION)_linux_386
	tar -czf release/$(APPNAME)_$(VERSION)_linux_arm64.tar.gz $(APPNAME)_$(VERSION)_linux_arm64
	tar -czf release/$(APPNAME)_$(VERSION)_darwin_amd64.tar.gz $(APPNAME)_$(VERSION)_darwin_amd64
	tar -czf release/$(APPNAME)_$(VERSION)_darwin_arm64.tar.gz $(APPNAME)_$(VERSION)_darwin_arm64 
