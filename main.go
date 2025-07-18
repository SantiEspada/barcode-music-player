package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"barcode-music-player/auth"
	"barcode-music-player/config"
	"barcode-music-player/musicbrainz"
	"barcode-music-player/spotify"
)

var (
	spotifyClient     *spotify.Client
	musicbrainzClient *musicbrainz.Client
)

func main() {
	fmt.Println("🎵 Barcode Music Player")
	fmt.Println("=====================")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Configuration error:", err)
	}

	// Initialize clients
	spotifyClient = spotify.NewClient(cfg.SpotifyClientID, cfg.SpotifyClientSecret, cfg.SpotifyRedirectURI)
	musicbrainzClient = musicbrainz.NewClient(cfg.MusicBrainzURL)

	// Authenticate with Spotify
	fmt.Println("🔐 Authenticating with Spotify...")
	if err := authenticateSpotify(); err != nil {
		log.Fatal("Authentication failed:", err)
	}

	fmt.Println("✅ Successfully authenticated with Spotify!")
	fmt.Println()
	fmt.Println("Ready to scan barcodes! 🎵")
	fmt.Println("Scan a barcode to play the album on Spotify!")
	fmt.Println("Press Ctrl+C to exit")
	fmt.Println()

	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("Scan barcode (or type 'quit' to exit): ")

		if !scanner.Scan() {
			break
		}

		barcode := strings.TrimSpace(scanner.Text())

		if barcode == "quit" {
			fmt.Println("Goodbye! 👋")
			break
		}

		if barcode == "" {
			continue
		}

		fmt.Printf("🔍 Processing barcode: %s\n", barcode)

		if err := processBarcode(barcode); err != nil {
			fmt.Printf("❌ Error: %v\n", err)
		}

		fmt.Println()
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}

func authenticateSpotify() error {
	// First, try to load a stored token
	if spotifyClient.LoadStoredToken() {
		fmt.Println("✅ Using stored authentication token")
		return nil
	}

	fmt.Println("🔐 No stored token found, starting OAuth flow...")

	// Create OAuth handler
	oauthHandler := auth.NewOAuthHandler(spotifyClient.RedirectURI)

	// Start local server
	if err := oauthHandler.StartServer("8080"); err != nil {
		return fmt.Errorf("failed to start OAuth server: %w", err)
	}

	// Get authorization URL and open browser
	authURL := spotifyClient.GetAuthURL()
	fmt.Printf("Opening browser for authorization: %s\n", authURL)

	if err := auth.OpenBrowser(authURL); err != nil {
		fmt.Printf("Failed to open browser automatically. Please open this URL manually:\n%s\n", authURL)
	}

	// Wait for authorization code
	code, err := oauthHandler.WaitForCode(2 * time.Minute)
	if err != nil {
		return fmt.Errorf("failed to get authorization code: %w", err)
	}

	// Exchange code for access token
	if err := spotifyClient.ExchangeCodeForToken(code); err != nil {
		return fmt.Errorf("failed to exchange code for token: %w", err)
	}

	return nil
}

func processBarcode(barcode string) error {
	// Step 1: Look up album in MusicBrainz
	fmt.Println("🔍 Looking up album in MusicBrainz...")
	release, err := musicbrainzClient.SearchByBarcode(barcode)
	if err != nil {
		return fmt.Errorf("failed to find album for barcode %s: %w", barcode, err)
	}

	fmt.Printf("📀 Found album: \"%s\" by %s\n", release.Title, release.GetMainArtist())

	// Step 2: Search for album on Spotify
	fmt.Println("🎵 Searching for album on Spotify...")
	searchQuery := release.GetSearchQuery()
	fmt.Printf("🔍 Search query: %s\n", searchQuery)
	albums, err := spotifyClient.SearchAlbums(searchQuery)
	if err != nil {
		return fmt.Errorf("failed to search Spotify: %w", err)
	}

	if len(albums) == 0 {
		return fmt.Errorf("no albums found on Spotify for: %s", searchQuery)
	}

	// Use the first (most relevant) result
	album := albums[0]
	fmt.Printf("🎯 Found on Spotify: \"%s\" by %s\n", album.Name, album.GetMainArtist())

	// Step 3: Play the album
	fmt.Println("▶️  Playing album...")
	if err := spotifyClient.PlayAlbum(album.URI); err != nil {
		return fmt.Errorf("failed to play album: %w", err)
	}

	fmt.Printf("🎉 Successfully playing: \"%s\" by %s\n", album.Name, album.GetMainArtist())
	return nil
}
