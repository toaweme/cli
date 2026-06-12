package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/toaweme/cli"
)

// StatusConfig takes no inputs.
type StatusConfig struct{}

// StatusCommand checks if the server is running by hitting /health.
type StatusCommand struct {
	cli.BaseCommand[StatusConfig]
}

var _ cli.Command[StatusConfig] = (*StatusCommand)(nil)

func (c *StatusCommand) Run(options cli.GlobalFlags, _ cli.Unknowns) error {
	client := &http.Client{Timeout: 2 * time.Second}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://localhost:8080/health", http.NoBody)
	if err != nil {
		return fmt.Errorf("failed to build status request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("server is not running")
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		fmt.Println("server is running")
	} else {
		fmt.Printf("server returned status %d\n", resp.StatusCode)
	}

	return nil
}

func (c *StatusCommand) Help() string {
	return "Check if the server is running"
}
