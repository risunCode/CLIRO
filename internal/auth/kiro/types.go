package kiro

// Wire types and constants for Kiro auth flows.

const (
	RegisterClientURL   = "https://oidc.us-east-1.amazonaws.com/client/register"
	DeviceAuthURL       = "https://oidc.us-east-1.amazonaws.com/device_authorization"
	DeviceTokenURL      = "https://oidc.us-east-1.amazonaws.com/token"
	BuilderStartURL     = "https://view.awsapps.com/start"
	BuilderClientName   = "kiro-oauth-client"
	RuntimeUserAgent    = "aws-sdk-js/1.2.15 ua/2.1 os/linux lang/js md/nodejs#22.21.1 api/codewhispererstreaming#1.2.15 m/E KiroIDE-0.11.107"
	RuntimeAmzUserAgent = "aws-sdk-js/1.2.15 KiroIDE 0.11.107"
	SocialAuthURL       = "https://prod.us-east-1.auth.desktop.kiro.dev"
)

var BuilderScopes = []string{
	"codewhisperer:completions",
	"codewhisperer:analysis",
	"codewhisperer:conversations",
	"codewhisperer:transformations",
	"codewhisperer:taskassist",
}

type SocialProvider string

const (
	SocialProviderGoogle SocialProvider = "Google"
	SocialProviderGitHub SocialProvider = "Github"
)

type AuthStart struct {
	SessionID       string `json:"sessionId"`
	AuthURL         string `json:"authUrl"`
	VerificationURL string `json:"verificationUrl,omitempty"`
	UserCode        string `json:"userCode,omitempty"`
	ExpiresAt       int64  `json:"expiresAt,omitempty"`
	Status          string `json:"status"`
	AuthMethod      string `json:"authMethod,omitempty"`
	Provider        string `json:"provider,omitempty"`
}

type AuthSessionView struct {
	SessionID       string `json:"sessionId"`
	AuthURL         string `json:"authUrl"`
	VerificationURL string `json:"verificationUrl,omitempty"`
	UserCode        string `json:"userCode,omitempty"`
	ExpiresAt       int64  `json:"expiresAt,omitempty"`
	Status          string `json:"status"`
	Error           string `json:"error,omitempty"`
	AccountID       string `json:"accountId,omitempty"`
	Email           string `json:"email,omitempty"`
	AuthMethod      string `json:"authMethod,omitempty"`
	Provider        string `json:"provider,omitempty"`
}

type SocialCallbackResult struct {
	Code  string
	State string
	Error string
}

type ClientRegistrationResponse struct {
	ClientID              string `json:"clientId"`
	ClientSecret          string `json:"clientSecret"`
	ClientSecretExpiresAt int64  `json:"clientSecretExpiresAt"`
}

type DeviceAuthorizationResponse struct {
	DeviceCode              string `json:"deviceCode"`
	UserCode                string `json:"userCode"`
	VerificationURI         string `json:"verificationUri"`
	VerificationURIComplete string `json:"verificationUriComplete"`
	ExpiresIn               int    `json:"expiresIn"`
	Interval                int    `json:"interval"`
}

type DeviceTokenResponse struct {
	AccessToken      string `json:"accessToken"`
	RefreshToken     string `json:"refreshToken"`
	ExpiresIn        int    `json:"expiresIn"`
	TokenType        string `json:"tokenType"`
	ProfileARN       string `json:"profileArn"`
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

type TokenData struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int
	TokenType    string
	ProfileARN   string
	Email        string
	ClientID     string
	ClientSecret string
}

type SocialTokenResponse struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	ProfileARN   string `json:"profileArn"`
	ExpiresIn    int    `json:"expiresIn"`
}
