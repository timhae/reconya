package nicidentifier

import (
	"fmt"
	"reconya-ai/models"
	"strings"
)

// extractCIDR extracts the CIDR from an IPv4 address
func extractCIDR(ipv4 string) string {
	parts := strings.Split(ipv4, ".")
	if len(parts) < 3 {
		return ""
	}
	return fmt.Sprintf("%s.%s.%s.0/24", parts[0], parts[1], parts[2])
}

// extractIpRange generates IP range from a given NIC
func extractIpRange(nic models.NIC) []string {
	var addresses []string
	parts := strings.Split(nic.IPv4, ".")
	base := strings.Join(parts[:3], ".")

	for i := 1; i <= 254; i++ {
		addresses = append(addresses, fmt.Sprintf("%s.%d", base, i))
	}
	return addresses
}
