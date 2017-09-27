package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/schu/rktup"
)

var (
	addr        string
	hostname    string
	githubToken string
)

func init() {
	flag.StringVar(&addr, "addr", "127.0.0.1:33333", "addr to listen on")
	flag.StringVar(&hostname, "hostname", "localhost", "hostname to use (e.g. rktup.org)")
	flag.StringVar(&githubToken, "github-token", "", "GitHub API token")
}

func main() {
	flag.Parse()

	server, err := rktup.NewServer(&rktup.ServerConfig{
		Addr:        addr,
		Hostname:    hostname,
		GithubToken: githubToken,
	})
	if err != nil {
		log.Fatalf("failed to get server: %v\n", err)
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, os.Kill)

	go func() {
		log.Printf("listening on %s\n", addr)
		if err := server.ListenAndServe(); err != nil {
			if err != http.ErrServerClosed {
				log.Fatalf("http server error: %v\n", err)
			}
		}
	}()

	<-sigChan

	log.Printf("shutting down ...\n")

	ctx, cancelCtx := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelCtx()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("clean shutdown failed: %v\n", err)
	}
}
