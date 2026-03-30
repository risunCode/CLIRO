package cloudflared

import "testing"

func TestExtractTunnelURL_QuickTunnel(t *testing.T) {
	line := "INF https://gentle-breeze-demo.trycloudflare.com registered tunnel connection"
	if got := extractTunnelURL(line); got != "https://gentle-breeze-demo.trycloudflare.com" {
		t.Fatalf("quick tunnel url = %q", got)
	}
}

func TestExtractTunnelURL_NamedTunnel(t *testing.T) {
	line := `INF Updated to new configuration config="{\"ingress\":[{\"hostname\":\"api.example.com\",\"service\":\"http://localhost:8095\"}]}"`
	if got := extractTunnelURL(line); got != "https://api.example.com" {
		t.Fatalf("named tunnel url = %q", got)
	}
}

func TestIsCloudflaredArchiveName(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		wanted bool
	}{
		{name: "binary", input: "cloudflared.exe", wanted: false},
		{name: "zip", input: "cloudflared-windows-amd64.zip", wanted: true},
		{name: "tgz", input: "cloudflared-darwin-amd64.tgz", wanted: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := isCloudflaredArchiveName(tc.input); got != tc.wanted {
				t.Fatalf("isCloudflaredArchiveName(%q) = %t, want %t", tc.input, got, tc.wanted)
			}
		})
	}
}

func TestFirstNonEmptyLine(t *testing.T) {
	input := "\n\ncloudflared version 2026.1.0\n"
	if got := firstNonEmptyLine(input); got != "cloudflared version 2026.1.0" {
		t.Fatalf("firstNonEmptyLine = %q", got)
	}
}
