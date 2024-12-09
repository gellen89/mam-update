package mamupdater

import (
	"context"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"strings"
	"time"
)

type Config struct {
	DataDir     string
	CookiePath  string
	IpPath      string
	LastRunPath string
	MamId       *string
	Force       bool
	Logger      *slog.Logger
}

type MamUpdater struct {
	config     *Config
	httpClient *http.Client
	logger     *slog.Logger
	updatUrl   *url.URL
}

type dynamicSeedboxResponse struct {
	Success bool   `json:"Success"`
	Msg     string `json:"msg"`
}

const (
	ipUrl         = "https://ifconfig.io/ip"
	dynSeedBoxUrl = "https://t.myanonamouse.net/json/dynamicSeedbox.php"
	minWaitPeriod = time.Hour
)

func NewMamUpdater(config *Config) (*MamUpdater, error) {
	jar, err := cookiejar.New(&cookiejar.Options{})
	if err != nil {
		return nil, fmt.Errorf("failed to create cookie jar: %w", err)
	}
	u, err := url.Parse(dynSeedBoxUrl)
	if err != nil {
		return nil, err
	}

	return &MamUpdater{
		config:     config,
		httpClient: &http.Client{Jar: jar, Timeout: 10 * time.Second},
		logger:     config.Logger,
		updatUrl:   u,
	}, nil
}

func (m *MamUpdater) Run(ctx context.Context) error {
	// Check for cookie file or MAM_ID
	hasCookie := m.hasCookieFile()
	if !hasCookie && m.config.MamId == nil || *m.config.MamId == "" {
		return fmt.Errorf("no cookie file found and MAM_ID not provided")
	}

	m.logger.Debug("retrieving ip address...")
	// Get current IP
	currentIP, err := m.getCurrentIP(ctx)
	if err != nil {
		return fmt.Errorf("failed to get current IP: %w", err)
	}

	// First run without cookie
	if !hasCookie {
		m.logger.Debug("no cookie found, handling first run")
		return m.handleFirstRun(ctx, *m.config.MamId)
	}

	// Check if IP has changed
	if changed, err := m.hasIPChanged(currentIP); err != nil {
		return fmt.Errorf("failed to check IP change: %w", err)
	} else if !changed {
		m.logger.Info("IP address hasn't changed, skipping update")
		return nil
	}

	m.logger.Debug("ip address changed, checking if should should update")

	// Check last run time
	if shouldSkip, err := m.shouldSkipUpdate(); err != nil {
		return fmt.Errorf("failed to check last run time: %w", err)
	} else if shouldSkip {
		m.logger.Info("Last run was too recent, skipping update")
		return nil
	}

	m.logger.Debug("handling ip address update")

	// Load existing cookies
	if err := m.loadCookies(); err != nil {
		return fmt.Errorf("failed to load cookies: %w", err)
	}
	m.logger.Debug("cookies loaded")
	// Update IP
	if err := m.updateIP(ctx); err != nil {
		return fmt.Errorf("failed to update ip address: %w", err)
	}
	m.logger.Debug("ip address updated successfully")
	// Write the current ip address to disk
	if err := m.writeFile(m.config.IpPath, currentIP); err != nil {
		return fmt.Errorf("failed to write new ip address: %w", err)
	}
	m.logger.Debug("successfully wrote new ip address to disk")
	return nil
}

func (m *MamUpdater) getCurrentIP(ctx context.Context) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, ipUrl, http.NoBody)
	if err != nil {
		return "", err
	}
	resp, err := m.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	ip, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(ip)), nil
}

func (m *MamUpdater) hasIPChanged(currentIP string) (bool, error) {
	oldIP, exists, err := m.readFile(m.config.IpPath)
	if err != nil {
		return false, err
	}
	return !exists || oldIP != currentIP, nil
}

func (m *MamUpdater) shouldSkipUpdate() (bool, error) {
	lastRunStr, exists, err := m.readFile(m.config.LastRunPath)
	if err != nil {
		return false, err
	}
	if !exists || m.config.Force {
		return false, nil
	}

	lastRun, err := time.Parse(time.RFC3339, lastRunStr)
	if err != nil {
		return false, err
	}

	return time.Since(lastRun) < minWaitPeriod, nil
}

func (m *MamUpdater) updateIP(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, dynSeedBoxUrl, http.NoBody)
	if err != nil {
		return err
	}
	resp, err := m.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	// Parse response
	var result dynamicSeedboxResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	if !result.Success {
		return fmt.Errorf("IP update failed: %s", result.Msg)
	}

	// Save cookies and update timestamp
	if err := m.saveCookies(resp.Cookies()); err != nil {
		return fmt.Errorf("failed to save cookies: %w", err)
	}

	if err := m.writeFile(m.config.LastRunPath, time.Now().Format(time.RFC3339)); err != nil {
		return fmt.Errorf("failed to update last run time: %w", err)
	}

	m.logger.Info("IP address updated successfully", "message", result.Msg)
	return nil
}

func (m *MamUpdater) handleFirstRun(ctx context.Context, mamID string) error {
	// Create initial cookie with MAM_ID
	initialCookies := []*http.Cookie{
		{
			Name:  "mam_id",
			Value: mamID,
		},
	}
	m.httpClient.Jar.SetCookies(m.updatUrl, initialCookies)

	if err := m.updateIP(ctx); err != nil {
		return fmt.Errorf("first run failed: %w", err)
	}
	return nil
}

func (m *MamUpdater) hasCookieFile() bool {
	_, err := os.Stat(m.config.CookiePath)
	return !os.IsNotExist(err)
}

func (m *MamUpdater) loadCookies() error {
	file, err := os.Open(m.config.CookiePath)
	if err != nil {
		return err
	}
	defer file.Close()

	var cookies []*http.Cookie
	if err := gob.NewDecoder(file).Decode(&cookies); err != nil {
		return err
	}

	m.httpClient.Jar.SetCookies(m.updatUrl, cookies)
	return nil
}

func (m *MamUpdater) saveCookies(cookies []*http.Cookie) error {
	file, err := os.Create(m.config.CookiePath)
	if err != nil {
		return err
	}
	defer file.Close()

	return gob.NewEncoder(file).Encode(cookies)
}

func (m *MamUpdater) readFile(path string) (contents string, exists bool, err error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	return string(data), true, nil
}

func (m *MamUpdater) writeFile(path, content string) error {
	return os.WriteFile(path, []byte(content), 0600)
}
