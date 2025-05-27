package netutil

import (
	"fmt"
	"net"
	"strings"
)

// Default private IP ranges in CIDR notation
const DefaultPrivateIPRanges = "10.0.0.0/8,172.16.0.0/12,192.168.0.0/16,127.0.0.0/8,169.254.0.0/16,fc00::/7,fe80::/10"

// PrivateCidrRanges holds the default private IP ranges
var PrivateCidrRanges CidrRanges

func init() {
	// Initialize PrivateCidrRanges with default private IP ranges
	var err error
	PrivateCidrRanges, err = NewCidrRanges(
		strings.Split(DefaultPrivateIPRanges, ","),
	)
	if err != nil {
		panic(fmt.Errorf("failed to initialize private CIDR ranges: %w", err))
	}
}

// CidrRanges holds a list of parsed CIDR ranges
type CidrRanges []*net.IPNet

// NewCidrRanges initializes CidrRanges from a list of CIDR strings
func NewCidrRanges(cidrs []string) (CidrRanges, error) {
	var ranges CidrRanges
	for _, cidr := range cidrs {
		_, ipnet, err := net.ParseCIDR(cidr)
		if err != nil {
			return nil, fmt.Errorf("invalid CIDR %q: %w", cidr, err)
		}
		ranges = append(ranges, ipnet)
	}
	return ranges, nil
}

// Contains returns true if the given IP is within any of the CIDR ranges
func (c CidrRanges) Contains(ip net.IP) bool {
	for _, ipnet := range c {
		if ipnet.Contains(ip) {
			return true
		}
	}
	return false
}
