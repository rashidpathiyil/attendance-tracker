# Publishing Updates to GitHub

This guide explains how to publish new versions of Attendance Tracker that will be automatically detected by the auto-update system.

## Prerequisites

- You have push access to the [attendance-tracker repository](https://github.com/rashidpathiyil/attendance-tracker)
- You have Git installed locally
- You have Go and required build tools installed

## Publishing a New Release

### 1. Update Version Number

1. Edit the `build-windows.sh` file and update the `VERSION` variable:

```bash
# Configuration
APP_NAME="Attendance Tracker"
VERSION="1.0.1"  # Update this for each release
```

2. Commit the version change:

```bash
git add build-windows.sh
git commit -m "Bump version to 1.0.1"
```

### 2. Create a Git Tag

Create and push a tag matching your version number:

```bash
git tag v1.0.1
git push origin v1.0.1
```

**Important**: The tag must start with `v` followed by the version number (e.g., `v1.0.1`).

### 3. Wait for Automated Build

The GitHub Actions workflow will automatically:
1. Build the application with the new version number
2. Create a release with the executable
3. Generate SHA-256 checksums for verification
4. Publish the release on GitHub

You can monitor the progress in the [Actions tab](https://github.com/rashidpathiyil/attendance-tracker/actions).

### 4. Verify the Release

1. Check that the release appears on the [Releases page](https://github.com/rashidpathiyil/attendance-tracker/releases)
2. Verify that the release includes:
   - The Windows executable
   - SHA-256 checksum file
   - Release notes (auto-generated from commit messages)

## Manual Release (Alternative)

If you prefer to manually create a release:

1. Build the executable locally:
   ```bash
   ./build-windows.sh
   ```

2. Go to [GitHub Releases](https://github.com/rashidpathiyil/attendance-tracker/releases) and click "Draft a new release"

3. Enter the tag version (e.g., `v1.0.1`)

4. Add a title and description
   - Use bullet points for release notes (•, -, or *)
   - This format will be properly parsed by the auto-update system

5. Upload the executable from `./releases/windows/`

6. Publish the release

## Release Notes Best Practices

Format your release notes with bullet points to ensure they're properly displayed in the update notification:

```
## What's New

- Fixed idle time detection on Windows
- Improved startup performance
- Added better error handling
```

## Versioning Guidelines

Follow [Semantic Versioning](https://semver.org/) for version numbers:

- **Major** (x.0.0): Incompatible API changes
- **Minor** (0.x.0): New functionality in a backward-compatible manner
- **Patch** (0.0.x): Backward-compatible bug fixes

## Testing a Release

After publishing, test the update feature:

1. Install an older version of the application
2. Launch the application
3. Verify that the update notification appears
4. Test the download and installation process

You can also use the `check-update.ps1` script for local testing before publishing. 
