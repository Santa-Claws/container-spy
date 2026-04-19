package web

import (
	"context"
	"embed"
	"io/fs"
	"net/http"
	"time"

	"github.com/tmac/container-spy/internal/config"
	"github.com/tmac/container-spy/internal/store"
)

//go:embed static
var staticFiles embed.FS

// Server is the HTTP server for the WebUI.
type Server struct {
	cfg    *config.Config
	store  *store.Store
	server *http.Server
}

// NewServer creates a web server bound to addr (e.g. ":8080").
func NewServer(addr string, cfg *config.Config, st *store.Store) *Server {
	s := &Server{cfg: cfg, store: st}

	mux := http.NewServeMux()

	// API endpoints.
	mux.HandleFunc("/api/containers", containersHandler(cfg, st))

	// Static files.
	staticFS, _ := fs.Sub(staticFiles, "static")
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(staticFS))))

	// Root → index.html.
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		data, err := staticFiles.ReadFile("static/index.html")
		if err != nil {
			http.Error(w, "not found", 404)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(data) //nolint:errcheck
	})

	s.server = &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	return s
}

// Start begins serving. It blocks until the context is cancelled.
func (s *Server) Start(ctx context.Context) error {
	go func() {
		<-ctx.Done()
		shutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		s.server.Shutdown(shutCtx) //nolint:errcheck
	}()
	if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}
