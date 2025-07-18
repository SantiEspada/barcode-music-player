package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	SpotifyClientID     string
	SpotifyClientSecret string
	SpotifyRedirectURI  string
	MusicBrainzURL      string
}

func Load() (*Config, error) {
	// Load .env file if it exists (ignore error if file doesn't exist)
	_ = godotenv.Load()

	config := &Config{
		SpotifyClientID:     os.Getenv("SPOTIFY_CLIENT_ID"),
		SpotifyClientSecret: os.Getenv("SPOTIFY_CLIENT_SECRET"),
		SpotifyRedirectURI:  getEnvOrDefault("SPOTIFY_REDIRECT_URI", "http://127.0.0.1:8080/callback"),
		MusicBrainzURL:      getEnvOrDefault("MUSICBRAINZ_URL", "https://musicbrainz.org/ws/2"),
	}

	// Validate required configuration
	if config.SpotifyClientID == "" {
		return nil, fmt.Errorf("SPOTIFY_CLIENT_ID environment variable is required")
	}

	if config.SpotifyClientSecret == "" {
		return nil, fmt.Errorf("SPOTIFY_CLIENT_SECRET environment variable is required")
	}

	return config, nil
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
