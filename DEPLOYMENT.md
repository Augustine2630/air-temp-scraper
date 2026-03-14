# Deployment Guide

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│  macOS System                                               │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  ┌──────────────────────────────────────────────────────┐  │
│  │  temp-scraper (Go binary)                            │  │
│  ├──────────────────────────────────────────────────────┤  │
│  │  • SMC Reader (native IOKit, no subprocess)          │  │
│  │  • Persistent SMC connection (reused every scrape)   │  │
│  │  • Type conversion (fp*/sp*/flt/ioft)                │  │
│  │  • 190+ temperature sensor keys                      │  │
│  │  • Prometheus metric export (port 9100)              │  │
│  └─────────────────┬──────────────────────────────────┘  │
│                    │                                       │
│  ┌─────────────────▼──────────────────────────────────┐  │
│  │  IOKit Framework                                   │  │
│  ├────────────────────────────────────────────────────┤  │
│  │  • SMC key info cache (100 entries, thread-safe)   │  │
│  │  • Direct SMC key/value reads                      │  │
│  └─────────────────┬──────────────────────────────────┘  │
│                    │                                       │
│  ┌─────────────────▼──────────────────────────────────┐  │
│  │  Apple SMC (Hardware)                              │  │
│  ├────────────────────────────────────────────────────┤  │
│  │  • Temperature sensor readouts (CPU, GPU, etc)     │  │
│  │  • Real-time hardware values                       │  │
│  └────────────────────────────────────────────────────┘  │
│                                                            │
└────────────────────────────────────────────────────────────┘

                    HTTP GET /metrics
                           ↓
                  http://localhost:9100/metrics
```

## Installation Methods

### Method 1: One-Line Install (Recommended)

```bash
curl -fsSL https://raw.githubusercontent.com/Augustine2630/air-temp-scraper/main/install.sh | bash
```

**What it does:**
- Detects macOS architecture (arm64 or amd64)
- Downloads latest binary from GitHub releases
- Verifies binary integrity (IOKit linkage)
- Installs to `~/.local/bin/temp-scraper`
- Creates launchd plist for auto-start
- Tests metrics endpoint
- Shows service management commands

### Method 2: Build from Source

```bash
git clone https://github.com/Augustine2630/air-temp-scraper
cd air-temp-scraper
make build
./install-systemd.sh
```

### Method 3: Manual Binary Download

```bash
# Download from releases
curl -L -o temp-scraper_darwin_arm64 \
  https://github.com/Augustine2630/air-temp-scraper/releases/download/v1.0.0/temp-scraper_darwin_arm64

chmod +x temp-scraper_darwin_arm64
TEMP_SCRAPER_PORT=9100 ./temp-scraper_darwin_arm64
```

---

## Release Process

### 1. Build Release Binaries

```bash
make release
# Creates ./bin/temp-scraper_darwin_arm64 and ./bin/temp-scraper_darwin_amd64
```

### 2. Create GitHub Release

```bash
gh release create v1.0.0 \
  --title "Temperature Scraper v1.0.0" \
  --notes "Native SMC temperature reading, zero subprocess overhead" \
  bin/temp-scraper_darwin_arm64 \
  bin/temp-scraper_darwin_amd64
```

### 3. Users Run:

```bash
curl -fsSL https://raw.githubusercontent.com/Augustine2630/air-temp-scraper/main/install.sh | bash
```

---

## Service Lifecycle

### Automatic (launchd)

1. **Install** – `install.sh` creates `~/Library/LaunchAgents/com.local.temp-scraper.plist`
2. **Load** – `launchctl load` starts the service
3. **Run** – Service runs in background, auto-restarts on crash
4. **Monitor** – `tail -f /tmp/temp-scraper.log`

### Manual Commands

```bash
# Check status
launchctl list | grep temp-scraper

# Stop
launchctl unload ~/Library/LaunchAgents/com.local.temp-scraper.plist

# Start
launchctl load ~/Library/LaunchAgents/com.local.temp-scraper.plist

# Restart
launchctl unload ~/Library/LaunchAgents/com.local.temp-scraper.plist && \
sleep 1 && \
launchctl load ~/Library/LaunchAgents/com.local.temp-scraper.plist

