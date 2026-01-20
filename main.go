package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"time"
)

// IPResponse represents the response from ipify API
type IPResponse struct {
	IP string `json:"ip"`
}

// CloudflareDNSRecord represents a Cloudflare DNS record
type CloudflareDNSRecord struct {
	Type    string `json:"type"`
	Name    string `json:"name"`
	Content string `json:"content"`
	TTL     int    `json:"ttl"`
	Proxied bool   `json:"proxied"`
	Comment string `json:"comment"`
}

// CloudflareResponse represents the response from Cloudflare API
type CloudflareResponse struct {
	Success bool          `json:"success"`
	Errors  []interface{} `json:"errors"`
	Result  interface{}   `json:"result"`
}

// GetPublicIPv4 fetches the public IPv4 address
func GetPublicIPv4() (string, error) {
	return getIP("https://api.ipify.org/?format=json")
}

// GetPublicIPv6 fetches the public IPv6 address
func GetPublicIPv6() (string, error) {
	return getIP("https://api64.ipify.org/?format=json")
}

// getIP is a helper function to fetch IP from ipify API with retry logic
func getIP(url string) (string, error) {
	maxRetries := 5
	minWaitTime := 5 * time.Second

	var lastErr error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		resp, err := http.Get(url)
		if err != nil {
			lastErr = err
		} else if resp.StatusCode != http.StatusOK {
			lastErr = fmt.Errorf("unexpected status code %d from %s", resp.StatusCode, url)
			resp.Body.Close()
		} else {
			var ipResp IPResponse
			if err := json.NewDecoder(resp.Body).Decode(&ipResp); err != nil {
				lastErr = fmt.Errorf("failed to decode IP response: %w", err)
			} else {
				resp.Body.Close()
				return ipResp.IP, nil
			}
			resp.Body.Close()
		}

		if attempt < maxRetries {
			// Calculate exponential backoff: 5s, 10s, 20s, 40s, 80s
			waitTime := minWaitTime * time.Duration(math.Pow(2, float64(attempt)))
			fmt.Printf("Retry attempt %d/%d in %v seconds...\n", attempt+1, maxRetries, waitTime.Seconds())
			time.Sleep(waitTime)
		}
	}

	return "", fmt.Errorf("failed to fetch IP from %s after %d retries: %w", url, maxRetries, lastErr)
}

// UpdateCloudflareDNSRecord updates a DNS record on Cloudflare with retry logic
func UpdateCloudflareDNSRecord(zoneID, recordID, apiToken string, record CloudflareDNSRecord) error {
	maxRetries := 5
	minWaitTime := 5 * time.Second
	url := fmt.Sprintf("https://api.cloudflare.com/client/v4/zones/%s/dns_records/%s", zoneID, recordID)

	var lastErr error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		payload, err := json.Marshal(record)
		if err != nil {
			return fmt.Errorf("failed to marshal DNS record: %w", err)
		}

		req, err := http.NewRequest("PUT", url, bytes.NewBuffer(payload))
		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiToken))

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			lastErr = err
		} else {
			body, err := io.ReadAll(resp.Body)
			resp.Body.Close()

			if err != nil {
				lastErr = fmt.Errorf("failed to read response body: %w", err)
			} else {
				var cfResp CloudflareResponse
				if err := json.Unmarshal(body, &cfResp); err != nil {
					lastErr = fmt.Errorf("failed to decode Cloudflare response: %w", err)
				} else if !cfResp.Success {
					errors := "unknown error"
					if len(cfResp.Errors) > 0 {
						errors = fmt.Sprintf("%v", cfResp.Errors)
					}
					lastErr = fmt.Errorf("cloudflare API error: %s", errors)
				} else {
					return nil
				}
			}
		}

		if attempt < maxRetries {
			// Calculate exponential backoff: 5s, 10s, 20s, 40s, 80s
			waitTime := minWaitTime * time.Duration(math.Pow(2, float64(attempt)))
			fmt.Printf("Retry attempt %d/%d in %v seconds...\n", attempt+1, maxRetries, waitTime.Seconds())
			time.Sleep(waitTime)
		}
	}

	return fmt.Errorf("failed to update DNS record after %d retries: %w", maxRetries, lastErr)
}

func main() {
	// Parse command-line flags
	zoneID := flag.String("zone-id", "", "Cloudflare Zone ID (required)")
	recordIDv4 := flag.String("record-id-v4", "", "Cloudflare DNS Record ID for A record (required)")
	recordIDv6 := flag.String("record-id-v6", "", "Cloudflare DNS Record ID for AAAA record (required)")
	apiToken := flag.String("api-token", "", "Cloudflare API Token (from env var CF_API_TOKEN or flag)")
	recordName := flag.String("name", "", "DNS record name (required)")
	ttl := flag.Int("ttl", 1, "DNS record TTL (default: auto)")
	proxied := flag.Bool("proxied", true, "Whether the records are proxied (default: true)")
	comment := flag.String("comment", "", "DNS record comment")

	flag.Parse()

	// Get API token from environment variable if not provided via flag
	if *apiToken == "" {
		*apiToken = os.Getenv("CF_API_TOKEN")
	}

	// Validate required flags
	if *zoneID == "" || *recordIDv4 == "" || *recordIDv6 == "" || *apiToken == "" || *recordName == "" {
		fmt.Fprintf(os.Stderr, "Error: Missing required parameters\n")
		fmt.Fprintf(os.Stderr, "Usage: %s -zone-id ZONE_ID -record-id-v4 RECORD_ID_V4 -record-id-v6 RECORD_ID_V6 -name DOMAIN_NAME\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\nSet CF_API_TOKEN environment variable or use -api-token flag\n")
		flag.PrintDefaults()
		os.Exit(1)
	}

	// Fetch both IPv4 and IPv6 addresses
	fmt.Println("Fetching public IPv4 address...")
	ipv4, err := GetPublicIPv4()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error fetching IPv4: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Got IPv4: %s\n", ipv4)

	fmt.Println("Fetching public IPv6 address...")
	ipv6, err := GetPublicIPv6()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error fetching IPv6: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Got IPv6: %s\n", ipv6)

	// Create DNS records for both A and AAAA
	recordA := CloudflareDNSRecord{
		Type:    "A",
		Name:    *recordName,
		Content: ipv4,
		TTL:     *ttl,
		Proxied: *proxied,
		Comment: *comment,
	}

	recordAAAA := CloudflareDNSRecord{
		Type:    "AAAA",
		Name:    *recordName,
		Content: ipv6,
		TTL:     *ttl,
		Proxied: *proxied,
		Comment: *comment,
	}

	// Update both Cloudflare DNS records
	fmt.Println("Updating Cloudflare DNS records...")
	if err := UpdateCloudflareDNSRecord(*zoneID, *recordIDv4, *apiToken, recordA); err != nil {
		fmt.Fprintf(os.Stderr, "Error updating A record: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✓ A record updated successfully")

	if err := UpdateCloudflareDNSRecord(*zoneID, *recordIDv6, *apiToken, recordAAAA); err != nil {
		fmt.Fprintf(os.Stderr, "Error updating AAAA record: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✓ AAAA record updated successfully")

	fmt.Println("\nSuccessfully updated all DNS records!")
	fmt.Printf("A record:    %s -> %s\n", *recordName, ipv4)
	fmt.Printf("AAAA record: %s -> %s\n", *recordName, ipv6)
}
