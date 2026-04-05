package config

import (
	"sort"
	"strings"
	"sync"
)

type Manager struct {
	mu      sync.RWMutex
	storage *Storage

	settings        AppSettings
	accounts        []Account
	stats           ProxyStats
	startupWarnings []StartupWarning
}

func NewManager(dataDir string) (*Manager, error) {
	storage, err := NewStorage(dataDir)
	if err != nil {
		return nil, err
	}

	m := &Manager{storage: storage}
	if err := m.load(); err != nil {
		return nil, err
	}
	return m, nil
}

func (m *Manager) load() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	settings, warnings, err := m.storage.LoadSettings()
	if err != nil {
		return err
	}
	m.startupWarnings = append(m.startupWarnings, warnings...)

	if settings.ProxyPort == 0 {
		settings.ProxyPort = defaultProxyPort
	}
	settings.SchedulingMode = string(normalizeSchedulingMode(SchedulingMode(settings.SchedulingMode)))
	settings.Cloudflared = normalizeCloudflaredSettings(settings.Cloudflared)
	m.settings = settings

	accounts, warnings, err := m.storage.LoadAccounts()
	if err != nil {
		return err
	}
	m.startupWarnings = append(m.startupWarnings, warnings...)
	if accounts == nil {
		accounts = []Account{}
	}
	sort.Slice(accounts, func(i, j int) bool { return accounts[i].CreatedAt > accounts[j].CreatedAt })
	m.accounts = accounts

	stats, warnings, err := m.storage.LoadStats()
	if err != nil {
		return err
	}
	m.startupWarnings = append(m.startupWarnings, warnings...)
	m.stats = stats

	return nil
}

func (m *Manager) saveSettings() error {
	return m.storage.SaveSettings(m.settings)
}

func (m *Manager) saveAccounts() error {
	return m.storage.SaveAccounts(m.accounts)
}

func (m *Manager) saveStats() error {
	return m.storage.SaveStats(m.stats)
}

func (m *Manager) Snapshot() AppConfig {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return AppConfig{
		ProxyPort:         m.settings.ProxyPort,
		AllowLAN:          m.settings.AllowLAN,
		AutoStartProxy:    m.settings.AutoStartProxy,
		ProxyAPIKey:       m.settings.ProxyAPIKey,
		AuthorizationMode: m.settings.AuthorizationMode,
		SchedulingMode:    normalizeSchedulingMode(SchedulingMode(m.settings.SchedulingMode)),
		Cloudflared:       normalizeCloudflaredSettings(m.settings.Cloudflared),
		Thinking:          cloneThinkingSettings(m.settings.Thinking),
		ModelAliases:      cloneModelAliases(m.settings.ModelAliases),
		Accounts:          cloneAccounts(m.accounts),
		Stats:             m.stats,
		StartupWarnings:   cloneStartupWarnings(m.startupWarnings),
	}
}

func (m *Manager) StartupWarnings() []StartupWarning {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return cloneStartupWarnings(m.startupWarnings)
}

func (m *Manager) Accounts() []Account {
	m.mu.RLock()
	defer m.mu.RUnlock()
	accounts := cloneAccounts(m.accounts)
	sort.Slice(accounts, func(i, j int) bool { return accounts[i].CreatedAt > accounts[j].CreatedAt })
	return accounts
}

func (m *Manager) ProxyPort() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.settings.ProxyPort == 0 {
		return defaultProxyPort
	}
	return m.settings.ProxyPort
}

func (m *Manager) SetProxyPort(port int) error {
	if port <= 0 {
		port = defaultProxyPort
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.settings.ProxyPort = port
	return m.saveSettings()
}

func (m *Manager) AllowLAN() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.settings.AllowLAN
}

func (m *Manager) SetAllowLAN(enabled bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.settings.AllowLAN = enabled
	return m.saveSettings()
}

func (m *Manager) AutoStartProxy() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.settings.AutoStartProxy
}

func (m *Manager) SetAutoStartProxy(enabled bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.settings.AutoStartProxy = enabled
	return m.saveSettings()
}

func (m *Manager) ProxyAPIKey() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return strings.TrimSpace(m.settings.ProxyAPIKey)
}

func (m *Manager) SetProxyAPIKey(apiKey string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.settings.ProxyAPIKey = strings.TrimSpace(apiKey)
	return m.saveSettings()
}

func (m *Manager) AuthorizationMode() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.settings.AuthorizationMode
}

func (m *Manager) SetAuthorizationMode(enabled bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.settings.AuthorizationMode = enabled
	return m.saveSettings()
}

func (m *Manager) SchedulingMode() SchedulingMode {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return normalizeSchedulingMode(SchedulingMode(m.settings.SchedulingMode))
}

func (m *Manager) SetSchedulingMode(mode SchedulingMode) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.settings.SchedulingMode = string(normalizeSchedulingMode(mode))
	return m.saveSettings()
}

func (m *Manager) Cloudflared() CloudflaredSettings {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return normalizeCloudflaredSettings(m.settings.Cloudflared)
}

func (m *Manager) SetCloudflaredConfig(mode CloudflaredMode, token string, useHTTP2 bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	next := normalizeCloudflaredSettings(m.settings.Cloudflared)
	next.Mode = normalizeCloudflaredMode(mode)
	next.Token = strings.TrimSpace(token)
	next.UseHTTP2 = useHTTP2
	m.settings.Cloudflared = next
	return m.saveSettings()
}

func (m *Manager) SetCloudflaredEnabled(enabled bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	next := normalizeCloudflaredSettings(m.settings.Cloudflared)
	next.Enabled = enabled
	m.settings.Cloudflared = next
	return m.saveSettings()
}

func (m *Manager) ModelAliases() map[string]string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return cloneModelAliases(m.settings.ModelAliases)
}

func (m *Manager) SetModelAliases(aliases map[string]string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.settings.ModelAliases = cloneModelAliases(aliases)
	return m.saveSettings()
}
