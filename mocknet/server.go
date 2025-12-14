package mocknet

import (
	"context"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
	"path"
	"strconv"
	"strings"
	"time"
)

//go:embed web/templates/*.html web/static/* web/seed/*.json
var webFS embed.FS

type Config struct {
	Port int

	// When enabled, successful login routes users to /checkpoint first.
	CheckpointEnabled bool

	SeedCredentials Credentials
}

type Credentials struct {
	Username string
	Password string
}

type Server struct {
	cfg   Config
	state *State
	tpl   *template.Template
}

func New(cfg Config) (*Server, error) {
	if cfg.Port == 0 {
		cfg.Port = 8080
	}
	if strings.TrimSpace(cfg.SeedCredentials.Username) == "" {
		cfg.SeedCredentials.Username = "demo"
	}
	if strings.TrimSpace(cfg.SeedCredentials.Password) == "" {
		cfg.SeedCredentials.Password = "demo"
	}

	profiles, err := loadSeedProfiles()
	if err != nil {
		return nil, err
	}
	st := NewState(profiles)

	tpl, err := template.ParseFS(webFS, "web/templates/*.html")
	if err != nil {
		return nil, fmt.Errorf("mocknet: parse templates: %w", err)
	}

	return &Server{cfg: cfg, state: st, tpl: tpl}, nil
}

func (s *Server) Addr() string {
	return fmt.Sprintf("127.0.0.1:%d", s.cfg.Port)
}

func (s *Server) BaseURL() string {
	return fmt.Sprintf("http://localhost:%d", s.cfg.Port)
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /", s.handleHome)
	mux.HandleFunc("GET /login", s.handleLoginGet)
	mux.HandleFunc("POST /login", s.handleLoginPost)
	mux.HandleFunc("GET /logout", s.handleLogout)

	mux.HandleFunc("GET /checkpoint", s.handleCheckpointGet)
	mux.HandleFunc("POST /checkpoint", s.handleCheckpointPost)

	mux.HandleFunc("GET /search", s.requireAuth(s.handleSearchGet))
	mux.HandleFunc("GET /profile/{id}", s.requireAuth(s.handleProfileGet))
	mux.HandleFunc("POST /profile/{id}/connect", s.requireAuth(s.handleConnectPost))

	mux.HandleFunc("GET /connections", s.requireAuth(s.handleConnectionsGet))

	mux.HandleFunc("GET /messages", s.requireAuth(s.handleMessagesGet))
	mux.HandleFunc("POST /messages/send", s.requireAuth(s.handleMessagesSendPost))

	// Admin toggle: simulate acceptance of pending connection.
	mux.HandleFunc("POST /admin/accept", s.requireAuth(s.handleAdminAcceptPost))

	staticFS, _ := fs.Sub(webFS, "web/static")
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServerFS(staticFS)))

	return withSecurityHeaders(mux)
}

func (s *Server) Run(ctx context.Context) error {
	h := s.Handler()
	srv := &http.Server{
		Addr:              s.Addr(),
		Handler:           h,
		ReadHeaderTimeout: 5 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutdownCtx)
		return ctx.Err()
	case err := <-errCh:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	}
}

func withSecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("Cache-Control", "no-store")
		next.ServeHTTP(w, r)
	})
}

func loadSeedProfiles() ([]Profile, error) {
	b, err := webFS.ReadFile("web/seed/profiles.json")
	if err != nil {
		return nil, fmt.Errorf("mocknet: read seed profiles: %w", err)
	}
	var profiles []Profile
	if err := json.Unmarshal(b, &profiles); err != nil {
		return nil, fmt.Errorf("mocknet: parse seed profiles: %w", err)
	}
	return profiles, nil
}

// ---- Helpers ----

func qInt(r *http.Request, key string, def int) int {
	v := strings.TrimSpace(r.URL.Query().Get(key))
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return n
}

func safeRedirectPath(p string) string {
	if p == "" {
		return "/"
	}
	if !strings.HasPrefix(p, "/") {
		return "/"
	}
	clean := path.Clean(p)
	if clean == "." {
		return "/"
	}
	return clean
}
