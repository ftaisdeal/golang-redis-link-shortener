package util

import (
	"fmt"
	"net"
	"net/netip"
	"net/url"
	"regexp"
	"strings"
)

var codePattern = regexp.MustCompile(`^[a-zA-Z0-9_-]{3,32}$`)

func ValidateCustomCode(code string) error {
	if !codePattern.MatchString(code) {
		return fmt.Errorf("custom_code must be 3-32 chars: a-z, A-Z, 0-9, _ or -")
	}
	return nil
}

func ValidateAndNormalizeURL(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", fmt.Errorf("url is required")
	}
	if len(raw) > 2048 {
		return "", fmt.Errorf("url is too long")
	}

	parsed, err := url.Parse(raw)
	if err != nil {
		return "", fmt.Errorf("invalid url")
	}

	scheme := strings.ToLower(parsed.Scheme)
	if scheme != "http" && scheme != "https" {
		return "", fmt.Errorf("only http and https URLs are supported")
	}
	if parsed.Hostname() == "" {
		return "", fmt.Errorf("url host is required")
	}

	if hostIsLocalOrPrivate(parsed.Hostname()) {
		return "", fmt.Errorf("local/private network targets are not allowed")
	}

	return parsed.String(), nil
}

func hostIsLocalOrPrivate(host string) bool {
	lower := strings.ToLower(host)
	if lower == "localhost" {
		return true
	}

	if ip, err := netip.ParseAddr(host); err == nil {
		return ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast()
	}

	addrs, err := net.LookupIP(host)
	if err != nil {
		return false
	}
	for _, ip := range addrs {
		addr, ok := netip.AddrFromSlice(ip)
		if !ok {
			continue
		}
		if addr.IsLoopback() || addr.IsPrivate() || addr.IsLinkLocalUnicast() || addr.IsLinkLocalMulticast() {
			return true
		}
	}
	return false
}
