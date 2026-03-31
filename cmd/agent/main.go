package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/your-org/ai-k8s-ops/internal/agent"
	"github.com/your-org/ai-k8s-ops/pkg/version"
)

func main() {
	fmt.Printf("AI-K8S-OPS Agent v%s\n", version.Version)

	// Parse flags
	serverURL := flag.String("server", "http://localhost:8080", "Server URL")
	clusterID := flag.String("cluster", "", "Cluster ID (required)")
	token := flag.String("token", "", "Auth token (required)")
	interval := flag.Int("interval", 30, "Heartbeat interval in seconds")
	flag.Parse()

	if *clusterID == "" {
		log.Fatal("Error: --cluster is required")
	}
	if *token == "" {
		log.Fatal("Error: --token is required")
	}

	log.Printf("Starting agent for cluster: %s", *clusterID)
	log.Printf("Server: %s", *serverURL)

	// Create agent
	ag := agent.NewAgent(agent.Config{
		ServerURL:  *serverURL,
		ClusterID:  *clusterID,
		Token:      *token,
		Interval:   time.Duration(*interval) * time.Second,
	})

	// Handle shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Shutting down agent...")
		cancel()
	}()

	// Start agent
	if err := ag.Run(ctx); err != nil {
		log.Fatalf("Agent failed: %v", err)
	}

	log.Println("Agent stopped")
}
