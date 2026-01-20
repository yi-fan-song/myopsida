# myopsida

A Go program that fetches your current public IP address and automatically updates a Cloudflare DNS record with it.

## Features

- Fetches public IPv4 address from https://api.ipify.org/
- Fetches public IPv6 address from https://api64.ipify.org/
- Updates both A and AAAA Cloudflare DNS records simultaneously
- Command-line interface with flexible configuration

## Prerequisites

- Go 1.21 or later
- Cloudflare account with API token
- Zone ID and DNS Record ID from Cloudflare

## Building

### Manual Build

```bash
go build -o myopsida
```

### Arch Linux with makepkg

```bash
# Clone the repository
git clone https://github.com/yi-fan-song/myopsida.git
cd myopsida/dist

# Build and install the package
makepkg -si
```

This will build the binary, install it to `/usr/local/bin/myopsida`, and set up the systemd service and timer files.

## Usage

```bash
./myopsida -zone-id ZONE_ID -record-id-v4 RECORD_ID_V4 -record-id-v6 RECORD_ID_V6 -name example.com
```

The API token is read from the `CF_API_TOKEN` environment variable. You can also override it with the `-api-token` flag if needed.

### Systemd Timer Setup

To automatically run myopsida every 30 minutes via systemd timer:

1. Copy the example configuration:
   ```bash
   sudo cp /etc/myopsida/config.env.example /etc/myopsida/config.env
   ```

2. Edit the configuration with your Cloudflare credentials:
   ```bash
   sudo nano /etc/myopsida/config.env
   ```

3. Reload systemd and enable the timer:
   ```bash
   sudo systemctl daemon-reload
   sudo systemctl enable myopsida.timer
   sudo systemctl start myopsida.timer
   ```

4. Check timer status:
   ```bash
   sudo systemctl status myopsida.timer
   sudo systemctl list-timers myopsida.timer
   ```

5. View logs:
   ```bash
   sudo journalctl -u myopsida.service -f
   ```

### Required Flags

- `-zone-id`: Your Cloudflare Zone ID
- `-record-id-v4`: The DNS Record ID for the A record (IPv4)
- `-record-id-v6`: The DNS Record ID for the AAAA record (IPv6)
- `-name`: The DNS record name (e.g., example.com)

### Required Environment Variable

- `CF_API_TOKEN`: Your Cloudflare API Token

### Optional Flags

- `-ttl`: TTL in seconds (default: `3600`)
- `-proxied`: Whether the records should be proxied by Cloudflare (default: `true`)
- `-comment`: Comment for the DNS records (default: empty)

## Examples

Update both A and AAAA records:
```bash
export CF_API_TOKEN="token123"
./myopsida -zone-id abc123 -record-id-v4 def456 -record-id-v6 ghi789 -name example.com
```

With custom TTL and no proxying:
```bash
export CF_API_TOKEN="token123"
./myopsida -zone-id abc123 -record-id-v4 def456 -record-id-v6 ghi789 -name example.com -ttl 300 -proxied false
```

## Environment Variables

You can also set the required values via environment variables and pass them as flags:

```bash
export CF_API_TOKEN="token123"

./myopsida -zone-id abc123 -record-id-v4 def456 -record-id-v6 ghi789 -name example.com
```

## Error Handling

The program will exit with code 1 if:
- Required flags are missing
- Failed to fetch IP addresses after 5 retries (with exponential backoff, minimum 5s wait)
- Failed to update Cloudflare DNS record after 5 retries (with exponential backoff, minimum 5s wait)
- Invalid parameters are provided

## Retry Logic

The program implements automatic retry with exponential backoff for both IP fetching and DNS updates:
- **Maximum retries**: 5 attempts
- **Backoff times**: 5s, 10s, 20s, 40s, 80s
- **Minimum wait**: 5 seconds between retries

## License

See LICENSE file for details
