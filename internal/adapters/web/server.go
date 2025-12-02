// Package web provides an HTTP server for the ProtoDiff dashboard.
//
// The dashboard displays real-time schema drift detection results with a traffic
// light UI (green=sync, red=mismatch, yellow=unknown). It serves an embedded HTML
// template and provides both a main dashboard and a health check endpoint.
//
// Endpoints:
//   - GET /: Main dashboard showing all scan results with statistics
//   - GET /health: Health check endpoint returning {"status":"healthy"}
//
// The server reads scan results from the in-memory store and renders them using
// Go's html/template package with an embedded template file.
//
// Example usage:
//
//	server, err := web.NewServer(dataStore, ":18080")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	server.Start()
package web

import (
	_ "embed"
	"html/template"
	"log"
	"net/http"
	"time"

	"github.com/uzdada/protodiff/internal/core/domain"
	"github.com/uzdada/protodiff/internal/core/store"
)

//go:embed templates/index.html
var indexTemplate string

// Server provides the HTTP server for the dashboard
type Server struct {
	store    *store.Store
	template *template.Template
	addr     string
}

// NewServer creates a new web server instance
func NewServer(store *store.Store, addr string) (*Server, error) {
	// Create template with custom functions
	funcMap := template.FuncMap{
		"add": func(a, b int) int {
			return a + b
		},
	}

	tmpl, err := template.New("index").Funcs(funcMap).Parse(indexTemplate)
	if err != nil {
		return nil, err
	}

	return &Server{
		store:    store,
		template: tmpl,
		addr:     addr,
	}, nil
}

// Statistics holds aggregated stats for the dashboard
type Statistics struct {
	TotalCount    int
	SyncCount     int
	MismatchCount int
	UnknownCount  int
}

// TemplateData represents the data passed to the HTML template
type TemplateData struct {
	Results    []*domain.ScanResult
	Stats      Statistics
	LastUpdate string
}

// Start begins serving HTTP requests
func (s *Server) Start() error {
	http.HandleFunc("/", s.handleDashboard)
	http.HandleFunc("/health", s.handleHealth)

	log.Printf("Starting web server on %s", s.addr)
	return http.ListenAndServe(s.addr, nil)
}

// handleDashboard renders the main dashboard
func (s *Server) handleDashboard(w http.ResponseWriter, r *http.Request) {
	results := s.store.GetAll()
	stats := calculateStatistics(results)

	data := TemplateData{
		Results:    results,
		Stats:      stats,
		LastUpdate: time.Now().Format("2006-01-02 15:04:05"),
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := s.template.Execute(w, data); err != nil {
		log.Printf("Error rendering template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// calculateStatistics aggregates scan result statistics for dashboard display.
// It counts the total results and breaks them down by status (sync, mismatch, unknown).
func calculateStatistics(results []*domain.ScanResult) Statistics {
	stats := Statistics{
		TotalCount: len(results),
	}

	for _, result := range results {
		switch result.Status {
		case domain.StatusSync:
			stats.SyncCount++
		case domain.StatusMismatch:
			stats.MismatchCount++
		case domain.StatusUnknown:
			stats.UnknownCount++
		}
	}

	return stats
}

// handleHealth provides a health check endpoint
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte(`{"status":"healthy"}`)); err != nil {
		log.Printf("Error writing health response: %v", err)
	}
}
