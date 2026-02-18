package proxy

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

type HTTPServer struct {
	server *http.Server
	config ServerConfig
}

func NewHTTPServer(config ServerConfig) *HTTPServer {
	mux := http.NewServeMux()
	mux.HandleFunc("/live", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "OK")
	})

	return &HTTPServer{
		server: &http.Server{
			Addr:         config.HTTPAddr(),
			Handler:      mux,
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second,
		},
		config: config,
	}
}

func (s *HTTPServer) Serve() error {
	fmt.Printf("HTTP server listening on %s\n", s.config.HTTPAddr())
	return s.server.ListenAndServe()
}

func (s *HTTPServer) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return s.server.Shutdown(ctx)
}
