package musicbrainz

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Client struct {
	BaseURL    string
	HTTPClient *http.Client
}

type ReleaseGroup struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	Type  string `json:"primary-type"`
}

type Artist struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Release struct {
	ID           string       `json:"id"`
	Title        string       `json:"title"`
	Artists      []Artist     `json:"artist-credit"`
	ReleaseGroup ReleaseGroup `json:"release-group"`
	Date         string       `json:"date"`
	Barcode      string       `json:"barcode"`
}

type SearchResponse struct {
	Releases []Release `json:"releases"`
	Count    int       `json:"count"`
}

func NewClient(baseURL string) *Client {
	return &Client{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *Client) SearchByBarcode(barcode string) (*Release, error) {
	// MusicBrainz API endpoint for release search by barcode
	endpoint := fmt.Sprintf("%s/release", c.BaseURL)

	// Build query parameters
	params := url.Values{}
	params.Add("query", fmt.Sprintf("barcode:%s", barcode))
	params.Add("fmt", "json")
	params.Add("inc", "artists+release-groups")

	// Create request
	req, err := http.NewRequest("GET", endpoint+"?"+params.Encode(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set User-Agent header (required by MusicBrainz)
	req.Header.Set("User-Agent", "barcode-music-player/1.0 (https://github.com/user/barcode-music-player)")

	// Make request
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status: %d", resp.StatusCode)
	}

	// Parse response
	var searchResp SearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&searchResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Check if we found any releases
	if len(searchResp.Releases) == 0 {
		return nil, fmt.Errorf("no releases found for barcode: %s", barcode)
	}

	// Return the first release (most relevant)
	return &searchResp.Releases[0], nil
}

func (r *Release) GetMainArtist() string {
	if len(r.Artists) > 0 {
		return r.Artists[0].Name
	}
	return "Unknown Artist"
}

func (r *Release) GetSearchQuery() string {
	// Create a more flexible search query
	artist := r.GetMainArtist()
	title := r.Title

	// Clean up common issues in titles and artist names
	title = strings.ReplaceAll(title, " (disc 1)", "")
	title = strings.ReplaceAll(title, " (disc 2)", "")
	title = strings.ReplaceAll(title, " [disc 1]", "")
	title = strings.ReplaceAll(title, " [disc 2]", "")

	// Remove common suffixes that might cause issues
	title = strings.ReplaceAll(title, " (bonus track version)", "")
	title = strings.ReplaceAll(title, " (deluxe edition)", "")
	title = strings.ReplaceAll(title, " (remastered)", "")
	title = strings.ReplaceAll(title, " (expanded edition)", "")

	// Return a simple search query without field specifiers
	return fmt.Sprintf("%s %s", title, artist)
}
