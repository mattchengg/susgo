# Android Build Setup Guide for sfgo

## Overview
This document describes the Android APK build setup for sfgo using Fyne.

## Issues Resolved

### 1. Icon File Issue ✅
**Problem**: Icon.png was actually a JPEG file, causing the error: "Failed to decode app icon - png: invalid format: not a PNG file"

**Solution**: 
- Renamed the file to Icon.jpg
- Converted it to proper PNG format using ImageMagick
- Verified the conversion: `file Icon.png` shows "PNG image data, 1397 x 1397"

### 2. FyneApp.toml Configuration ✅
**Problem**: No FyneApp.toml configuration file existed

**Solution**: Created FyneApp.toml with proper configuration:
- App ID: com.samsung.firmware.sfgo
- App Name: Samsung Firmware Downloader
- Icon: Icon.png
- Android SDK settings: MinSDK 23, TargetSDK 34
- Android NDK Version: 25

## Android NDK Setup Status

### Current Limitation
Android NDK is **not available** as a direct package in Termux. The `android-ndk` package does not exist in Termux repositories.

### Available Android Tools in Termux
The following Android development tools are available and installed:
- aapt (Android Asset Packaging Tool)
- aapt2
- android-tools (includes adb)

### Building APK from Termux
Building Android APKs directly from Termux is challenging due to:
1. No pre-built Android NDK package for Termux
2. Android NDK requires significant disk space (>5GB)
3. Cross-compilation complexity

### Recommended Approaches

#### Option 1: Use a Desktop/Laptop Computer (Recommended)
Build the Android APK on a regular computer with:
1. Install Fyne command-line tools: `go install fyne.io/fyne/v2/cmd/fyne@latest`
2. Install Android SDK and NDK
3. Set environment variables:
   ```bash
   export ANDROID_HOME=/path/to/android/sdk
   export ANDROID_NDK_HOME=$ANDROID_HOME/ndk/25.2.9519653
   ```
4. Build APK:
   ```bash
   fyne package -os android -appID com.samsung.firmware.sfgo -icon Icon.png
   ```

#### Option 2: Use GitHub Actions / CI/CD
Set up automated builds using GitHub Actions with Fyne's Android build support.

#### Option 3: Use proot-distro in Termux
Install a full Linux distribution in Termux using proot-distro and set up Android SDK/NDK there:
```bash
pkg install proot-distro
proot-distro install ubuntu
proot-distro login ubuntu
# Then install Android SDK/NDK in the Ubuntu environment
```

#### Option 4: Manual NDK Installation (Advanced)
Download Android NDK manually from Google's website and set it up:
```bash
# Download NDK (example URL - check for latest version)
# wget https://dl.google.com/android/repository/android-ndk-r25c-linux.zip
# unzip android-ndk-r25c-linux.zip -d ~/android-ndk
# export ANDROID_NDK_HOME=~/android-ndk/android-ndk-r25c
```

## Files Created/Modified

### New Files
1. `FyneApp.toml` - Fyne application configuration
2. `ANDROID_BUILD_SETUP.md` - This documentation file

### Modified Files
1. `Icon.png` - Converted from JPEG to proper PNG format
2. `Icon.jpg` - Original JPEG file (renamed from Icon.png)

## Testing
To verify the configuration without building:
```bash
# Check if icon is valid
file Icon.png
# Should output: PNG image data, 1397 x 1397

# View FyneApp.toml
cat FyneApp.toml
```

## Next Steps
1. Choose one of the recommended build approaches above
2. Set up the Android NDK environment
3. Build the APK using: `fyne package -os android`
4. Test the APK on an Android device

## Notes
- The current setup on Termux can build the Linux binary successfully
- For Android APK, a proper Android NDK environment is required
- The FyneApp.toml configuration is ready and properly configured
- The Icon.png file is now valid and ready for packaging
