package web

import (
	"net"
	"net/http"

	"github.com/xdm67x/gh-markdown-viewer/internal/ghpr"
)

// Server holds the PR context and wires HTTP routes.
type Server struct {
	info   ghpr.PRInfo
	files  []ghpr.File
	byPath map[string]ghpr.File
}

// NewServer creates a Server for the given PR and file list.
func NewServer(info ghpr.PRInfo, files []ghpr.File) *Server {
	byPath := make(map[string]ghpr.File, len(files))
	for _, f := range files {
		byPath[f.Filename] = f
	}
	return &Server{info: info, files: files, byPath: byPath}
}

// Listen binds to a random free port on localhost and returns the listener.
func (s *Server) Listen() (net.Listener, error) {
	return net.Listen("tcp", "127.0.0.1:0")
}

// Handler returns the HTTP handler for the server.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()

	// Serve index.html for the root path.
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		data, err := staticFiles.ReadFile("static/index.html")
		if err != nil {
			http.Error(w, "index.html not found", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(data) //nolint:errcheck
	})

	mux.HandleFunc("GET /api/info", s.handleInfo)
	mux.HandleFunc("GET /api/files", s.handleFiles)
	mux.HandleFunc("GET /api/file/rendered", s.handleRendered)
	mux.HandleFunc("GET /api/file/source", s.handleSource)
	mux.HandleFunc("POST /api/comment", s.handleComment)
	mux.HandleFunc("GET /api/threads", s.handleThreads)
	mux.HandleFunc("POST /api/reply", s.handleReply)
	mux.HandleFunc("POST /api/resolve", s.handleResolveThread)

	return mux
}
