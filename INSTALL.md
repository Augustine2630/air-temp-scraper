# Installation

## Quick Install (One-Line)

**macOS only** – downloads latest release and sets up launchd service:

```bash
curl -fsSL https://raw.githubusercontent.com/Augustine2630/air-temp-scraper/main/install.sh | bash
```

Or with custom port and interval:

```bash
curl -fsSL https://raw.githubusercontent.com/Augustine2630/air-temp-scraper/main/install.sh | bash -s 9100 30s
```

### What the installer does:

1. **Detects architecture** (arm64 or x86_64)
2. **Downloads binary** from latest GitHub release
3. **Verifies binary** (Mach-O format, IOKit linked)
4. **Installs to** `~/.local/bin/temp-scraper`
5. **Creates launchd plist** at `~/Library/LaunchAgents/com.local.temp-scraper.plist`
6. **Loads service** and starts automatically
7. **Tests metrics endpoint** and verifies temperature readings

---

## Manual Build & Install

### Build from source:

```bash
git clone https://github.com/Augustine2630/air-temp-scraper
cd air-temp-scraper
make build
```

Binary is at `./bin/temp-scraper_darwin_arm64` (or `_amd64` on Intel).

### Install as launchd service:

```bash
./install-systemd.sh 9100 30s
launchctl load ~/Library/LaunchAgents/com.local.temp-scraper.plist
```

### Or run directly:

```bash
TEMP_SCRAPER_PORT=9100 ./bin/temp-scraper_darwin_arm64
```

---

## Build release binaries for both architectures:

```bash
make release
# Creates:
#   ./bin/temp-scraper_darwin_arm64
#   ./bin/temp-scraper_darwin_amd64
```

---

## Service Management

### Check status:

```bash
launchctl list | grep temp-scraper
tail -f /tmp/temp-scraper.log
```

### Stop:

```bash
launchctl unload ~/Library/LaunchAgents/com.local.temp-scraper.plist
```

### Start:

```bash
launchctl load ~/Library/LaunchAgents/com.local.temp-scraper.plist
```

### Restart:

```bash
launchctl unload ~/Library/LaunchAgents/com.local.temp-scraper.plist
sleep 1
launchctl load ~/Library/LaunchAgents/com.local.temp-scraper.plist
```

---

## Metrics endpoint

Once running:

```bash
curl http://localhost:9100/metrics | grep temp_scraper_sensor_temperature_celsius
```

Sample output:

```
temp_scraper_sensor_temperature_celsius{sensor_key="TC0C",sensor_name="CPU Core 1",sensor_type="sp78"} 52.1
temp_scraper_sensor_temperature_celsius{sensor_key="TC0D",sensor_name="CPU Core 2",sensor_type="sp78"} 51.9
temp_scraper_sensor_temperature_celsius{sensor_key="TC0E",sensor_name="CPU Proximity",sensor_type="sp78"} 50.2
```

---

## Environment Variables

- `TEMP_SCRAPER_PORT` – HTTP metrics port (required, default: 9100)
- `TEMP_SCRAPER_INTERVAL` – scrape interval (optional, default: 30s)

Example:

```bash
export TEMP_SCRAPER_PORT=9100
export TEMP_SCRAPER_INTERVAL=15s
./bin/temp-scraper_darwin_arm64
```

---

## Troubleshooting

### Binary not found after install

Check that the latest release includes your architecture:

```bash
curl -s https://api.github.com/repos/Augustine2630/air-temp-scraper/releases/latest \
  | grep browser_download_url | cut -d'"' -f4
```

### Service won't start

Check the error log:

```bash
tail -f /tmp/temp-scraper-error.log
```

### Port already in use

Change the port in the plist file and reload:

```bash
# Edit the plist:
nano ~/Library/LaunchAgents/com.local.temp-scraper.plist
# Change TEMP_SCRAPER_PORT to a different value, e.g., 9200

# Reload:
launchctl unload ~/Library/LaunchAgents/com.local.temp-scraper.plist
launchctl load ~/Library/LaunchAgents/com.local.temp-scraper.plist
```

### Metrics endpoint 404

Verify the service is running:

```bash
launchctl list | grep temp-scraper
```

If listed, check that it's listening on the correct port:

```bash
lsof -i :9100
```
