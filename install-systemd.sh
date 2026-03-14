#!/bin/bash
#
# Install temp-scraper as a systemd service on Linux/macOS with launchd
# Usage: sudo ./install-systemd.sh [PORT] [INTERVAL]
#   PORT defaults to 9100
#   INTERVAL defaults to 30s

set -e

PORT="${1:-9100}"
INTERVAL="${2:-30s}"
SERVICE_NAME="temp-scraper"
BINARY_PATH="$(cd "$(dirname "$0")" && pwd)/bin/temp-scraper"
USER="${SUDO_USER:-$(whoami)}"

if [ ! -f "$BINARY_PATH" ]; then
    echo "Error: Binary not found at $BINARY_PATH"
    echo "Run 'make build' first"
    exit 1
fi

# Detect OS
if [[ "$OSTYPE" == "darwin"* ]]; then
    install_launchd
else
    install_systemd
fi

install_systemd() {
    echo "Installing systemd service..."
    SERVICE_FILE="/etc/systemd/system/${SERVICE_NAME}.service"

    if [ ! -w /etc/systemd/system ]; then
        echo "Error: Need root/sudo to write to /etc/systemd/system"
        exit 1
    fi

    cat > "$SERVICE_FILE" <<EOF
[Unit]
Description=Temperature Sensor Scraper (SMC) for Prometheus
Documentation=file:///opt/temp-scraper/README.md
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
User=$USER
WorkingDirectory=$(dirname "$BINARY_PATH")
ExecStart=$BINARY_PATH
Restart=on-failure
RestartSec=10
StandardOutput=journal
StandardError=journal
SyslogIdentifier=$SERVICE_NAME

Environment="TEMP_SCRAPER_PORT=$PORT"
Environment="TEMP_SCRAPER_INTERVAL=$INTERVAL"

# Security hardening
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/var/lib/prometheus-node-exporter
ReadOnlyPaths=$BINARY_PATH

[Install]
WantedBy=multi-user.target
EOF

    chmod 644 "$SERVICE_FILE"
    systemctl daemon-reload
    echo "Service installed: $SERVICE_FILE"
    echo ""
    echo "Enable and start with:"
    echo "  sudo systemctl enable $SERVICE_NAME"
    echo "  sudo systemctl start $SERVICE_NAME"
    echo ""
    echo "Check status:"
    echo "  sudo systemctl status $SERVICE_NAME"
    echo "  sudo journalctl -u $SERVICE_NAME -f"
}

install_launchd() {
    echo "Installing launchd service..."
    PLIST_FILE="$HOME/Library/LaunchAgents/com.local.$SERVICE_NAME.plist"

    mkdir -p "$(dirname "$PLIST_FILE")"

    cat > "$PLIST_FILE" <<'EOFPLIST'
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.local.temp-scraper</string>

    <key>ProgramArguments</key>
    <array>
        <string>BINARY_PLACEHOLDER</string>
    </array>

    <key>WorkingDirectory</key>
    <string>WORKDIR_PLACEHOLDER</string>

    <key>EnvironmentVariables</key>
    <dict>
        <key>TEMP_SCRAPER_PORT</key>
        <string>PORT_PLACEHOLDER</string>
        <key>TEMP_SCRAPER_INTERVAL</key>
        <string>INTERVAL_PLACEHOLDER</string>
    </dict>

    <key>StandardOutPath</key>
    <string>/tmp/temp-scraper.log</string>

    <key>StandardErrorPath</key>
    <string>/tmp/temp-scraper-error.log</string>

    <key>RunAtLoad</key>
    <true/>

    <key>KeepAlive</key>
    <true/>
</dict>
</plist>
EOFPLIST

    # Substitute placeholders
    sed -i '' "s|BINARY_PLACEHOLDER|$BINARY_PATH|g" "$PLIST_FILE"
    sed -i '' "s|WORKDIR_PLACEHOLDER|$(dirname "$BINARY_PATH")|g" "$PLIST_FILE"
    sed -i '' "s|PORT_PLACEHOLDER|$PORT|g" "$PLIST_FILE"
    sed -i '' "s|INTERVAL_PLACEHOLDER|$INTERVAL|g" "$PLIST_FILE"

    chmod 644 "$PLIST_FILE"

    echo "Service installed: $PLIST_FILE"
    echo ""
    echo "Load and start with:"
    echo "  launchctl load $PLIST_FILE"
    echo ""
    echo "Check status:"
    echo "  launchctl list | grep temp-scraper"
    echo "  tail -f /tmp/temp-scraper.log"
    echo ""
    echo "To unload:"
    echo "  launchctl unload $PLIST_FILE"
}

echo "Installation complete!"
