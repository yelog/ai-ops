package main

import (
	"fmt"
	"os"

	"github.com/your-org/ai-k8s-ops/pkg/version"
)

func main() {
	fmt.Printf("AI-K8S-OPS CLI v%s\n", version.Version)
	fmt.Println("CLI not implemented yet")
	os.Exit(0)
}
