package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/toaweme/cli"
	"github.com/toaweme/cli/config"
)

// StartConfig demonstrates env-driven server configuration with defaults.
type StartConfig struct {
	Port    int    `arg:"port" short:"p" env:"SERVER_PORT" help:"Port to listen on" default:"8080"`
	Host    string `arg:"host" env:"SERVER_HOST" help:"Host to bind to" default:"0.0.0.0"`
	Timeout int    `arg:"timeout" env:"SERVER_TIMEOUT" help:"Shutdown timeout in seconds" default:"5"`
}

// StartCommand starts an HTTP server with graceful shutdown on SIGINT/SIGTERM.
type StartCommand struct {
	cli.BaseCommand[StartConfig]
	store *config.FileStore
}

var _ cli.Command[StartConfig] = (*StartCommand)(nil)

func (c *StartCommand) Run(_ cli.GlobalOptions, _ cli.Unknowns) error {
	addr := fmt.Sprintf("%s:%d", c.Inputs.Host, c.Inputs.Port)

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "ok\n")
	})
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `{"status":"healthy"}`+"\n")
	})

	srv := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	// save last used config for the status command
	c.store.Save("last", map[string]any{
		"addr":       addr,
		"started_at": time.Now().Format(time.RFC3339),
	})

	// graceful shutdown on SIGINT/SIGTERM
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		fmt.Printf("listening on %s\n", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Fprintf(os.Stderr, "failed to start server: %v\n", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	fmt.Println("\nshutting down...")

	timeout := time.Duration(c.Inputs.Timeout) * time.Second
	shutdownCtx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("failed to shutdown server: %w", err)
	}

	fmt.Println("stopped")
	return nil
}

func (c *StartCommand) Help() string {
	return "Start the HTTP server"
}

func (c *StartCommand) Examples() [][]string {
	return [][]string{
		{"server start"},
		{"server start -p 3000"},
		{"server start --host=127.0.0.1 --timeout=10"},
	}
}
