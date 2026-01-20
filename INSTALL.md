# Myopsida Installation Guide

## Arch Linux Installation (Recommended)

### Using makepkg

1. **Clone and build the package:**
   ```bash
   git clone https://github.com/yi-fan-song/myopsida.git
   cd myopsida/dist
   makepkg -si
   ```

   This will:
   - Build the binary with optimizations
   - Install it to `/usr/local/bin/myopsida`
   - Install systemd service and timer files
   - Create configuration file at `/etc/myopsida/config.env`

2. **Configure credentials:**
   ```bash
   sudo nano /etc/myopsida/config.env
   ```

   Fill in your Cloudflare credentials:
   - `CF_API_TOKEN`: Your Cloudflare API token
   - `CF_ZONE_ID`: Your domain's zone ID
   - `CF_RECORD_ID_V4`: DNS record ID for A record (IPv4)
   - `CF_RECORD_ID_V6`: DNS record ID for AAAA record (IPv6)
   - `CF_RECORD_NAME`: Your domain name (e.g., example.com)

3. **Enable and start the timer:**
   ```bash
   sudo systemctl daemon-reload
   sudo systemctl enable myopsida.timer
   sudo systemctl start myopsida.timer
   ```

4. **Verify it's running:**
   ```bash
   sudo systemctl status myopsida.timer
   sudo systemctl list-timers myopsida.timer
   ```

## Manual Installation (Any Linux Distribution)

1. **Build from source:**
   ```bash
   git clone https://github.com/yi-fan-song/myopsida.git
   cd myopsida
   go build -o myopsida
   ```

2. **Install the binary:**
   ```bash
   sudo install -Dm755 myopsida /usr/local/bin/myopsida
   ```

3. **Install systemd files:**
   ```bash
   sudo install -Dm644 systemd/myopsida.service /usr/lib/systemd/system/myopsida.service
   sudo install -Dm644 systemd/myopsida.timer /usr/lib/systemd/system/myopsida.timer
   ```

4. **Create configuration directory and file:**
   ```bash
   sudo mkdir -p /etc/myopsida
   sudo cp example.env /etc/myopsida/config.env
   sudo nano /etc/myopsida/config.env
   ```

5. **Enable and start the timer:**
   ```bash
   sudo systemctl daemon-reload
   sudo systemctl enable myopsida.timer
   sudo systemctl start myopsida.timer
   ```

## Finding Your Cloudflare Credentials

### Zone ID
1. Log in to [Cloudflare Dashboard](https://dash.cloudflare.com)
2. Select your domain
3. On the Overview page, scroll down to find "Zone ID" on the right side

### API Token
1. Go to [API Tokens](https://dash.cloudflare.com/profile/api-tokens)
2. Click "Create Token"
3. Choose "Edit zone DNS" template
4. Select your zone
5. Create the token and copy it

### DNS Record ID
1. Go to DNS > Records in your Cloudflare dashboard
2. For each record you want to update (A and AAAA), right-click and select "Inspect"
3. In the browser DevTools, look for the record data or API response
4. Alternatively, use the Cloudflare API to list records:
   ```bash
   curl -X GET "https://api.cloudflare.com/client/v4/zones/ZONE_ID/dns_records" \
     -H "Authorization: Bearer API_TOKEN" \
     -H "Content-Type: application/json" | jq '.result[] | {id, name, type, content}'
   ```

## Systemd Timer Usage

### View Timer Status
```bash
sudo systemctl status myopsida.timer
sudo systemctl list-timers myopsida.timer
```

### View Recent Runs
```bash
sudo journalctl -u myopsida.service -n 20
```

### Follow Logs in Real-time
```bash
sudo journalctl -u myopsida.service -f
```

### Manual Trigger
To manually run the service without waiting for the timer:
```bash
sudo systemctl start myopsida.service
```

### Disable the Timer
```bash
sudo systemctl stop myopsida.timer
sudo systemctl disable myopsida.timer
```

## Troubleshooting

### Timer not running
```bash
sudo systemctl daemon-reload
sudo systemctl start myopsida.timer
```

### Check timer log
```bash
sudo journalctl -u myopsida.timer -n 50
```

### View full service logs
```bash
sudo journalctl -u myopsida.service -n 100 --no-pager
```

### Verify configuration is loaded
```bash
sudo systemctl cat myopsida.service | grep -A 20 "EnvironmentFile"
```

## Uninstallation

### On Arch Linux
```bash
sudo systemctl stop myopsida.timer
sudo systemctl disable myopsida.timer
sudo pacman -R myopsida
```

### Manual Installation
```bash
sudo systemctl stop myopsida.timer
sudo systemctl disable myopsida.timer
sudo rm /usr/local/bin/myopsida
sudo rm /usr/lib/systemd/system/myopsida.{service,timer}
sudo rm -rf /etc/myopsida
sudo systemctl daemon-reload
```
