package main

import (
	"context"
	"io"
	"net/http"
	"net/http/pprof"
	"os"
	"os/signal"
	"syscall"

	"golang.org/x/sync/errgroup" //indirect
)

type serverHandler struct {
	body string
}

func (s *serverHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "Hello from a Handle")
}

type debugHandler struct{}

func (d *debugHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	pprof.Index(w, r)
}

func newAppServer() *http.Server {
	server := &http.Server{
		Addr:    "127.0.0.1:8080",
		Handler: &serverHandler{"nihao"},
	}
	return server
}

func newDebugServer() *http.Server {
	server := &http.Server{
		Addr:    "127.0.0.1:6060",
		Handler: &debugHandler{},
	}
	return server
}

func app(done chan bool, closed chan struct{}) {
	g := &errgroup.Group{}
	server := newAppServer()
	debugServer := newDebugServer()
	g.Go(func() error {
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			return err
		}
		return nil
	})
	g.Go(func() error {
		if err := debugServer.ListenAndServe(); err != http.ErrServerClosed {
			return err
		}
		return nil
	})
	go func() {
		<-done
		server.Shutdown(context.Background())
		debugServer.Shutdown(context.Background())
		close(closed)
	}()
}

func main() {

	quit := make(chan os.Signal, 1)
	shutdown := make(chan bool)
	closed := make(chan struct{})
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM, syscall.SIGUSR1)
	go app(shutdown, closed)
	select {
	case <-quit:
		print("shot down")
		shutdown <- true
	}
	<-closed

}
