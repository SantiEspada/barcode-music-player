# 🎵 Barcode Music Player

A Go terminal application that plays your physical albums on Spotify using barcode scanning. Simply scan the barcode on your CD/vinyl and the app will automatically find and play the album on Spotify!

## Features

- 🔍 **Barcode Recognition**: Uses any barcode scanner that works as a keyboard input
- 📀 **Album Lookup**: Searches MusicBrainz database to identify albums by barcode
- 🎵 **Spotify Integration**: Automatically searches and plays albums on Spotify
- 🔐 **OAuth Authentication**: Secure authentication with Spotify Web API
- 📱 **Device Detection**: Automatically finds and uses available Spotify devices
- 🔀 **Shuffle Control**: Automatically disables shuffle to play albums in track order
- 🚀 **Fast & Lightweight**: Terminal-based application with minimal dependencies

## Prerequisites

- Go 1.19 or higher
- Spotify Premium account (required for playback control)
- Barcode scanner (or manual barcode entry)
- Active Spotify session (desktop app, web player, or mobile app)

## Setup

### 1. Get Spotify API Credentials

1. Go to [Spotify Developer Dashboard](https://developer.spotify.com/dashboard/applications)
2. Create a new app
3. Note down your **Client ID** and **Client Secret**
4. Add `http://127.0.0.1:8080/callback` to your app's redirect URIs

### 2. Configure Environment Variables

1. Copy the example environment file:

   ```bash
   cp .env.example .env
   ```

2. Edit `.env` and add your Spotify credentials:
   ```bash
   SPOTIFY_CLIENT_ID=your_actual_client_id
   SPOTIFY_CLIENT_SECRET=your_actual_client_secret
   ```

Note: The application automatically loads the `.env` file if it exists.

### 3. Build and Run

```bash
# Build the application
go build -o barcode-music-player

# Run the application
./barcode-music-player
```

Or run directly:

```bash
go run main.go
```

## Usage

1. **Start the application** - It will automatically open your browser for Spotify authentication (first time only)
2. **Authorize the app** - Grant the necessary permissions in your browser
3. **Scan barcodes** - Use your barcode scanner to scan CD/vinyl barcodes
4. **Enjoy your music** - The app will automatically find and play the album on Spotify

### Manual Barcode Entry

If you don't have a barcode scanner, you can manually type the barcode numbers (UPC/EAN codes) found on your albums.

## How It Works

1. **Barcode Input**: Reads barcode from stdin (works with any USB barcode scanner)
2. **Album Lookup**: Queries MusicBrainz API to find album information by barcode
3. **Spotify Search**: Searches Spotify for the identified album using multiple search strategies
4. **Device Detection**: Automatically finds available Spotify devices
5. **Shuffle Control**: Disables shuffle and starts from track 1 for proper album experience
6. **Playback**: Plays the album on your selected Spotify device

## Troubleshooting

### "No Spotify devices found"

- **Open Spotify** on your computer, phone, or web browser
- **Start playing any song** to activate the device
- The app will automatically detect and use available devices

### "Device not found or not available"

- Make sure Spotify is running and actively playing music
- Ensure you have **Spotify Premium** (required for remote playback control)
- Try playing a song manually in Spotify first

### "Album not found"

- Some older or regional releases might not be in MusicBrainz database
- Try searching manually in Spotify to see if the album exists
- Independent releases might not have barcode data

### "Authentication failed"

- Check that your Spotify credentials are correct in the `.env` file
- Ensure redirect URI is set to `http://127.0.0.1:8080/callback`
- Make sure no other application is using port 8080

### "Playback failed - you need Spotify Premium"

- This app requires **Spotify Premium** to control playback remotely
- Free Spotify accounts cannot use the Web API for playback control

## API Rate Limits

- **MusicBrainz**: 1 request per second (automatically respected)
- **Spotify**: 100 requests per minute (rarely reached in normal usage)

## Dependencies

This project uses minimal external dependencies:

- `github.com/joho/godotenv` - For loading .env files
- Go standard library packages:
  - `net/http` - HTTP client and server
  - `encoding/json` - JSON parsing
  - `os/exec` - Browser launching

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Test thoroughly
5. Submit a pull request

## License

This project is open source and available under the MIT License.

## Disclaimer

This application is not affiliated with Spotify. It uses the official Spotify Web API under their terms of service. Make sure you comply with Spotify's API usage policies.
