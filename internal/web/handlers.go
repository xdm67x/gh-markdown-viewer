package web

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/xdm67x/gh-markdown-viewer/internal/ghpr"
)

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}

func (s *Server) lookupFile(w http.ResponseWriter, r *http.Request) (ghpr.File, bool) {
	path := r.URL.Query().Get("path")
	if path == "" {
		http.Error(w, "missing path", http.StatusBadRequest)
		return ghpr.File{}, false
	}
	f, ok := s.byPath[path]
	if !ok {
		http.Error(w, "unknown path", http.StatusNotFound)
		return ghpr.File{}, false
	}
	return f, true
}

func (s *Server) handleFiles(w http.ResponseWriter, r *http.Request) {
	type fileItem struct {
		Path      string `json:"path"`
		Status    string `json:"status"`
		Additions int    `json:"additions"`
		Deletions int    `json:"deletions"`
	}
	items := make([]fileItem, len(s.files))
	for i, f := range s.files {
		items[i] = fileItem{
			Path:      f.Filename,
			Status:    f.Status,
			Additions: f.Additions,
			Deletions: f.Deletions,
		}
	}
	writeJSON(w, items)
}

func (s *Server) handleRendered(w http.ResponseWriter, r *http.Request) {
	f, ok := s.lookupFile(w, r)
	if !ok {
		return
	}
	content, err := ghpr.FileContent(s.info, f.Filename)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	commentable := ghpr.CommentableLines(f.Patch)
	html, injected, err := ghpr.RenderMarkdownAnnotated(s.info, content, commentable)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, map[string]any{"html": html, "commentableLines": injected})
}

func (s *Server) handleSource(w http.ResponseWriter, r *http.Request) {
	f, ok := s.lookupFile(w, r)
	if !ok {
		return
	}
	content, err := ghpr.FileContent(s.info, f.Filename)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	commentable := ghpr.CommentableLines(f.Patch)
	type lineItem struct {
		N           int    `json:"n"`
		Text        string `json:"text"`
		Commentable bool   `json:"commentable"`
	}
	rawLines := strings.Split(content, "\n")
	lines := make([]lineItem, len(rawLines))
	for i, text := range rawLines {
		n := i + 1
		lines[i] = lineItem{N: n, Text: text, Commentable: commentable[n]}
	}
	writeJSON(w, map[string]any{"path": f.Filename, "lines": lines})
}

func (s *Server) handleThreads(w http.ResponseWriter, r *http.Request) {
	threads, err := ghpr.FetchThreads(s.info)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	writeJSON(w, threads)
}

func (s *Server) handleReply(w http.ResponseWriter, r *http.Request) {
	var body struct {
		CommentID int64  `json:"commentId"`
		Body      string `json:"body"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	if err := ghpr.ReplyToComment(s.info, body.CommentID, body.Body); err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	writeJSON(w, map[string]string{"status": "ok"})
}

func (s *Server) handleResolveThread(w http.ResponseWriter, r *http.Request) {
	var body struct {
		NodeID string `json:"nodeId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	if err := ghpr.ResolveThread(body.NodeID); err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	writeJSON(w, map[string]string{"status": "ok"})
}

func (s *Server) handleComment(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Path string `json:"path"`
		Line int    `json:"line"`
		Body string `json:"body"`
	}
	data, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "cannot read body", http.StatusBadRequest)
		return
	}
	if err := json.Unmarshal(data, &body); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	if _, ok := s.byPath[body.Path]; !ok {
		http.Error(w, "unknown path", http.StatusNotFound)
		return
	}
	if err := ghpr.PostLineComment(s.info, body.Path, body.Body, body.Line); err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	writeJSON(w, map[string]string{"status": "ok"})
}
