package auth

import (
	"context"
	"fmt"
	"net/http"
	"os/exec"
	"runtime"
	"time"
)

type OAuthHandler struct {
	server *http.Server
	code   string
	err    error
	done   chan bool
}

func NewOAuthHandler(redirectURI string) *OAuthHandler {
	return &OAuthHandler{
		done: make(chan bool),
	}
}

func (h *OAuthHandler) StartServer(port string) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/callback", h.handleCallback)
	mux.HandleFunc("/", h.handleRoot)

	h.server = &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	go func() {
		if err := h.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			h.err = err
			h.done <- true
		}
	}()

	return nil
}

func (h *OAuthHandler) handleCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	errorParam := r.URL.Query().Get("error")

	if errorParam != "" {
		h.err = fmt.Errorf("authorization error: %s", errorParam)
		fmt.Fprintf(w, `
			<html>
			<body>
				<h1>Authorization Error</h1>
				<p>Error: %s</p>
				<p>You can close this window.</p>
			</body>
			</html>
		`, errorParam)
		h.done <- true
		return
	}

	if code == "" {
		h.err = fmt.Errorf("no authorization code received")
		fmt.Fprintf(w, `
			<html>
			<body>
				<h1>Authorization Error</h1>
				<p>No authorization code received.</p>
				<p>You can close this window.</p>
			</body>
			</html>
		`)
		h.done <- true
		return
	}

	h.code = code
	fmt.Fprintf(w, `
		<html>
		<body>
			<h1>🎵 Authorization Successful!</h1>
			<p>You have successfully authorized the Barcode Music Player.</p>
			<p>You can close this window and return to the terminal.</p>
		</body>
		</html>
	`)
	h.done <- true
}

func (h *OAuthHandler) handleRoot(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, `
		<html>
		<body>
			<h1>🎵 Barcode Music Player</h1>
			<p>Waiting for authorization...</p>
		</body>
		</html>
	`)
}

func (h *OAuthHandler) WaitForCode(timeout time.Duration) (string, error) {
	select {
	case <-h.done:
		h.shutdown()
		if h.err != nil {
			return "", h.err
		}
		return h.code, nil
	case <-time.After(timeout):
		h.shutdown()
		return "", fmt.Errorf("authorization timeout after %v", timeout)
	}
}

func (h *OAuthHandler) shutdown() {
	if h.server != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		h.server.Shutdown(ctx)
	}
}

func OpenBrowser(url string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start"}
	case "darwin":
		cmd = "open"
	default: // "linux", "freebsd", "openbsd", "netbsd"
		cmd = "xdg-open"
	}

	args = append(args, url)
	return exec.Command(cmd, args...).Start()
}
