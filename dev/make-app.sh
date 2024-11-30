#!/bin/bash

APP_NAME="Secrets"
BINARY_PATH="./secrets" 

mkdir -p "build/${APP_NAME}.app/Contents/"{MacOS,Resources}

cp "$BINARY_PATH" "build/${APP_NAME}.app/Contents/MacOS/${APP_NAME}"
chmod +x "build/${APP_NAME}.app/Contents/MacOS/${APP_NAME}"

cat > "build/${APP_NAME}.app/Contents/Info.plist" << EOF
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>CFBundleExecutable</key>
    <string>${APP_NAME}</string>
    <key>CFBundleIdentifier</key>
    <string>com.theleaplog.${APP_NAME}</string>
    <key>CFBundleName</key>
    <string>${APP_NAME}</string>
    <key>CFBundlePackageType</key>
    <string>APPL</string>
    <key>CFBundleShortVersionString</key>
    <string>1.0</string>
    <key>LSMinimumSystemVersion</key>
    <string>10.10</string>
    <key>CFBundleIconFile</key>
    <string>AppIcon</string>
</dict>
</plist>
EOF
