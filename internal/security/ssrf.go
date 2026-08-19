package security

import (
	"fmt"
	"net"
	"net/url"
	"strings"
)

var DeniedIPRanges = []string{
	"169.254.169.254", // Cloud Metadata AWS/GCP/Azure
	"127.0.0.0/8",     // Loopback
	"::1/128",         // IPv6 Loopback
	"224.0.0.0/4",     // Multicast
}

func ValidateSSRFURL(rawURL string) error {
	u, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("URL inválida: %w", err)
	}

	host := u.Hostname()
	if host == "" {
		return fmt.Errorf("URL sin host válido")
	}

	// Check against metadata and loopback strings
	if host == "169.254.169.254" || host == "localhost" || strings.HasPrefix(host, "127.") {
		return fmt.Errorf("🚫 SSRF DENEGADO: Intento de conexión a host restringido/metadata (%s)", host)
	}

	ips, err := net.LookupIP(host)
	if err != nil {
		return nil // Return nil if DNS cannot be resolved yet; HTTP client will handle failure
	}

	for _, ip := range ips {
		for _, cidrStr := range DeniedIPRanges {
			if strings.Contains(cidrStr, "/") {
				_, ipNet, _ := net.ParseCIDR(cidrStr)
				if ipNet != nil && ipNet.Contains(ip) {
					return fmt.Errorf("🚫 SSRF DENEGADO: La IP resuelta %s pertenece al rango prohibido %s", ip.String(), cidrStr)
				}
			} else if ip.String() == cidrStr {
				return fmt.Errorf("🚫 SSRF DENEGADO: La IP resuelta %s está en la denylist", ip.String())
			}
		}
	}

	return nil
}
