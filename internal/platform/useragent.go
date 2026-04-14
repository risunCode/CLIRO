package platform

import "fmt"

const (
	codexTUIVersion    = "0.118.0"
	codexTUIApp        = "iTerm.app/3.6.9"
	codexTUISpoofOS    = "Mac OS 26.3.1"
	codexTUISpoofArch  = "arm64"
)

// BuildCodexTUIUserAgent mocks the codex-tui user-agent format used by the
// official Codex CLI client.
// Format: codex-tui/{version} ({os} {osVersion}; {arch}) {app} (codex-tui; {version})
// Example: codex-tui/0.118.0 (Mac OS 26.3.1; arm64) iTerm.app/3.6.9 (codex-tui; 0.118.0)
func BuildCodexTUIUserAgent() string {
	return fmt.Sprintf("codex-tui/%s (%s; %s) %s (codex-tui; %s)",
		codexTUIVersion,
		codexTUISpoofOS,
		codexTUISpoofArch,
		codexTUIApp,
		codexTUIVersion,
	)
}