# View logs
tail -f /tmp/temp-scraper.log
tail -f /tmp/temp-scraper-error.log
```

---

## Configuration

### Environment Variables

Set in the launchd plist or pass on command line:

| Variable | Default | Description |
|----------|---------|-------------|
| `TEMP_SCRAPER_PORT` | *(required)* | HTTP metrics port (e.g. 9100) |
| `TEMP_SCRAPER_INTERVAL` | `30s` | Scrape interval (e.g. 15s, 1m) |

### Edit Service Configuration

```bash
# Edit plist
nano ~/Library/LaunchAgents/com.local.temp-scraper.plist

# Change PORT or INTERVAL in <dict><key>EnvironmentVariables</key>...

# Reload
launchctl unload ~/Library/LaunchAgents/com.local.temp-scraper.plist
launchctl load ~/Library/LaunchAgents/com.local.temp-scraper.plist
```

---

## Metrics

### Endpoint

```
GET http://localhost:9100/metrics
```

### Example Metrics

```prometheus
# Scrape runtime
temp_scraper_scrape_attempts_total 1205
temp_scraper_scrape_errors_total 0
temp_scraper_scrape_success_total 1205
temp_scraper_scrape_duration_seconds{le="0.005"} 1200
temp_scraper_scrape_last_duration_seconds 0.003

# Parse runtime
temp_scraper_parse_duration_seconds{le="0.001"} 1195
temp_scraper_parse_sensors_total 98

# Per-sensor temperatures (Celsius)
temp_scraper_sensor_temperature_celsius{sensor_key="TC0C",sensor_name="CPU Core 1",sensor_type="sp78"} 52.1
temp_scraper_sensor_temperature_celsius{sensor_key="TC0D",sensor_name="CPU Core 2",sensor_type="sp78"} 51.9
temp_scraper_sensor_temperature_celsius{sensor_key="TC0E",sensor_name="CPU Proximity",sensor_type="sp78"} 50.2
temp_scraper_sensor_temperature_celsius{sensor_key="Tp01",sensor_name="CPU Performance Core 1",sensor_type="sp78"} 48.7

# Go runtime metrics
process_resident_memory_bytes 5242880
process_cpu_seconds_total 0.42
go_goroutines 3
```

### Prometheus Scrape Configuration

```yaml
# prometheus.yml
global:
  scrape_interval: 30s

scrape_configs:
  - job_name: 'temp-scraper'
    static_configs:
      - targets: ['localhost:9100']
```

---

## Troubleshooting

### Service won't start

1. Check error log:
   ```bash
   tail -f /tmp/temp-scraper-error.log
   ```

2. Verify binary permissions:
   ```bash
   ls -la ~/.local/bin/temp-scraper
   ```

3. Test binary manually:
   ```bash
   TEMP_SCRAPER_PORT=9100 ~/.local/bin/temp-scraper
   ```

### Port already in use

Change port in plist and reload:
```bash
# Find what's using port 9100
lsof -i :9100

# Edit plist with new port
nano ~/Library/LaunchAgents/com.local.temp-scraper.plist

# Restart service
launchctl unload ~/Library/LaunchAgents/com.local.temp-scraper.plist
launchctl load ~/Library/LaunchAgents/com.local.temp-scraper.plist
```

### No temperature readings

Verify IOKit access (requires no special permissions on macOS):
```bash
# Test manually
TEMP_SCRAPER_PORT=9100 ~/.local/bin/temp-scraper &
sleep 2
curl http://localhost:9100/metrics | grep temp_scraper_sensor_temperature_celsius
```

### Metrics endpoint 404

Verify service is running:
```bash
curl -v http://localhost:9100/metrics
```

If timeout, check:
```bash
launchctl list | grep temp-scraper
lsof -i :9100
```

---

## Performance Notes

- **CPU**: <1% idle (only at scrape intervals)
- **Memory**: ~5-6 MB resident
- **Scrape latency**: 2-4ms (10 ms p99)
- **IOKit round-trips**: ~100-150 keys per scrape, cached (no re-reads within interval)
- **Process spawning**: Zero (all native code)

---

## Uninstall

```bash
# Stop service
launchctl unload ~/Library/LaunchAgents/com.local.temp-scraper.plist

# Remove files
rm ~/Library/LaunchAgents/com.local.temp-scraper.plist
rm ~/.local/bin/temp-scraper

# Verify
launchctl list | grep temp-scraper  # should show nothing
```
