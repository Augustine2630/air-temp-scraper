#!/bin/bash
#
# One-line installer for temp-scraper
# Usage: curl -fsSL https://raw.githubusercontent.com/Augustine2630/air-temp-scraper/main/install.sh | bash
# or: bash install.sh [PORT] [INTERVAL]

set -e

PORT="${1:-9100}"
INTERVAL="${2:-30s}"
SERVICE_NAME="temp-scraper"
INSTALL_DIR="${HOME}/.local/bin"
LOG_DIR="/tmp"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m'

log() { echo -e "${BLUE}→${NC} $*"; }
success() { echo -e "${GREEN}✓${NC} $*"; }
error() { echo -e "${RED}✗${NC} $*" >&2; exit 1; }

# ── Detect OS ────────────────────────────────────────────────────────────────

if [[ "$OSTYPE" != "darwin"* ]]; then
    error "This script only works on macOS. This is an Apple SMC tool."
fi

# Detect architecture
if [[ $(uname -m) == "arm64" ]]; then
    ARCH="arm64"
elif [[ $(uname -m) == "x86_64" ]]; then
    ARCH="amd64"
else
    error "Unsupported architecture: $(uname -m)"
fi

BINARY_NAME="temp-scraper_darwin_${ARCH}"
log "Detected macOS/$ARCH, will download $BINARY_NAME"

# ── Download latest release ────────────────────────────────────────────────────

log "Fetching latest release from GitHub..."
RELEASE_URL=$(curl -fsSL \
    https://api.github.com/repos/Augustine2630/air-temp-scraper/releases/latest \
    | grep -o "\"browser_download_url\": \"[^\"]*${BINARY_NAME}[^\"]*\"" \
    | head -1 \
    | cut -d'"' -f4)

if [ -z "$RELEASE_URL" ]; then
    error "Could not find $BINARY_NAME in latest release. Check GitHub releases."
fi

success "Found release: $RELEASE_URL"

# ── Download binary ───────────────────────────────────────────────────────────

log "Downloading binary..."
mkdir -p "$INSTALL_DIR"
TEMP_BIN=$(mktemp)
trap "rm -f $TEMP_BIN" EXIT

if ! curl -fsSL -o "$TEMP_BIN" "$RELEASE_URL"; then
    error "Failed to download binary"
fi

chmod +x "$TEMP_BIN"
success "Downloaded successfully"

# ── Verify binary ────────────────────────────────────────────────────────────

log "Verifying binary..."
if ! file "$TEMP_BIN" | grep -q "Mach-O"; then
    error "Downloaded file is not a valid macOS binary"
fi

if ! otool -L "$TEMP_BIN" 2>/dev/null | grep -q "IOKit"; then
    error "Binary is not linked against IOKit framework"
fi

success "Binary verified (Mach-O, IOKit linked)"

# ── Install binary ──────────────────────────────────────────────────────────

BINARY_PATH="$INSTALL_DIR/$SERVICE_NAME"
mv "$TEMP_BIN" "$BINARY_PATH"
success "Installed to $BINARY_PATH"

# ── Create launchd plist ────────────────────────────────────────────────────

PLIST_DIR="$HOME/Library/LaunchAgents"
PLIST_FILE="$PLIST_DIR/com.local.$SERVICE_NAME.plist"
mkdir -p "$PLIST_DIR"

log "Creating launchd service..."
cat > "$PLIST_FILE" <<EOFPLIST
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.local.$SERVICE_NAME</string>

    <key>ProgramArguments</key>
    <array>
        <string>$BINARY_PATH</string>
    </array>

    <key>WorkingDirectory</key>
    <string>$INSTALL_DIR</string>

    <key>EnvironmentVariables</key>
    <dict>
        <key>TEMP_SCRAPER_PORT</key>
        <string>$PORT</string>
        <key>TEMP_SCRAPER_INTERVAL</key>
        <string>$INTERVAL</string>
    </dict>

    <key>StandardOutPath</key>
    <string>$LOG_DIR/${SERVICE_NAME}.log</string>

    <key>StandardErrorPath</key>
    <string>$LOG_DIR/${SERVICE_NAME}-error.log</string>

    <key>RunAtLoad</key>
    <true/>

    <key>KeepAlive</key>
    <dict>
        <key>Crashed</key>
        <true/>
    </dict>
</dict>
</plist>
EOFPLIST

chmod 644 "$PLIST_FILE"
success "Created launchd plist: $PLIST_FILE"

# ── Load service ────────────────────────────────────────────────────────────

log "Loading launchd service..."
if launchctl list | grep -q "com.local.$SERVICE_NAME"; then
    log "Service already loaded, restarting..."
    launchctl unload "$PLIST_FILE" 2>/dev/null || true
    sleep 1
fi

launchctl load "$PLIST_FILE"
sleep 2

# ── Verify service is running ────────────────────────────────────────────────

log "Verifying service..."
if launchctl list | grep -q "com.local.$SERVICE_NAME"; then
    success "Service is loaded and running"
else
    error "Service failed to start. Check: launchctl list | grep $SERVICE_NAME"
fi

# ── Test metrics endpoint ────────────────────────────────────────────────────

log "Testing metrics endpoint (http://localhost:$PORT/metrics)..."
sleep 2

RESPONSE=$(curl -s -w "\n%{http_code}" "http://localhost:$PORT/metrics" 2>/dev/null || echo "000")
HTTP_CODE=$(echo "$RESPONSE" | tail -1)
BODY=$(echo "$RESPONSE" | head -n -1)

if [ "$HTTP_CODE" = "200" ]; then
    success "Metrics endpoint is responding"
    if echo "$BODY" | grep -q "temp_scraper_sensor_temperature_celsius"; then
        success "Temperature metrics are being published"
        echo ""
        echo "Sample metrics:"
        echo "$BODY" | grep "temp_scraper_sensor_temperature_celsius" | head -3
    else
        error "No temperature metrics found in response"
    fi
else
    error "Metrics endpoint returned HTTP $HTTP_CODE (expected 200)"
fi

# ── Summary ──────────────────────────────────────────────────────────────────

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
success "Installation complete!"
echo ""
echo "Service Details:"
echo "  Binary:     $BINARY_PATH"
echo "  Config:     $PLIST_FILE"
echo "  Port:       $PORT"
echo "  Interval:   $INTERVAL"
echo "  Metrics:    http://localhost:$PORT/metrics"
echo ""
echo "Logs:"
echo "  Output: tail -f $LOG_DIR/${SERVICE_NAME}.log"
echo "  Errors: tail -f $LOG_DIR/${SERVICE_NAME}-error.log"
echo ""
echo "Service Management:"
echo "  Status:  launchctl list | grep $SERVICE_NAME"
echo "  Stop:    launchctl unload $PLIST_FILE"
echo "  Start:   launchctl load $PLIST_FILE"
echo "  Restart: launchctl unload $PLIST_FILE && sleep 1 && launchctl load $PLIST_FILE"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
