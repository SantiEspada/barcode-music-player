package spotify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Client struct {
	ClientID     string
	ClientSecret string
	RedirectURI  string
	AccessToken  string
	RefreshToken string
	ExpiresAt    time.Time
	HTTPClient   *http.Client
}

type Album struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	URI    string `json:"uri"`
	Images []struct {
		URL    string `json:"url"`
		Height int    `json:"height"`
		Width  int    `json:"width"`
	} `json:"images"`
	Artists []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"artists"`
}

type Device struct {
	ID            string `json:"id"`
	IsActive      bool   `json:"is_active"`
	IsRestricted  bool   `json:"is_restricted"`
	Name          string `json:"name"`
	Type          string `json:"type"`
	VolumePercent int    `json:"volume_percent"`
}

type DevicesResponse struct {
	Devices []Device `json:"devices"`
}

type SearchResponse struct {
	Albums struct {
		Items []Album `json:"items"`
		Total int     `json:"total"`
	} `json:"albums"`
}

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	Scope        string `json:"scope"`
}

type StoredToken struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
}

func NewClient(clientID, clientSecret, redirectURI string) *Client {
	return &Client{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURI:  redirectURI,
		HTTPClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *Client) getTokenFilePath() string {
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, ".barcode-music-player-token.json")
}

func (c *Client) LoadStoredToken() bool {
	tokenFile := c.getTokenFilePath()

	data, err := os.ReadFile(tokenFile)
	if err != nil {
		return false
	}

	var stored StoredToken
	if err := json.Unmarshal(data, &stored); err != nil {
		return false
	}

	// Check if token is still valid (with 5 minute buffer)
	if time.Now().Add(5 * time.Minute).After(stored.ExpiresAt) {
		return false
	}

	c.AccessToken = stored.AccessToken
	c.RefreshToken = stored.RefreshToken
	c.ExpiresAt = stored.ExpiresAt

	return true
}

func (c *Client) SaveToken() error {
	tokenFile := c.getTokenFilePath()

	stored := StoredToken{
		AccessToken:  c.AccessToken,
		RefreshToken: c.RefreshToken,
		ExpiresAt:    c.ExpiresAt,
	}

	data, err := json.MarshalIndent(stored, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(tokenFile, data, 0600)
}

func (c *Client) GetAuthURL() string {
	params := url.Values{}
	params.Add("client_id", c.ClientID)
	params.Add("response_type", "code")
	params.Add("redirect_uri", c.RedirectURI)
	params.Add("scope", "user-read-playback-state user-modify-playback-state")

	return "https://accounts.spotify.com/authorize?" + params.Encode()
}

func (c *Client) ExchangeCodeForToken(code string) error {
	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)
	data.Set("redirect_uri", c.RedirectURI)

	req, err := http.NewRequest("POST", "https://accounts.spotify.com/api/token", strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create token request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(c.ClientID, c.ClientSecret)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to exchange code for token: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("token exchange failed with status: %d", resp.StatusCode)
	}

	var tokenResp TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return fmt.Errorf("failed to decode token response: %w", err)
	}

	c.AccessToken = tokenResp.AccessToken
	c.RefreshToken = tokenResp.RefreshToken
	c.ExpiresAt = time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)

	// Save token for future use
	if err := c.SaveToken(); err != nil {
		fmt.Printf("Warning: Failed to save token: %v\n", err)
	}

	return nil
}

func (c *Client) GetAvailableDevices() ([]Device, error) {
	if c.AccessToken == "" {
		return nil, fmt.Errorf("not authenticated - access token required")
	}

	req, err := http.NewRequest("GET", "https://api.spotify.com/v1/me/player/devices", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create devices request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.AccessToken)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get devices: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("devices request failed with status: %d", resp.StatusCode)
	}

	var devicesResp DevicesResponse
	if err := json.NewDecoder(resp.Body).Decode(&devicesResp); err != nil {
		return nil, fmt.Errorf("failed to decode devices response: %w", err)
	}

	return devicesResp.Devices, nil
}

func (c *Client) SetShuffle(state bool) error {
	if c.AccessToken == "" {
		return fmt.Errorf("not authenticated - access token required")
	}

	params := url.Values{}
	params.Add("state", fmt.Sprintf("%t", state))

	req, err := http.NewRequest("PUT", "https://api.spotify.com/v1/me/player/shuffle?"+params.Encode(), nil)
	if err != nil {
		return fmt.Errorf("failed to create shuffle request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.AccessToken)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to set shuffle: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("shuffle request failed with status: %d", resp.StatusCode)
	}

	return nil
}

func (c *Client) SearchAlbums(query string) ([]Album, error) {
	if c.AccessToken == "" {
		return nil, fmt.Errorf("not authenticated - access token required")
	}

	// Try multiple search strategies
	searchStrategies := []string{
		query,                                   // Original query
		strings.ReplaceAll(query, "album:", ""), // Remove album: prefix
		strings.ReplaceAll(strings.ReplaceAll(query, "album:", ""), "artist:", ""), // Remove both prefixes
	}

	for i, searchQuery := range searchStrategies {
		fmt.Printf("🔍 Search attempt %d: %s\n", i+1, searchQuery)

		albums, err := c.performSearch(searchQuery)
		if err != nil {
			fmt.Printf("   ❌ Search failed: %v\n", err)
			continue
		}

		if len(albums) > 0 {
			fmt.Printf("   ✅ Found %d albums\n", len(albums))
			return albums, nil
		}

		fmt.Printf("   ⚠️  No results\n")
	}

	return nil, fmt.Errorf("no albums found after trying multiple search strategies")
}

func (c *Client) performSearch(query string) ([]Album, error) {
	params := url.Values{}
	params.Add("q", query)
	params.Add("type", "album")
	params.Add("limit", "10")

	req, err := http.NewRequest("GET", "https://api.spotify.com/v1/search?"+params.Encode(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create search request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.AccessToken)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to search albums: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("search request failed with status: %d", resp.StatusCode)
	}

	var searchResp SearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&searchResp); err != nil {
		return nil, fmt.Errorf("failed to decode search response: %w", err)
	}

	return searchResp.Albums.Items, nil
}

func (c *Client) PlayAlbum(albumURI string) error {
	if c.AccessToken == "" {
		return fmt.Errorf("not authenticated - access token required")
	}

	// First, check for available devices
	fmt.Println("🔍 Checking for available Spotify devices...")
	devices, err := c.GetAvailableDevices()
	if err != nil {
		return fmt.Errorf("failed to get available devices: %w", err)
	}

	if len(devices) == 0 {
		return fmt.Errorf("no Spotify devices found. Please:\n" +
			"1. Open Spotify on your computer, phone, or web browser\n" +
			"2. Start playing any song to activate the device\n" +
			"3. Try scanning the barcode again")
	}

	// Find an active device or use the first available one
	var activeDevice *Device
	for _, device := range devices {
		if device.IsActive {
			activeDevice = &device
			break
		}
	}

	if activeDevice == nil {
		activeDevice = &devices[0]
		fmt.Printf("🔄 No active device found, using: %s (%s)\n", activeDevice.Name, activeDevice.Type)
	} else {
		fmt.Printf("🎵 Using active device: %s (%s)\n", activeDevice.Name, activeDevice.Type)
	}

	// Disable shuffle to ensure album plays in order
	fmt.Println("🔀 Disabling shuffle to play album in order...")
	if err := c.SetShuffle(false); err != nil {
		// Don't fail if shuffle can't be disabled, just warn
		fmt.Printf("⚠️  Warning: Could not disable shuffle: %v\n", err)
	}

	playData := map[string]interface{}{
		"context_uri": albumURI,
		"offset": map[string]interface{}{
			"position": 0, // Start from the first track
		},
	}

	// If we're using a non-active device, specify it
	if !activeDevice.IsActive {
		playData["device_id"] = activeDevice.ID
	}

	jsonData, err := json.Marshal(playData)
	if err != nil {
		return fmt.Errorf("failed to marshal play data: %w", err)
	}

	req, err := http.NewRequest("PUT", "https://api.spotify.com/v1/me/player/play", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create play request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.AccessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to play album: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("device not found or not available. Please:\n" +
			"1. Make sure Spotify is running on your device\n" +
			"2. Play any song to activate the device\n" +
			"3. Ensure you have Spotify Premium (required for playback control)")
	}

	if resp.StatusCode == http.StatusForbidden {
		return fmt.Errorf("playback failed - you need Spotify Premium to control playback remotely")
	}

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("play request failed with status: %d", resp.StatusCode)
	}

	return nil
}

func (a *Album) GetMainArtist() string {
	if len(a.Artists) > 0 {
		return a.Artists[0].Name
	}
	return "Unknown Artist"
}
