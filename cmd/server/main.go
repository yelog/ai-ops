package main

import (
	"fmt"
	"os"

	"github.com/your-org/ai-k8s-ops/pkg/version"
)

func main() {
	fmt.Printf("AI-K8S-OPS Server v%s\n", version.Version)
	fmt.Println("Server not implemented yet")
	os.Exit(0)
}
