package http

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	// enable http profiling
	_ "net/http/pprof"
)

// Server handles app's http requests.
type Server struct {
	addr        string
	profileAddr string
	handler     http.Handler
}

// NewServer creates new Server instance.
func NewServer(addr string, profileAddr string, handler http.Handler) *Server {
	return &Server{
		addr:        addr,
		profileAddr: profileAddr,
		handler:     handler,
	}
}

// Run runs the server. Waits SIGINT is received, then gracefully shutdowns.
// Blocks until shutdown is complete.
func (s *Server) Run() {
	srv := http.Server{
		Addr: s.addr,

		// For timeouts explanation see: https://blog.cloudflare.com/the-complete-guide-to-golang-net-http-timeouts/
		ReadHeaderTimeout: time.Second,
		ReadTimeout:       60 * time.Second,
		WriteTimeout:      70 * time.Second,
		IdleTimeout:       10 * time.Second,

		Handler: s.handler,
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("server returned error: %v", err)
		}
	}()

	if s.profileAddr != "" {
		profilingServer := http.Server{
			Addr:    s.profileAddr,
			Handler: nil,
		}
		go func() {
			if err := profilingServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Printf("profiling server returned error: %v", err)
			}
		}()
		defer profilingServer.Close()
	}

	<-stop
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil && err != http.ErrServerClosed {
		log.Printf("server shutdown returned error: %v", err)
	}
}
