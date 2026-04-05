package kiro

// OAuth and PKCE utility functions for Kiro auth flows.

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/google/uuid"
)

func NormalizeSocialProvider(provider string) (SocialProvider, error) {
	switch strings.ToLower(strings.TrimSpace(provider)) {
	case "google", "":
		return SocialProviderGoogle, nil
	case "github":
		return SocialProviderGitHub, nil
	default:
		return "", fmt.Errorf("unsupported Kiro social provider: %s", strings.TrimSpace(provider))
	}
}

func GeneratePKCE() (string, string, error) {
	raw := make([]byte, 64)
	if _, err := rand.Read(raw); err != nil {
		return "", "", err
	}
	verifier := base64.RawURLEncoding.EncodeToString(raw)
	if len(verifier) > 128 {
		verifier = verifier[:128]
	}
	hashed := sha256.Sum256([]byte(verifier))
	challenge := base64.RawURLEncoding.EncodeToString(hashed[:])
	return verifier, challenge, nil
}

func GenerateState() (string, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(raw), nil
}

func BuildSocialLoginURL(provider SocialProvider, codeChallenge string, state string) string {
	customRedirectURI := "kiro://kiro.kiroAgent/authenticate-success"
	return fmt.Sprintf("%s/login?idp=%s&redirect_uri=%s&code_challenge=%s&code_challenge_method=S256&state=%s",
		SocialAuthURL,
		url.QueryEscape(string(provider)),
		url.QueryEscape(customRedirectURI),
		url.QueryEscape(strings.TrimSpace(codeChallenge)),
		url.QueryEscape(strings.TrimSpace(state)),
	)
}

func BuildSocialUserAgent() string {
	return fmt.Sprintf("KiroIDE-0.11.107-%s", strings.ReplaceAll(uuid.NewString(), "-", ""))
}

func DetermineAuthMethod(tokens *TokenData) string {
	if tokens == nil {
		return "social"
	}
	if strings.TrimSpace(tokens.ClientID) != "" && strings.TrimSpace(tokens.ClientSecret) != "" {
		return "idc"
	}
	return "social"
}

func ExtractEmailFromJWT(token string) string {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return ""
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return ""
	}
	var claims struct {
		Email string `json:"email"`
	}
	if err := json.Unmarshal(payload, &claims); err != nil {
		return ""
	}
	return strings.TrimSpace(claims.Email)
}
