package netutil

import (
	"net"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewCidrRanges(t *testing.T) {
	tests := []struct {
		name      string
		cidrs     []string
		wantErr   bool
		wantCount int
	}{
		{
			name:      "valid IPv4 and IPv6 CIDRs",
			cidrs:     []string{"10.0.0.0/8", "fc00::/7"},
			wantErr:   false,
			wantCount: 2,
		},
		{
			name:      "invalid CIDR",
			cidrs:     []string{"10.0.0.0/8", "invalid"},
			wantErr:   true,
			wantCount: 0,
		},
		{
			name:      "empty CIDR list",
			cidrs:     []string{},
			wantErr:   false,
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ranges, err := NewCidrRanges(tt.cidrs)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Len(t, ranges, tt.wantCount)
			}
		})
	}
}

func TestCidrRanges_Contains(t *testing.T) {
	cidrs := []string{"10.0.0.0/8", "192.168.0.0/16", "fc00::/7"}
	ranges, err := NewCidrRanges(cidrs)
	require.NoError(t, err)

	tests := []struct {
		ip      string
		want    bool
		comment string
	}{
		{"10.1.2.3", true, "in 10.0.0.0/8"},
		{"192.168.1.1", true, "in 192.168.0.0/16"},
		{"8.8.8.8", false, "not in any range"},
		{"fc00::1", true, "in fc00::/7"},
		{"fe80::1", false, "not in any range"},
	}

	for _, tt := range tests {
		t.Run(tt.ip, func(t *testing.T) {
			ip := net.ParseIP(tt.ip)
			require.Equal(t, tt.want, ranges.Contains(ip), tt.comment)
		})
	}
}

func TestPrivateCidrRanges_Defaults(t *testing.T) {
	// Test that the default PrivateCidrRanges contains a known private IP
	require.True(t, PrivateCidrRanges.Contains(net.ParseIP("10.0.0.1")))
	require.True(t, PrivateCidrRanges.Contains(net.ParseIP("192.168.1.1")))
	require.True(t, PrivateCidrRanges.Contains(net.ParseIP("fc00::1")))
	require.False(t, PrivateCidrRanges.Contains(net.ParseIP("8.8.8.8")))
	require.False(t, PrivateCidrRanges.Contains(net.ParseIP("2001:4860:4860::8888"))) // Google DNS IPv6
}
